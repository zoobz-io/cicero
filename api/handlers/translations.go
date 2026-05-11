package handlers

import (
	"context"
	"log"

	"github.com/zoobz-io/cicero/api/contracts"
	"github.com/zoobz-io/cicero/api/transformers"
	"github.com/zoobz-io/cicero/api/wire"
	"github.com/zoobz-io/cicero/internal/auth"
	"github.com/zoobz-io/cicero/models"
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"
)

// translate handles the core translate-and-store flow for a single text.
func translate(ctx context.Context, text, sourceLang, targetLang, tenantID string) (*models.Source, *models.Translation, error) {
	hash := models.HashText(text)

	translationsStore := sum.MustUse[contracts.Translations](ctx)

	existing, err := translationsStore.GetBySourceAndLang(ctx, hash, sourceLang, targetLang, tenantID)
	if err == nil && existing != nil {
		sourcesStore := sum.MustUse[contracts.Sources](ctx)
		src, _ := sourcesStore.Get(ctx, hash)
		if src == nil {
			src = &models.Source{Hash: hash, Text: text, TenantID: tenantID}
		}
		return src, existing, nil
	}

	translator := sum.MustUse[contracts.Translator](ctx)
	translatedText, provider, err := translator.Translate(ctx, text, sourceLang, targetLang)
	if err != nil {
		return nil, nil, err
	}

	src := &models.Source{Hash: hash, Text: text, TenantID: tenantID}
	sourcesStore := sum.MustUse[contracts.Sources](ctx)
	if err := sourcesStore.Set(ctx, hash, src); err != nil {
		return nil, nil, err
	}

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
	tenantID, err := auth.TenantFromIdentity(req.Identity)
	if err != nil {
		return wire.TranslateResponse{}, ErrUnauthorized.WithCause(err)
	}

	src, translation, err := translate(req.Context, req.Body.Text, req.Body.SourceLang, req.Body.TargetLang, tenantID)
	if err != nil {
		log.Printf("translation error: %v", err)
		return wire.TranslateResponse{}, ErrTranslationFailed.WithCause(err)
	}

	return transformers.SourceAndTranslationToResponse(src, translation), nil
}).WithAuthentication().
	WithSummary("Submit translation").
	WithDescription("Submits text for translation. Tenant is derived from the authenticated user.").
	WithTags("Translations").
	WithErrors(ErrTranslationFailed, ErrUnauthorized).
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
}).WithAuthentication().
	WithPathParams("hash").
	WithSummary("Get translations by hash").
	WithDescription("Retrieves the source text and all translations for a given content hash.").
	WithTags("Translations").
	WithErrors(ErrSourceNotFound)
