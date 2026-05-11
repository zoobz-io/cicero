// Package auth provides authentication and authorization for Cicero via the aegis mesh.
package auth

import "github.com/zoobz-io/rocco"

// Compile-time assertion.
var _ rocco.Identity = (*Identity)(nil)

// AuthorizedTenant represents a tenant the user is authorized to access for this application.
type AuthorizedTenant struct {
	TenantID   string
	TenantName string
	Role       string
}

// Identity represents an authenticated user with their authorized tenants.
type Identity struct {
	UserID   string
	TenantAt string // Resolved tenant ID for this request.
	Tenants  []AuthorizedTenant
}

// ID implements rocco.Identity.
func (id *Identity) ID() string {
	return id.UserID
}

// TenantID implements rocco.Identity.
func (id *Identity) TenantID() string {
	return id.TenantAt
}

// Email implements rocco.Identity.
func (id *Identity) Email() string {
	return ""
}

// Scopes implements rocco.Identity.
func (id *Identity) Scopes() []string {
	return nil
}

// Roles implements rocco.Identity.
func (id *Identity) Roles() []string {
	for _, t := range id.Tenants {
		if t.TenantID == id.TenantAt {
			return []string{t.Role}
		}
	}
	return nil
}

// HasScope implements rocco.Identity.
func (id *Identity) HasScope(_ string) bool {
	return false
}

// HasRole implements rocco.Identity.
func (id *Identity) HasRole(role string) bool {
	for _, r := range id.Roles() {
		if r == role {
			return true
		}
	}
	return false
}

// Stats implements rocco.Identity.
func (id *Identity) Stats() map[string]int {
	return nil
}
