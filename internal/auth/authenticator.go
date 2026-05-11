package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/zoobz-io/aegis"
	sessionpb "github.com/zoobz-io/aegis/proto/session/v1"
	"github.com/zoobz-io/rocco"
)

// NewAuthenticator returns a rocco authenticator function that validates session
// tokens against Janus via the aegis mesh.
func NewAuthenticator(sessionClient *aegis.ServiceClient[sessionpb.SessionServiceClient]) func(context.Context, *http.Request) (rocco.Identity, error) {
	return func(ctx context.Context, r *http.Request) (rocco.Identity, error) {
		// Extract bearer token.
		token := extractBearerToken(r)
		if token == "" {
			return nil, fmt.Errorf("missing authorization token")
		}

		// Validate via Janus.
		client, err := sessionClient.Get(ctx)
		if err != nil {
			return nil, fmt.Errorf("session service unavailable: %w", err)
		}

		resp, err := client.ValidateSession(ctx, &sessionpb.ValidateSessionRequest{Token: token})
		if err != nil {
			return nil, fmt.Errorf("session validation failed: %w", err)
		}
		if !resp.Valid {
			return nil, fmt.Errorf("invalid or expired session")
		}

		// Map authorized tenants.
		tenants := make([]AuthorizedTenant, len(resp.Tenants))
		for i, t := range resp.Tenants {
			tenants[i] = AuthorizedTenant{
				TenantID:   t.TenantId,
				TenantName: t.TenantName,
				Role:       t.Role,
			}
		}

		// Resolve which tenant this request is for.
		tenantID, err := resolveTenantFromRequest(r, tenants)
		if err != nil {
			return nil, err
		}

		return &Identity{
			UserID:   resp.UserId,
			TenantAt: tenantID,
			Tenants:  tenants,
		}, nil
	}
}

// extractBearerToken extracts a bearer token from the Authorization header.
func extractBearerToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return ""
	}
	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return ""
	}
	return parts[1]
}

// resolveTenantFromRequest determines which tenant the request is for.
// Single-tenant users auto-select. Multi-tenant users must provide X-Tenant-ID.
func resolveTenantFromRequest(r *http.Request, tenants []AuthorizedTenant) (string, error) {
	if len(tenants) == 0 {
		return "", fmt.Errorf("user has no authorized tenants for this application")
	}

	if len(tenants) == 1 {
		return tenants[0].TenantID, nil
	}

	// Multi-tenant: require explicit selection.
	selected := r.Header.Get("X-Tenant-ID")
	if selected == "" {
		return "", fmt.Errorf("multi-tenant user must provide X-Tenant-ID header")
	}

	for _, t := range tenants {
		if t.TenantID == selected {
			return selected, nil
		}
	}

	return "", fmt.Errorf("user is not authorized for tenant %s", selected)
}
