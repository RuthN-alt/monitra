package monitor

import (
	"Monitra/pkg/models"
	"testing"
)

func TestNew(t *testing.T) {
	m := New()
	if m == nil {
		t.Error("Expected non-nil monitor")
	}
	if m.client == nil {
		t.Error("Expected non-nil HTTP client")
	}
}

func TestCheck(t *testing.T) {
	m := New()

	// Test basic functionality with a valid URL structure
	target := models.Target{
		Name:     "Example",
		URL:      "https://example.com",
		CheckSSL: true,
	}

	result := m.Check(target)

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	if result.TargetName != target.Name {
		t.Errorf("Expected target name %s, got %s", target.Name, result.TargetName)
	}

	if result.TargetURL != target.URL {
		t.Errorf("Expected target URL %s, got %s", target.URL, result.TargetURL)
	}

	if result.CheckedAt.IsZero() {
		t.Error("Expected non-zero checked_at time")
	}

	// Response time should be set even if there's an error
	if result.ResponseTime < 0 {
		t.Error("Expected non-negative response time")
	}

	// Test with definitely invalid URL
	invalidTarget := models.Target{
		Name:     "Invalid",
		URL:      "not-a-valid-url",
		CheckSSL: false,
	}

	invalidResult := m.Check(invalidTarget)
	if invalidResult == nil {
		t.Fatal("Expected non-nil result for invalid URL")
	}

	if invalidResult.IsUp {
		t.Error("Expected invalid URL to be down")
	}

	if invalidResult.Error == "" {
		t.Error("Expected error message for invalid URL")
	}
}

func TestCheckAll(t *testing.T) {
	m := New()

	targets := []models.Target{
		{
			Name:     "Example",
			URL:      "https://example.com",
			CheckSSL: true,
		},
		{
			Name:     "Invalid",
			URL:      "http://invalid-url-12345.com",
			CheckSSL: false,
		},
	}

	results := m.CheckAll(targets)

	if len(results) != len(targets) {
		t.Errorf("Expected %d results, got %d", len(targets), len(results))
	}

	// Verify all results are populated
	for i, result := range results {
		if result == nil {
			t.Errorf("Result %d is nil", i)
			continue
		}
		if result.TargetName == "" {
			t.Errorf("Result %d has empty target name", i)
		}
		if result.CheckedAt.IsZero() {
			t.Errorf("Result %d has zero checked_at time", i)
		}
	}
}
