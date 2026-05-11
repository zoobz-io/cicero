package auth

import (
	"fmt"

	"github.com/zoobz-io/rocco"
)

// TenantFromIdentity extracts the resolved tenant ID from the authenticated identity.
func TenantFromIdentity(identity rocco.Identity) (string, error) {
	if identity == nil {
		return "", fmt.Errorf("not authenticated")
	}
	tid := identity.TenantID()
	if tid == "" {
		return "", fmt.Errorf("no tenant resolved")
	}
	return tid, nil
}

// RequireTenantAccess verifies the authenticated user has access to the specified tenant.
func RequireTenantAccess(identity rocco.Identity, tenantID string) error {
	id, ok := identity.(*Identity)
	if !ok {
		return fmt.Errorf("not authenticated")
	}

	for _, t := range id.Tenants {
		if t.TenantID == tenantID {
			return nil
		}
	}

	return fmt.Errorf("not authorized for tenant %s", tenantID)
}
