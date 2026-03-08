package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server    ServerConfig    `yaml:"server"`
	Database  DatabaseConfig  `yaml:"database"`
	Storage   StorageConfig   `yaml:"storage"`
	Auth      AuthConfig      `yaml:"auth"`
	Heartbeat HeartbeatConfig `yaml:"heartbeat"`
	Deploy    DeployConfig    `yaml:"deployment"`
	Compiler  CompilerConfig  `yaml:"compiler"`
	NATS      NATSConfig      `yaml:"nats"`
	Tracing      TracingConfig      `yaml:"tracing"`
	Integrations IntegrationsConfig `yaml:"integrations"`
	Billing      BillingConfig      `yaml:"billing"`
	Logging      LoggingConfig      `yaml:"logging"`
}

type ServerConfig struct {
	RESTPort int       `yaml:"rest_port"`
	GRPCPort int       `yaml:"grpc_port"`
	TLS      TLSConfig `yaml:"tls"`
}

type TLSConfig struct {
	Enabled  bool   `yaml:"enabled"`
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
	CAFile   string `yaml:"ca_file"`
}

type DatabaseConfig struct {
	URL            string `yaml:"url"`
	MaxConnections int    `yaml:"max_connections"`
}

type StorageConfig struct {
	Type      string `yaml:"type"`
	Endpoint  string `yaml:"endpoint"`
	Bucket    string `yaml:"bucket"`
	AccessKey string `yaml:"access_key"`
	SecretKey string `yaml:"secret_key"`
	Region    string `yaml:"region"`
}

type AuthConfig struct {
	JWTSecret string        `yaml:"jwt_secret"`
	JWTExpiry time.Duration `yaml:"jwt_expiry"`
}

type HeartbeatConfig struct {
	OfflineThreshold time.Duration `yaml:"offline_threshold"`
}

type DeployConfig struct {
	DefaultTimeout          time.Duration `yaml:"default_timeout"`
	ConcurrentDeploysPerDev int           `yaml:"concurrent_deploys_per_device"`
}

type CompilerConfig struct {
	URL string `yaml:"url"`
}

type NATSConfig struct {
	URL string `yaml:"url"`
}

type IntegrationsConfig struct {
	MLflowURL string `yaml:"mlflow_url"`
	HFToken   string `yaml:"hf_token"`
}

type TracingConfig struct {
	Enabled    bool    `yaml:"enabled"`
	Endpoint   string  `yaml:"endpoint"`
	SampleRate float64 `yaml:"sample_rate"`
}

type BillingConfig struct {
	DodoAPIKey       string `yaml:"dodo_api_key"`
	DodoWebhookKey   string `yaml:"dodo_webhook_key"`
	DodoEnvironment  string `yaml:"dodo_environment"` // test or live
	StarterProductID string `yaml:"starter_product_id"`
	ProProductID     string `yaml:"pro_product_id"`
	SuccessURL       string `yaml:"success_url"`
	DashboardURL     string `yaml:"dashboard_url"`
}

type LoggingConfig struct {
	Level string `yaml:"level"`
}

func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			RESTPort: 8080,
			GRPCPort: 50051,
		},
		Database: DatabaseConfig{
			URL:            "postgres://fleetml:password@localhost:5432/fleetml?sslmode=disable",
			MaxConnections: 50,
		},
		Storage: StorageConfig{
			Type:      "s3",
			Endpoint:  "http://localhost:9000",
			Bucket:    "fleetml-models",
			AccessKey: "minioadmin",
			SecretKey: "minioadmin",
			Region:    "us-east-1",
		},
		Auth: AuthConfig{
			JWTSecret: "changeme",
			JWTExpiry: 24 * time.Hour,
		},
		Heartbeat: HeartbeatConfig{
			OfflineThreshold: 90 * time.Second,
		},
		Deploy: DeployConfig{
			DefaultTimeout:          10 * time.Minute,
			ConcurrentDeploysPerDev: 1,
		},
		Compiler: CompilerConfig{
			URL: "",
		},
		NATS: NATSConfig{
			URL: "",
		},
		Tracing: TracingConfig{
			Enabled:    false,
			Endpoint:   "localhost:4318",
			SampleRate: 1.0,
		},
		Billing: BillingConfig{
			DodoEnvironment: "test",
			SuccessURL:      "http://localhost:3000/dashboard/billing?success=true",
			DashboardURL:    "http://localhost:3000/dashboard",
		},
		Logging: LoggingConfig{
			Level: "info",
		},
	}
}

func Load(path string) (*Config, error) {
	cfg := DefaultConfig()

	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, err
		}
	}

	// Environment variable overrides
	if v := os.Getenv("DATABASE_URL"); v != "" {
		cfg.Database.URL = v
	}
	if v := os.Getenv("S3_ENDPOINT"); v != "" {
		cfg.Storage.Endpoint = v
	}
	if v := os.Getenv("S3_ACCESS_KEY"); v != "" {
		cfg.Storage.AccessKey = v
	}
	if v := os.Getenv("S3_SECRET_KEY"); v != "" {
		cfg.Storage.SecretKey = v
	}
	if v := os.Getenv("S3_BUCKET"); v != "" {
		cfg.Storage.Bucket = v
	}
	if v := os.Getenv("JWT_SECRET"); v != "" {
		cfg.Auth.JWTSecret = v
	}
	if v := os.Getenv("COMPILER_URL"); v != "" {
		cfg.Compiler.URL = v
	}
	if v := os.Getenv("NATS_URL"); v != "" {
		cfg.NATS.URL = v
	}
	if v := os.Getenv("MLFLOW_TRACKING_URI"); v != "" {
		cfg.Integrations.MLflowURL = v
	}
	if v := os.Getenv("HF_TOKEN"); v != "" {
		cfg.Integrations.HFToken = v
	}
	if v := os.Getenv("DODO_API_KEY"); v != "" {
		cfg.Billing.DodoAPIKey = v
	}
	if v := os.Getenv("DODO_WEBHOOK_KEY"); v != "" {
		cfg.Billing.DodoWebhookKey = v
	}
	if v := os.Getenv("DODO_ENVIRONMENT"); v != "" {
		cfg.Billing.DodoEnvironment = v
	}
	if v := os.Getenv("DODO_STARTER_PRODUCT_ID"); v != "" {
		cfg.Billing.StarterProductID = v
	}
	if v := os.Getenv("DODO_PRO_PRODUCT_ID"); v != "" {
		cfg.Billing.ProProductID = v
	}
	if v := os.Getenv("BILLING_SUCCESS_URL"); v != "" {
		cfg.Billing.SuccessURL = v
	}
	if v := os.Getenv("BILLING_DASHBOARD_URL"); v != "" {
		cfg.Billing.DashboardURL = v
	}
	if v := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"); v != "" {
		cfg.Tracing.Endpoint = v
		cfg.Tracing.Enabled = true
	}

	return cfg, nil
}
