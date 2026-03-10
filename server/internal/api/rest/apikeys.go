package rest

import (
	"encoding/json"
	"net/http"

	"github.com/fleetml/fleetml/server/internal/auth"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// APIKeyHandler handles API key management endpoints.
type APIKeyHandler struct {
	db     *pgxpool.Pool
	logger *zap.SugaredLogger
}

func NewAPIKeyHandler(db *pgxpool.Pool, logger *zap.SugaredLogger) *APIKeyHandler {
	return &APIKeyHandler{db: db, logger: logger}
}

// Get returns the API key for the current user's organization.
func (h *APIKeyHandler) Get(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())
	if claims == nil || claims.OrgID == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var apiKey *string
	err := h.db.QueryRow(r.Context(),
		`SELECT api_key FROM organizations WHERE id = $1`, claims.OrgID,
	).Scan(&apiKey)
	if err != nil {
		h.logger.Errorw("failed to get API key", "org_id", claims.OrgID, "error", err)
		http.Error(w, `{"error":"failed to get API key"}`, http.StatusInternalServerError)
		return
	}

	key := ""
	if apiKey != nil {
		key = *apiKey
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"api_key": key,
	})
}

// Regenerate creates a new API key, invalidating the old one.
func (h *APIKeyHandler) Regenerate(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())
	if claims == nil || claims.OrgID == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	if claims.Role != "admin" {
		http.Error(w, `{"error":"only admins can regenerate API keys"}`, http.StatusForbidden)
		return
	}

	newKey := generateAPIKey()

	_, err := h.db.Exec(r.Context(),
		`UPDATE organizations SET api_key = $1, updated_at = NOW() WHERE id = $2`,
		newKey, claims.OrgID,
	)
	if err != nil {
		h.logger.Errorw("failed to regenerate API key", "org_id", claims.OrgID, "error", err)
		http.Error(w, `{"error":"failed to regenerate API key"}`, http.StatusInternalServerError)
		return
	}

	h.logger.Infow("API key regenerated", "org_id", claims.OrgID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"api_key": newKey,
	})
}
