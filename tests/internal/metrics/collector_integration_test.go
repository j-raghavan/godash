package metrics

import (
	m "github.com/j-raghavan/godash/internal/metrics"
	"testing"
	"time"
)

// TestErrorHandlingInSystemCollector tests error handling in the real SystemCollector
func TestErrorHandlingInSystemCollector(t *testing.T) {
	collector := m.NewSystemCollector()
	metric, err := collector.Collect()

	// Since we're using the real collector, we expect no errors and a valid metric
	if err != nil {
		t.Errorf("Expected no error from real collector, got %v", err)
	}

	if metric == nil {
		t.Error("Expected non-nil metric from real collector")
	}
}

// TestMultipleCollectors tests that multiple collectors can run simultaneously
func TestMultipleCollectors(t *testing.T) {
	collector1 := m.NewSystemCollector()
	collector2 := m.NewSystemCollector()

	metricsChan1 := make(chan m.Metric, 10)
	metricsChan2 := make(chan m.Metric, 10)

	// Start both collectors
	collector1.Start(100*time.Millisecond, metricsChan1)
	collector2.Start(100*time.Millisecond, metricsChan2)

	// Wait for metrics to be collected - increased wait time
	time.Sleep(350 * time.Millisecond)

	// Stop both collectors
	collector1.Stop()
	collector2.Stop()

	// Check that both channels received metrics
	if len(metricsChan1) == 0 {
		t.Error("Collector 1 didn't collect any metrics")
	}

	if len(metricsChan2) == 0 {
		t.Error("Collector 2 didn't collect any metrics")
	}
}

// TestCollectionWithZeroInterval tests collector behavior with a very small interval
func TestCollectionWithZeroInterval(t *testing.T) {
	collector := m.NewSystemCollector()
	metricsChan := make(chan m.Metric, 100)

	// Start with very small interval instead of zero
	// Using 1 millisecond as the smallest reasonable interval
	collector.Start(1*time.Millisecond, metricsChan)

	// Wait briefly
	time.Sleep(10 * time.Millisecond)

	// Stop collection
	collector.Stop()

	// We might get several metrics with a very small interval
	count := len(metricsChan)
	t.Logf("Collected %d metrics with 1ms interval", count)

	// Just verify we got at least one metric
	if count < 1 {
		t.Error("Expected at least one metric with small interval")
	}
}
