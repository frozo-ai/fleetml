package rest

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	"github.com/fleetml/fleetml/server/internal/auth"
	"github.com/fleetml/fleetml/server/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// generateAPIKey creates a random API key with flml_ prefix.
func generateAPIKey() string {
	b := make([]byte, 24)
	rand.Read(b)
	return "flml_" + hex.EncodeToString(b)
}

// AuthHandler handles authentication endpoints.
type AuthHandler struct {
	jwt    *auth.JWTService
	db     *pgxpool.Pool
	logger *zap.SugaredLogger
}

func NewAuthHandler(jwt *auth.JWTService, db *pgxpool.Pool, logger *zap.SugaredLogger) *AuthHandler {
	return &AuthHandler{jwt: jwt, db: db, logger: logger}
}

var slugRegex = regexp.MustCompile(`[^a-z0-9]+`)

func toSlug(name string) string {
	slug := strings.ToLower(strings.TrimSpace(name))
	slug = slugRegex.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	if slug == "" {
		slug = "org"
	}
	return slug
}

// Register handles user registration and creates an organization.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email        string `json:"email"`
		Password     string `json:"password"`
		Name         string `json:"name"`
		Organization string `json:"organization"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Errorw("register: failed to decode body", "error", err, "content_type", r.Header.Get("Content-Type"), "content_length", r.ContentLength)
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		http.Error(w, `{"error":"email and password are required"}`, http.StatusBadRequest)
		return
	}

	if len(req.Password) < 8 {
		http.Error(w, `{"error":"password must be at least 8 characters"}`, http.StatusBadRequest)
		return
	}

	// Check if user already exists
	var existingID string
	err := h.db.QueryRow(r.Context(), `SELECT id FROM users WHERE email = $1`, req.Email).Scan(&existingID)
	if err == nil {
		http.Error(w, `{"error":"user with this email already exists"}`, http.StatusConflict)
		return
	}

	// Hash password with bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		h.logger.Errorw("failed to hash password", "error", err)
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	// Create organization
	orgName := req.Organization
	if orgName == "" {
		orgName = strings.Split(req.Email, "@")[0] + "'s Org"
	}
	slug := toSlug(orgName)

	// Ensure slug uniqueness by appending a suffix if needed
	baseSlug := slug
	for i := 1; ; i++ {
		var existingOrgID string
		err = h.db.QueryRow(r.Context(), `SELECT id FROM organizations WHERE slug = $1`, slug).Scan(&existingOrgID)
		if err == pgx.ErrNoRows {
			break
		}
		slug = baseSlug + "-" + strings.Repeat("x", i)
		if i > 10 {
			slug = baseSlug + "-" + existingOrgID[:8]
			break
		}
	}

	deviceLimit, fleetLimit, logRetention, features := domain.PlanLimits("free")
	featuresJSON, _ := json.Marshal(features)

	apiKey := generateAPIKey()

	var orgID string
	err = h.db.QueryRow(r.Context(), `
		INSERT INTO organizations (name, slug, plan, device_limit, fleet_limit, log_retention_days, features, api_key)
		VALUES ($1, $2, 'free', $3, $4, $5, $6, $7)
		RETURNING id`,
		orgName, slug, deviceLimit, fleetLimit, logRetention, string(featuresJSON), apiKey,
	).Scan(&orgID)
	if err != nil {
		h.logger.Errorw("failed to create organization", "error", err)
		http.Error(w, `{"error":"failed to create organization"}`, http.StatusInternalServerError)
		return
	}

	// First user in org is admin
	role := "admin"

	// Create user in database
	var userID string
	err = h.db.QueryRow(r.Context(), `
		INSERT INTO users (email, password_hash, name, role, org_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`,
		req.Email, string(hashedPassword), req.Name, role, orgID,
	).Scan(&userID)
	if err != nil {
		h.logger.Errorw("failed to create user", "error", err)
		http.Error(w, `{"error":"failed to create user"}`, http.StatusInternalServerError)
		return
	}

	// Create free subscription record
	h.db.Exec(r.Context(), `
		INSERT INTO subscriptions (org_id, plan, status)
		VALUES ($1, 'free', 'active')`, orgID)

	h.logger.Infow("user registered", "email", req.Email, "role", role, "org_id", orgID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "user created",
		"user": map[string]string{
			"id":    userID,
			"email": req.Email,
			"name":  req.Name,
			"role":  role,
		},
		"organization": map[string]interface{}{
			"id":   orgID,
			"name": orgName,
			"slug": slug,
			"plan": "free",
		},
	})
}

// Login handles user login.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		http.Error(w, `{"error":"email and password are required"}`, http.StatusBadRequest)
		return
	}

	// Look up user in database (with org info)
	var userID, name, role, passwordHash string
	var orgID *string
	err := h.db.QueryRow(r.Context(), `
		SELECT id, name, role, password_hash, org_id FROM users WHERE email = $1`,
		req.Email,
	).Scan(&userID, &name, &role, &passwordHash, &orgID)
	if err != nil {
		if err == pgx.ErrNoRows {
			http.Error(w, `{"error":"invalid email or password"}`, http.StatusUnauthorized)
			return
		}
		h.logger.Errorw("failed to query user", "error", err)
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		http.Error(w, `{"error":"invalid email or password"}`, http.StatusUnauthorized)
		return
	}

	// Generate JWT token with org_id
	orgIDStr := ""
	if orgID != nil {
		orgIDStr = *orgID
	}
	token, expiresAt, err := h.jwt.GenerateToken(userID, req.Email, role, orgIDStr)
	if err != nil {
		http.Error(w, `{"error":"failed to generate token"}`, http.StatusInternalServerError)
		return
	}

	// Get org info if available
	resp := map[string]interface{}{
		"token":      token,
		"expires_at": expiresAt,
		"user": map[string]interface{}{
			"id":     userID,
			"email":  req.Email,
			"name":   name,
			"role":   role,
			"org_id": orgIDStr,
		},
	}

	if orgIDStr != "" {
		var orgName, orgSlug, orgPlan string
		var deviceLimit, fleetLimit int
		err = h.db.QueryRow(r.Context(), `
			SELECT name, slug, plan, device_limit, fleet_limit FROM organizations WHERE id = $1`,
			orgIDStr,
		).Scan(&orgName, &orgSlug, &orgPlan, &deviceLimit, &fleetLimit)
		if err == nil {
			resp["organization"] = map[string]interface{}{
				"id":           orgIDStr,
				"name":         orgName,
				"slug":         orgSlug,
				"plan":         orgPlan,
				"device_limit": deviceLimit,
				"fleet_limit":  fleetLimit,
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Me returns the current authenticated user.
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())
	if claims == nil {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	resp := map[string]interface{}{
		"id":     claims.UserID,
		"email":  claims.Email,
		"role":   claims.Role,
		"org_id": claims.OrgID,
	}

	// Include org details
	if claims.OrgID != "" {
		var orgName, orgSlug, orgPlan string
		var deviceLimit, fleetLimit int
		err := h.db.QueryRow(r.Context(), `
			SELECT name, slug, plan, device_limit, fleet_limit FROM organizations WHERE id = $1`,
			claims.OrgID,
		).Scan(&orgName, &orgSlug, &orgPlan, &deviceLimit, &fleetLimit)
		if err == nil {
			resp["organization"] = map[string]interface{}{
				"id":           claims.OrgID,
				"name":         orgName,
				"slug":         orgSlug,
				"plan":         orgPlan,
				"device_limit": deviceLimit,
				"fleet_limit":  fleetLimit,
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
