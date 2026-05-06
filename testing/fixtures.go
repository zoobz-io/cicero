//go:build testing

// Package testing provides test infrastructure for cicero.
package testing

import (
	"testing"
	"time"

	"github.com/zoobz-io/cicero/models"
)

// NewSource returns a Source with sensible defaults.
func NewSource(t *testing.T) *models.Source {
	t.Helper()
	return &models.Source{
		Hash:      "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4",
		Text:      "Hello, world!",
		TenantID:  "zoobzio",
		CreatedAt: time.Now(),
	}
}

// NewTranslation returns a Translation with sensible defaults.
func NewTranslation(t *testing.T) *models.Translation {
	t.Helper()
	return &models.Translation{
		ID:         "test-uuid-1",
		SourceHash: "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4",
		SourceLang: "en",
		TargetLang: "es",
		Text:       "¡Hola, mundo!",
		Provider:   "sidecar",
		Status:     "completed",
		TenantID:   "zoobzio",
		CreatedAt:  time.Now(),
	}
}
