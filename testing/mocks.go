//go:build testing

package testing

import (
	"context"

	"github.com/zoobz-io/cicero/api/contracts"
	"github.com/zoobz-io/cicero/models"
)

// Compile-time assertions: mocks satisfy their contracts.
var _ contracts.Sources = (*MockSources)(nil)
var _ contracts.Translations = (*MockTranslations)(nil)
var _ contracts.Translator = (*MockTranslator)(nil)

// MockSources is a mock implementation of api/contracts.Sources.
type MockSources struct {
	OnGet          func(ctx context.Context, hash string) (*models.Source, error)
	OnSet          func(ctx context.Context, hash string, source *models.Source) error
	OnListByTenant func(ctx context.Context, tenantID string, limit, offset int) ([]*models.Source, int, error)
}

func (m *MockSources) Get(ctx context.Context, hash string) (*models.Source, error) {
	if m.OnGet != nil {
		return m.OnGet(ctx, hash)
	}
	return &models.Source{}, nil
}

func (m *MockSources) Set(ctx context.Context, hash string, source *models.Source) error {
	if m.OnSet != nil {
		return m.OnSet(ctx, hash, source)
	}
	return nil
}

func (m *MockSources) ListByTenant(ctx context.Context, tenantID string, limit, offset int) ([]*models.Source, int, error) {
	if m.OnListByTenant != nil {
		return m.OnListByTenant(ctx, tenantID, limit, offset)
	}
	return nil, 0, nil
}

// MockTranslations is a mock implementation of api/contracts.Translations.
type MockTranslations struct {
	OnGetBySourceAndLang  func(ctx context.Context, sourceHash, sourceLang, targetLang, tenantID string) (*models.Translation, error)
	OnListBySourceHash    func(ctx context.Context, sourceHash string) ([]*models.Translation, error)
	OnListByTenant        func(ctx context.Context, tenantID string, limit, offset int) ([]*models.Translation, int, error)
	OnListByTenantAndHash func(ctx context.Context, tenantID, sourceHash string) ([]*models.Translation, error)
	OnSet                 func(ctx context.Context, key string, translation *models.Translation) error
	OnUpdate              func(ctx context.Context, id, tenantID, text string) (*models.Translation, error)
	OnDeleteByID          func(ctx context.Context, id, tenantID string) error
}

func (m *MockTranslations) GetBySourceAndLang(ctx context.Context, sourceHash, sourceLang, targetLang, tenantID string) (*models.Translation, error) {
	if m.OnGetBySourceAndLang != nil {
		return m.OnGetBySourceAndLang(ctx, sourceHash, sourceLang, targetLang, tenantID)
	}
	return nil, nil
}

func (m *MockTranslations) ListBySourceHash(ctx context.Context, sourceHash string) ([]*models.Translation, error) {
	if m.OnListBySourceHash != nil {
		return m.OnListBySourceHash(ctx, sourceHash)
	}
	return nil, nil
}

func (m *MockTranslations) ListByTenant(ctx context.Context, tenantID string, limit, offset int) ([]*models.Translation, int, error) {
	if m.OnListByTenant != nil {
		return m.OnListByTenant(ctx, tenantID, limit, offset)
	}
	return nil, 0, nil
}

func (m *MockTranslations) ListByTenantAndHash(ctx context.Context, tenantID, sourceHash string) ([]*models.Translation, error) {
	if m.OnListByTenantAndHash != nil {
		return m.OnListByTenantAndHash(ctx, tenantID, sourceHash)
	}
	return nil, nil
}

func (m *MockTranslations) Set(ctx context.Context, key string, translation *models.Translation) error {
	if m.OnSet != nil {
		return m.OnSet(ctx, key, translation)
	}
	return nil
}

func (m *MockTranslations) Update(ctx context.Context, id, tenantID, text string) (*models.Translation, error) {
	if m.OnUpdate != nil {
		return m.OnUpdate(ctx, id, tenantID, text)
	}
	return &models.Translation{}, nil
}

func (m *MockTranslations) DeleteByID(ctx context.Context, id, tenantID string) error {
	if m.OnDeleteByID != nil {
		return m.OnDeleteByID(ctx, id, tenantID)
	}
	return nil
}

// MockTranslator is a mock implementation of api/contracts.Translator.
type MockTranslator struct {
	OnTranslate func(ctx context.Context, text, sourceLang, targetLang string) (string, string, error)
}

func (m *MockTranslator) Translate(ctx context.Context, text, sourceLang, targetLang string) (string, string, error) {
	if m.OnTranslate != nil {
		return m.OnTranslate(ctx, text, sourceLang, targetLang)
	}
	return "", "", nil
}
