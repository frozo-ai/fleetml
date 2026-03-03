package rest

import (
	"encoding/json"
	"net/http"

	"github.com/fleetml/fleetml/server/internal/auth"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// AuthHandler handles authentication endpoints.
type AuthHandler struct {
	jwt    *auth.JWTService
	db     *pgxpool.Pool
	logger *zap.SugaredLogger
}

func NewAuthHandler(jwt *auth.JWTService, db *pgxpool.Pool, logger *zap.SugaredLogger) *AuthHandler {
	return &AuthHandler{jwt: jwt, db: db, logger: logger}
}

// Register handles user registration.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
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

	// Determine role: first user is admin, rest are viewer
	var userCount int
	h.db.QueryRow(r.Context(), `SELECT COUNT(*) FROM users`).Scan(&userCount)
	role := "viewer"
	if userCount == 0 {
		role = "admin"
	}

	// Create user in database
	var userID string
	err = h.db.QueryRow(r.Context(), `
		INSERT INTO users (email, password_hash, name, role)
		VALUES ($1, $2, $3, $4)
		RETURNING id`,
		req.Email, string(hashedPassword), req.Name, role,
	).Scan(&userID)
	if err != nil {
		h.logger.Errorw("failed to create user", "error", err)
		http.Error(w, `{"error":"failed to create user"}`, http.StatusInternalServerError)
		return
	}

	h.logger.Infow("user registered", "email", req.Email, "role", role)

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

	// Look up user in database
	var userID, name, role, passwordHash string
	err := h.db.QueryRow(r.Context(), `
		SELECT id, name, role, password_hash FROM users WHERE email = $1`,
		req.Email,
	).Scan(&userID, &name, &role, &passwordHash)
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

	// Generate JWT token
	token, expiresAt, err := h.jwt.GenerateToken(userID, req.Email, role)
	if err != nil {
		http.Error(w, `{"error":"failed to generate token"}`, http.StatusInternalServerError)
		return
	}

	resp := map[string]interface{}{
		"token":      token,
		"expires_at": expiresAt,
		"user": map[string]string{
			"id":    userID,
			"email": req.Email,
			"name":  name,
			"role":  role,
		},
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

	resp := map[string]string{
		"id":    claims.UserID,
		"email": claims.Email,
		"role":  claims.Role,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
