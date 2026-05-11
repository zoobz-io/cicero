package handlers

import (
	"strconv"

	"github.com/zoobz-io/cicero/api/contracts"
	"github.com/zoobz-io/cicero/api/transformers"
	"github.com/zoobz-io/cicero/api/wire"
	"github.com/zoobz-io/cicero/internal/auth"
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"
)

// ListTenantTranslations lists all translations for a tenant with pagination.
var ListTenantTranslations = rocco.GET("/tenants/{tenant}/translations", func(req *rocco.Request[rocco.NoBody]) (wire.ListTranslationsResponse, error) {
	tenant := req.Params.Path["tenant"]
	if err := auth.RequireTenantAccess(req.Identity, tenant); err != nil {
		return wire.ListTranslationsResponse{}, ErrForbidden.WithCause(err)
	}

	limit, _ := strconv.Atoi(req.Params.Query["limit"])
	offset, _ := strconv.Atoi(req.Params.Query["offset"])
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	translations := sum.MustUse[contracts.Translations](req.Context)
	list, total, err := translations.ListByTenant(req.Context, tenant, limit, offset)
	if err != nil {
		return wire.ListTranslationsResponse{}, err
	}

	return transformers.TranslationsToList(list, total, limit, offset), nil
}).WithAuthentication().
	WithPathParams("tenant").
	WithQueryParams("limit", "offset").
	WithSummary("List tenant translations").
	WithDescription("Lists all translations for a tenant with pagination.").
	WithTags("Tenants").
	WithErrors(ErrForbidden)

// GetTenantTranslationsByHash retrieves all translations for a hash within a tenant.
var GetTenantTranslationsByHash = rocco.GET("/tenants/{tenant}/translations/{hash}", func(req *rocco.Request[rocco.NoBody]) (wire.TenantTranslationsResponse, error) {
	tenant := req.Params.Path["tenant"]
	if err := auth.RequireTenantAccess(req.Identity, tenant); err != nil {
		return wire.TenantTranslationsResponse{}, ErrForbidden.WithCause(err)
	}

	hash := req.Params.Path["hash"]

	sources := sum.MustUse[contracts.Sources](req.Context)
	src, err := sources.Get(req.Context, hash)
	if err != nil {
		return wire.TenantTranslationsResponse{}, ErrSourceNotFound
	}

	translations := sum.MustUse[contracts.Translations](req.Context)
	list, err := translations.ListByTenantAndHash(req.Context, tenant, hash)
	if err != nil {
		return wire.TenantTranslationsResponse{}, err
	}

	return transformers.TenantTranslationsToResponse(src, list, tenant), nil
}).WithAuthentication().
	WithPathParams("tenant", "hash").
	WithSummary("Get tenant translations by hash").
	WithDescription("Retrieves all translations for a content hash within a tenant.").
	WithTags("Tenants").
	WithErrors(ErrSourceNotFound, ErrForbidden)

// UpdateTranslation edits a translation's text by ID within a tenant.
var UpdateTranslation = rocco.PUT("/tenants/{tenant}/translations/{id}", func(req *rocco.Request[wire.UpdateTranslationRequest]) (wire.TranslationDetail, error) {
	tenant := req.Params.Path["tenant"]
	if err := auth.RequireTenantAccess(req.Identity, tenant); err != nil {
		return wire.TranslationDetail{}, ErrForbidden.WithCause(err)
	}

	id := req.Params.Path["id"]

	translations := sum.MustUse[contracts.Translations](req.Context)
	updated, err := translations.Update(req.Context, id, tenant, req.Body.Text)
	if err != nil {
		return wire.TranslationDetail{}, ErrTranslationNotFound.WithCause(err)
	}

	return transformers.TranslationToDetail(updated), nil
}).WithAuthentication().
	WithPathParams("tenant", "id").
	WithSummary("Update translation").
	WithDescription("Edits a translation's text by ID within a tenant. Invalidates the cache.").
	WithTags("Tenants").
	WithErrors(ErrTranslationNotFound, ErrForbidden)

// DeleteTranslation removes a translation by ID within a tenant.
var DeleteTranslation = rocco.DELETE("/tenants/{tenant}/translations/{id}", func(req *rocco.Request[rocco.NoBody]) (rocco.NoBody, error) {
	tenant := req.Params.Path["tenant"]
	if err := auth.RequireTenantAccess(req.Identity, tenant); err != nil {
		return rocco.NoBody{}, ErrForbidden.WithCause(err)
	}

	id := req.Params.Path["id"]

	translations := sum.MustUse[contracts.Translations](req.Context)
	if err := translations.DeleteByID(req.Context, id, tenant); err != nil {
		return rocco.NoBody{}, ErrTranslationNotFound.WithCause(err)
	}

	return rocco.NoBody{}, nil
}).WithAuthentication().
	WithPathParams("tenant", "id").
	WithSummary("Delete translation").
	WithDescription("Deletes a translation by ID within a tenant.").
	WithTags("Tenants").
	WithErrors(ErrTranslationNotFound, ErrForbidden).
	WithSuccessStatus(204)

// GetCoverage returns translation coverage for a tenant and target language.
var GetCoverage = rocco.GET("/tenants/{tenant}/coverage", func(req *rocco.Request[rocco.NoBody]) (wire.CoverageResponse, error) {
	tenant := req.Params.Path["tenant"]
	if err := auth.RequireTenantAccess(req.Identity, tenant); err != nil {
		return wire.CoverageResponse{}, ErrForbidden.WithCause(err)
	}

	targetLang := req.Params.Query["target_lang"]

	sources := sum.MustUse[contracts.Sources](req.Context)
	_, totalSources, err := sources.ListByTenant(req.Context, tenant, 0, 0)
	if err != nil {
		return wire.CoverageResponse{}, err
	}

	translations := sum.MustUse[contracts.Translations](req.Context)
	allTrans, _, err := translations.ListByTenant(req.Context, tenant, 0, 0)
	if err != nil {
		return wire.CoverageResponse{}, err
	}

	translated := 0
	for _, t := range allTrans {
		if targetLang == "" || t.TargetLang == targetLang {
			translated++
		}
	}

	var pct float64
	if totalSources > 0 {
		pct = float64(translated) / float64(totalSources) * 100
	}

	return wire.CoverageResponse{
		TenantID:   tenant,
		TargetLang: targetLang,
		Total:      totalSources,
		Translated: translated,
		Percentage: pct,
	}, nil
}).WithAuthentication().
	WithPathParams("tenant").
	WithQueryParams("target_lang").
	WithSummary("Translation coverage").
	WithDescription("Returns translation coverage percentage for a tenant and optional target language.").
	WithTags("Tenants").
	WithErrors(ErrForbidden)
