package dashboard

import (
	"Monitra/pkg/models"
	"fmt"
	"strings"
	"time"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
	colorWhite  = "\033[97m"
	bold        = "\033[1m"
)

// Terminal provides terminal dashboard functionality
type Terminal struct{}

// New creates a new terminal dashboard
func New() *Terminal {
	return &Terminal{}
}

// Clear clears the terminal screen
func (t *Terminal) Clear() {
	fmt.Print("\033[H\033[2J")
}

// DisplayResults shows the current check results in a formatted table
func (t *Terminal) DisplayResults(results []*models.CheckResult) {
	t.Clear()
	fmt.Println(bold + colorCyan + "╔══════════════════════════════════════════════════════════════════════════════╗" + colorReset)
	fmt.Println(bold + colorCyan + "║" + colorWhite + "                        MONITRA DASHBOARD                               " + colorCyan + "║" + colorReset)
	fmt.Println(bold + colorCyan + "╚══════════════════════════════════════════════════════════════════════════════╝" + colorReset)
	fmt.Println()

	if len(results) == 0 {
		fmt.Println(colorYellow + "No monitoring data available yet." + colorReset)
		return
	}

	fmt.Printf("%s%-25s %-10s %-12s %-12s %-20s%s\n",
		bold+colorWhite,
		"TARGET",
		"STATUS",
		"STATUS CODE",
		"RESPONSE",
		"SSL DAYS LEFT",
		colorReset,
	)
	fmt.Println(strings.Repeat("─", 90))

	for _, result := range results {
		t.displayResult(result)
	}

	fmt.Println()
	fmt.Println(colorCyan + "Last updated: " + time.Now().Format("2006-01-02 15:04:05") + colorReset)
}

// displayResult displays a single check result
func (t *Terminal) displayResult(result *models.CheckResult) {
	statusColor := colorRed
	statusText := "DOWN"
	if result.IsUp {
		statusColor = colorGreen
		statusText = "UP"
	}

	sslText := "N/A"
	sslColor := colorWhite
	if result.SSLDaysLeft != nil {
		sslText = fmt.Sprintf("%d days", *result.SSLDaysLeft)
		if *result.SSLDaysLeft <= 7 {
			sslColor = colorRed
		} else if *result.SSLDaysLeft <= 30 {
			sslColor = colorYellow
		} else {
			sslColor = colorGreen
		}
	}

	name := result.TargetName
	if len(name) > 24 {
		name = name[:21] + "..."
	}

	fmt.Printf("%-25s %s%-10s%s %-12d %-12s %s%-20s%s\n",
		name,
		statusColor, statusText, colorReset,
		result.StatusCode,
		fmt.Sprintf("%dms", result.ResponseTime),
		sslColor, sslText, colorReset,
	)

	if result.Error != "" {
		errorMsg := result.Error
		if len(errorMsg) > 70 {
			errorMsg = errorMsg[:67] + "..."
		}
		fmt.Printf("  %s└─ Error: %s%s\n", colorRed, errorMsg, colorReset)
	}
}

// DisplayStats shows uptime statistics
func (t *Terminal) DisplayStats(stats []*models.UptimeStats) {
	t.Clear()
	fmt.Println(bold + colorCyan + "╔══════════════════════════════════════════════════════════════════════════════╗" + colorReset)
	fmt.Println(bold + colorCyan + "║" + colorWhite + "                        UPTIME STATISTICS                                     " + colorCyan + "║" + colorReset)
	fmt.Println(bold + colorCyan + "╚══════════════════════════════════════════════════════════════════════════════╝" + colorReset)
	fmt.Println()

	if len(stats) == 0 {
		fmt.Println(colorYellow + "No statistics available yet." + colorReset)
		return
	}

	fmt.Printf("%s%-25s %-12s %-12s %-15s %-15s%s\n",
		bold+colorWhite,
		"TARGET",
		"UPTIME %",
		"CHECKS",
		"AVG RESPONSE",
		"LAST CHECK",
		colorReset,
	)
	fmt.Println(strings.Repeat("─", 90))

	for _, stat := range stats {
		t.displayStat(stat)
	}

	fmt.Println()
	fmt.Println(colorCyan + "Generated at: " + time.Now().Format("2006-01-02 15:04:05") + colorReset)
}

// displayStat displays a single uptime statistic
func (t *Terminal) displayStat(stat *models.UptimeStats) {
	uptimeColor := colorRed
	if stat.UptimePercent >= 99.0 {
		uptimeColor = colorGreen
	} else if stat.UptimePercent >= 95.0 {
		uptimeColor = colorYellow
	}

	name := stat.TargetName
	if len(name) > 24 {
		name = name[:21] + "..."
	}

	lastCheck := "N/A"
	if stat.LastCheck != nil {
		lastCheck = stat.LastCheck.Format("15:04:05")
	}

	fmt.Printf("%-25s %s%-12s%s %-12d %-15s %-15s\n",
		name,
		uptimeColor, fmt.Sprintf("%.2f%%", stat.UptimePercent), colorReset,
		stat.TotalChecks,
		fmt.Sprintf("%.0fms", stat.AvgResponseTime),
		lastCheck,
	)
}

// DisplayLiveMonitoring shows live monitoring with auto-refresh
func (t *Terminal) DisplayLiveMonitoring(results []*models.CheckResult, interval int) {
	t.DisplayResults(results)
	fmt.Println()
	fmt.Printf("%sPress Ctrl+C to stop monitoring (refresh every %ds)%s\n",
		colorYellow, interval, colorReset)
}
