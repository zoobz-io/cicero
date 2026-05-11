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

func setupTenantRegistry(t *testing.T, ms *cicerotest.MockSources, mt *cicerotest.MockTranslations) {
	t.Helper()
	sum.Reset()
	k := sum.Start()
	sum.Register[contracts.Sources](k, ms)
	sum.Register[contracts.Translations](k, mt)
	sum.Register[contracts.Translator](k, &cicerotest.MockTranslator{})
	sum.Freeze(k)
	t.Cleanup(sum.Reset)
}

// unauthorizedAuthenticator returns an identity for tenant "other", not "zoobzio".
func unauthorizedAuthenticator() func(context.Context, *http.Request) (rocco.Identity, error) {
	return func(_ context.Context, _ *http.Request) (rocco.Identity, error) {
		return &auth.Identity{
			UserID:   "user-1",
			TenantAt: "other",
			Tenants:  []auth.AuthorizedTenant{{TenantID: "other", TenantName: "Other", Role: "admin"}},
		}, nil
	}
}

func TestListTenantTranslations_Success(t *testing.T) {
	mt := &cicerotest.MockTranslations{
		OnListByTenant: func(_ context.Context, _ string, _, _ int) ([]*models.Translation, int, error) {
			return []*models.Translation{
				{ID: "uuid-1", SourceHash: "h1", SourceLang: "en", TargetLang: "es", Text: "Hola", Provider: "libretranslate", Status: "completed", TenantID: "zoobzio", CreatedAt: time.Now()},
			}, 1, nil
		},
	}

	setupTenantRegistry(t, &cicerotest.MockSources{}, mt)

	engine := roccotest.TestEngineWithAuth(testAuthenticator())
	engine.WithHandlers(ListTenantTranslations)

	resp := roccotest.ServeRequest(engine, http.MethodGet, "/tenants/zoobzio/translations?limit=10", nil)

	if resp.StatusCode() != http.StatusOK {
		t.Errorf("status: got %d, want %d (body: %s)", resp.StatusCode(), http.StatusOK, resp.BodyString())
	}

	var body map[string]any
	if err := resp.DecodeJSON(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if total, _ := body["total"].(float64); int(total) != 1 {
		t.Errorf("total: got %v, want 1", total)
	}
}

func TestListTenantTranslations_Forbidden(t *testing.T) {
	setupTenantRegistry(t, &cicerotest.MockSources{}, &cicerotest.MockTranslations{})

	engine := roccotest.TestEngineWithAuth(unauthorizedAuthenticator())
	engine.WithHandlers(ListTenantTranslations)

	resp := roccotest.ServeRequest(engine, http.MethodGet, "/tenants/zoobzio/translations", nil)

	if resp.StatusCode() != http.StatusForbidden {
		t.Errorf("status: got %d, want %d (body: %s)", resp.StatusCode(), http.StatusForbidden, resp.BodyString())
	}
}

func TestGetTenantTranslationsByHash_Success(t *testing.T) {
	src := &models.Source{Hash: "h1", Text: "Hello", TenantID: "zoobzio"}
	ms := &cicerotest.MockSources{
		OnGet: func(_ context.Context, _ string) (*models.Source, error) { return src, nil },
	}
	mt := &cicerotest.MockTranslations{
		OnListByTenantAndHash: func(_ context.Context, _, _ string) ([]*models.Translation, error) {
			return []*models.Translation{
				{ID: "uuid-1", SourceLang: "en", TargetLang: "es", Text: "Hola", Provider: "libretranslate", Status: "completed", TenantID: "zoobzio"},
			}, nil
		},
	}

	setupTenantRegistry(t, ms, mt)

	engine := roccotest.TestEngineWithAuth(testAuthenticator())
	engine.WithHandlers(GetTenantTranslationsByHash)

	resp := roccotest.ServeRequest(engine, http.MethodGet, "/tenants/zoobzio/translations/h1", nil)

	if resp.StatusCode() != http.StatusOK {
		t.Errorf("status: got %d, want %d (body: %s)", resp.StatusCode(), http.StatusOK, resp.BodyString())
	}
}

func TestUpdateTranslation_Success(t *testing.T) {
	updated := &models.Translation{
		ID: "uuid-1", SourceHash: "h1", SourceLang: "en", TargetLang: "es",
		Text: "Edited", Provider: "libretranslate", Status: "completed", TenantID: "zoobzio",
		CreatedAt: time.Now(),
	}
	mt := &cicerotest.MockTranslations{
		OnUpdate: func(_ context.Context, _, _, _ string) (*models.Translation, error) {
			return updated, nil
		},
	}

	setupTenantRegistry(t, &cicerotest.MockSources{}, mt)

	engine := roccotest.TestEngineWithAuth(testAuthenticator())
	engine.WithHandlers(UpdateTranslation)

	resp := roccotest.ServeRequest(engine, http.MethodPut, "/tenants/zoobzio/translations/uuid-1", map[string]string{
		"text": "Edited",
	})

	if resp.StatusCode() != http.StatusOK {
		t.Errorf("status: got %d, want %d (body: %s)", resp.StatusCode(), http.StatusOK, resp.BodyString())
	}
}

func TestUpdateTranslation_NotFound(t *testing.T) {
	mt := &cicerotest.MockTranslations{
		OnUpdate: func(_ context.Context, _, _, _ string) (*models.Translation, error) {
			return nil, errors.New("not found")
		},
	}

	setupTenantRegistry(t, &cicerotest.MockSources{}, mt)

	engine := roccotest.TestEngineWithAuth(testAuthenticator())
	engine.WithHandlers(UpdateTranslation)

	resp := roccotest.ServeRequest(engine, http.MethodPut, "/tenants/zoobzio/translations/uuid-999", map[string]string{
		"text": "Edited",
	})

	if resp.StatusCode() != http.StatusNotFound {
		t.Errorf("status: got %d, want %d (body: %s)", resp.StatusCode(), http.StatusNotFound, resp.BodyString())
	}
}

func TestUpdateTranslation_Forbidden(t *testing.T) {
	setupTenantRegistry(t, &cicerotest.MockSources{}, &cicerotest.MockTranslations{})

	engine := roccotest.TestEngineWithAuth(unauthorizedAuthenticator())
	engine.WithHandlers(UpdateTranslation)

	resp := roccotest.ServeRequest(engine, http.MethodPut, "/tenants/zoobzio/translations/uuid-1", map[string]string{
		"text": "Edited",
	})

	if resp.StatusCode() != http.StatusForbidden {
		t.Errorf("status: got %d, want %d (body: %s)", resp.StatusCode(), http.StatusForbidden, resp.BodyString())
	}
}

func TestDeleteTranslation_Success(t *testing.T) {
	mt := &cicerotest.MockTranslations{
		OnDeleteByID: func(_ context.Context, _, _ string) error { return nil },
	}

	setupTenantRegistry(t, &cicerotest.MockSources{}, mt)

	engine := roccotest.TestEngineWithAuth(testAuthenticator())
	engine.WithHandlers(DeleteTranslation)

	resp := roccotest.ServeRequest(engine, http.MethodDelete, "/tenants/zoobzio/translations/uuid-1", nil)

	if resp.StatusCode() != http.StatusNoContent {
		t.Errorf("status: got %d, want %d (body: %s)", resp.StatusCode(), http.StatusNoContent, resp.BodyString())
	}
}

func TestDeleteTranslation_Forbidden(t *testing.T) {
	setupTenantRegistry(t, &cicerotest.MockSources{}, &cicerotest.MockTranslations{})

	engine := roccotest.TestEngineWithAuth(unauthorizedAuthenticator())
	engine.WithHandlers(DeleteTranslation)

	resp := roccotest.ServeRequest(engine, http.MethodDelete, "/tenants/zoobzio/translations/uuid-1", nil)

	if resp.StatusCode() != http.StatusForbidden {
		t.Errorf("status: got %d, want %d (body: %s)", resp.StatusCode(), http.StatusForbidden, resp.BodyString())
	}
}
