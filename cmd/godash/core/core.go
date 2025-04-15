package core

import (
	"fmt"
	"time"

	"github.com/j-raghavan/godash/internal/metrics"
	"github.com/j-raghavan/godash/internal/tui"
)

// Config holds application configuration
type Config struct {
	ConfigFile      string
	RefreshInterval int
	WebPort         int
	EnableGoRuntime bool
}

// RunMonitor contains the actual monitor logic
func RunMonitor(cfg Config) {
	fmt.Printf("Starting GoDash monitor with refresh interval: %ds\n", cfg.RefreshInterval)
	if cfg.EnableGoRuntime {
		fmt.Println("Go runtime metrics enabled.")
	} else {
		fmt.Println("Go runtime metrics disabled.")
	}

	// Create a new metrics collector
	collector := metrics.NewSystemCollector()

	// Create a new UI instance
	ui := tui.NewUI(collector, cfg.EnableGoRuntime)

	// Start the UI with the configured refresh interval
	refreshInterval := time.Duration(cfg.RefreshInterval) * time.Second
	if err := ui.Start(refreshInterval); err != nil {
		fmt.Printf("Error starting UI: %v\n", err)
		return
	}
}

// RunServer contains the actual server logic
func RunServer(cfg Config) {
	fmt.Printf("Starting GoDash web server on port %d\n", cfg.WebPort)
	fmt.Printf("Refresh interval: %ds\n", cfg.RefreshInterval)
	if cfg.EnableGoRuntime {
		fmt.Println("Go runtime metrics enabled")
	}

	// This is where you would initialize and start the web server
	fmt.Println("Web server would start here (implementation pending)")
}

// ShowVersion displays version info
func ShowVersion() string {
	return "GoDash v0.1.0"
}
