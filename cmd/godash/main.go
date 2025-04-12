package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

// Global variables
// This file contains global variables that are used throughout the application.
var (
	configFile      string = "config.toml" // Default config file name
	refreshInterval int    = 5             // Default refresh interval in seconds
	webPort         int    = 8080          // Default web server port
	enableGoRuntime bool   = true          // Enable Go runtime metrics
)

// This is the main entry point of the application.
func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// rootCmd represents the root command of the application.
// It is used to define the command-line interface (CLI) for the application.
var rootCmd = &cobra.Command{
	Use:   "godash",
	Short: "GoDash - Cross platform system monitoring tool",
	Long: `GoDash is a self-contained, cross-platform system monitoring tool 
	that provides real-time system resource metrics via both a CLI 
	and a lightweight local web dashboard.
	
	It's designed for developers, DevOps engineers, and homelab enthusiasts
	who need a portable and install-free performance monitor.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

// monitorCmd represents the monitor subcommand for CLI
var monitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Start the interactive CLI monitor",
	Long: `Start GoDash in terminal UI mode, displaying real-time system metrics.
Press 'q' to quit, 'g' to toggle Go runtime stats.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Starting GoDash monitor with refresh interval: %ds\n", refreshInterval)
		if enableGoRuntime {
			fmt.Println("Go runtime metrics enabled.")
		} else {
			fmt.Println("Go runtime metrics disabled.")
		}
		// This is where you would initialize and start the TUI
		fmt.Println("CLI monitor would start here (implementation pending)")
	},
}

// serverCmd represents the server subcommand for the web dashboard
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the web dashboard server",
	Long: `Start GoDash web server, providing a dashboard accessible via browser
at http://localhost:<port> and metrics via REST API and WebSocket.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Starting GoDash web server on port %d\n", webPort)
		fmt.Printf("Refresh interval: %ds\n", refreshInterval)
		if enableGoRuntime {
			fmt.Println("Go runtime metrics enabled")
		}

		// This is where you would initialize and start the web server
		fmt.Println("Web server would start here (implementation pending)")
	},
}

// versionCmd represents the version subcommand
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of GoDash",
	Long:  `All software has versions. This is GoDash's.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("GoDash v0.1.0")
	},
}

func init() {
	// Define global flags that apply to all commands
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "config file (default is $HOME/.godash.yaml)")
	rootCmd.PersistentFlags().IntVarP(&refreshInterval, "interval", "i", 1, "Metrics refresh interval in seconds")
	rootCmd.PersistentFlags().BoolVarP(&enableGoRuntime, "go-runtime", "g", false, "Enable Go runtime metrics")

	// Add flags specific to the server command
	serverCmd.Flags().IntVarP(&webPort, "port", "p", 8080, "Port to serve dashboard on")

	// Add subcommands to root command
	rootCmd.AddCommand(monitorCmd)
	rootCmd.AddCommand(serverCmd)
	rootCmd.AddCommand(versionCmd)
}
