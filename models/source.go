// Package models provides domain model types for cicero.
package models

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/zoobz-io/check"
)

// HashText computes the SHA-256 content hash of text,
// truncated to 32 hex characters. This is the content-addressable
// key used for deduplication and retrieval.
func HashText(text string) string {
	sum := sha256.Sum256([]byte(text))
	return hex.EncodeToString(sum[:])[:32]
}

// Source represents content-addressable source text.
// The Hash field is derived from the source text and serves as the primary key,
// ensuring that identical text from any caller produces the same record.
type Source struct {
	CreatedAt time.Time `json:"created_at" db:"created_at" default:"now()" description:"Record creation timestamp"`
	Hash      string    `json:"hash" db:"hash" constraints:"primarykey" description:"SHA-256 truncated to 32 hex chars" example:"a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4"`
	Text      string    `json:"text" db:"text" constraints:"notnull" description:"Original source text" example:"Hello, world!"`
	TenantID  string    `json:"tenant_id" db:"tenant_id" constraints:"notnull" description:"Tenant identifier" example:"zoobzio"`
}

// Clone returns a shallow copy of the source.
func (s Source) Clone() Source {
	return s
}

// Validate checks that required fields are present.
func (s Source) Validate() error {
	return check.All(
		check.Str(s.Hash, "hash").Required().V(),
		check.Str(s.Text, "text").Required().V(),
		check.Str(s.TenantID, "tenant_id").Required().V(),
	).Err()
}
