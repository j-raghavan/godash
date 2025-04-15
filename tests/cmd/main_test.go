package cmd_test

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/j-raghavan/godash/cmd/godash/core"
	"github.com/j-raghavan/godash/internal/config"
)

func TestRunMonitor(t *testing.T) {
	// Setup
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Call the function directly for testing
	testConfig := config.Config{
		RefreshInterval: 10,
		EnableGoRuntime: true,
	}
	core.RunMonitor(testConfig)

	// Reset stdout
	if err := w.Close(); err != nil {
		t.Fatalf("Failed to close writer: %v", err)
	}
	os.Stdout = old

	// Read the output
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("Failed to copy: %v", err)
	}
	output := buf.String()

	// Assertions
	assert.Contains(t, output, "refresh interval: 10s")
	assert.Contains(t, output, "Go runtime metrics enabled")
}

func TestRunServer(t *testing.T) {
	// Setup
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Call the function directly for testing
	testConfig := config.Config{
		RefreshInterval: 5,
		WebPort:         9090,
		EnableGoRuntime: true,
	}
	core.RunServer(testConfig)

	// Reset stdout
	if err := w.Close(); err != nil {
		t.Fatalf("Failed to close writer: %v", err)
	}
	os.Stdout = old

	// Read the output
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("Failed to copy: %v", err)
	}
	output := buf.String()

	// Assertions
	assert.Contains(t, output, "port 9090")
	assert.Contains(t, output, "refresh interval: 5s")
	assert.Contains(t, output, "Go runtime metrics enabled")
}

func TestShowVersion(t *testing.T) {
	version := core.ShowVersion()
	assert.Equal(t, "GoDash v0.1.0", version)
}
