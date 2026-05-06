// Package transformers provides pure functions for mapping between domain models and API wire types.
package transformers

import (
	"github.com/zoobz-io/cicero/api/wire"
	"github.com/zoobz-io/cicero/models"
)

// SourceAndTranslationToResponse maps a source and translation model to a TranslateResponse.
func SourceAndTranslationToResponse(source *models.Source, translation *models.Translation) wire.TranslateResponse {
	return wire.TranslateResponse{
		Hash:           source.Hash,
		SourceText:     source.Text,
		TranslatedText: translation.Text,
		SourceLang:     translation.SourceLang,
		TargetLang:     translation.TargetLang,
		Provider:       translation.Provider,
		Status:         translation.Status,
	}
}

// SourceAndTranslationsToHashResponse maps a source and its translations to a TranslationsByHashResponse.
func SourceAndTranslationsToHashResponse(source *models.Source, translations []*models.Translation) wire.TranslationsByHashResponse {
	details := make([]wire.TranslationDetail, len(translations))
	for i, t := range translations {
		details[i] = TranslationToDetail(t)
	}
	return wire.TranslationsByHashResponse{
		Hash:         source.Hash,
		SourceText:   source.Text,
		Translations: details,
	}
}

// TranslationToDetail maps a single translation model to a TranslationDetail.
func TranslationToDetail(t *models.Translation) wire.TranslationDetail {
	return wire.TranslationDetail{
		ID:             t.ID,
		SourceHash:     t.SourceHash,
		SourceLang:     t.SourceLang,
		TargetLang:     t.TargetLang,
		TranslatedText: t.Text,
		Provider:       t.Provider,
		Status:         t.Status,
		TenantID:       t.TenantID,
		CreatedAt:      t.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// TranslationsToList maps a slice of translations with pagination to a ListTranslationsResponse.
func TranslationsToList(translations []*models.Translation, total, limit, offset int) wire.ListTranslationsResponse {
	details := make([]wire.TranslationDetail, len(translations))
	for i, t := range translations {
		details[i] = TranslationToDetail(t)
	}
	return wire.ListTranslationsResponse{
		Translations: details,
		Total:        total,
		Limit:        limit,
		Offset:       offset,
	}
}

// TenantTranslationsToResponse maps a source and tenant-scoped translations.
func TenantTranslationsToResponse(source *models.Source, translations []*models.Translation, tenantID string) wire.TenantTranslationsResponse {
	details := make([]wire.TranslationDetail, len(translations))
	for i, t := range translations {
		details[i] = TranslationToDetail(t)
	}
	return wire.TenantTranslationsResponse{
		Hash:         source.Hash,
		SourceText:   source.Text,
		TenantID:     tenantID,
		Translations: details,
	}
}
