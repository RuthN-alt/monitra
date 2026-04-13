package db

import (
	"Monitra/pkg/models"
	"os"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	tmpfile := "test_sentinel.db"
	defer os.Remove(tmpfile)

	db, err := New(tmpfile)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	if db.conn == nil {
		t.Error("Expected non-nil database connection")
	}
}

func TestSaveAndGetCheckResult(t *testing.T) {
	tmpfile := "test_sentinel.db"
	defer os.Remove(tmpfile)

	db, err := New(tmpfile)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Create a test result
	now := time.Now()
	sslDays := 90
	result := &models.CheckResult{
		TargetName:   "TestSite",
		TargetURL:    "https://example.com",
		StatusCode:   200,
		ResponseTime: 150,
		IsUp:         true,
		SSLDaysLeft:  &sslDays,
		CheckedAt:    now,
	}

	// Save the result
	err = db.SaveCheckResult(result)
	if err != nil {
		t.Fatalf("Failed to save check result: %v", err)
	}

	if result.ID == 0 {
		t.Error("Expected non-zero ID after save")
	}

	// Get latest results
	results, err := db.GetLatestResults()
	if err != nil {
		t.Fatalf("Failed to get latest results: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	r := results[0]
	if r.TargetName != result.TargetName {
		t.Errorf("Expected target name %s, got %s", result.TargetName, r.TargetName)
	}

	if r.StatusCode != result.StatusCode {
		t.Errorf("Expected status code %d, got %d", result.StatusCode, r.StatusCode)
	}

	if r.IsUp != result.IsUp {
		t.Errorf("Expected IsUp %v, got %v", result.IsUp, r.IsUp)
	}
}

func TestGetUptimeStats(t *testing.T) {
	tmpfile := "test_sentinel.db"
	defer os.Remove(tmpfile)

	db, err := New(tmpfile)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Create test results
	results := []*models.CheckResult{
		{
			TargetName:   "Site1",
			TargetURL:    "https://example1.com",
			StatusCode:   200,
			ResponseTime: 100,
			IsUp:         true,
			CheckedAt:    time.Now().Add(-3 * time.Minute),
		},
		{
			TargetName:   "Site1",
			TargetURL:    "https://example1.com",
			StatusCode:   200,
			ResponseTime: 120,
			IsUp:         true,
			CheckedAt:    time.Now().Add(-2 * time.Minute),
		},
		{
			TargetName:   "Site1",
			TargetURL:    "https://example1.com",
			StatusCode:   500,
			ResponseTime: 50,
			IsUp:         false,
			CheckedAt:    time.Now().Add(-1 * time.Minute),
		},
	}

	for _, result := range results {
		if err := db.SaveCheckResult(result); err != nil {
			t.Fatalf("Failed to save result: %v", err)
		}
	}

	// Get stats
	stats, err := db.GetUptimeStats()
	if err != nil {
		t.Fatalf("Failed to get uptime stats: %v", err)
	}

	if len(stats) != 1 {
		t.Fatalf("Expected 1 stat entry, got %d", len(stats))
	}

	s := stats[0]
	if s.TargetName != "Site1" {
		t.Errorf("Expected target name Site1, got %s", s.TargetName)
	}

	if s.TotalChecks != 3 {
		t.Errorf("Expected 3 total checks, got %d", s.TotalChecks)
	}

	if s.SuccessChecks != 2 {
		t.Errorf("Expected 2 success checks, got %d", s.SuccessChecks)
	}

	if s.FailedChecks != 1 {
		t.Errorf("Expected 1 failed check, got %d", s.FailedChecks)
	}

	expectedUptime := 66.67
	if s.UptimePercent < expectedUptime-1 || s.UptimePercent > expectedUptime+1 {
		t.Errorf("Expected uptime around %.2f%%, got %.2f%%", expectedUptime, s.UptimePercent)
	}
}

func TestGetRecentResults(t *testing.T) {
	tmpfile := "test_sentinel.db"
	defer os.Remove(tmpfile)

	db, err := New(tmpfile)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Create multiple results
	for i := 0; i < 5; i++ {
		result := &models.CheckResult{
			TargetName:   "TestSite",
			TargetURL:    "https://example.com",
			StatusCode:   200,
			ResponseTime: int64(100 + i*10),
			IsUp:         true,
			CheckedAt:    time.Now().Add(time.Duration(-i) * time.Minute),
		}
		if err := db.SaveCheckResult(result); err != nil {
			t.Fatalf("Failed to save result: %v", err)
		}
	}

	// Get recent results with limit
	results, err := db.GetRecentResults("TestSite", 3)
	if err != nil {
		t.Fatalf("Failed to get recent results: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}

	// Verify they're in descending order by time
	for i := 0; i < len(results)-1; i++ {
		if results[i].CheckedAt.Before(results[i+1].CheckedAt) {
			t.Error("Results not in descending order by time")
		}
	}
}
