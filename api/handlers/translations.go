package handlers

import (
	"context"
	"log"

	"github.com/zoobz-io/cicero/api/contracts"
	"github.com/zoobz-io/cicero/api/transformers"
	"github.com/zoobz-io/cicero/api/wire"
	"github.com/zoobz-io/cicero/models"
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"
)

// translate handles the core translate-and-store flow for a single text.
// Check cache → miss → call LibreTranslate → store source + translation → return.
func translate(ctx context.Context, text, sourceLang, targetLang, tenantID string) (*models.Source, *models.Translation, error) {
	hash := models.HashText(text)

	translationsStore := sum.MustUse[contracts.Translations](ctx)

	// Dedup: check if this translation already exists for this tenant
	existing, err := translationsStore.GetBySourceAndLang(ctx, hash, sourceLang, targetLang, tenantID)
	if err == nil && existing != nil {
		sourcesStore := sum.MustUse[contracts.Sources](ctx)
		src, _ := sourcesStore.Get(ctx, hash)
		if src == nil {
			src = &models.Source{Hash: hash, Text: text, TenantID: tenantID}
		}
		return src, existing, nil
	}

	// Translate via LibreTranslate
	translator := sum.MustUse[contracts.Translator](ctx)
	translatedText, provider, err := translator.Translate(ctx, text, sourceLang, targetLang)
	if err != nil {
		return nil, nil, err
	}

	// Store source (idempotent by hash)
	src := &models.Source{Hash: hash, Text: text, TenantID: tenantID}
	sourcesStore := sum.MustUse[contracts.Sources](ctx)
	if err := sourcesStore.Set(ctx, hash, src); err != nil {
		return nil, nil, err
	}

	// Store translation
	translation := &models.Translation{
		SourceHash: hash,
		SourceLang: sourceLang,
		TargetLang: targetLang,
		Text:       translatedText,
		Provider:   provider,
		Status:     "completed",
		TenantID:   tenantID,
	}
	if err := translationsStore.Set(ctx, "", translation); err != nil {
		return nil, nil, err
	}

	return src, translation, nil
}

// CreateTranslation submits text for translation.
var CreateTranslation = rocco.POST("/translations", func(req *rocco.Request[wire.TranslateRequest]) (wire.TranslateResponse, error) {
	src, translation, err := translate(req.Context, req.Body.Text, req.Body.SourceLang, req.Body.TargetLang, req.Body.TenantID)
	if err != nil {
		log.Printf("translation error: %v", err)
		return wire.TranslateResponse{}, ErrTranslationFailed.WithCause(err)
	}

	return transformers.SourceAndTranslationToResponse(src, translation), nil
}).WithSummary("Submit translation").
	WithDescription("Submits text for translation. Returns the content hash and translation result.").
	WithTags("Translations").
	WithErrors(ErrTranslationFailed).
	WithSuccessStatus(201)

// GetTranslationsByHash retrieves the source text and all translations for a given content hash.
var GetTranslationsByHash = rocco.GET("/translations/{hash}", func(req *rocco.Request[rocco.NoBody]) (wire.TranslationsByHashResponse, error) {
	hash := req.Params.Path["hash"]

	sources := sum.MustUse[contracts.Sources](req.Context)
	src, err := sources.Get(req.Context, hash)
	if err != nil {
		return wire.TranslationsByHashResponse{}, ErrSourceNotFound
	}

	translations := sum.MustUse[contracts.Translations](req.Context)
	list, err := translations.ListBySourceHash(req.Context, hash)
	if err != nil {
		return wire.TranslationsByHashResponse{}, err
	}

	return transformers.SourceAndTranslationsToHashResponse(src, list), nil
}).WithPathParams("hash").
	WithSummary("Get translations by hash").
	WithDescription("Retrieves the source text and all translations for a given content hash.").
	WithTags("Translations").
	WithErrors(ErrSourceNotFound)
