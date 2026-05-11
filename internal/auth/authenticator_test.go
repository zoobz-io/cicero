//go:build testing

package auth

import (
	"net/http"
	"testing"
)

func TestExtractBearerToken_Valid(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Authorization", "Bearer my-token-123")

	token := extractBearerToken(r)
	if token != "my-token-123" {
		t.Errorf("got %q, want %q", token, "my-token-123")
	}
}

func TestExtractBearerToken_CaseInsensitive(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Authorization", "bearer my-token")

	token := extractBearerToken(r)
	if token != "my-token" {
		t.Errorf("got %q, want %q", token, "my-token")
	}
}

func TestExtractBearerToken_Missing(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	token := extractBearerToken(r)
	if token != "" {
		t.Errorf("got %q, want empty", token)
	}
}

func TestExtractBearerToken_NotBearer(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Authorization", "Basic abc123")

	token := extractBearerToken(r)
	if token != "" {
		t.Errorf("got %q, want empty", token)
	}
}

func TestExtractBearerToken_BearerOnly(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Authorization", "Bearer")

	token := extractBearerToken(r)
	if token != "" {
		t.Errorf("got %q, want empty", token)
	}
}

func TestResolveTenantFromRequest_SingleTenant(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	tenants := []AuthorizedTenant{{TenantID: "t1", TenantName: "Tenant 1"}}

	tid, err := resolveTenantFromRequest(r, tenants)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tid != "t1" {
		t.Errorf("got %q, want %q", tid, "t1")
	}
}

func TestResolveTenantFromRequest_MultiTenant_WithHeader(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("X-Tenant-ID", "t2")
	tenants := []AuthorizedTenant{
		{TenantID: "t1"},
		{TenantID: "t2"},
	}

	tid, err := resolveTenantFromRequest(r, tenants)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tid != "t2" {
		t.Errorf("got %q, want %q", tid, "t2")
	}
}

func TestResolveTenantFromRequest_MultiTenant_NoHeader(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	tenants := []AuthorizedTenant{
		{TenantID: "t1"},
		{TenantID: "t2"},
	}

	_, err := resolveTenantFromRequest(r, tenants)
	if err == nil {
		t.Fatal("expected error for multi-tenant without header")
	}
}

func TestResolveTenantFromRequest_MultiTenant_UnauthorizedTenant(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("X-Tenant-ID", "t3")
	tenants := []AuthorizedTenant{
		{TenantID: "t1"},
		{TenantID: "t2"},
	}

	_, err := resolveTenantFromRequest(r, tenants)
	if err == nil {
		t.Fatal("expected error for unauthorized tenant")
	}
}

func TestResolveTenantFromRequest_NoTenants(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	_, err := resolveTenantFromRequest(r, nil)
	if err == nil {
		t.Fatal("expected error for no tenants")
	}
}
