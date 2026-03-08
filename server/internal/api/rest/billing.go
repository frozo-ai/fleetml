package rest

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/fleetml/fleetml/server/internal/auth"
	"github.com/fleetml/fleetml/server/internal/billing"
	"go.uber.org/zap"
)

// BillingHandler handles billing and subscription endpoints.
type BillingHandler struct {
	client *billing.Client
	logger *zap.SugaredLogger
}

func NewBillingHandler(client *billing.Client, logger *zap.SugaredLogger) *BillingHandler {
	return &BillingHandler{client: client, logger: logger}
}

// CreateCheckout creates a Dodo Payments checkout session for plan upgrade.
func (h *BillingHandler) CreateCheckout(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())
	if claims == nil || claims.OrgID == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req struct {
		Plan string `json:"plan"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Plan != "starter" && req.Plan != "pro" {
		http.Error(w, `{"error":"plan must be starter or pro"}`, http.StatusBadRequest)
		return
	}

	session, err := h.client.CreateCheckoutSession(r.Context(), claims.OrgID, req.Plan, claims.Email, claims.Email)
	if err != nil {
		h.logger.Errorw("failed to create checkout session", "error", err, "org_id", claims.OrgID)
		http.Error(w, `{"error":"failed to create checkout session"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"checkout_url": session.CheckoutURL,
		"session_id":   session.SessionID,
	})
}

// GetSubscription returns the current subscription for the authenticated org.
func (h *BillingHandler) GetSubscription(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())
	if claims == nil || claims.OrgID == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	sub, err := h.client.GetOrgSubscription(r.Context(), claims.OrgID)
	if err != nil {
		// No subscription found — return free plan info
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"plan":   "free",
			"status": "active",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sub)
}

// Webhook handles Dodo Payments webhook events.
func (h *BillingHandler) Webhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, `{"error":"failed to read body"}`, http.StatusBadRequest)
		return
	}

	// Parse the webhook event
	var event struct {
		Type string          `json:"type"`
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(body, &event); err != nil {
		http.Error(w, `{"error":"invalid webhook payload"}`, http.StatusBadRequest)
		return
	}

	h.logger.Infow("received webhook", "type", event.Type)

	if err := h.client.HandleSubscriptionWebhook(r.Context(), event.Type, event.Data); err != nil {
		h.logger.Errorw("webhook processing failed", "type", event.Type, "error", err)
		http.Error(w, `{"error":"webhook processing failed"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"received":true}`))
}
