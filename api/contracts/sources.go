// Package contracts defines the interfaces consumed by the public API handlers.
package contracts

import (
	"context"

	"github.com/zoobz-io/cicero/models"
)

// Sources defines the contract for source text storage operations.
type Sources interface {
	// Get retrieves a source by its content hash.
	Get(ctx context.Context, hash string) (*models.Source, error)
	// Set stores a source record, inserting or updating by hash.
	Set(ctx context.Context, hash string, source *models.Source) error
	// ListByTenant retrieves all sources for a tenant with pagination.
	ListByTenant(ctx context.Context, tenantID string, limit, offset int) ([]*models.Source, int, error)
}
