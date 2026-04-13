package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const UserContextKey contextKey = "user"

// Claims represents JWT claims.
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	OrgID  string `json:"org_id,omitempty"`
	jwt.RegisteredClaims
}

// TokenPair contains an access token and refresh token.
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	TokenType    string    `json:"token_type"`
}

// JWTService handles JWT token creation and validation.
type JWTService struct {
	secret        []byte
	expiry        time.Duration
	refreshExpiry time.Duration
	revokedTokens map[string]time.Time // token ID -> revoked at
	mu            sync.RWMutex
}

func NewJWTService(secret string, expiry time.Duration) *JWTService {
	return &JWTService{
		secret:        []byte(secret),
		expiry:        expiry,
		refreshExpiry: 7 * 24 * time.Hour, // 7 days
		revokedTokens: make(map[string]time.Time),
	}
}

// GenerateToken creates a new JWT token.
func (j *JWTService) GenerateToken(userID, email, role string, orgID ...string) (string, time.Time, error) {
	expiresAt := time.Now().Add(j.expiry)

	claims := &Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "fleetml",
		},
	}
	if len(orgID) > 0 && orgID[0] != "" {
		claims.OrgID = orgID[0]
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(j.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign token: %w", err)
	}

	return tokenStr, expiresAt, nil
}

// ValidateToken validates a JWT token and returns claims.
func (j *JWTService) ValidateToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

// AuthMiddleware validates JWT tokens or API keys on incoming requests.
func (j *JWTService) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		// Try API key from X-API-Key header
		if authHeader == "" {
			apiKey := r.Header.Get("X-API-Key")
			if apiKey == "" {
				http.Error(w, `{"error":"authorization header required"}`, http.StatusUnauthorized)
				return
			}
			// API key validation would be done via database lookup
			// For now, treat as unauthorized
			http.Error(w, `{"error":"invalid API key"}`, http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, `{"error":"invalid authorization header format"}`, http.StatusUnauthorized)
			return
		}

		tokenStr := parts[1]

		// Check revocation
		if j.IsRevoked(tokenStr) {
			http.Error(w, `{"error":"token has been revoked"}`, http.StatusUnauthorized)
			return
		}

		claims, err := j.ValidateToken(tokenStr)
		if err != nil {
			http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GenerateTokenPair creates both access and refresh tokens.
func (j *JWTService) GenerateTokenPair(userID, email, role string, orgID ...string) (*TokenPair, error) {
	accessToken, expiresAt, err := j.GenerateToken(userID, email, role, orgID...)
	if err != nil {
		return nil, err
	}

	// Generate refresh token with longer expiry
	refreshClaims := &Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.refreshExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "fleetml",
			Subject:   "refresh",
		},
	}
	if len(orgID) > 0 && orgID[0] != "" {
		refreshClaims.OrgID = orgID[0]
	}
	refreshTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshToken, err := refreshTokenObj.SignedString(j.secret)
	if err != nil {
		return nil, fmt.Errorf("sign refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		TokenType:    "Bearer",
	}, nil
}

// RefreshAccessToken validates a refresh token and issues a new access token.
func (j *JWTService) RefreshAccessToken(refreshToken string) (*TokenPair, error) {
	claims, err := j.ValidateToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	if claims.Subject != "refresh" {
		return nil, fmt.Errorf("token is not a refresh token")
	}

	return j.GenerateTokenPair(claims.UserID, claims.Email, claims.Role, claims.OrgID)
}

// RevokeToken marks a token as revoked.
func (j *JWTService) RevokeToken(tokenStr string) error {
	claims, err := j.ValidateToken(tokenStr)
	if err != nil {
		return err
	}
	_ = claims
	j.mu.Lock()
	j.revokedTokens[tokenStr] = time.Now()
	j.mu.Unlock()
	return nil
}

// IsRevoked checks if a token has been revoked.
func (j *JWTService) IsRevoked(tokenStr string) bool {
	j.mu.RLock()
	defer j.mu.RUnlock()
	_, revoked := j.revokedTokens[tokenStr]
	return revoked
}

// StartRevocationCleanup periodically removes expired tokens from the revocation list.
func (j *JWTService) StartRevocationCleanup(ctx context.Context) {
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			j.cleanupExpiredRevocations()
		}
	}
}

func (j *JWTService) cleanupExpiredRevocations() {
	j.mu.Lock()
	defer j.mu.Unlock()
	now := time.Now()
	for token, revokedAt := range j.revokedTokens {
		if now.Sub(revokedAt) > j.expiry+time.Hour {
			delete(j.revokedTokens, token)
		}
	}
}

// ValidateAPIKey checks if the provided API key is valid.
// API keys are stored as-is in the database (not JWT).
func ValidateAPIKey(apiKey string, lookup func(key string) (*Claims, error)) (*Claims, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("empty API key")
	}
	return lookup(apiKey)
}

// GetClaims extracts claims from the request context.
func GetClaims(ctx context.Context) *Claims {
	claims, _ := ctx.Value(UserContextKey).(*Claims)
	return claims
}
