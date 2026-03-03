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
	Logging   LoggingConfig   `yaml:"logging"`
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

	return cfg, nil
}
