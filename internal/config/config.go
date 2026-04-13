package config

import (
	"Monitra/pkg/models"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Load reads and parses the YAML configuration file
func Load(path string) (*models.Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg models.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set defaults
	if cfg.CheckInterval == 0 {
		cfg.CheckInterval = 60 // default: check every 60 seconds
	}
	if cfg.Database.Path == "" {
		cfg.Database.Path = "sentinel.db"
	}
	if cfg.Metrics.Path == "" {
		cfg.Metrics.Path = "/metrics"
	}
	if cfg.Metrics.Port == 0 {
		cfg.Metrics.Port = 9090
	}

	return &cfg, nil
}

// Validate checks if the configuration is valid
func Validate(cfg *models.Config) error {
	if len(cfg.Targets) == 0 {
		return fmt.Errorf("no targets defined in configuration")
	}

	for i, target := range cfg.Targets {
		if target.URL == "" {
			return fmt.Errorf("target %d: URL is required", i)
		}
		if target.Name == "" {
			return fmt.Errorf("target %d: name is required", i)
		}
	}

	return nil
}
