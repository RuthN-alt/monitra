package models

import "time"

// Config represents the application configuration
type Config struct {
	Targets       []Target           `yaml:"targets"`
	CheckInterval int                `yaml:"check_interval"` // seconds
	Database      DatabaseConfig     `yaml:"database"`
	Notifications NotificationConfig `yaml:"notifications"`
	Metrics       MetricsConfig      `yaml:"metrics"`
}

// Target represents a URL to monitor
type Target struct {
	Name     string `yaml:"name"`
	URL      string `yaml:"url"`
	CheckSSL bool   `yaml:"check_ssl"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Path string `yaml:"path"`
}

// NotificationConfig holds notification settings
type NotificationConfig struct {
	Webhooks []WebhookConfig `yaml:"webhooks"`
	Email    EmailConfig     `yaml:"email"`
}

// WebhookConfig holds webhook configuration
type WebhookConfig struct {
	URL     string   `yaml:"url"`
	Events  []string `yaml:"events"` // up, down, ssl_expiring
	Enabled bool     `yaml:"enabled"`
}

// EmailConfig holds email configuration
type EmailConfig struct {
	Enabled bool       `yaml:"enabled"`
	SMTP    SMTPConfig `yaml:"smtp"`
	From    string     `yaml:"from"`
	To      []string   `yaml:"to"`
	Events  []string   `yaml:"events"`
}

// SMTPConfig holds SMTP server configuration
type SMTPConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	UseTLS   bool   `yaml:"use_tls"`
}

// MetricsConfig holds Prometheus metrics configuration
type MetricsConfig struct {
	Enabled bool   `yaml:"enabled"`
	Port    int    `yaml:"port"`
	Path    string `yaml:"path"`
}

// CheckResult represents the result of a health check
type CheckResult struct {
	ID            int64
	TargetName    string
	TargetURL     string
	StatusCode    int
	ResponseTime  int64 // milliseconds
	IsUp          bool
	SSLExpiration *time.Time
	SSLDaysLeft   *int
	Error         string
	CheckedAt     time.Time
}

// UptimeStats represents uptime statistics for a target
type UptimeStats struct {
	TargetName      string
	TotalChecks     int64
	SuccessChecks   int64
	FailedChecks    int64
	UptimePercent   float64
	AvgResponseTime float64
	LastCheck       *time.Time
	LastStatus      string
}
