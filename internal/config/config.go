package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

// Config holds the application configuration
type Config struct {
	RefreshInterval int    `toml:"refresh_interval"`
	WebPort         int    `toml:"web_port"`
	EnableGoRuntime bool   `toml:"enable_go_runtime"`
	ConfigFile      string `toml:"-"`
}

// DefaultConfig returns a Config with default values
func DefaultConfig() Config {
	return Config{
		RefreshInterval: 1,
		WebPort:         8080,
		EnableGoRuntime: false,
	}
}

// LoadConfig loads configuration from a TOML file
func LoadConfig(configFile string) (Config, error) {
	cfg := DefaultConfig()
	cfg.ConfigFile = configFile

	// If no config file is specified, try to find one in default locations
	if configFile == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return cfg, fmt.Errorf("failed to get user home directory: %w", err)
		}

		// Try to find config in default locations
		possiblePaths := []string{
			filepath.Join(homeDir, ".godash.toml"),
			"godash.toml",
			".godash.toml",
		}

		for _, path := range possiblePaths {
			if _, err := os.Stat(path); err == nil {
				configFile = path
				break
			}
		}
	}

	// If we found a config file, load it
	if configFile != "" {
		data, err := os.ReadFile(configFile)
		if err != nil {
			return cfg, fmt.Errorf("failed to read config file: %w", err)
		}

		if err := toml.Unmarshal(data, &cfg); err != nil {
			return cfg, fmt.Errorf("failed to parse config file: %w", err)
		}
	}

	return cfg, nil
}

// SaveConfig saves the configuration to a TOML file
func SaveConfig(cfg Config) error {
	if cfg.ConfigFile == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get user home directory: %w", err)
		}
		cfg.ConfigFile = filepath.Join(homeDir, ".godash.toml")
	}

	data, err := toml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(cfg.ConfigFile, data, 0o644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
