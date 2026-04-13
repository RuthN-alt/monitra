package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// Create a temporary config file
	configContent := `
check_interval: 30
database:
  path: test.db
targets:
  - name: TestSite
    url: https://example.com
    check_ssl: true
metrics:
  enabled: true
  port: 9091
  path: /metrics
`
	tmpfile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(configContent)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	// Test loading
	cfg, err := Load(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.CheckInterval != 30 {
		t.Errorf("Expected check_interval 30, got %d", cfg.CheckInterval)
	}

	if cfg.Database.Path != "test.db" {
		t.Errorf("Expected database path 'test.db', got '%s'", cfg.Database.Path)
	}

	if len(cfg.Targets) != 1 {
		t.Errorf("Expected 1 target, got %d", len(cfg.Targets))
	}

	if cfg.Targets[0].Name != "TestSite" {
		t.Errorf("Expected target name 'TestSite', got '%s'", cfg.Targets[0].Name)
	}
}

func TestLoadDefaults(t *testing.T) {
	configContent := `
targets:
  - name: Test
    url: https://example.com
    check_ssl: false
`
	tmpfile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(configContent)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Check defaults
	if cfg.CheckInterval != 60 {
		t.Errorf("Expected default check_interval 60, got %d", cfg.CheckInterval)
	}

	if cfg.Database.Path != "sentinel.db" {
		t.Errorf("Expected default database path 'sentinel.db', got '%s'", cfg.Database.Path)
	}

	if cfg.Metrics.Path != "/metrics" {
		t.Errorf("Expected default metrics path '/metrics', got '%s'", cfg.Metrics.Path)
	}

	if cfg.Metrics.Port != 9090 {
		t.Errorf("Expected default metrics port 9090, got %d", cfg.Metrics.Port)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  string
		wantErr bool
	}{
		{
			name: "valid config",
			config: `
targets:
  - name: Test
    url: https://example.com
`,
			wantErr: false,
		},
		{
			name: "no targets",
			config: `
targets: []
`,
			wantErr: true,
		},
		{
			name: "missing URL",
			config: `
targets:
  - name: Test
`,
			wantErr: true,
		},
		{
			name: "missing name",
			config: `
targets:
  - url: https://example.com
`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpfile, err := os.CreateTemp("", "config-*.yaml")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tmpfile.Name())

			if _, err := tmpfile.Write([]byte(tt.config)); err != nil {
				t.Fatal(err)
			}
			if err := tmpfile.Close(); err != nil {
				t.Fatal(err)
			}

			cfg, err := Load(tmpfile.Name())
			if err != nil {
				t.Fatalf("Failed to load config: %v", err)
			}

			err = Validate(cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
