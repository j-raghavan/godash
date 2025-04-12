// cmd/godash/main.go
package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"

	"github.com/j-raghavan/godash/cmd/godash/core"
)

// Global config
var config core.Config

// OsExit for testing - allows tests to override os.Exit
var OsExit = os.Exit

// Execute runs the CLI application and returns an error if any
func Execute() error {
	return rootCmd.Execute()
}

func main() {
	if err := Execute(); err != nil {
		fmt.Println(err)
		OsExit(1)
	}
}

// rootCmd represents the root command of the application.
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
		core.RunMonitor(config)
	},
}

// serverCmd represents the server subcommand for the web dashboard
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the web dashboard server",
	Long: `Start GoDash web server, providing a dashboard accessible via browser
at http://localhost:<port> and metrics via REST API and WebSocket.`,
	Run: func(cmd *cobra.Command, args []string) {
		core.RunServer(config)
	},
}

// versionCmd represents the version subcommand
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of GoDash",
	Long:  `All software has versions. This is GoDash's.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(core.ShowVersion())
	},
}

func init() {
	// Define global flags that apply to all commands
	rootCmd.PersistentFlags().StringVarP(&config.ConfigFile, "config", "c", "", "config file (default is $HOME/.godash.toml)")
	rootCmd.PersistentFlags().IntVarP(&config.RefreshInterval, "interval", "i", 1, "Metrics refresh interval in seconds")
	rootCmd.PersistentFlags().BoolVarP(&config.EnableGoRuntime, "go-runtime", "g", false, "Enable Go runtime metrics")

	// Add flags specific to the server command
	serverCmd.Flags().IntVarP(&config.WebPort, "port", "p", 8080, "Port to serve dashboard on")

	// Add subcommands to root command
	rootCmd.AddCommand(monitorCmd)
	rootCmd.AddCommand(serverCmd)
	rootCmd.AddCommand(versionCmd)
}
