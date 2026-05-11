//go:build testing

package handlers

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/zoobz-io/cicero/api/contracts"
	"github.com/zoobz-io/cicero/internal/auth"
	"github.com/zoobz-io/cicero/models"
	cicerotest "github.com/zoobz-io/cicero/testing"
	"github.com/zoobz-io/rocco"
	roccotest "github.com/zoobz-io/rocco/testing"
	"github.com/zoobz-io/sum"
)

func testAuthenticator() func(context.Context, *http.Request) (rocco.Identity, error) {
	return func(_ context.Context, _ *http.Request) (rocco.Identity, error) {
		return &auth.Identity{
			UserID:   "user-1",
			TenantAt: "zoobzio",
			Tenants:  []auth.AuthorizedTenant{{TenantID: "zoobzio", TenantName: "Zoobzio", Role: "admin"}},
		}, nil
	}
}

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

	engine := roccotest.TestEngineWithAuth(testAuthenticator())
	engine.WithHandlers(CreateTranslation)

	resp := roccotest.ServeRequest(engine, http.MethodPost, "/translations", map[string]string{
		"text":        "Hello, world!",
		"source_lang": "en",
		"target_lang": "es",
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
}

func TestCreateTranslation_Unauthenticated_Returns401(t *testing.T) {
	ms := &cicerotest.MockSources{}
	mt := &cicerotest.MockTranslations{}
	mtr := &cicerotest.MockTranslator{}

	setupTranslationRegistry(t, ms, mt, mtr)

	engine := roccotest.TestEngine()
	engine.WithHandlers(CreateTranslation)

	resp := roccotest.ServeRequest(engine, http.MethodPost, "/translations", map[string]string{
		"text":        "Hello",
		"source_lang": "en",
		"target_lang": "es",
	})

	if resp.StatusCode() != http.StatusUnauthorized {
		t.Errorf("status: got %d, want %d", resp.StatusCode(), http.StatusUnauthorized)
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

	engine := roccotest.TestEngineWithAuth(testAuthenticator())
	engine.WithHandlers(GetTranslationsByHash)

	resp := roccotest.ServeRequest(engine, http.MethodGet, "/translations/315f5bdb76d078c43b8ac0064e4a0164", nil)

	if resp.StatusCode() != http.StatusOK {
		t.Errorf("status: got %d, want %d (body: %s)", resp.StatusCode(), http.StatusOK, resp.BodyString())
	}
}
