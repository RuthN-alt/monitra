package db

import (
	"Monitra/pkg/models"
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// DB represents the database connection
type DB struct {
	conn *sql.DB
}

// New creates a new database connection and initializes schema
func New(path string) (*DB, error) {
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db := &DB{conn: conn}
	if err := db.initSchema(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return db, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.conn.Close()
}

// initSchema creates the required database tables
func (db *DB) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS check_results (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		target_name TEXT NOT NULL,
		target_url TEXT NOT NULL,
		status_code INTEGER NOT NULL,
		response_time INTEGER NOT NULL,
		is_up BOOLEAN NOT NULL,
		ssl_expiration TEXT,
		ssl_days_left INTEGER,
		error TEXT,
		checked_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	
	CREATE INDEX IF NOT EXISTS idx_target_name ON check_results(target_name);
	CREATE INDEX IF NOT EXISTS idx_checked_at ON check_results(checked_at);
	`

	_, err := db.conn.Exec(schema)
	return err
}

// SaveCheckResult saves a check result to the database
func (db *DB) SaveCheckResult(result *models.CheckResult) error {
	var sslExpiration *string
	if result.SSLExpiration != nil {
		exp := result.SSLExpiration.Format(time.RFC3339)
		sslExpiration = &exp
	}

	query := `
		INSERT INTO check_results 
		(target_name, target_url, status_code, response_time, is_up, ssl_expiration, ssl_days_left, error, checked_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	res, err := db.conn.Exec(query,
		result.TargetName,
		result.TargetURL,
		result.StatusCode,
		result.ResponseTime,
		result.IsUp,
		sslExpiration,
		result.SSLDaysLeft,
		result.Error,
		result.CheckedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save check result: %w", err)
	}

	id, _ := res.LastInsertId()
	result.ID = id
	return nil
}

// GetLatestResults retrieves the latest check result for each target
func (db *DB) GetLatestResults() ([]*models.CheckResult, error) {
	query := `
		SELECT id, target_name, target_url, status_code, response_time, is_up, 
		       ssl_expiration, ssl_days_left, error, checked_at
		FROM check_results
		WHERE id IN (
			SELECT MAX(id) FROM check_results GROUP BY target_name
		)
		ORDER BY target_name
	`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query latest results: %w", err)
	}
	defer rows.Close()

	var results []*models.CheckResult
	for rows.Next() {
		result := &models.CheckResult{}
		var sslExpiration *string

		err := rows.Scan(
			&result.ID,
			&result.TargetName,
			&result.TargetURL,
			&result.StatusCode,
			&result.ResponseTime,
			&result.IsUp,
			&sslExpiration,
			&result.SSLDaysLeft,
			&result.Error,
			&result.CheckedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan result: %w", err)
		}

		if sslExpiration != nil {
			t, err := time.Parse(time.RFC3339, *sslExpiration)
			if err == nil {
				result.SSLExpiration = &t
			}
		}

		results = append(results, result)
	}

	return results, nil
}

// GetUptimeStats calculates uptime statistics for all targets
func (db *DB) GetUptimeStats() ([]*models.UptimeStats, error) {
	query := `
		WITH latest_checks AS (
			SELECT target_name, is_up
			FROM check_results
			WHERE id IN (
				SELECT MAX(id) FROM check_results GROUP BY target_name
			)
		)
		SELECT 
			cr.target_name,
			COUNT(*) as total_checks,
			SUM(CASE WHEN cr.is_up THEN 1 ELSE 0 END) as success_checks,
			SUM(CASE WHEN NOT cr.is_up THEN 1 ELSE 0 END) as failed_checks,
			CAST(SUM(CASE WHEN cr.is_up THEN 1 ELSE 0 END) AS REAL) * 100.0 / COUNT(*) as uptime_percent,
			AVG(cr.response_time) as avg_response_time,
			MAX(cr.checked_at) as last_check,
			lc.is_up as last_status
		FROM check_results cr
		JOIN latest_checks lc ON cr.target_name = lc.target_name
		GROUP BY cr.target_name, lc.is_up
		ORDER BY cr.target_name
	`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query uptime stats: %w", err)
	}
	defer rows.Close()

	var stats []*models.UptimeStats
	for rows.Next() {
		stat := &models.UptimeStats{}
		var lastCheckStr string
		var lastStatusBool bool

		err := rows.Scan(
			&stat.TargetName,
			&stat.TotalChecks,
			&stat.SuccessChecks,
			&stat.FailedChecks,
			&stat.UptimePercent,
			&stat.AvgResponseTime,
			&lastCheckStr,
			&lastStatusBool,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan stat: %w", err)
		}

		if lastCheckStr != "" {
			t, err := time.Parse("2006-01-02 15:04:05", lastCheckStr)
			if err == nil {
				stat.LastCheck = &t
			}
		}

		if lastStatusBool {
			stat.LastStatus = "UP"
		} else {
			stat.LastStatus = "DOWN"
		}

		stats = append(stats, stat)
	}

	return stats, nil
}

// GetRecentResults retrieves recent check results for a target
func (db *DB) GetRecentResults(targetName string, limit int) ([]*models.CheckResult, error) {
	query := `
		SELECT id, target_name, target_url, status_code, response_time, is_up, 
		       ssl_expiration, ssl_days_left, error, checked_at
		FROM check_results
		WHERE target_name = ?
		ORDER BY checked_at DESC
		LIMIT ?
	`

	rows, err := db.conn.Query(query, targetName, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query recent results: %w", err)
	}
	defer rows.Close()

	var results []*models.CheckResult
	for rows.Next() {
		result := &models.CheckResult{}
		var sslExpiration *string

		err := rows.Scan(
			&result.ID,
			&result.TargetName,
			&result.TargetURL,
			&result.StatusCode,
			&result.ResponseTime,
			&result.IsUp,
			&sslExpiration,
			&result.SSLDaysLeft,
			&result.Error,
			&result.CheckedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan result: %w", err)
		}

		if sslExpiration != nil {
			t, err := time.Parse(time.RFC3339, *sslExpiration)
			if err == nil {
				result.SSLExpiration = &t
			}
		}

		results = append(results, result)
	}

	return results, nil
}
