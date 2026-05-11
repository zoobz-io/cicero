// Package handlers provides HTTP handlers for the public API.
package handlers

import "github.com/zoobz-io/rocco"

// Domain errors for the public API.
var (
	// ErrUnauthorized is returned when the user is not authenticated or has no tenant access.
	ErrUnauthorized = rocco.ErrUnauthorized.WithMessage("unauthorized")
	// ErrForbidden is returned when the user does not have access to the requested tenant.
	ErrForbidden = rocco.ErrForbidden.WithMessage("forbidden")
	// ErrSourceNotFound is returned when a source hash has no matching record.
	ErrSourceNotFound = rocco.ErrNotFound.WithMessage("source not found")
	// ErrTranslationNotFound is returned when a translation ID has no matching record.
	ErrTranslationNotFound = rocco.ErrNotFound.WithMessage("translation not found")
	// ErrTranslationFailed is returned when the translation provider returns an error.
	ErrTranslationFailed = rocco.ErrInternalServer.WithMessage("translation failed")
)
