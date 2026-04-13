package billing

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/fleetml/fleetml/server/internal/config"
	"github.com/fleetml/fleetml/server/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// Client handles Dodo Payments API interactions.
type Client struct {
	cfg    config.BillingConfig
	db     *pgxpool.Pool
	http   *http.Client
	logger *zap.SugaredLogger
}

func NewClient(cfg config.BillingConfig, db *pgxpool.Pool, logger *zap.SugaredLogger) *Client {
	return &Client{
		cfg:    cfg,
		db:     db,
		http:   &http.Client{},
		logger: logger,
	}
}

// VerifyWebhookSignature validates the Dodo webhook HMAC-SHA256 signature.
func (c *Client) VerifyWebhookSignature(body []byte, signature string) error {
	if c.cfg.DodoWebhookKey == "" {
		c.logger.Warn("webhook key not configured, skipping signature verification")
		return nil
	}

	mac := hmac.New(sha256.New, []byte(c.cfg.DodoWebhookKey))
	mac.Write(body)
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(signature), []byte(expectedSig)) {
		return fmt.Errorf("invalid webhook signature")
	}
	return nil
}

func (c *Client) baseURL() string {
	if c.cfg.DodoEnvironment == "live" {
		return "https://live.dodopayments.com"
	}
	return "https://test.dodopayments.com"
}

// CheckoutSession represents a Dodo Payments checkout session response.
type CheckoutSession struct {
	SessionID   string `json:"session_id"`
	CheckoutURL string `json:"checkout_url"`
}

// CreateCheckoutSession creates a Dodo Payments checkout session for upgrading.
func (c *Client) CreateCheckoutSession(ctx context.Context, orgID, plan, email, customerName string) (*CheckoutSession, error) {
	productID := ""
	switch plan {
	case "starter":
		productID = c.cfg.StarterProductID
	case "pro":
		productID = c.cfg.ProProductID
	default:
		return nil, fmt.Errorf("invalid plan: %s", plan)
	}

	if productID == "" {
		return nil, fmt.Errorf("product ID not configured for plan: %s", plan)
	}

	payload := map[string]interface{}{
		"product_cart": []map[string]interface{}{
			{"product_id": productID, "quantity": 1},
		},
		"customer": map[string]string{
			"email": email,
			"name":  customerName,
		},
		"return_url": c.cfg.SuccessURL,
		"metadata": map[string]string{
			"org_id": orgID,
			"plan":   plan,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal checkout payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL()+"/checkout-sessions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.cfg.DodoAPIKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("dodo API request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("dodo API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var session CheckoutSession
	if err := json.Unmarshal(respBody, &session); err != nil {
		return nil, fmt.Errorf("parse checkout response: %w", err)
	}

	return &session, nil
}

// HandleSubscriptionWebhook processes Dodo Payments subscription webhook events.
func (c *Client) HandleSubscriptionWebhook(ctx context.Context, eventType string, data json.RawMessage) error {
	var payload struct {
		SubscriptionID string            `json:"subscription_id"`
		CustomerID     string            `json:"customer_id"`
		Status         string            `json:"status"`
		Metadata       map[string]string `json:"metadata"`
	}

	if err := json.Unmarshal(data, &payload); err != nil {
		return fmt.Errorf("parse webhook data: %w", err)
	}

	orgID := payload.Metadata["org_id"]
	plan := payload.Metadata["plan"]

	if orgID == "" {
		c.logger.Warnw("webhook missing org_id in metadata", "event", eventType, "subscription_id", payload.SubscriptionID)
		return nil
	}

	c.logger.Infow("processing subscription webhook",
		"event", eventType,
		"subscription_id", payload.SubscriptionID,
		"org_id", orgID,
		"plan", plan,
	)

	switch eventType {
	case "subscription.active", "subscription.renewed":
		return c.activateSubscription(ctx, orgID, plan, payload.SubscriptionID, payload.CustomerID)
	case "subscription.cancelled", "subscription.expired":
		return c.cancelSubscription(ctx, orgID)
	case "subscription.on_hold":
		return c.holdSubscription(ctx, orgID)
	case "subscription.plan_changed":
		return c.activateSubscription(ctx, orgID, plan, payload.SubscriptionID, payload.CustomerID)
	default:
		c.logger.Debugw("unhandled webhook event", "event", eventType)
	}

	return nil
}

func (c *Client) activateSubscription(ctx context.Context, orgID, plan, subscriptionID, customerID string) error {
	if plan == "" {
		plan = "starter"
	}

	deviceLimit, fleetLimit, logRetention, features := domain.PlanLimits(plan)
	featuresJSON, _ := json.Marshal(features)

	// Update organization plan and limits
	_, err := c.db.Exec(ctx, `
		UPDATE organizations
		SET plan = $1, device_limit = $2, fleet_limit = $3, log_retention_days = $4, features = $5, updated_at = NOW()
		WHERE id = $6`,
		plan, deviceLimit, fleetLimit, logRetention, string(featuresJSON), orgID,
	)
	if err != nil {
		return fmt.Errorf("update org plan: %w", err)
	}

	// Upsert subscription record
	_, err = c.db.Exec(ctx, `
		INSERT INTO subscriptions (org_id, dodo_subscription_id, dodo_customer_id, plan, status, current_period_start, updated_at)
		VALUES ($1, $2, $3, $4, 'active', NOW(), NOW())
		ON CONFLICT (dodo_subscription_id) DO UPDATE SET
			plan = $4, status = 'active', updated_at = NOW()`,
		orgID, subscriptionID, customerID, plan,
	)
	if err != nil {
		return fmt.Errorf("upsert subscription: %w", err)
	}

	c.logger.Infow("subscription activated", "org_id", orgID, "plan", plan)
	return nil
}

func (c *Client) cancelSubscription(ctx context.Context, orgID string) error {
	// Downgrade to free plan
	deviceLimit, fleetLimit, logRetention, features := domain.PlanLimits("free")
	featuresJSON, _ := json.Marshal(features)

	_, err := c.db.Exec(ctx, `
		UPDATE organizations
		SET plan = 'free', device_limit = $1, fleet_limit = $2, log_retention_days = $3, features = $4, updated_at = NOW()
		WHERE id = $5`,
		deviceLimit, fleetLimit, logRetention, string(featuresJSON), orgID,
	)
	if err != nil {
		return fmt.Errorf("downgrade org: %w", err)
	}

	_, err = c.db.Exec(ctx, `
		UPDATE subscriptions SET status = 'cancelled', cancelled_at = NOW(), updated_at = NOW()
		WHERE org_id = $1 AND status = 'active'`, orgID)
	if err != nil {
		return fmt.Errorf("cancel subscription record: %w", err)
	}

	c.logger.Infow("subscription cancelled, downgraded to free", "org_id", orgID)
	return nil
}

func (c *Client) holdSubscription(ctx context.Context, orgID string) error {
	_, err := c.db.Exec(ctx, `
		UPDATE subscriptions SET status = 'on_hold', updated_at = NOW()
		WHERE org_id = $1 AND status = 'active'`, orgID)
	if err != nil {
		return fmt.Errorf("hold subscription: %w", err)
	}

	c.logger.Warnw("subscription on hold", "org_id", orgID)
	return nil
}

// GetOrgSubscription returns the current subscription for an organization.
func (c *Client) GetOrgSubscription(ctx context.Context, orgID string) (*domain.Subscription, error) {
	var sub domain.Subscription
	err := c.db.QueryRow(ctx, `
		SELECT id, org_id, COALESCE(dodo_subscription_id, ''), COALESCE(dodo_customer_id, ''),
			plan, status, current_period_start, current_period_end, cancelled_at, created_at, updated_at
		FROM subscriptions
		WHERE org_id = $1
		ORDER BY created_at DESC LIMIT 1`, orgID,
	).Scan(
		&sub.ID, &sub.OrgID, &sub.DodoSubscriptionID, &sub.DodoCustomerID,
		&sub.Plan, &sub.Status, &sub.CurrentPeriodStart, &sub.CurrentPeriodEnd,
		&sub.CancelledAt, &sub.CreatedAt, &sub.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &sub, nil
}
