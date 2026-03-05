package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func newTestJWTService() *JWTService {
	return NewJWTService("test-secret-key-for-unit-tests", 15*time.Minute)
}

// ---------------------------------------------------------------------------
// GenerateToken
// ---------------------------------------------------------------------------

func TestGenerateToken_ValidClaims(t *testing.T) {
	svc := newTestJWTService()
	token, expiresAt, err := svc.GenerateToken("user-1", "alice@example.com", RoleAdmin)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}
	if expiresAt.IsZero() {
		t.Fatal("expected non-zero expiry")
	}
	if expiresAt.Before(time.Now()) {
		t.Fatal("expected expiry in the future")
	}
}

func TestGenerateToken_EmptyUserID(t *testing.T) {
	svc := newTestJWTService()
	token, _, err := svc.GenerateToken("", "bob@example.com", RoleViewer)
	// The implementation does not reject empty user IDs; it still produces a token.
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token even with empty user ID")
	}
	claims, err := svc.ValidateToken(token)
	if err != nil {
		t.Fatalf("expected valid token, got %v", err)
	}
	if claims.UserID != "" {
		t.Fatalf("expected empty UserID in claims, got %q", claims.UserID)
	}
}

func TestGenerateToken_EmptyRole(t *testing.T) {
	svc := newTestJWTService()
	token, _, err := svc.GenerateToken("user-2", "carol@example.com", "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	claims, err := svc.ValidateToken(token)
	if err != nil {
		t.Fatalf("expected valid token, got %v", err)
	}
	if claims.Role != "" {
		t.Fatalf("expected empty role, got %q", claims.Role)
	}
}

// ---------------------------------------------------------------------------
// ValidateToken
// ---------------------------------------------------------------------------

func TestValidateToken_ValidToken(t *testing.T) {
	svc := newTestJWTService()
	token, _, err := svc.GenerateToken("user-3", "dave@example.com", RoleDeployer)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	claims, err := svc.ValidateToken(token)
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if claims.UserID != "user-3" {
		t.Fatalf("expected UserID user-3, got %q", claims.UserID)
	}
	if claims.Email != "dave@example.com" {
		t.Fatalf("expected Email dave@example.com, got %q", claims.Email)
	}
	if claims.Role != RoleDeployer {
		t.Fatalf("expected Role deployer, got %q", claims.Role)
	}
	if claims.Issuer != "fleetml" {
		t.Fatalf("expected Issuer fleetml, got %q", claims.Issuer)
	}
}

func TestValidateToken_ExpiredToken(t *testing.T) {
	// Create a service with a negative expiry to guarantee expiration.
	svc := NewJWTService("test-secret-key-for-unit-tests", -1*time.Second)
	token, _, err := svc.GenerateToken("user-exp", "exp@example.com", RoleViewer)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	_, err = svc.ValidateToken(token)
	if err == nil {
		t.Fatal("expected error for expired token")
	}
}

func TestValidateToken_InvalidSigningMethod(t *testing.T) {
	// Create a token signed with RSA (none method simulation via unsigned).
	// Build a token with "none" alg which should be rejected.
	claims := &Claims{
		UserID: "user-bad",
		Email:  "bad@example.com",
		Role:   RoleViewer,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "fleetml",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	tokenStr, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("sign none token: %v", err)
	}

	svc := newTestJWTService()
	_, err = svc.ValidateToken(tokenStr)
	if err == nil {
		t.Fatal("expected error for none signing method")
	}
}

func TestValidateToken_MalformedToken(t *testing.T) {
	svc := newTestJWTService()
	_, err := svc.ValidateToken("this.is.not.a.valid.jwt")
	if err == nil {
		t.Fatal("expected error for malformed token")
	}
}

func TestValidateToken_EmptyString(t *testing.T) {
	svc := newTestJWTService()
	_, err := svc.ValidateToken("")
	if err == nil {
		t.Fatal("expected error for empty token string")
	}
}

// ---------------------------------------------------------------------------
// GenerateTokenPair
// ---------------------------------------------------------------------------

func TestGenerateTokenPair_Success(t *testing.T) {
	svc := newTestJWTService()
	pair, err := svc.GenerateTokenPair("user-pair", "pair@example.com", RoleAdmin)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if pair.AccessToken == "" {
		t.Fatal("expected non-empty access token")
	}
	if pair.RefreshToken == "" {
		t.Fatal("expected non-empty refresh token")
	}
	if pair.TokenType != "Bearer" {
		t.Fatalf("expected token type Bearer, got %q", pair.TokenType)
	}
	if pair.ExpiresAt.IsZero() {
		t.Fatal("expected non-zero expiry")
	}
}

func TestGenerateTokenPair_ClaimsMatch(t *testing.T) {
	svc := newTestJWTService()
	pair, err := svc.GenerateTokenPair("user-claims", "claims@example.com", RoleDeployer)
	if err != nil {
		t.Fatalf("generate pair: %v", err)
	}

	accessClaims, err := svc.ValidateToken(pair.AccessToken)
	if err != nil {
		t.Fatalf("validate access: %v", err)
	}
	if accessClaims.UserID != "user-claims" {
		t.Fatalf("access UserID: got %q", accessClaims.UserID)
	}
	if accessClaims.Role != RoleDeployer {
		t.Fatalf("access Role: got %q", accessClaims.Role)
	}

	refreshClaims, err := svc.ValidateToken(pair.RefreshToken)
	if err != nil {
		t.Fatalf("validate refresh: %v", err)
	}
	if refreshClaims.UserID != "user-claims" {
		t.Fatalf("refresh UserID: got %q", refreshClaims.UserID)
	}
	if refreshClaims.Subject != "refresh" {
		t.Fatalf("refresh Subject: expected 'refresh', got %q", refreshClaims.Subject)
	}
}

// ---------------------------------------------------------------------------
// RefreshAccessToken
// ---------------------------------------------------------------------------

func TestRefreshAccessToken_ValidRefresh(t *testing.T) {
	svc := newTestJWTService()
	pair, err := svc.GenerateTokenPair("user-refresh", "refresh@example.com", RoleViewer)
	if err != nil {
		t.Fatalf("generate pair: %v", err)
	}

	newPair, err := svc.RefreshAccessToken(pair.RefreshToken)
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if newPair.AccessToken == "" {
		t.Fatal("expected non-empty new access token")
	}
	if newPair.AccessToken == pair.AccessToken {
		t.Fatal("expected a different access token after refresh")
	}

	claims, err := svc.ValidateToken(newPair.AccessToken)
	if err != nil {
		t.Fatalf("validate new access: %v", err)
	}
	if claims.UserID != "user-refresh" {
		t.Fatalf("expected UserID user-refresh, got %q", claims.UserID)
	}
}

func TestRefreshAccessToken_ExpiredRefresh(t *testing.T) {
	// Use a service with negative expiry + negative refresh expiry to make
	// both tokens expired immediately. We craft the refresh manually.
	svc := newTestJWTService()
	refreshClaims := &Claims{
		UserID: "user-exp-refresh",
		Email:  "exp-refresh@example.com",
		Role:   RoleViewer,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			Issuer:    "fleetml",
			Subject:   "refresh",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	tokenStr, _ := token.SignedString(svc.secret)

	_, err := svc.RefreshAccessToken(tokenStr)
	if err == nil {
		t.Fatal("expected error for expired refresh token")
	}
}

func TestRefreshAccessToken_AccessTokenUsedAsRefresh(t *testing.T) {
	svc := newTestJWTService()
	pair, err := svc.GenerateTokenPair("user-wrong", "wrong@example.com", RoleAdmin)
	if err != nil {
		t.Fatalf("generate pair: %v", err)
	}

	// The access token does not have Subject = "refresh", so this should fail.
	_, err = svc.RefreshAccessToken(pair.AccessToken)
	if err == nil {
		t.Fatal("expected error when using access token as refresh token")
	}
}

// ---------------------------------------------------------------------------
// RevokeToken + IsRevoked
// ---------------------------------------------------------------------------

func TestRevokeToken_ThenIsRevoked(t *testing.T) {
	svc := newTestJWTService()
	token, _, err := svc.GenerateToken("user-revoke", "revoke@example.com", RoleAdmin)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	if svc.IsRevoked(token) {
		t.Fatal("token should not be revoked initially")
	}

	if err := svc.RevokeToken(token); err != nil {
		t.Fatalf("revoke: %v", err)
	}

	if !svc.IsRevoked(token) {
		t.Fatal("token should be revoked after RevokeToken")
	}
}

func TestIsRevoked_NonRevokedToken(t *testing.T) {
	svc := newTestJWTService()
	token, _, _ := svc.GenerateToken("user-check", "check@example.com", RoleViewer)
	if svc.IsRevoked(token) {
		t.Fatal("brand new token should not be revoked")
	}
}

func TestRevokeToken_InvalidToken(t *testing.T) {
	svc := newTestJWTService()
	err := svc.RevokeToken("invalid-token-string")
	if err == nil {
		t.Fatal("expected error when revoking an invalid token")
	}
}

// ---------------------------------------------------------------------------
// AuthMiddleware
// ---------------------------------------------------------------------------

func dummyHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
}

func TestAuthMiddleware_ValidBearerToken(t *testing.T) {
	svc := newTestJWTService()
	token, _, _ := svc.GenerateToken("user-mid", "mid@example.com", RoleAdmin)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	svc.AuthMiddleware(dummyHandler()).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestAuthMiddleware_MissingHeader(t *testing.T) {
	svc := newTestJWTService()

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	svc.AuthMiddleware(dummyHandler()).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	svc := newTestJWTService()

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rr := httptest.NewRecorder()

	svc.AuthMiddleware(dummyHandler()).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestAuthMiddleware_XAPIKeyHeader(t *testing.T) {
	svc := newTestJWTService()

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-API-Key", "some-api-key")
	rr := httptest.NewRecorder()

	svc.AuthMiddleware(dummyHandler()).ServeHTTP(rr, req)

	// Current implementation treats all API keys as unauthorized.
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	svc := newTestJWTService()

	// Create an already-expired token.
	claims := &Claims{
		UserID: "user-expired",
		Email:  "expired@example.com",
		Role:   RoleAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			Issuer:    "fleetml",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, _ := token.SignedString(svc.secret)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	rr := httptest.NewRecorder()

	svc.AuthMiddleware(dummyHandler()).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for expired token, got %d", rr.Code)
	}
}

func TestAuthMiddleware_RevokedToken(t *testing.T) {
	svc := newTestJWTService()
	token, _, _ := svc.GenerateToken("user-rev-mid", "rev-mid@example.com", RoleAdmin)
	svc.RevokeToken(token)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	svc.AuthMiddleware(dummyHandler()).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for revoked token, got %d", rr.Code)
	}
}

func TestAuthMiddleware_InvalidFormat_NoBearerPrefix(t *testing.T) {
	svc := newTestJWTService()
	token, _, _ := svc.GenerateToken("user-fmt", "fmt@example.com", RoleAdmin)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Token "+token)
	rr := httptest.NewRecorder()

	svc.AuthMiddleware(dummyHandler()).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for non-Bearer prefix, got %d", rr.Code)
	}
}

func TestAuthMiddleware_ContextContainsClaims(t *testing.T) {
	svc := newTestJWTService()
	token, _, _ := svc.GenerateToken("user-ctx", "ctx@example.com", RoleDeployer)

	var captured *Claims
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = GetClaims(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	svc.AuthMiddleware(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if captured == nil {
		t.Fatal("expected claims in context")
	}
	if captured.UserID != "user-ctx" {
		t.Fatalf("expected user-ctx, got %q", captured.UserID)
	}
	if captured.Role != RoleDeployer {
		t.Fatalf("expected deployer, got %q", captured.Role)
	}
}

// ---------------------------------------------------------------------------
// GetClaims
// ---------------------------------------------------------------------------

func TestGetClaims_ValidContext(t *testing.T) {
	claims := &Claims{UserID: "user-gc", Email: "gc@example.com", Role: RoleAdmin}
	ctx := context.WithValue(context.Background(), UserContextKey, claims)

	got := GetClaims(ctx)
	if got == nil {
		t.Fatal("expected non-nil claims")
	}
	if got.UserID != "user-gc" {
		t.Fatalf("expected user-gc, got %q", got.UserID)
	}
}

func TestGetClaims_EmptyContext(t *testing.T) {
	got := GetClaims(context.Background())
	if got != nil {
		t.Fatal("expected nil claims from empty context")
	}
}

func TestGetClaims_WrongType(t *testing.T) {
	ctx := context.WithValue(context.Background(), UserContextKey, "not-claims")
	got := GetClaims(ctx)
	if got != nil {
		t.Fatal("expected nil when context value is wrong type")
	}
}

// ---------------------------------------------------------------------------
// ValidateAPIKey
// ---------------------------------------------------------------------------

func TestValidateAPIKey_EmptyKey(t *testing.T) {
	_, err := ValidateAPIKey("", func(key string) (*Claims, error) {
		return nil, nil
	})
	if err == nil {
		t.Fatal("expected error for empty API key")
	}
}

func TestValidateAPIKey_ValidKey(t *testing.T) {
	expected := &Claims{UserID: "api-user", Role: RoleAdmin}
	claims, err := ValidateAPIKey("valid-key-123", func(key string) (*Claims, error) {
		if key == "valid-key-123" {
			return expected, nil
		}
		return nil, nil
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if claims.UserID != "api-user" {
		t.Fatalf("expected api-user, got %q", claims.UserID)
	}
}
