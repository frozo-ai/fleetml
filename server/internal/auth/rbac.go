package auth

import (
	"net/http"
)

// Role constants.
const (
	RoleAdmin    = "admin"
	RoleDeployer = "deployer"
	RoleViewer   = "viewer"
)

// permissions maps roles to allowed actions.
var permissions = map[string]map[string]bool{
	RoleAdmin: {
		"models:read": true, "models:write": true, "models:delete": true,
		"devices:read": true, "devices:write": true, "devices:delete": true,
		"fleets:read": true, "fleets:write": true, "fleets:delete": true,
		"deployments:read": true, "deployments:write": true, "deployments:cancel": true,
		"users:read": true, "users:write": true,
		"settings:read": true, "settings:write": true,
	},
	RoleDeployer: {
		"models:read": true, "models:write": true,
		"devices:read": true, "devices:write": true,
		"fleets:read": true, "fleets:write": true,
		"deployments:read": true, "deployments:write": true, "deployments:cancel": true,
	},
	RoleViewer: {
		"models:read":      true,
		"devices:read":     true,
		"fleets:read":      true,
		"deployments:read": true,
	},
}

// HasPermission checks if a role has a specific permission.
func HasPermission(role, permission string) bool {
	rolePerms, ok := permissions[role]
	if !ok {
		return false
	}
	return rolePerms[permission]
}

// RequirePermission returns middleware that checks for a specific permission.
func RequirePermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := GetClaims(r.Context())
			if claims == nil {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			if !HasPermission(claims.Role, permission) {
				http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
