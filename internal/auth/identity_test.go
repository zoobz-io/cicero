//go:build testing

package auth

import (
	"testing"

	"github.com/zoobz-io/rocco"
)

var _ rocco.Identity = (*Identity)(nil)

func TestIdentity_ID(t *testing.T) {
	id := &Identity{UserID: "user-1"}
	if id.ID() != "user-1" {
		t.Errorf("ID: got %q, want %q", id.ID(), "user-1")
	}
}

func TestIdentity_TenantID(t *testing.T) {
	id := &Identity{TenantAt: "tenant-1"}
	if id.TenantID() != "tenant-1" {
		t.Errorf("TenantID: got %q, want %q", id.TenantID(), "tenant-1")
	}
}

func TestIdentity_TenantID_Empty(t *testing.T) {
	id := &Identity{}
	if id.TenantID() != "" {
		t.Errorf("TenantID: got %q, want empty", id.TenantID())
	}
}

func TestIdentity_Roles_MatchesTenant(t *testing.T) {
	id := &Identity{
		TenantAt: "t1",
		Tenants: []AuthorizedTenant{
			{TenantID: "t1", Role: "admin"},
			{TenantID: "t2", Role: "viewer"},
		},
	}
	roles := id.Roles()
	if len(roles) != 1 || roles[0] != "admin" {
		t.Errorf("Roles: got %v, want [admin]", roles)
	}
}

func TestIdentity_Roles_NoMatch(t *testing.T) {
	id := &Identity{
		TenantAt: "t3",
		Tenants:  []AuthorizedTenant{{TenantID: "t1", Role: "admin"}},
	}
	if len(id.Roles()) != 0 {
		t.Errorf("Roles: got %v, want empty", id.Roles())
	}
}

func TestIdentity_HasRole(t *testing.T) {
	id := &Identity{
		TenantAt: "t1",
		Tenants:  []AuthorizedTenant{{TenantID: "t1", Role: "editor"}},
	}
	if !id.HasRole("editor") {
		t.Error("HasRole(editor) should be true")
	}
	if id.HasRole("admin") {
		t.Error("HasRole(admin) should be false")
	}
}

func TestIdentity_HasScope_AlwaysFalse(t *testing.T) {
	id := &Identity{}
	if id.HasScope("anything") {
		t.Error("HasScope should always return false")
	}
}

func TestIdentity_Stats_Nil(t *testing.T) {
	id := &Identity{}
	if id.Stats() != nil {
		t.Error("Stats should return nil")
	}
}
