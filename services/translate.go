// Package services provides resilient clients for external service dependencies.
package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/zoobz-io/pipz"
)

const libreTranslateProvider = "libretranslate"

// Resilience configuration.
const (
	translateTimeout          = 30 * time.Second
	translateMaxAttempts      = 3
	translateBackoffDelay     = 200 * time.Millisecond
	translateFailureThreshold = 5
	translateResetTimeout     = 30 * time.Second
)

// Pipeline identities for the resilience stack.
var (
	translateProcessorID = pipz.NewIdentity("translate.call", "HTTP call to LibreTranslate")
	translateTimeoutID   = pipz.NewIdentity("translate.timeout", "Timeout for translation calls")
	translateBackoffID   = pipz.NewIdentity("translate.backoff", "Backoff retry for translation calls")
	translateBreakerID   = pipz.NewIdentity("translate.breaker", "Circuit breaker for LibreTranslate")
)

type translateCall struct {
	text       string
	sourceLang string
	targetLang string
	result     string
	provider   string
}

func (c *translateCall) Clone() *translateCall {
	clone := *c
	return &clone
}

type libreTranslateRequest struct {
	Q      string `json:"q"`
	Source string `json:"source"`
	Target string `json:"target"`
}

type libreTranslateResponse struct {
	TranslatedText string `json:"translatedText"`
}

type libreTranslateError struct {
	Error string `json:"error"`
}

// TranslateService calls LibreTranslate over HTTP with a resilient pipeline.
type TranslateService struct {
	pipeline   pipz.Chainable[*translateCall]
	httpClient *http.Client
	addr       string
}

// NewTranslateService creates a LibreTranslate client targeting the given address.
func NewTranslateService(addr string) *TranslateService {
	s := &TranslateService{
		addr:       addr,
		httpClient: &http.Client{},
	}
	s.pipeline = s.buildPipeline()
	return s
}

func (s *TranslateService) buildPipeline() pipz.Chainable[*translateCall] {
	processor := pipz.Apply(translateProcessorID, s.doTranslate)

	return pipz.NewCircuitBreaker(translateBreakerID,
		pipz.NewBackoff(translateBackoffID,
			pipz.NewTimeout(translateTimeoutID, processor, translateTimeout),
			translateMaxAttempts, translateBackoffDelay,
		),
		translateFailureThreshold, translateResetTimeout,
	)
}

func (s *TranslateService) doTranslate(ctx context.Context, call *translateCall) (*translateCall, error) {
	body, err := json.Marshal(libreTranslateRequest{
		Q:      call.text,
		Source: call.sourceLang,
		Target: call.targetLang,
	})
	if err != nil {
		return call, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.addr+"/translate", bytes.NewReader(body))
	if err != nil {
		return call, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return call, fmt.Errorf("libretranslate unavailable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadRequest {
		var errResp libreTranslateError
		if jsonErr := json.NewDecoder(resp.Body).Decode(&errResp); jsonErr == nil && errResp.Error != "" {
			return call, fmt.Errorf("libretranslate: %s", errResp.Error)
		}
		return call, fmt.Errorf("bad request to libretranslate: status %d", resp.StatusCode)
	}

	if resp.StatusCode != http.StatusOK {
		return call, fmt.Errorf("unexpected libretranslate status %d", resp.StatusCode)
	}

	var result libreTranslateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return call, fmt.Errorf("decode response: %w", err)
	}

	call.result = result.TranslatedText
	call.provider = libreTranslateProvider
	return call, nil
}

// Translate sends text to LibreTranslate and returns the translated string and provider name.
func (s *TranslateService) Translate(ctx context.Context, text, sourceLang, targetLang string) (string, string, error) {
	call := &translateCall{
		text:       text,
		sourceLang: sourceLang,
		targetLang: targetLang,
	}
	out, err := s.pipeline.Process(ctx, call)
	if err != nil {
		return "", "", err
	}
	return out.result, out.provider, nil
}

// Close shuts down the resilience pipeline.
func (s *TranslateService) Close() error {
	if s.pipeline != nil {
		return s.pipeline.Close()
	}
	return nil
}
