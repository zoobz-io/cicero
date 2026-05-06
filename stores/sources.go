// Package stores provides shared data access layer implementations for cicero.
package stores

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/zoobz-io/astql"
	"github.com/zoobz-io/cicero/models"
	"github.com/zoobz-io/sum"
)

// Sources provides database access for source text records.
type Sources struct {
	*sum.Database[models.Source]
}

// NewSources creates a new sources store.
func NewSources(db *sqlx.DB, renderer astql.Renderer) *Sources {
	return &Sources{Database: sum.NewDatabase[models.Source](db, "sources", renderer)}
}

// ListByTenant retrieves all sources for a tenant with pagination.
func (s *Sources) ListByTenant(ctx context.Context, tenantID string, limit, offset int) ([]*models.Source, int, error) {
	results, err := s.Query().
		Where("tenant_id", "=", "tenant_id").
		Limit(limit).
		Offset(offset).
		Exec(ctx, map[string]any{"tenant_id": tenantID})
	if err != nil {
		return nil, 0, err
	}

	count, err := s.Count().
		Where("tenant_id", "=", "tenant_id").
		Exec(ctx, map[string]any{"tenant_id": tenantID})
	if err != nil {
		return nil, 0, err
	}

	return results, int(count), nil
}
