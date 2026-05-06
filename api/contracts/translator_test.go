//go:build testing

package contracts_test

import (
	"context"
	"testing"

	"github.com/zoobz-io/cicero/api/contracts"
	cicerotest "github.com/zoobz-io/cicero/testing"
)

// Compile-time assertion: MockTranslator satisfies the Translator interface.
var _ contracts.Translator = (*cicerotest.MockTranslator)(nil)

func TestTranslatorInterface_Success(t *testing.T) {
	mock := &cicerotest.MockTranslator{
		OnTranslate: func(_ context.Context, _, _, _ string) (string, string, error) {
			return "¡Hola, mundo!", "libretranslate", nil
		},
	}

	ctx := context.Background()
	result, provider, err := mock.Translate(ctx, "Hello", "en", "es")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "¡Hola, mundo!" {
		t.Errorf("result: got %q, want %q", result, "¡Hola, mundo!")
	}
	if provider != "libretranslate" {
		t.Errorf("provider: got %q, want %q", provider, "libretranslate")
	}
}

func TestTranslatorInterface_ZeroDefault(t *testing.T) {
	mock := &cicerotest.MockTranslator{}

	ctx := context.Background()
	result, provider, err := mock.Translate(ctx, "Hello", "en", "es")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "" {
		t.Errorf("zero result: got %q, want empty", result)
	}
	if provider != "" {
		t.Errorf("zero provider: got %q, want empty", provider)
	}
}
