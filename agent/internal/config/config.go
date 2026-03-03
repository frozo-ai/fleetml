package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	DeviceID           string        `yaml:"device_id"`
	ServerAddress      string        `yaml:"server_address"`
	HeartbeatInterval  time.Duration `yaml:"heartbeat_interval"`
	ModelStorageDir    string        `yaml:"model_storage_dir"`
	MaxModelVersions   int           `yaml:"max_model_versions"`
	LogLevel           string        `yaml:"log_level"`
	Mode               string        `yaml:"mode"` // production, test
}

func DefaultConfig() *Config {
	return &Config{
		DeviceID:          "",
		ServerAddress:     "localhost:50051",
		HeartbeatInterval: 30 * time.Second,
		ModelStorageDir:   "/var/lib/fleetml/models",
		MaxModelVersions:  3,
		LogLevel:          "info",
		Mode:              "production",
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
	if v := os.Getenv("FLEETML_SERVER"); v != "" {
		cfg.ServerAddress = v
	}
	if v := os.Getenv("DEVICE_ID"); v != "" {
		cfg.DeviceID = v
	}
	if v := os.Getenv("FLEETML_MODE"); v != "" {
		cfg.Mode = v
	}

	return cfg, nil
}
