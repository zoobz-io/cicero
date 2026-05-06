//go:build testing

package models

import (
	"testing"
	"time"
)

func TestTranslation_Clone(t *testing.T) {
	original := Translation{
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

	clone := original.Clone()

	if clone.ID != original.ID {
		t.Errorf("ID: got %q, want %q", clone.ID, original.ID)
	}
	if clone.SourceHash != original.SourceHash {
		t.Errorf("SourceHash: got %q, want %q", clone.SourceHash, original.SourceHash)
	}
	if clone.SourceLang != original.SourceLang {
		t.Errorf("SourceLang: got %q, want %q", clone.SourceLang, original.SourceLang)
	}
	if clone.TargetLang != original.TargetLang {
		t.Errorf("TargetLang: got %q, want %q", clone.TargetLang, original.TargetLang)
	}
	if clone.Text != original.Text {
		t.Errorf("Text: got %q, want %q", clone.Text, original.Text)
	}
	if clone.Provider != original.Provider {
		t.Errorf("Provider: got %q, want %q", clone.Provider, original.Provider)
	}
	if clone.Status != original.Status {
		t.Errorf("Status: got %q, want %q", clone.Status, original.Status)
	}
	if clone.TenantID != original.TenantID {
		t.Errorf("TenantID: got %q, want %q", clone.TenantID, original.TenantID)
	}

	// Mutations to clone do not affect original.
	clone.Status = "failed"
	if original.Status == "failed" {
		t.Error("mutating clone affected original Status")
	}
}

func TestTranslation_Validate(t *testing.T) {
	valid := Translation{
		SourceHash: "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4",
		SourceLang: "en",
		TargetLang: "es",
		Text:       "¡Hola, mundo!",
		Provider:   "sidecar",
		Status:     "completed",
		TenantID:   "zoobzio",
	}

	tests := []struct {
		name    string
		mutate  func(*Translation)
		wantErr bool
	}{
		{
			name:    "valid",
			mutate:  nil,
			wantErr: false,
		},
		{
			name:    "missing source_hash",
			mutate:  func(tr *Translation) { tr.SourceHash = "" },
			wantErr: true,
		},
		{
			name:    "missing source_lang",
			mutate:  func(tr *Translation) { tr.SourceLang = "" },
			wantErr: true,
		},
		{
			name:    "missing target_lang",
			mutate:  func(tr *Translation) { tr.TargetLang = "" },
			wantErr: true,
		},
		{
			name:    "missing text",
			mutate:  func(tr *Translation) { tr.Text = "" },
			wantErr: true,
		},
		{
			name:    "missing provider",
			mutate:  func(tr *Translation) { tr.Provider = "" },
			wantErr: true,
		},
		{
			name:    "missing status",
			mutate:  func(tr *Translation) { tr.Status = "" },
			wantErr: true,
		},
		{
			name:    "missing tenant_id",
			mutate:  func(tr *Translation) { tr.TenantID = "" },
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tr := valid
			if tc.mutate != nil {
				tc.mutate(&tr)
			}
			err := tr.Validate()
			if tc.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
