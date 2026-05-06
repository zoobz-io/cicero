package handlers

import "github.com/zoobz-io/rocco"

// All returns all public API handlers for registration with the router.
func All() []rocco.Endpoint {
	return []rocco.Endpoint{
		CreateTranslation,
		GetTranslationsByHash,
		BatchTranslate,
		ListTenantTranslations,
		GetTenantTranslationsByHash,
		UpdateTranslation,
		DeleteTranslation,
		GetCoverage,
	}
}
