package metrics

import (
	m "github.com/j-raghavan/godash/internal/metrics"
	"runtime"
	"testing"
	"time"
)

// TestStartStopWithTableDrivenTests uses table-driven tests to verify Start/Stop behavior
func TestStartStopWithTableDrivenTests(t *testing.T) {
	tests := []struct {
		name            string
		collectionCount int
		interval        time.Duration
		waitTime        time.Duration
		expectedMin     int
		expectedMax     int
	}{
		{
			name:        "Short interval multiple collections",
			interval:    50 * time.Millisecond,
			waitTime:    175 * time.Millisecond,
			expectedMin: 3,
			expectedMax: 4,
		},
		{
			name:        "Longer interval fewer collections",
			interval:    100 * time.Millisecond,
			waitTime:    250 * time.Millisecond,
			expectedMin: 2,
			expectedMax: 3,
		},
		{
			name:        "Very short wait time",
			interval:    50 * time.Millisecond,
			waitTime:    60 * time.Millisecond,
			expectedMin: 1,
			expectedMax: 2,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			collector := m.NewSystemCollector()
			metricsChan := make(chan m.Metric, 10)

			// Start the collector
			collector.Start(tc.interval, metricsChan)

			// Wait for specified time
			time.Sleep(tc.waitTime)

			// Stop the collector
			collector.Stop()

			// Check collection count is within expected range
			count := len(metricsChan)
			if count < tc.expectedMin || count > tc.expectedMax {
				t.Errorf("Expected %d-%d metrics, got %d",
					tc.expectedMin, tc.expectedMax, count)
			}
		})
	}
}

// TestCollectFunctionality tests the Collect method with different scenarios
func TestCollectFunctionality(t *testing.T) {
	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "Basic collection succeeds",
			testFunc: func(t *testing.T) {
				collector := m.NewSystemCollector()
				metric, err := collector.Collect()
				if err != nil {
					t.Fatalf("Expected no error, got %v", err)
				}
				if metric == nil {
					t.Fatal("Expected non-nil metric")
				}
				if metric.Timestamp.IsZero() {
					t.Error("Expected non-zero timestamp")
				}
			},
		},
		{
			name: "Collected CPU metrics length equals NumCPU",
			testFunc: func(t *testing.T) {
				collector := m.NewSystemCollector()
				metric, _ := collector.Collect()

				if len(metric.CPU) != runtime.NumCPU() {
					t.Errorf("Expected %d CPU metrics, got %d",
						runtime.NumCPU(), len(metric.CPU))
				}
			},
		},
		{
			name: "Memory metrics have reasonable values",
			testFunc: func(t *testing.T) {
				collector := m.NewSystemCollector()
				metric, _ := collector.Collect()

				if metric.Memory.Total <= 0 {
					t.Error("Expected positive Total memory")
				}

				if metric.Memory.Used <= 0 {
					t.Error("Expected positive Used memory")
				}

				if metric.Memory.UsedPercentage < 0 || metric.Memory.UsedPercentage > 100 {
					t.Errorf("Expected memory usage between 0-100%%, got %.2f%%",
						metric.Memory.UsedPercentage)
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, tc.testFunc)
	}
}

// TestCollectorBehavior tests various behaviors of the collector
func TestCollectorBehavior(t *testing.T) {
	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "Second Start call doesn't affect collection",
			testFunc: func(t *testing.T) {
				collector := m.NewSystemCollector()
				metricsChan := make(chan m.Metric, 5)

				collector.Start(100*time.Millisecond, metricsChan)
				goroutinesBefore := runtime.NumGoroutine()

				collector.Start(100*time.Millisecond, metricsChan) // Second call should do nothing
				goroutinesAfter := runtime.NumGoroutine()

				if goroutinesAfter > goroutinesBefore {
					t.Errorf("Expected no additional goroutines, before: %d, after: %d",
						goroutinesBefore, goroutinesAfter)
				}

				collector.Stop()
			},
		},
		{
			name: "Stop can be called multiple times without errors",
			testFunc: func(t *testing.T) {
				collector := m.NewSystemCollector()
				metricsChan := make(chan m.Metric, 5)

				collector.Start(100*time.Millisecond, metricsChan)
				time.Sleep(50 * time.Millisecond)

				// These should not panic
				collector.Stop()
				collector.Stop() // Second stop call
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, tc.testFunc)
	}
}
