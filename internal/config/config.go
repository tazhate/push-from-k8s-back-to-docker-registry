package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	// Kubernetes configuration (slices first for alignment)
	Namespaces  []string
	Deployments []string

	// Sync settings
	SyncPeriod time.Duration
	RetryDelay time.Duration

	// Target registry configuration
	RegistryURL          string
	RegistryUsername     string
	RegistryPassword     string
	ContainerdSocketPath string

	// Server settings
	MetricsAddr string
	HealthAddr  string
	LogLevel    string

	// Retry settings
	MaxRetries int
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		RegistryURL:          getEnv("TARGET_REGISTRY_URL", ""),
		RegistryUsername:     getEnv("TARGET_REGISTRY_USERNAME", ""),
		RegistryPassword:     getEnv("TARGET_REGISTRY_PASSWORD", ""),
		MaxRetries:           3,
		RetryDelay:           10 * time.Second,
		MetricsAddr:          getEnv("METRICS_ADDR", ":8080"),
		HealthAddr:           getEnv("HEALTH_ADDR", ":8081"),
		LogLevel:             getEnv("LOG_LEVEL", "info"),
		ContainerdSocketPath: getEnv("CONTAINERD_SOCKET_PATH", ""),
	}

	// Parse namespaces
	namespacesStr := getEnv("NAMESPACES", "")
	if namespacesStr != "" {
		cfg.Namespaces = strings.Split(namespacesStr, ",")
		for i := range cfg.Namespaces {
			cfg.Namespaces[i] = strings.TrimSpace(cfg.Namespaces[i])
		}
	}

	// Parse deployments (optional)
	deploymentsStr := getEnv("DEPLOYMENTS", "")
	if deploymentsStr != "" {
		cfg.Deployments = strings.Split(deploymentsStr, ",")
		for i := range cfg.Deployments {
			cfg.Deployments[i] = strings.TrimSpace(cfg.Deployments[i])
		}
	}

	// Parse sync period
	syncPeriodStr := getEnv("SYNC_PERIOD", "10m")
	syncPeriod, err := time.ParseDuration(syncPeriodStr)
	if err != nil {
		return nil, fmt.Errorf("invalid SYNC_PERIOD: %w", err)
	}
	cfg.SyncPeriod = syncPeriod

	// Validate required fields
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate checks if all required configuration is present
func (c *Config) Validate() error {
	if c.RegistryURL == "" {
		return fmt.Errorf("TARGET_REGISTRY_URL is required")
	}
	if c.RegistryUsername == "" {
		return fmt.Errorf("TARGET_REGISTRY_USERNAME is required")
	}
	if c.RegistryPassword == "" {
		return fmt.Errorf("TARGET_REGISTRY_PASSWORD is required")
	}
	if len(c.Namespaces) == 0 {
		return fmt.Errorf("NAMESPACES is required")
	}
	if c.SyncPeriod <= 0 {
		return fmt.Errorf("SYNC_PERIOD must be positive")
	}
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
