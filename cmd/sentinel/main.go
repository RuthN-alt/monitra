package main

import (
	"Monitra/internal/config"
	"Monitra/internal/dashboard"
	"Monitra/internal/db"
	"Monitra/internal/metrics"
	"Monitra/internal/monitor"
	"Monitra/internal/notifier"
	"Monitra/pkg/models"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "monitor":
		runMonitor()
	case "dashboard":
		runDashboard()
	case "stats":
		runStats()
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Monitra - URL Monitoring Tool")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  Monitra monitor [config-file]  - Start monitoring URLs")
	fmt.Println("  Monitra dashboard [config-file] - Show live dashboard")
	fmt.Println("  Monitra stats [config-file]    - Show uptime statistics")
	fmt.Println("  Monitra help                   - Show this help message")
	fmt.Println()
	fmt.Println("Default config file: config.yaml")
}

func getConfigPath() string {
	if len(os.Args) > 2 {
		return os.Args[2]
	}
	return "config.yaml"
}

func runMonitor() {
	configPath := getConfigPath()
	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	if err := config.Validate(cfg); err != nil {
		fmt.Printf("Invalid config: %v\n", err)
		os.Exit(1)
	}

	// Initialize database
	database, err := db.New(cfg.Database.Path)
	if err != nil {
		fmt.Printf("Error initializing database: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	// Initialize monitor
	mon := monitor.New()

	// Initialize notifier
	notif := notifier.New(cfg.Notifications)

	// Initialize metrics server
	metricsServer := metrics.New(cfg.Metrics)
	if err := metricsServer.Start(); err != nil {
		fmt.Printf("Error starting metrics server: %v\n", err)
	}

	// Initialize dashboard
	dash := dashboard.New()

	fmt.Printf("Starting monitoring of %d targets...\n", len(cfg.Targets))
	fmt.Printf("Check interval: %d seconds\n", cfg.CheckInterval)
	fmt.Printf("Database: %s\n", cfg.Database.Path)
	fmt.Println()

	// Set up channels
	resultChan := make(chan []*models.CheckResult)
	stopChan := make(chan bool)

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start monitoring in a goroutine
	go mon.MonitorLoop(cfg.Targets, time.Duration(cfg.CheckInterval)*time.Second, resultChan, stopChan)

	// Process results
	for {
		select {
		case results := <-resultChan:
			// Save to database
			for _, result := range results {
				if err := database.SaveCheckResult(result); err != nil {
					fmt.Printf("Error saving result: %v\n", err)
				}

				// Send notifications
				notif.Notify(result)
			}

			// Update metrics
			metricsServer.Update(results)

			// Display in terminal
			dash.DisplayLiveMonitoring(results, cfg.CheckInterval)

		case <-sigChan:
			fmt.Println("\nShutting down amazing...")
			stopChan <- true
			time.Sleep(1 * time.Second)
			return
		}
	}
}

func runDashboard() {
	configPath := getConfigPath()
	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Initialize database
	database, err := db.New(cfg.Database.Path)
	if err != nil {
		fmt.Printf("Error initializing database: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	dash := dashboard.New()

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// Display initial results
	results, err := database.GetLatestResults()
	if err != nil {
		fmt.Printf("Error getting results: %v\n", err)
		os.Exit(1)
	}
	dash.DisplayResults(results)

	for {
		select {
		case <-ticker.C:
			results, err := database.GetLatestResults()
			if err != nil {
				fmt.Printf("Error getting results: %v\n", err)
				continue
			}
			dash.DisplayResults(results)

		case <-sigChan:
			fmt.Println("\nExiting...")
			return
		}
	}
}

func runStats() {
	configPath := getConfigPath()
	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Initialize database
	database, err := db.New(cfg.Database.Path)
	if err != nil {
		fmt.Printf("Error initializing database: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	dash := dashboard.New()

	stats, err := database.GetUptimeStats()
	if err != nil {
		fmt.Printf("Error getting stats: %v\n", err)
		os.Exit(1)
	}

	dash.DisplayStats(stats)
}
