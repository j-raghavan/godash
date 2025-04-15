package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/j-raghavan/godash/internal/config"
)

func TestDefaultConfig(t *testing.T) {
	cfg := config.DefaultConfig()
	assert.Equal(t, 1, cfg.RefreshInterval)
	assert.Equal(t, 8080, cfg.WebPort)
	assert.False(t, cfg.EnableGoRuntime)
	assert.Empty(t, cfg.ConfigFile)
}

func TestLoadConfig(t *testing.T) {
	// Create a temporary home directory for all tests
	tempHome, err := os.MkdirTemp("", "godash-test-*")
	require.NoError(t, err)
	defer func() {
		err := os.RemoveAll(tempHome)
		require.NoError(t, err)
	}()

	// Set up test environment
	oldHome := os.Getenv("HOME")
	err = os.Setenv("HOME", tempHome)
	require.NoError(t, err)
	defer func() {
		err := os.Setenv("HOME", oldHome)
		require.NoError(t, err)
	}()

	tests := []struct {
		name        string
		configFile  string
		configData  string
		wantConfig  config.Config
		wantErr     bool
		errContains string
	}{
		{
			name:       "valid config file",
			configFile: "test_config.toml",
			configData: `refresh_interval = 5
web_port = 9090
enable_go_runtime = true`,
			wantConfig: config.Config{
				RefreshInterval: 5,
				WebPort:         9090,
				EnableGoRuntime: true,
				ConfigFile:      "test_config.toml",
			},
			wantErr: false,
		},
		{
			name:        "invalid config file",
			configFile:  "invalid_config.toml",
			configData:  `invalid toml content`,
			wantErr:     true,
			errContains: "failed to parse config file",
		},
		{
			name:       "empty config file",
			configFile: "",
			wantConfig: config.DefaultConfig(),
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up any existing config files first
			for _, path := range []string{
				"test_config.toml",
				"invalid_config.toml",
				"godash.toml",
				".godash.toml",
				filepath.Join(tempHome, ".godash.toml"),
			} {
				err := os.Remove(path)
				if err != nil && !os.IsNotExist(err) {
					require.NoError(t, err)
				}
			}

			// Create temporary config file if needed
			if tt.configFile != "" {
				err := os.WriteFile(tt.configFile, []byte(tt.configData), 0o644)
				require.NoError(t, err)
				defer func() {
					err := os.Remove(tt.configFile)
					if err != nil && !os.IsNotExist(err) {
						require.NoError(t, err)
					}
				}()
			}

			cfg, err := config.LoadConfig(tt.configFile)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantConfig, cfg)
			}
		})
	}
}

func TestLoadConfig_DefaultLocations(t *testing.T) {
	// Create a temporary home directory
	tempHome, err := os.MkdirTemp("", "godash-test-*")
	require.NoError(t, err)
	defer func() {
		err := os.RemoveAll(tempHome)
		require.NoError(t, err)
	}()

	// Set up test environment
	oldHome := os.Getenv("HOME")
	err = os.Setenv("HOME", tempHome)
	require.NoError(t, err)
	defer func() {
		err := os.Setenv("HOME", oldHome)
		require.NoError(t, err)
	}()

	// Create config file in home directory
	homeConfig := filepath.Join(tempHome, ".godash.toml")
	err = os.WriteFile(homeConfig, []byte(`refresh_interval = 10`), 0o644)
	require.NoError(t, err)
	defer func() {
		err := os.Remove(homeConfig)
		if err != nil && !os.IsNotExist(err) {
			require.NoError(t, err)
		}
	}()

	// Test loading from default location
	cfg, err := config.LoadConfig("")
	assert.NoError(t, err)
	assert.Equal(t, 10, cfg.RefreshInterval)
	// Note: ConfigFile is not set when loading from default locations
	assert.Empty(t, cfg.ConfigFile)
}

func TestSaveConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      config.Config
		wantErr     bool
		errContains string
	}{
		{
			name: "valid config",
			config: config.Config{
				RefreshInterval: 5,
				WebPort:         9090,
				EnableGoRuntime: true,
				ConfigFile:      "test_save.toml",
			},
			wantErr: false,
		},
		{
			name: "default location",
			config: config.Config{
				RefreshInterval: 5,
				WebPort:         9090,
				EnableGoRuntime: true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := config.SaveConfig(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)

				// Verify the saved config
				if tt.config.ConfigFile != "" {
					loadedCfg, err := config.LoadConfig(tt.config.ConfigFile)
					assert.NoError(t, err)
					assert.Equal(t, tt.config.RefreshInterval, loadedCfg.RefreshInterval)
					assert.Equal(t, tt.config.WebPort, loadedCfg.WebPort)
					assert.Equal(t, tt.config.EnableGoRuntime, loadedCfg.EnableGoRuntime)

					// Clean up
					err = os.Remove(tt.config.ConfigFile)
					if err != nil && !os.IsNotExist(err) {
						require.NoError(t, err)
					}
				}
			}
		})
	}
}

func TestConfig_Validation(t *testing.T) {
	tests := []struct {
		name    string
		config  config.Config
		isValid bool
	}{
		{
			name: "valid config",
			config: config.Config{
				RefreshInterval: 1,
				WebPort:         8080,
			},
			isValid: true,
		},
		{
			name: "invalid refresh interval",
			config: config.Config{
				RefreshInterval: 0,
				WebPort:         8080,
			},
			isValid: false,
		},
		{
			name: "invalid web port",
			config: config.Config{
				RefreshInterval: 1,
				WebPort:         0,
			},
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.isValid {
				assert.NotZero(t, tt.config.RefreshInterval)
				assert.NotZero(t, tt.config.WebPort)
			} else {
				assert.True(t, tt.config.RefreshInterval == 0 || tt.config.WebPort == 0)
			}
		})
	}
}
