package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// ---------------------------------------------------------------------------
// HasPermission
// ---------------------------------------------------------------------------

func TestHasPermission_AdminHasAllPermissions(t *testing.T) {
	allPerms := []string{
		"models:read", "models:write", "models:delete",
		"devices:read", "devices:write", "devices:delete",
		"fleets:read", "fleets:write", "fleets:delete",
		"deployments:read", "deployments:write", "deployments:cancel",
		"users:read", "users:write",
		"settings:read", "settings:write",
	}
	for _, perm := range allPerms {
		if !HasPermission(RoleAdmin, perm) {
			t.Errorf("admin should have permission %q", perm)
		}
	}
}

func TestHasPermission_DeployerCanReadWriteModels(t *testing.T) {
	if !HasPermission(RoleDeployer, "models:read") {
		t.Error("deployer should have models:read")
	}
	if !HasPermission(RoleDeployer, "models:write") {
		t.Error("deployer should have models:write")
	}
}

func TestHasPermission_DeployerCannotDeleteUsers(t *testing.T) {
	if HasPermission(RoleDeployer, "users:write") {
		t.Error("deployer should not have users:write")
	}
	if HasPermission(RoleDeployer, "models:delete") {
		t.Error("deployer should not have models:delete")
	}
	if HasPermission(RoleDeployer, "settings:write") {
		t.Error("deployer should not have settings:write")
	}
}

func TestHasPermission_DeployerCanDeployAndCancel(t *testing.T) {
	if !HasPermission(RoleDeployer, "deployments:write") {
		t.Error("deployer should have deployments:write")
	}
	if !HasPermission(RoleDeployer, "deployments:cancel") {
		t.Error("deployer should have deployments:cancel")
	}
}

func TestHasPermission_ViewerReadOnly(t *testing.T) {
	readPerms := []string{"models:read", "devices:read", "fleets:read", "deployments:read"}
	for _, perm := range readPerms {
		if !HasPermission(RoleViewer, perm) {
			t.Errorf("viewer should have permission %q", perm)
		}
	}

	writePerms := []string{
		"models:write", "models:delete",
		"devices:write", "devices:delete",
		"fleets:write", "fleets:delete",
		"deployments:write", "deployments:cancel",
		"users:read", "users:write",
		"settings:read", "settings:write",
	}
	for _, perm := range writePerms {
		if HasPermission(RoleViewer, perm) {
			t.Errorf("viewer should NOT have permission %q", perm)
		}
	}
}

func TestHasPermission_UnknownRoleHasNoPermissions(t *testing.T) {
	perms := []string{
		"models:read", "models:write", "devices:read",
		"deployments:write", "users:read", "settings:write",
	}
	for _, perm := range perms {
		if HasPermission("unknown-role", perm) {
			t.Errorf("unknown role should not have permission %q", perm)
		}
	}
}

func TestHasPermission_EmptyRole(t *testing.T) {
	if HasPermission("", "models:read") {
		t.Error("empty role should not have any permissions")
	}
}

func TestHasPermission_NonexistentPermission(t *testing.T) {
	if HasPermission(RoleAdmin, "nonexistent:action") {
		t.Error("admin should not have a nonexistent permission")
	}
}

// ---------------------------------------------------------------------------
// RequirePermission middleware
// ---------------------------------------------------------------------------

func rbacDummyHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
}

func TestRequirePermission_AdminPasses(t *testing.T) {
	claims := &Claims{UserID: "admin-1", Role: RoleAdmin}
	ctx := context.WithValue(context.Background(), UserContextKey, claims)

	req := httptest.NewRequest(http.MethodPost, "/models", nil).WithContext(ctx)
	rr := httptest.NewRecorder()

	middleware := RequirePermission("models:write")
	middleware(rbacDummyHandler()).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 for admin, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestRequirePermission_AdminCanDelete(t *testing.T) {
	claims := &Claims{UserID: "admin-2", Role: RoleAdmin}
	ctx := context.WithValue(context.Background(), UserContextKey, claims)

	req := httptest.NewRequest(http.MethodDelete, "/models/1", nil).WithContext(ctx)
	rr := httptest.NewRecorder()

	middleware := RequirePermission("models:delete")
	middleware(rbacDummyHandler()).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 for admin delete, got %d", rr.Code)
	}
}

func TestRequirePermission_ViewerBlockedOnWrite(t *testing.T) {
	claims := &Claims{UserID: "viewer-1", Role: RoleViewer}
	ctx := context.WithValue(context.Background(), UserContextKey, claims)

	req := httptest.NewRequest(http.MethodPost, "/models", nil).WithContext(ctx)
	rr := httptest.NewRecorder()

	middleware := RequirePermission("models:write")
	middleware(rbacDummyHandler()).ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for viewer on write, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestRequirePermission_ViewerAllowedRead(t *testing.T) {
	claims := &Claims{UserID: "viewer-2", Role: RoleViewer}
	ctx := context.WithValue(context.Background(), UserContextKey, claims)

	req := httptest.NewRequest(http.MethodGet, "/models", nil).WithContext(ctx)
	rr := httptest.NewRecorder()

	middleware := RequirePermission("models:read")
	middleware(rbacDummyHandler()).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 for viewer on read, got %d", rr.Code)
	}
}

func TestRequirePermission_MissingClaimsContext(t *testing.T) {
	// No claims in context at all.
	req := httptest.NewRequest(http.MethodGet, "/models", nil)
	rr := httptest.NewRecorder()

	middleware := RequirePermission("models:read")
	middleware(rbacDummyHandler()).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for missing claims, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestRequirePermission_DeployerBlockedOnSettings(t *testing.T) {
	claims := &Claims{UserID: "deployer-1", Role: RoleDeployer}
	ctx := context.WithValue(context.Background(), UserContextKey, claims)

	req := httptest.NewRequest(http.MethodPut, "/settings", nil).WithContext(ctx)
	rr := httptest.NewRecorder()

	middleware := RequirePermission("settings:write")
	middleware(rbacDummyHandler()).ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for deployer on settings:write, got %d", rr.Code)
	}
}

func TestRequirePermission_DeployerAllowedDeployments(t *testing.T) {
	claims := &Claims{UserID: "deployer-2", Role: RoleDeployer}
	ctx := context.WithValue(context.Background(), UserContextKey, claims)

	req := httptest.NewRequest(http.MethodPost, "/deployments", nil).WithContext(ctx)
	rr := httptest.NewRecorder()

	middleware := RequirePermission("deployments:write")
	middleware(rbacDummyHandler()).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 for deployer on deployments:write, got %d", rr.Code)
	}
}

func TestRequirePermission_NilClaimsValue(t *testing.T) {
	// Set the context key but with a nil value.
	ctx := context.WithValue(context.Background(), UserContextKey, (*Claims)(nil))
	req := httptest.NewRequest(http.MethodGet, "/models", nil).WithContext(ctx)
	rr := httptest.NewRecorder()

	middleware := RequirePermission("models:read")
	middleware(rbacDummyHandler()).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for nil claims value, got %d", rr.Code)
	}
}
