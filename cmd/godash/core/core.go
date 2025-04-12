package core

import (
	"fmt"
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
	// This is where you would initialize and start the TUI
	fmt.Println("CLI monitor would start here (implementation pending)")
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
