package handlers

import (
	"log"

	"github.com/zoobz-io/cicero/api/transformers"
	"github.com/zoobz-io/cicero/api/wire"
	"github.com/zoobz-io/rocco"
)

// BatchTranslate translates multiple texts in a single request.
var BatchTranslate = rocco.POST("/translations/batch", func(req *rocco.Request[wire.BatchTranslateRequest]) (wire.BatchTranslateResponse, error) {
	results := make([]wire.TranslateResponse, 0, len(req.Body.Texts))

	for _, text := range req.Body.Texts {
		src, translation, err := translate(req.Context, text, req.Body.SourceLang, req.Body.TargetLang, req.Body.TenantID)
		if err != nil {
			log.Printf("batch translation error: %v", err)
			return wire.BatchTranslateResponse{}, ErrTranslationFailed.WithCause(err)
		}

		results = append(results, transformers.SourceAndTranslationToResponse(src, translation))
	}

	return wire.BatchTranslateResponse{Translations: results}, nil
}).WithSummary("Batch translate").
	WithDescription("Translates multiple texts in a single request. Each text is deduplicated and cached.").
	WithTags("Translations").
	WithErrors(ErrTranslationFailed).
	WithSuccessStatus(201)
