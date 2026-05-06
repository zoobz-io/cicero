package contracts

import (
	"context"

	"github.com/zoobz-io/cicero/models"
)

// Translations defines the contract for translation storage operations.
type Translations interface {
	// GetBySourceAndLang retrieves a translation by source hash, language pair, and tenant.
	GetBySourceAndLang(ctx context.Context, sourceHash, sourceLang, targetLang, tenantID string) (*models.Translation, error)
	// ListBySourceHash retrieves all translations associated with a source hash.
	ListBySourceHash(ctx context.Context, sourceHash string) ([]*models.Translation, error)
	// ListByTenant retrieves all translations for a tenant with pagination.
	ListByTenant(ctx context.Context, tenantID string, limit, offset int) ([]*models.Translation, int, error)
	// ListByTenantAndHash retrieves all translations for a hash within a tenant.
	ListByTenantAndHash(ctx context.Context, tenantID, sourceHash string) ([]*models.Translation, error)
	// Set stores a translation record.
	Set(ctx context.Context, key string, translation *models.Translation) error
	// Update modifies a translation's text by ID within a tenant.
	Update(ctx context.Context, id, tenantID, text string) (*models.Translation, error)
	// DeleteByID removes a translation by ID within a tenant.
	DeleteByID(ctx context.Context, id, tenantID string) error
}
