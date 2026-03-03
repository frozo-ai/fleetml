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
	APIKey       *string   `json:"api_key,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
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
