package stores

import (
	"github.com/jmoiron/sqlx"
	"github.com/zoobz-io/astql"
	"github.com/zoobz-io/cicero/models"
	"github.com/zoobz-io/grub"
)

// Stores aggregates all data access implementations.
type Stores struct {
	Sources      *Sources
	Translations *Translations
}

// New creates all stores and returns the aggregate.
func New(db *sqlx.DB, renderer astql.Renderer, translationCache *grub.Store[models.Translation]) *Stores {
	return &Stores{
		Sources:      NewSources(db, renderer),
		Translations: NewTranslations(db, renderer, translationCache),
	}
}
