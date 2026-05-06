//go:build testing

package transformers

import (
	"testing"
	"time"

	"github.com/zoobz-io/cicero/models"
)

func TestSourceAndTranslationToResponse(t *testing.T) {
	source := &models.Source{
		Hash:     "315f5bdb76d078c43b8ac0064e4a0164",
		Text:     "Hello, world!",
		TenantID: "zoobzio",
	}
	translation := &models.Translation{
		SourceHash: "315f5bdb76d078c43b8ac0064e4a0164",
		SourceLang: "en",
		TargetLang: "es",
		Text:       "¡Hola, mundo!",
		Provider:   "sidecar",
		Status:     "completed",
	}

	resp := SourceAndTranslationToResponse(source, translation)

	if resp.Hash != source.Hash {
		t.Errorf("Hash: got %q, want %q", resp.Hash, source.Hash)
	}
	if resp.SourceText != source.Text {
		t.Errorf("SourceText: got %q, want %q", resp.SourceText, source.Text)
	}
	if resp.TranslatedText != translation.Text {
		t.Errorf("TranslatedText: got %q, want %q", resp.TranslatedText, translation.Text)
	}
	if resp.SourceLang != translation.SourceLang {
		t.Errorf("SourceLang: got %q, want %q", resp.SourceLang, translation.SourceLang)
	}
	if resp.TargetLang != translation.TargetLang {
		t.Errorf("TargetLang: got %q, want %q", resp.TargetLang, translation.TargetLang)
	}
	if resp.Provider != translation.Provider {
		t.Errorf("Provider: got %q, want %q", resp.Provider, translation.Provider)
	}
	if resp.Status != translation.Status {
		t.Errorf("Status: got %q, want %q", resp.Status, translation.Status)
	}
}

func TestSourceAndTranslationsToHashResponse(t *testing.T) {
	source := &models.Source{
		Hash:     "315f5bdb76d078c43b8ac0064e4a0164",
		Text:     "Hello, world!",
		TenantID: "zoobzio",
	}
	translations := []*models.Translation{
		{
			SourceLang: "en",
			TargetLang: "es",
			Text:       "¡Hola, mundo!",
			Provider:   "sidecar",
			Status:     "completed",
			CreatedAt:  time.Date(2026, 3, 3, 0, 0, 0, 0, time.UTC),
		},
		{
			SourceLang: "en",
			TargetLang: "fr",
			Text:       "Bonjour, monde!",
			Provider:   "sidecar",
			Status:     "completed",
			CreatedAt:  time.Date(2026, 3, 3, 0, 0, 0, 0, time.UTC),
		},
	}

	resp := SourceAndTranslationsToHashResponse(source, translations)

	if resp.Hash != source.Hash {
		t.Errorf("Hash: got %q, want %q", resp.Hash, source.Hash)
	}
	if resp.SourceText != source.Text {
		t.Errorf("SourceText: got %q, want %q", resp.SourceText, source.Text)
	}
	if len(resp.Translations) != 2 {
		t.Fatalf("Translations length: got %d, want 2", len(resp.Translations))
	}
	if resp.Translations[0].TargetLang != "es" {
		t.Errorf("Translations[0].TargetLang: got %q, want %q", resp.Translations[0].TargetLang, "es")
	}
	if resp.Translations[1].TargetLang != "fr" {
		t.Errorf("Translations[1].TargetLang: got %q, want %q", resp.Translations[1].TargetLang, "fr")
	}
}

func TestSourceAndTranslationsToHashResponse_Empty(t *testing.T) {
	source := &models.Source{
		Hash: "abc",
		Text: "Hello",
	}

	resp := SourceAndTranslationsToHashResponse(source, nil)

	if resp.Hash != source.Hash {
		t.Errorf("Hash: got %q, want %q", resp.Hash, source.Hash)
	}
	if len(resp.Translations) != 0 {
		t.Errorf("Translations length: got %d, want 0", len(resp.Translations))
	}
}

func TestTranslationToDetail(t *testing.T) {
	ts := time.Date(2026, 3, 3, 12, 0, 0, 0, time.UTC)
	tr := &models.Translation{
		SourceLang: "en",
		TargetLang: "es",
		Text:       "¡Hola!",
		Provider:   "sidecar",
		Status:     "completed",
		CreatedAt:  ts,
	}

	detail := TranslationToDetail(tr)

	if detail.SourceLang != tr.SourceLang {
		t.Errorf("SourceLang: got %q, want %q", detail.SourceLang, tr.SourceLang)
	}
	if detail.TargetLang != tr.TargetLang {
		t.Errorf("TargetLang: got %q, want %q", detail.TargetLang, tr.TargetLang)
	}
	if detail.TranslatedText != tr.Text {
		t.Errorf("TranslatedText: got %q, want %q", detail.TranslatedText, tr.Text)
	}
	if detail.Provider != tr.Provider {
		t.Errorf("Provider: got %q, want %q", detail.Provider, tr.Provider)
	}
	if detail.Status != tr.Status {
		t.Errorf("Status: got %q, want %q", detail.Status, tr.Status)
	}
	wantCreatedAt := "2026-03-03T12:00:00Z"
	if detail.CreatedAt != wantCreatedAt {
		t.Errorf("CreatedAt: got %q, want %q", detail.CreatedAt, wantCreatedAt)
	}
}
