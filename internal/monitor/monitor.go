package monitor

import (
	"Monitra/pkg/models"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"
)

// Monitor handles health checks for targets
type Monitor struct {
	client *http.Client
}

// New creates a new monitor instance
func New() *Monitor {
	return &Monitor{
		client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: false,
				},
			},
		},
	}
}

// Check performs a health check on a target
func (m *Monitor) Check(target models.Target) *models.CheckResult {
	result := &models.CheckResult{
		TargetName: target.Name,
		TargetURL:  target.URL,
		CheckedAt:  time.Now(),
	}

	startTime := time.Now()
	resp, err := m.client.Get(target.URL)
	responseTime := time.Since(startTime).Milliseconds()

	result.ResponseTime = responseTime

	if err != nil {
		result.IsUp = false
		result.Error = err.Error()
		result.StatusCode = 0
		return result
	}
	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode
	result.IsUp = resp.StatusCode >= 200 && resp.StatusCode < 400

	// Check SSL certificate if requested and HTTPS
	if target.CheckSSL && resp.TLS != nil && len(resp.TLS.PeerCertificates) > 0 {
		cert := resp.TLS.PeerCertificates[0]
		result.SSLExpiration = &cert.NotAfter
		daysLeft := int(time.Until(cert.NotAfter).Hours() / 24)
		result.SSLDaysLeft = &daysLeft
	}

	return result
}

// CheckAll performs health checks on multiple targets concurrently
func (m *Monitor) CheckAll(targets []models.Target) []*models.CheckResult {
	results := make([]*models.CheckResult, len(targets))
	done := make(chan bool)

	for i, target := range targets {
		go func(index int, tgt models.Target) {
			results[index] = m.Check(tgt)
			done <- true
		}(i, target)
	}

	// Wait for all checks to complete
	for range targets {
		<-done
	}

	return results
}

// MonitorLoop continuously monitors targets at specified intervals
func (m *Monitor) MonitorLoop(targets []models.Target, interval time.Duration, resultChan chan<- []*models.CheckResult, stopChan <-chan bool) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Perform initial check immediately
	results := m.CheckAll(targets)
	resultChan <- results

	for {
		select {
		case <-ticker.C:
			results := m.CheckAll(targets)
			resultChan <- results
		case <-stopChan:
			fmt.Println("Monitoring stopped")
			return
		}
	}
}
