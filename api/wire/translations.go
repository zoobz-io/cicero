// Package wire provides request and response types for the public API.
package wire

import "github.com/zoobz-io/check"

// TranslateRequest is the request body for submitting text for translation.
type TranslateRequest struct {
	Text       string `json:"text" description:"Source text to translate" example:"Hello, world!"`
	SourceLang string `json:"source_lang" description:"Source language code" example:"en"`
	TargetLang string `json:"target_lang" description:"Target language code" example:"es"`
	TenantID   string `json:"tenant_id" description:"Tenant identifier" example:"zoobzio"`
}

// Validate validates the translation request.
func (r TranslateRequest) Validate() error {
	return check.All(
		check.Str(r.Text, "text").Required().V(),
		check.Str(r.SourceLang, "source_lang").Required().V(),
		check.Str(r.TargetLang, "target_lang").Required().V(),
		check.Str(r.TenantID, "tenant_id").Required().V(),
		check.NotEqual(r.SourceLang, r.TargetLang, "target_lang"),
	).Err()
}

// Clone returns a shallow copy.
func (r TranslateRequest) Clone() TranslateRequest {
	return r
}

// TranslateResponse is the response returned after submitting a translation.
type TranslateResponse struct {
	Hash           string `json:"hash" description:"Content hash for retrieval"`
	SourceText     string `json:"source_text" description:"Original text"`
	TranslatedText string `json:"translated_text" description:"Translated text"`
	SourceLang     string `json:"source_lang" description:"Source language code"`
	TargetLang     string `json:"target_lang" description:"Target language code"`
	Provider       string `json:"provider" description:"Translation provider used"`
	Status         string `json:"status" description:"Translation status"`
}

// Clone returns a shallow copy.
func (r TranslateResponse) Clone() TranslateResponse {
	return r
}

// TranslationsByHashResponse is the response for retrieving translations by content hash.
type TranslationsByHashResponse struct {
	Hash         string              `json:"hash"`
	SourceText   string              `json:"source_text"`
	Translations []TranslationDetail `json:"translations"`
}

// Clone returns a deep copy.
func (r TranslationsByHashResponse) Clone() TranslationsByHashResponse {
	c := r
	if r.Translations != nil {
		c.Translations = make([]TranslationDetail, len(r.Translations))
		copy(c.Translations, r.Translations)
	}
	return c
}

// TranslationDetail is a single translation entry within a hash response.
type TranslationDetail struct {
	ID             string `json:"id"`
	SourceHash     string `json:"source_hash"`
	SourceLang     string `json:"source_lang"`
	TargetLang     string `json:"target_lang"`
	TranslatedText string `json:"translated_text"`
	Provider       string `json:"provider"`
	Status         string `json:"status"`
	TenantID       string `json:"tenant_id"`
	CreatedAt      string `json:"created_at"`
}

// Clone returns a shallow copy.
func (d TranslationDetail) Clone() TranslationDetail {
	return d
}

// BatchTranslateRequest is the request body for batch translation.
type BatchTranslateRequest struct {
	Texts      []string `json:"texts" description:"Source texts to translate"`
	SourceLang string   `json:"source_lang" description:"Source language code" example:"en"`
	TargetLang string   `json:"target_lang" description:"Target language code" example:"es"`
	TenantID   string   `json:"tenant_id" description:"Tenant identifier" example:"zoobzio"`
}

// Validate validates the batch request.
func (r BatchTranslateRequest) Validate() error {
	return check.All(
		check.NotEmpty(r.Texts, "texts"),
		check.Str(r.SourceLang, "source_lang").Required().V(),
		check.Str(r.TargetLang, "target_lang").Required().V(),
		check.Str(r.TenantID, "tenant_id").Required().V(),
		check.NotEqual(r.SourceLang, r.TargetLang, "target_lang"),
	).Err()
}

// Clone returns a deep copy.
func (r BatchTranslateRequest) Clone() BatchTranslateRequest {
	c := r
	c.Texts = make([]string, len(r.Texts))
	copy(c.Texts, r.Texts)
	return c
}

// BatchTranslateResponse is the response for batch translation.
type BatchTranslateResponse struct {
	Translations []TranslateResponse `json:"translations"`
}

// Clone returns a deep copy.
func (r BatchTranslateResponse) Clone() BatchTranslateResponse {
	c := r
	c.Translations = make([]TranslateResponse, len(r.Translations))
	copy(c.Translations, r.Translations)
	return c
}

// ListTranslationsResponse is a paginated list of translations.
type ListTranslationsResponse struct {
	Translations []TranslationDetail `json:"translations"`
	Total        int                 `json:"total"`
	Limit        int                 `json:"limit"`
	Offset       int                 `json:"offset"`
}

// Clone returns a deep copy.
func (r ListTranslationsResponse) Clone() ListTranslationsResponse {
	c := r
	c.Translations = make([]TranslationDetail, len(r.Translations))
	copy(c.Translations, r.Translations)
	return c
}

// TenantTranslationsResponse is the response for tenant-scoped translations by hash.
type TenantTranslationsResponse struct {
	Hash         string              `json:"hash"`
	SourceText   string              `json:"source_text"`
	TenantID     string              `json:"tenant_id"`
	Translations []TranslationDetail `json:"translations"`
}

// Clone returns a deep copy.
func (r TenantTranslationsResponse) Clone() TenantTranslationsResponse {
	c := r
	c.Translations = make([]TranslationDetail, len(r.Translations))
	copy(c.Translations, r.Translations)
	return c
}

// UpdateTranslationRequest is the request body for editing a translation.
type UpdateTranslationRequest struct {
	Text string `json:"text" description:"Updated translation text"`
}

// Validate validates the update request.
func (r UpdateTranslationRequest) Validate() error {
	return check.All(
		check.Str(r.Text, "text").Required().V(),
	).Err()
}

// Clone returns a shallow copy.
func (r UpdateTranslationRequest) Clone() UpdateTranslationRequest {
	return r
}

// CoverageResponse shows translation coverage for a tenant/locale.
type CoverageResponse struct {
	TenantID   string  `json:"tenant_id"`
	TargetLang string  `json:"target_lang"`
	Total      int     `json:"total"`
	Translated int     `json:"translated"`
	Percentage float64 `json:"percentage"`
}

// Clone returns a shallow copy.
func (r CoverageResponse) Clone() CoverageResponse {
	return r
}
