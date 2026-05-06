package stores

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/zoobz-io/astql"
	"github.com/zoobz-io/cicero/models"
	"github.com/zoobz-io/grub"
	"github.com/zoobz-io/sum"
)

const cacheTTL = 24 * time.Hour

// Translations provides database access for translation records with an optional Redis cache.
type Translations struct {
	*sum.Database[models.Translation]
	cache *grub.Store[models.Translation]
}

// NewTranslations creates a new translations store with optional cache.
func NewTranslations(db *sqlx.DB, renderer astql.Renderer, cache *grub.Store[models.Translation]) *Translations {
	return &Translations{
		Database: sum.NewDatabase[models.Translation](db, "translations", renderer),
		cache:    cache,
	}
}

func cacheKey(sourceHash, sourceLang, targetLang, tenantID string) string {
	return fmt.Sprintf("translations:%s:%s:%s:%s", tenantID, sourceHash, sourceLang, targetLang)
}

// GetBySourceAndLang retrieves a translation by source hash, language pair, and tenant.
// Checks cache first, falls back to Postgres, writes through on miss.
func (s *Translations) GetBySourceAndLang(ctx context.Context, sourceHash, sourceLang, targetLang, tenantID string) (*models.Translation, error) {
	key := cacheKey(sourceHash, sourceLang, targetLang, tenantID)

	// Cache check
	if s.cache != nil {
		if cached, err := s.cache.Get(ctx, key); err == nil {
			return cached, nil
		}
	}

	// Database fallback
	t, err := s.Select().
		Where("source_hash", "=", "source_hash").
		Where("source_lang", "=", "source_lang").
		Where("target_lang", "=", "target_lang").
		Where("tenant_id", "=", "tenant_id").
		Exec(ctx, map[string]any{
			"source_hash": sourceHash,
			"source_lang": sourceLang,
			"target_lang": targetLang,
			"tenant_id":   tenantID,
		})
	if err != nil {
		return nil, err
	}

	// Write-through
	if s.cache != nil {
		_ = s.cache.Set(ctx, key, t, cacheTTL)
	}

	return t, nil
}

// ListBySourceHash retrieves all translations for a given source hash.
func (s *Translations) ListBySourceHash(ctx context.Context, sourceHash string) ([]*models.Translation, error) {
	return s.Query().
		Where("source_hash", "=", "source_hash").
		Exec(ctx, map[string]any{"source_hash": sourceHash})
}

// ListByTenant retrieves all translations for a tenant with pagination.
func (s *Translations) ListByTenant(ctx context.Context, tenantID string, limit, offset int) ([]*models.Translation, int, error) {
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

// ListByTenantAndHash retrieves all translations for a hash within a tenant.
func (s *Translations) ListByTenantAndHash(ctx context.Context, tenantID, sourceHash string) ([]*models.Translation, error) {
	return s.Query().
		Where("tenant_id", "=", "tenant_id").
		Where("source_hash", "=", "source_hash").
		Exec(ctx, map[string]any{
			"tenant_id":   tenantID,
			"source_hash": sourceHash,
		})
}

// Update modifies a translation's text by ID within a tenant. Invalidates cache.
func (s *Translations) Update(ctx context.Context, id, tenantID, text string) (*models.Translation, error) {
	t, err := s.Select().
		Where("id", "=", "id").
		Where("tenant_id", "=", "tenant_id").
		Exec(ctx, map[string]any{"id": id, "tenant_id": tenantID})
	if err != nil {
		return nil, err
	}

	t.Text = text
	if err := s.Set(ctx, t.ID, t); err != nil {
		return nil, err
	}

	// Invalidate cache
	if s.cache != nil {
		key := cacheKey(t.SourceHash, t.SourceLang, t.TargetLang, t.TenantID)
		_ = s.cache.Delete(ctx, key)
	}

	return t, nil
}

// DeleteByID removes a translation by ID within a tenant. Invalidates cache.
func (s *Translations) DeleteByID(ctx context.Context, id, tenantID string) error {
	// Fetch first to get cache key
	t, err := s.Select().
		Where("id", "=", "id").
		Where("tenant_id", "=", "tenant_id").
		Exec(ctx, map[string]any{"id": id, "tenant_id": tenantID})
	if err != nil {
		return err
	}

	if err := s.Delete(ctx, t.ID); err != nil {
		return err
	}

	// Invalidate cache
	if s.cache != nil {
		key := cacheKey(t.SourceHash, t.SourceLang, t.TargetLang, t.TenantID)
		_ = s.cache.Delete(ctx, key)
	}

	return nil
}
