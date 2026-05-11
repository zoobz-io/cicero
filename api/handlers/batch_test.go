//go:build testing

package handlers

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/zoobz-io/cicero/models"
	cicerotest "github.com/zoobz-io/cicero/testing"
	roccotest "github.com/zoobz-io/rocco/testing"
)

func TestBatchTranslate_Success(t *testing.T) {
	callCount := 0
	ms := &cicerotest.MockSources{}
	mt := &cicerotest.MockTranslations{
		OnGetBySourceAndLang: func(_ context.Context, _, _, _, _ string) (*models.Translation, error) {
			return nil, errors.New("not found")
		},
	}
	mtr := &cicerotest.MockTranslator{
		OnTranslate: func(_ context.Context, text, _, _ string) (string, string, error) {
			callCount++
			return "translated:" + text, "libretranslate", nil
		},
	}

	setupTranslationRegistry(t, ms, mt, mtr)

	engine := roccotest.TestEngineWithAuth(testAuthenticator())
	engine.WithHandlers(BatchTranslate)

	resp := roccotest.ServeRequest(engine, http.MethodPost, "/translations/batch", map[string]any{
		"texts":       []string{"Hello", "World"},
		"source_lang": "en",
		"target_lang": "es",
	})

	if resp.StatusCode() != http.StatusCreated {
		t.Errorf("status: got %d, want %d (body: %s)", resp.StatusCode(), http.StatusCreated, resp.BodyString())
	}

	var body map[string]any
	if err := resp.DecodeJSON(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}

	txs, _ := body["translations"].([]any)
	if len(txs) != 2 {
		t.Errorf("translations length: got %d, want 2", len(txs))
	}
	if callCount != 2 {
		t.Errorf("translator called %d times, want 2", callCount)
	}
}

func TestBatchTranslate_Unauthenticated(t *testing.T) {
	ms := &cicerotest.MockSources{}
	mt := &cicerotest.MockTranslations{}
	mtr := &cicerotest.MockTranslator{}

	setupTranslationRegistry(t, ms, mt, mtr)

	engine := roccotest.TestEngine()
	engine.WithHandlers(BatchTranslate)

	resp := roccotest.ServeRequest(engine, http.MethodPost, "/translations/batch", map[string]any{
		"texts":       []string{"Hello"},
		"source_lang": "en",
		"target_lang": "es",
	})

	if resp.StatusCode() != http.StatusUnauthorized {
		t.Errorf("status: got %d, want %d", resp.StatusCode(), http.StatusUnauthorized)
	}
}

func TestBatchTranslate_TranslatorError(t *testing.T) {
	ms := &cicerotest.MockSources{}
	mt := &cicerotest.MockTranslations{
		OnGetBySourceAndLang: func(_ context.Context, _, _, _, _ string) (*models.Translation, error) {
			return nil, errors.New("not found")
		},
	}
	mtr := &cicerotest.MockTranslator{
		OnTranslate: func(_ context.Context, _, _, _ string) (string, string, error) {
			return "", "", errors.New("provider down")
		},
	}

	setupTranslationRegistry(t, ms, mt, mtr)

	engine := roccotest.TestEngineWithAuth(testAuthenticator())
	engine.WithHandlers(BatchTranslate)

	resp := roccotest.ServeRequest(engine, http.MethodPost, "/translations/batch", map[string]any{
		"texts":       []string{"Hello"},
		"source_lang": "en",
		"target_lang": "es",
	})

	if resp.StatusCode() != http.StatusInternalServerError {
		t.Errorf("status: got %d, want %d", resp.StatusCode(), http.StatusInternalServerError)
	}
}
