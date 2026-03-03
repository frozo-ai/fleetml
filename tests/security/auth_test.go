package security

import (
	"testing"
	"time"

	"github.com/fleetml/fleetml/server/internal/auth"
)

func TestJWT_GenerateAndValidate(t *testing.T) {
	svc := auth.NewJWTService("test-secret-key-32bytes!!", 1*time.Hour)

	token, expiresAt, err := svc.GenerateToken("user-1", "test@example.com", "admin")
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	if token == "" {
		t.Fatal("expected non-empty token")
	}
	if expiresAt.Before(time.Now()) {
		t.Error("token already expired")
	}

	claims, err := svc.ValidateToken(token)
	if err != nil {
		t.Fatalf("validate token: %v", err)
	}

	if claims.UserID != "user-1" {
		t.Errorf("expected user_id 'user-1', got %q", claims.UserID)
	}
	if claims.Email != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got %q", claims.Email)
	}
	if claims.Role != "admin" {
		t.Errorf("expected role 'admin', got %q", claims.Role)
	}
}

func TestJWT_ExpiredToken(t *testing.T) {
	svc := auth.NewJWTService("test-secret", -1*time.Hour) // Already expired

	token, _, err := svc.GenerateToken("user-1", "test@example.com", "admin")
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	_, err = svc.ValidateToken(token)
	if err == nil {
		t.Error("expected error for expired token")
	}
}

func TestJWT_WrongSecret(t *testing.T) {
	svc1 := auth.NewJWTService("secret-one", 1*time.Hour)
	svc2 := auth.NewJWTService("secret-two", 1*time.Hour)

	token, _, _ := svc1.GenerateToken("user-1", "test@example.com", "admin")

	_, err := svc2.ValidateToken(token)
	if err == nil {
		t.Error("expected error for token signed with different secret")
	}
}

func TestJWT_TokenPair(t *testing.T) {
	svc := auth.NewJWTService("test-secret-key-32bytes!!", 1*time.Hour)

	pair, err := svc.GenerateTokenPair("user-1", "test@example.com", "deployer")
	if err != nil {
		t.Fatalf("generate token pair: %v", err)
	}

	if pair.AccessToken == "" {
		t.Error("expected non-empty access token")
	}
	if pair.RefreshToken == "" {
		t.Error("expected non-empty refresh token")
	}
	if pair.TokenType != "Bearer" {
		t.Errorf("expected token type 'Bearer', got %q", pair.TokenType)
	}
	if pair.AccessToken == pair.RefreshToken {
		t.Error("access and refresh tokens should be different")
	}
}

func TestJWT_RefreshToken(t *testing.T) {
	svc := auth.NewJWTService("test-secret-key-32bytes!!", 1*time.Hour)

	pair, _ := svc.GenerateTokenPair("user-1", "test@example.com", "admin")

	// Refresh should return new tokens
	newPair, err := svc.RefreshAccessToken(pair.RefreshToken)
	if err != nil {
		t.Fatalf("refresh token: %v", err)
	}

	if newPair.AccessToken == pair.AccessToken {
		t.Error("refreshed access token should be different")
	}

	// Old access token should still be valid (not revoked)
	claims, err := svc.ValidateToken(pair.AccessToken)
	if err != nil {
		t.Fatalf("old access token should still be valid: %v", err)
	}
	if claims.UserID != "user-1" {
		t.Error("old token should have correct claims")
	}
}

func TestJWT_RefreshWithAccessToken_Rejected(t *testing.T) {
	svc := auth.NewJWTService("test-secret-key-32bytes!!", 1*time.Hour)

	pair, _ := svc.GenerateTokenPair("user-1", "test@example.com", "admin")

	// Using access token as refresh should fail
	_, err := svc.RefreshAccessToken(pair.AccessToken)
	if err == nil {
		t.Error("using access token as refresh token should fail")
	}
}

func TestJWT_RevokeToken(t *testing.T) {
	svc := auth.NewJWTService("test-secret-key-32bytes!!", 1*time.Hour)

	token, _, _ := svc.GenerateToken("user-1", "test@example.com", "admin")

	if svc.IsRevoked(token) {
		t.Error("token should not be revoked initially")
	}

	svc.RevokeToken(token)

	if !svc.IsRevoked(token) {
		t.Error("token should be revoked after RevokeToken")
	}
}

func TestRBAC_AdminHasAllPermissions(t *testing.T) {
	allPerms := []string{
		"models:read", "models:write", "models:delete",
		"devices:read", "devices:write", "devices:delete",
		"fleets:read", "fleets:write", "fleets:delete",
		"deployments:read", "deployments:write", "deployments:cancel",
		"users:read", "users:write",
		"settings:read", "settings:write",
	}

	for _, perm := range allPerms {
		if !auth.HasPermission(auth.RoleAdmin, perm) {
			t.Errorf("admin should have permission %q", perm)
		}
	}
}

func TestRBAC_ViewerReadOnly(t *testing.T) {
	readPerms := []string{"models:read", "devices:read", "fleets:read", "deployments:read"}
	writePerms := []string{"models:write", "devices:write", "deployments:write", "users:write", "settings:write"}

	for _, perm := range readPerms {
		if !auth.HasPermission(auth.RoleViewer, perm) {
			t.Errorf("viewer should have permission %q", perm)
		}
	}

	for _, perm := range writePerms {
		if auth.HasPermission(auth.RoleViewer, perm) {
			t.Errorf("viewer should NOT have permission %q", perm)
		}
	}
}

func TestRBAC_DeployerCanDeploy(t *testing.T) {
	if !auth.HasPermission(auth.RoleDeployer, "deployments:write") {
		t.Error("deployer should be able to create deployments")
	}
	if !auth.HasPermission(auth.RoleDeployer, "models:write") {
		t.Error("deployer should be able to upload models")
	}
	if auth.HasPermission(auth.RoleDeployer, "users:write") {
		t.Error("deployer should NOT be able to manage users")
	}
}

func TestRBAC_UnknownRoleHasNoPermissions(t *testing.T) {
	if auth.HasPermission("hacker", "models:read") {
		t.Error("unknown role should have no permissions")
	}
}
