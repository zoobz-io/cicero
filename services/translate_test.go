//go:build testing

package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/zoobz-io/cicero/api/contracts"
)

// Compile-time assertion: TranslateService satisfies contracts.Translator.
var _ contracts.Translator = (*TranslateService)(nil)

func setupTestServer(t *testing.T, handler http.HandlerFunc) (*TranslateService, func()) {
	t.Helper()
	srv := httptest.NewServer(handler)
	svc := NewTranslateService(srv.URL)
	return svc, func() {
		_ = svc.Close()
		srv.Close()
	}
}

func TestTranslateService_Success(t *testing.T) {
	svc, cleanup := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(libreTranslateResponse{TranslatedText: "¡Hola, mundo!"})
	})
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, prov, err := svc.Translate(ctx, "Hello, world!", "en", "es")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "¡Hola, mundo!" {
		t.Errorf("result: got %q, want %q", result, "¡Hola, mundo!")
	}
	if prov != libreTranslateProvider {
		t.Errorf("provider: got %q, want %q", prov, libreTranslateProvider)
	}
}

func TestTranslateService_RequestFields(t *testing.T) {
	var captured libreTranslateRequest

	svc, cleanup := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&captured)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(libreTranslateResponse{TranslatedText: "Bonjour"})
	})
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, _, err := svc.Translate(ctx, "Hello", "en", "fr")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if captured.Q != "Hello" {
		t.Errorf("Q: got %q, want %q", captured.Q, "Hello")
	}
	if captured.Source != "en" {
		t.Errorf("Source: got %q, want %q", captured.Source, "en")
	}
	if captured.Target != "fr" {
		t.Errorf("Target: got %q, want %q", captured.Target, "fr")
	}
}

func TestTranslateService_ServerError(t *testing.T) {
	svc, cleanup := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, _, err := svc.Translate(ctx, "Hello", "en", "es")
	if err == nil {
		t.Fatal("expected error from server, got nil")
	}
}

func TestTranslateService_BadRequest(t *testing.T) {
	svc, cleanup := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(libreTranslateError{Error: "unsupported language"})
	})
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, _, err := svc.Translate(ctx, "Hello", "en", "xx")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestTranslateService_Close(t *testing.T) {
	svc, cleanup := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {})
	_ = cleanup

	if err := svc.Close(); err != nil {
		t.Errorf("Close() returned error: %v", err)
	}
}
