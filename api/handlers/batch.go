package handlers

import (
	"log"

	"github.com/zoobz-io/cicero/api/transformers"
	"github.com/zoobz-io/cicero/api/wire"
	"github.com/zoobz-io/cicero/internal/auth"
	"github.com/zoobz-io/rocco"
)

// BatchTranslate translates multiple texts in a single request.
var BatchTranslate = rocco.POST("/translations/batch", func(req *rocco.Request[wire.BatchTranslateRequest]) (wire.BatchTranslateResponse, error) {
	tenantID, err := auth.TenantFromIdentity(req.Identity)
	if err != nil {
		return wire.BatchTranslateResponse{}, ErrUnauthorized.WithCause(err)
	}

	results := make([]wire.TranslateResponse, 0, len(req.Body.Texts))

	for _, text := range req.Body.Texts {
		src, translation, err := translate(req.Context, text, req.Body.SourceLang, req.Body.TargetLang, tenantID)
		if err != nil {
			log.Printf("batch translation error: %v", err)
			return wire.BatchTranslateResponse{}, ErrTranslationFailed.WithCause(err)
		}

		results = append(results, transformers.SourceAndTranslationToResponse(src, translation))
	}

	return wire.BatchTranslateResponse{Translations: results}, nil
}).WithAuthentication().
	WithSummary("Batch translate").
	WithDescription("Translates multiple texts in a single request. Tenant is derived from the authenticated user.").
	WithTags("Translations").
	WithErrors(ErrTranslationFailed, ErrUnauthorized).
	WithSuccessStatus(201)
