//go:build testing

package handlers

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/zoobz-io/cicero/api/contracts"
	"github.com/zoobz-io/cicero/models"
	cicerotest "github.com/zoobz-io/cicero/testing"
	roccotest "github.com/zoobz-io/rocco/testing"
	"github.com/zoobz-io/sum"
)

func setupTranslationRegistry(t *testing.T, ms *cicerotest.MockSources, mt *cicerotest.MockTranslations, mtr *cicerotest.MockTranslator) {
	t.Helper()
	sum.Reset()
	k := sum.Start()
	sum.Register[contracts.Sources](k, ms)
	sum.Register[contracts.Translations](k, mt)
	sum.Register[contracts.Translator](k, mtr)
	sum.Freeze(k)
	t.Cleanup(sum.Reset)
}

func TestCreateTranslation_Success(t *testing.T) {
	ms := &cicerotest.MockSources{}
	mt := &cicerotest.MockTranslations{
		OnGetBySourceAndLang: func(_ context.Context, _, _, _, _ string) (*models.Translation, error) {
			return nil, errors.New("not found")
		},
	}
	mtr := &cicerotest.MockTranslator{
		OnTranslate: func(_ context.Context, _, _, _ string) (string, string, error) {
			return "¡Hola, mundo!", "libretranslate", nil
		},
	}

	setupTranslationRegistry(t, ms, mt, mtr)

	engine := roccotest.TestEngine()
	engine.WithHandlers(CreateTranslation)

	resp := roccotest.ServeRequest(engine, http.MethodPost, "/translations", map[string]string{
		"text":        "Hello, world!",
		"source_lang": "en",
		"target_lang": "es",
		"tenant_id":   "zoobzio",
	})

	if resp.StatusCode() != http.StatusCreated {
		t.Errorf("status: got %d, want %d (body: %s)", resp.StatusCode(), http.StatusCreated, resp.BodyString())
	}

	var body map[string]any
	if err := resp.DecodeJSON(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if got, _ := body["translated_text"].(string); got != "¡Hola, mundo!" {
		t.Errorf("translated_text: got %q, want %q", got, "¡Hola, mundo!")
	}
	if got, _ := body["provider"].(string); got != "libretranslate" {
		t.Errorf("provider: got %q, want %q", got, "libretranslate")
	}
}

func TestCreateTranslation_DedupHit(t *testing.T) {
	existing := &models.Translation{
		ID:         "test-uuid-1",
		SourceHash: "315f5bdb76d078c43b8ac0064e4a0164",
		SourceLang: "en",
		TargetLang: "es",
		Text:       "cached translation",
		Provider:   "libretranslate",
		Status:     "completed",
		TenantID:   "zoobzio",
	}

	ms := &cicerotest.MockSources{
		OnGet: func(_ context.Context, _ string) (*models.Source, error) {
			return &models.Source{Hash: "315f5bdb76d078c43b8ac0064e4a0164", Text: "Hello, world!", TenantID: "zoobzio"}, nil
		},
	}
	mt := &cicerotest.MockTranslations{
		OnGetBySourceAndLang: func(_ context.Context, _, _, _, _ string) (*models.Translation, error) {
			return existing, nil
		},
	}

	translatorCalled := false
	mtr := &cicerotest.MockTranslator{
		OnTranslate: func(_ context.Context, _, _, _ string) (string, string, error) {
			translatorCalled = true
			return "", "", nil
		},
	}

	setupTranslationRegistry(t, ms, mt, mtr)

	engine := roccotest.TestEngine()
	engine.WithHandlers(CreateTranslation)

	resp := roccotest.ServeRequest(engine, http.MethodPost, "/translations", map[string]string{
		"text":        "Hello, world!",
		"source_lang": "en",
		"target_lang": "es",
		"tenant_id":   "zoobzio",
	})

	if resp.StatusCode() != http.StatusCreated {
		t.Errorf("status: got %d, want %d (body: %s)", resp.StatusCode(), http.StatusCreated, resp.BodyString())
	}
	if translatorCalled {
		t.Error("translator should not be called on dedup hit")
	}
}

func TestCreateTranslation_TranslatorError_Returns500(t *testing.T) {
	ms := &cicerotest.MockSources{}
	mt := &cicerotest.MockTranslations{
		OnGetBySourceAndLang: func(_ context.Context, _, _, _, _ string) (*models.Translation, error) {
			return nil, errors.New("not found")
		},
	}
	mtr := &cicerotest.MockTranslator{
		OnTranslate: func(_ context.Context, _, _, _ string) (string, string, error) {
			return "", "", errors.New("libretranslate unavailable")
		},
	}

	setupTranslationRegistry(t, ms, mt, mtr)

	engine := roccotest.TestEngine()
	engine.WithHandlers(CreateTranslation)

	resp := roccotest.ServeRequest(engine, http.MethodPost, "/translations", map[string]string{
		"text":        "Hello",
		"source_lang": "en",
		"target_lang": "es",
		"tenant_id":   "zoobzio",
	})

	if resp.StatusCode() != http.StatusInternalServerError {
		t.Errorf("status: got %d, want %d (body: %s)", resp.StatusCode(), http.StatusInternalServerError, resp.BodyString())
	}
}

func TestGetTranslationsByHash_Success(t *testing.T) {
	src := &models.Source{
		Hash:     "315f5bdb76d078c43b8ac0064e4a0164",
		Text:     "Hello, world!",
		TenantID: "zoobzio",
	}
	translationList := []*models.Translation{
		{
			SourceLang: "en",
			TargetLang: "es",
			Text:       "¡Hola, mundo!",
			Provider:   "libretranslate",
			Status:     "completed",
			CreatedAt:  time.Date(2026, 3, 3, 0, 0, 0, 0, time.UTC),
		},
	}

	ms := &cicerotest.MockSources{
		OnGet: func(_ context.Context, _ string) (*models.Source, error) {
			return src, nil
		},
	}
	mt := &cicerotest.MockTranslations{
		OnListBySourceHash: func(_ context.Context, _ string) ([]*models.Translation, error) {
			return translationList, nil
		},
	}

	setupTranslationRegistry(t, ms, mt, &cicerotest.MockTranslator{})

	engine := roccotest.TestEngine()
	engine.WithHandlers(GetTranslationsByHash)

	resp := roccotest.ServeRequest(engine, http.MethodGet, "/translations/315f5bdb76d078c43b8ac0064e4a0164", nil)

	if resp.StatusCode() != http.StatusOK {
		t.Errorf("status: got %d, want %d (body: %s)", resp.StatusCode(), http.StatusOK, resp.BodyString())
	}

	var body map[string]any
	if err := resp.DecodeJSON(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if body["hash"] != src.Hash {
		t.Errorf("hash: got %v, want %q", body["hash"], src.Hash)
	}
	txs, _ := body["translations"].([]any)
	if len(txs) != 1 {
		t.Errorf("translations length: got %d, want 1", len(txs))
	}
}

func TestGetTranslationsByHash_SourceNotFound_Returns404(t *testing.T) {
	ms := &cicerotest.MockSources{
		OnGet: func(_ context.Context, _ string) (*models.Source, error) {
			return nil, errors.New("not found")
		},
	}

	setupTranslationRegistry(t, ms, &cicerotest.MockTranslations{}, &cicerotest.MockTranslator{})

	engine := roccotest.TestEngine()
	engine.WithHandlers(GetTranslationsByHash)

	resp := roccotest.ServeRequest(engine, http.MethodGet, "/translations/doesnotexist", nil)

	if resp.StatusCode() != http.StatusNotFound {
		t.Errorf("status: got %d, want %d (body: %s)", resp.StatusCode(), http.StatusNotFound, resp.BodyString())
	}
}
