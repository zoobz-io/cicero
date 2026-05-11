//go:build testing

package auth

import (
	"testing"

	"github.com/zoobz-io/rocco"
)

func TestTenantFromIdentity_Success(t *testing.T) {
	id := &Identity{TenantAt: "tenant-1"}
	tid, err := TenantFromIdentity(id)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tid != "tenant-1" {
		t.Errorf("got %q, want %q", tid, "tenant-1")
	}
}

func TestTenantFromIdentity_NilIdentity(t *testing.T) {
	_, err := TenantFromIdentity(nil)
	if err == nil {
		t.Fatal("expected error for nil identity")
	}
}

func TestTenantFromIdentity_EmptyTenant(t *testing.T) {
	id := &Identity{UserID: "user-1"}
	_, err := TenantFromIdentity(id)
	if err == nil {
		t.Fatal("expected error for empty tenant")
	}
}

func TestTenantFromIdentity_NoIdentity(t *testing.T) {
	_, err := TenantFromIdentity(rocco.NoIdentity{})
	if err == nil {
		t.Fatal("expected error for NoIdentity")
	}
}

func TestRequireTenantAccess_Authorized(t *testing.T) {
	id := &Identity{
		Tenants: []AuthorizedTenant{
			{TenantID: "t1"},
			{TenantID: "t2"},
		},
	}
	if err := RequireTenantAccess(id, "t2"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRequireTenantAccess_Unauthorized(t *testing.T) {
	id := &Identity{
		Tenants: []AuthorizedTenant{{TenantID: "t1"}},
	}
	if err := RequireTenantAccess(id, "t2"); err == nil {
		t.Error("expected error for unauthorized tenant")
	}
}

func TestRequireTenantAccess_NilIdentity(t *testing.T) {
	if err := RequireTenantAccess(nil, "t1"); err == nil {
		t.Error("expected error for nil identity")
	}
}

func TestRequireTenantAccess_WrongType(t *testing.T) {
	if err := RequireTenantAccess(rocco.NoIdentity{}, "t1"); err == nil {
		t.Error("expected error for wrong identity type")
	}
}
