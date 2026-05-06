package models

import (
	"time"

	"github.com/zoobz-io/check"
)

// Translation represents a per-language translation of a source text.
// Each translation is associated with a source via SourceHash and is unique
// per (source_hash, source_lang, target_lang, tenant_id) combination.
type Translation struct {
	CreatedAt  time.Time `json:"created_at" db:"created_at" default:"now()" description:"Record creation timestamp"`
	SourceHash string    `json:"source_hash" db:"source_hash" constraints:"notnull" references:"sources(hash)" description:"Content hash of the source text"`
	SourceLang string    `json:"source_lang" db:"source_lang" constraints:"notnull" description:"Source language code" example:"en"`
	TargetLang string    `json:"target_lang" db:"target_lang" constraints:"notnull" description:"Target language code" example:"es"`
	Text       string    `json:"text" db:"text" constraints:"notnull" description:"Translated content"`
	Provider   string    `json:"provider" db:"provider" constraints:"notnull" description:"Provider that produced this translation" example:"libretranslate"`
	Status     string    `json:"status" db:"status" constraints:"notnull" default:"'completed'" description:"Translation status: completed or failed"`
	TenantID   string    `json:"tenant_id" db:"tenant_id" constraints:"notnull" description:"Tenant identifier" example:"zoobzio"`
	ID         string    `json:"id" db:"id" constraints:"primarykey" default:"gen_random_uuid()" description:"UUID primary key"`
}

// Clone returns a shallow copy of the translation.
func (t Translation) Clone() Translation {
	return t
}

// Validate checks that required fields are present and valid.
func (t Translation) Validate() error {
	return check.All(
		check.Str(t.SourceHash, "source_hash").Required().V(),
		check.Str(t.SourceLang, "source_lang").Required().V(),
		check.Str(t.TargetLang, "target_lang").Required().V(),
		check.Str(t.Text, "text").Required().V(),
		check.Str(t.Provider, "provider").Required().V(),
		check.Str(t.Status, "status").Required().V(),
		check.Str(t.TenantID, "tenant_id").Required().V(),
	).Err()
}
