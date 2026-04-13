package metrics

import (
	"Monitra/pkg/models"
	"fmt"
	"net/http"
	"strings"
	"sync"
)

// Server provides Prometheus metrics endpoint
type Server struct {
	config  models.MetricsConfig
	results map[string]*models.CheckResult
	mu      sync.RWMutex
}

// New creates a new metrics server
func New(config models.MetricsConfig) *Server {
	return &Server{
		config:  config,
		results: make(map[string]*models.CheckResult),
	}
}

// Update updates the metrics with new check results
func (s *Server) Update(results []*models.CheckResult) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, result := range results {
		s.results[result.TargetName] = result
	}
}

// Start starts the metrics HTTP server
func (s *Server) Start() error {
	if !s.config.Enabled {
		return nil
	}

	http.HandleFunc(s.config.Path, s.metricsHandler)

	addr := fmt.Sprintf(":%d", s.config.Port)
	fmt.Printf("Prometheus metrics endpoint available at http://localhost%s%s\n", addr, s.config.Path)

	go func() {
		if err := http.ListenAndServe(addr, nil); err != nil {
			fmt.Printf("Failed to start metrics server: %v\n", err)
		}
	}()

	return nil
}

// metricsHandler handles the /metrics HTTP endpoint
func (s *Server) metricsHandler(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var metrics strings.Builder

	// Write metrics header
	metrics.WriteString("# HELP site_Monitra_up Whether the target is up (1) or down (0)\n")
	metrics.WriteString("# TYPE site_Monitra_up gauge\n")

	for name, result := range s.results {
		upValue := 0
		if result.IsUp {
			upValue = 1
		}
		metrics.WriteString(fmt.Sprintf("site_Monitra_up{target=\"%s\",url=\"%s\"} %d\n",
			sanitizeLabel(name), sanitizeLabel(result.TargetURL), upValue))
	}

	metrics.WriteString("\n")
	metrics.WriteString("# HELP site_Monitra_response_time_ms Response time in milliseconds\n")
	metrics.WriteString("# TYPE site_Monitra_response_time_ms gauge\n")

	for name, result := range s.results {
		metrics.WriteString(fmt.Sprintf("site_Monitra_response_time_ms{target=\"%s\",url=\"%s\"} %d\n",
			sanitizeLabel(name), sanitizeLabel(result.TargetURL), result.ResponseTime))
	}

	metrics.WriteString("\n")
	metrics.WriteString("# HELP site_Monitra_status_code HTTP status code\n")
	metrics.WriteString("# TYPE site_Monitra_status_code gauge\n")

	for name, result := range s.results {
		metrics.WriteString(fmt.Sprintf("site_Monitra_status_code{target=\"%s\",url=\"%s\"} %d\n",
			sanitizeLabel(name), sanitizeLabel(result.TargetURL), result.StatusCode))
	}

	metrics.WriteString("\n")
	metrics.WriteString("# HELP site_Monitra_ssl_days_left Days until SSL certificate expiration\n")
	metrics.WriteString("# TYPE site_Monitra_ssl_days_left gauge\n")

	for name, result := range s.results {
		if result.SSLDaysLeft != nil {
			metrics.WriteString(fmt.Sprintf("site_Monitra_ssl_days_left{target=\"%s\",url=\"%s\"} %d\n",
				sanitizeLabel(name), sanitizeLabel(result.TargetURL), *result.SSLDaysLeft))
		}
	}

	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(metrics.String()))
}

// sanitizeLabel sanitizes a label value for Prometheus
func sanitizeLabel(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	return s
}
