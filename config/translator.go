package config

import "github.com/zoobz-io/check"

// Translator holds configuration for the LibreTranslate client.
type Translator struct {
	Addr string `env:"APP_LIBRETRANSLATE_ADDR" default:"http://localhost:5000"`
}

// Validate checks that the translator configuration is valid.
func (t Translator) Validate() error {
	return check.All(
		check.Str(t.Addr, "addr").Required().V(),
	).Err()
}
