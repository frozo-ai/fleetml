package domain

import "time"

// Device represents an edge device in the fleet.
type Device struct {
	ID                     string            `json:"id"`
	DeviceID               string            `json:"device_id"`
	Name                   string            `json:"name"`
	Status                 string            `json:"status"` // registered, healthy, warning, offline, decommissioned
	Arch                   string            `json:"arch"`
	GPUType                string            `json:"gpu_type"`
	Runtime                string            `json:"runtime"`
	RAMMB                  int               `json:"ram_mb"`
	DiskGB                 int               `json:"disk_gb"`
	OS                     string            `json:"os"`
	HardwareModel          string            `json:"hardware_model"`
	Labels                 map[string]string `json:"labels"`
	FleetID                *string           `json:"fleet_id,omitempty"`
	CertificateFingerprint string            `json:"certificate_fingerprint,omitempty"`
	LastHeartbeat          *time.Time        `json:"last_heartbeat,omitempty"`
	RegisteredAt           time.Time         `json:"registered_at"`
	UpdatedAt              time.Time         `json:"updated_at"`
	CPUPercent             *float64          `json:"cpu_percent,omitempty"`
	GPUPercent             *float64          `json:"gpu_percent,omitempty"`
	RAMMBUsed              *int              `json:"ram_mb_used,omitempty"`
	DiskPercent            *float64          `json:"disk_percent,omitempty"`
	TemperatureC           *float64          `json:"temperature_c,omitempty"`
	UptimeHours            *float64          `json:"uptime_hours,omitempty"`
}

// Fleet represents a device group.
type Fleet struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Labels      map[string]string `json:"labels"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// Model represents an ML model in the registry.
type Model struct {
	ID               string                 `json:"id"`
	Name             string                 `json:"name"`
	Version          string                 `json:"version"`
	Format           string                 `json:"format"`
	ArtifactURL      string                 `json:"artifact_url"`
	ArtifactSize     int64                  `json:"artifact_size"`
	Checksum         string                 `json:"checksum"`
	Description      string                 `json:"description"`
	Metadata         map[string]interface{} `json:"metadata"`
	Tags             []string               `json:"tags"`
	ParentModelID    *string                `json:"parent_model_id,omitempty"`
	CompiledVariants []CompiledVariant      `json:"compiled_variants"`
	CreatedAt        time.Time              `json:"created_at"`
	CreatedBy        *string                `json:"created_by,omitempty"`
}

// CompiledVariant represents a compiled model artifact.
type CompiledVariant struct {
	Runtime     string `json:"runtime"`
	ArtifactURL string `json:"artifact_url"`
	Checksum    string `json:"checksum"`
}

// Deployment represents a model deployment to devices.
type Deployment struct {
	ID               string          `json:"id"`
	ModelID          string          `json:"model_id"`
	TargetType       string          `json:"target_type"` // fleet, device, label_selector
	TargetFleetID    *string         `json:"target_fleet_id,omitempty"`
	TargetDeviceIDs  []string        `json:"target_device_ids,omitempty"`
	TargetLabels     map[string]string `json:"target_labels,omitempty"`
	State            string          `json:"state"` // pending, rolling_out, completed, failed, rolled_back, cancelled
	TotalDevices     int             `json:"total_devices"`
	CompletedDevices int             `json:"completed_devices"`
	FailedDevices    int             `json:"failed_devices"`
	QueuedDevices    int             `json:"queued_devices"`
	DeploymentPolicy string          `json:"deployment_policy"`
	CanaryConfig     *CanaryConfig   `json:"canary_config,omitempty"`
	RollbackModelID  *string         `json:"rollback_model_id,omitempty"`
	Error            string          `json:"error,omitempty"`
	StartedAt        *time.Time      `json:"started_at,omitempty"`
	CompletedAt      *time.Time      `json:"completed_at,omitempty"`
	CreatedAt        time.Time       `json:"created_at"`
	CreatedBy        *string         `json:"created_by,omitempty"`
}

// CanaryConfig defines canary deployment stages.
type CanaryConfig struct {
	Stages []CanaryStage `json:"stages"`
}

// CanaryStage defines a single canary stage.
type CanaryStage struct {
	Percent       int    `json:"percent"`
	Duration      string `json:"duration"`
	SuccessMetric string `json:"success_metric"`
}

// User represents an authenticated user.
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	Name         string    `json:"name"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"` // admin, deployer, viewer
	OrgID        *string   `json:"org_id,omitempty"`
	APIKey       *string   `json:"api_key,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ABTest represents an A/B test between two model variants.
type ABTest struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	ModelAID      string                 `json:"model_a_id"`
	ModelBID      string                 `json:"model_b_id"`
	SplitA        int                    `json:"split_a"`
	SplitB        int                    `json:"split_b"`
	TargetFleetID *string                `json:"target_fleet_id,omitempty"`
	TargetLabels  map[string]string      `json:"target_labels,omitempty"`
	Metric        string                 `json:"metric"`
	Duration      *string                `json:"duration,omitempty"`
	AutoPromote   bool                   `json:"auto_promote"`
	State         string                 `json:"state"` // pending, running, completed, stopped
	Winner        *string                `json:"winner,omitempty"`
	ModelAMetrics map[string]interface{} `json:"model_a_metrics,omitempty"`
	ModelBMetrics map[string]interface{} `json:"model_b_metrics,omitempty"`
	StartedAt     *time.Time             `json:"started_at,omitempty"`
	StoppedAt     *time.Time             `json:"stopped_at,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	CreatedBy     *string                `json:"created_by,omitempty"`
}

// Policy defines a deployment or operational policy.
type Policy struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	PolicyType    string                 `json:"policy_type"` // deployment, scaling, alerting, compliance
	Rules         map[string]interface{} `json:"rules"`
	Enabled       bool                   `json:"enabled"`
	Priority      int                    `json:"priority"`
	TargetFleetID *string                `json:"target_fleet_id,omitempty"`
	TargetLabels  map[string]string      `json:"target_labels,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
	CreatedBy     *string                `json:"created_by,omitempty"`
}

// Organization represents a tenant in the multi-tenant system.
type Organization struct {
	ID               string         `json:"id"`
	Name             string         `json:"name"`
	Slug             string         `json:"slug"`
	Plan             string         `json:"plan"` // free, starter, pro, enterprise
	DeviceLimit      int            `json:"device_limit"`
	FleetLimit       int            `json:"fleet_limit"`
	LogRetentionDays int            `json:"log_retention_days"`
	Features         map[string]bool `json:"features"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
}

// Subscription tracks a Dodo Payments subscription for an organization.
type Subscription struct {
	ID                  string     `json:"id"`
	OrgID               string     `json:"org_id"`
	DodoSubscriptionID  string     `json:"dodo_subscription_id,omitempty"`
	DodoCustomerID      string     `json:"dodo_customer_id,omitempty"`
	Plan                string     `json:"plan"`
	Status              string     `json:"status"` // active, on_hold, cancelled, expired
	CurrentPeriodStart  *time.Time `json:"current_period_start,omitempty"`
	CurrentPeriodEnd    *time.Time `json:"current_period_end,omitempty"`
	CancelledAt         *time.Time `json:"cancelled_at,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

// PlanLimits returns the limits for a given plan.
func PlanLimits(plan string) (deviceLimit, fleetLimit, logRetentionDays int, features map[string]bool) {
	switch plan {
	case "starter":
		return 25, 3, 7, map[string]bool{
			"ab_testing":      false,
			"drift_detection": false,
			"policy_engine":   false,
		}
	case "pro":
		return 100, -1, 30, map[string]bool{
			"ab_testing":      true,
			"drift_detection": true,
			"policy_engine":   true,
		}
	case "enterprise":
		return -1, -1, 90, map[string]bool{
			"ab_testing":      true,
			"drift_detection": true,
			"policy_engine":   true,
			"sso":             true,
			"audit_log":       true,
		}
	default: // free
		return 5, 1, 3, map[string]bool{
			"ab_testing":      false,
			"drift_detection": false,
			"policy_engine":   false,
		}
	}
}

// DeviceFilter provides filter criteria for listing devices.
type DeviceFilter struct {
	Status  string
	FleetID string
	Labels  map[string]string
	Runtime string
	Limit   int
	Offset  int
}

// ModelFilter provides filter criteria for listing models.
type ModelFilter struct {
	Name   string
	Tags   []string
	Limit  int
	Offset int
}
