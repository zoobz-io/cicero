package contracts

import "context"

// Translator defines the contract for external translation provider calls.
type Translator interface {
	// Translate sends text to the translation provider and returns the translated text
	// and the provider name that handled the request.
	Translate(ctx context.Context, text, sourceLang, targetLang string) (result string, provider string, err error)
}
