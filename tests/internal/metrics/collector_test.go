package metrics

import (
	m "github.com/j-raghavan/godash/internal/metrics"
	"reflect"
	"runtime"
	"testing"
	"time"
)

// TestNewSystemCollector tests the creation of a new SystemCollector
func TestNewSystemCollector(t *testing.T) {
	collector := m.NewSystemCollector()
	if collector == nil {
		t.Fatal("Expected non-nil SystemCollector")
	}

	// Use type assertion to access unexported fields
	// This is a test-only pattern and might break if implementation changes
	// Consider adding exported accessor methods if needed for testing

	// Note: We can't directly access unexported fields like stopChan and running
	// from a separate package. If needed, consider adding accessor methods for testing
	// or use reflection (not recommended for regular code but can be useful in tests)
}

// TestCollect tests the Collect method
func TestCollect(t *testing.T) {
	collector := m.NewSystemCollector()

	metric, err := collector.Collect()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if metric == nil {
		t.Fatal("Expected non-nil Metric")
	}

	// Test timestamp is reasonable (within the last minute)
	now := time.Now()
	if metric.Timestamp.After(now) || metric.Timestamp.Before(now.Add(-time.Minute)) {
		t.Errorf("Expected timestamp to be within the last minute, got %v", metric.Timestamp)
	}

	// Test CPU metrics
	if len(metric.CPU) != runtime.NumCPU() {
		t.Errorf("Expected %d CPU metrics, got %d", runtime.NumCPU(), len(metric.CPU))
	}

	// Test memory metrics
	if metric.Memory.Total == 0 {
		t.Error("Expected non-zero Total memory")
	}

	if metric.Memory.UsedPercentage < 0 || metric.Memory.UsedPercentage > 100 {
		t.Errorf("Expected UsedPercentage between 0 and 100, got %f", metric.Memory.UsedPercentage)
	}

	// Test Go runtime metrics
	if metric.GoRuntime.NumGoroutine <= 0 {
		t.Errorf("Expected positive number of goroutines, got %d", metric.GoRuntime.NumGoroutine)
	}

	if metric.GoRuntime.MemAlloc <= 0 {
		t.Errorf("Expected positive memory allocation, got %d", metric.GoRuntime.MemAlloc)
	}
}

// TestStartStop tests the Start and Stop methods
func TestStartStop(t *testing.T) {
	collector := m.NewSystemCollector()
	metricsChan := make(chan m.Metric, 10)

	// Test Start
	collector.Start(100*time.Millisecond, metricsChan)

	// Note: Can't directly access collector.running because it's in a different package

	// Wait for at least one metric to be collected
	time.Sleep(150 * time.Millisecond)

	// Test Stop
	collector.Stop()

	// Check if at least one metric was collected
	select {
	case <-metricsChan:
		// Success - we received a metric
	default:
		t.Error("Expected to receive at least one metric")
	}
}

// TestMetricTypes tests the structure of the Metric types
func TestMetricTypes(t *testing.T) {
	// Test Metric struct
	metricType := reflect.TypeOf(m.Metric{})

	fields := []struct {
		name string
		kind reflect.Kind
	}{
		{"Timestamp", reflect.Struct}, // Note: This matches the typo in the original code
		{"CPU", reflect.Slice},
		{"Memory", reflect.Struct},
		{"Disk", reflect.Slice},
		{"Network", reflect.Slice},
		{"GoRuntime", reflect.Struct},
	}

	for _, field := range fields {
		f, exists := metricType.FieldByName(field.name)
		if !exists {
			t.Errorf("Metric struct missing field: %s", field.name)
			continue
		}

		if f.Type.Kind() != field.kind {
			t.Errorf("Field %s has wrong kind. Expected %v, got %v", field.name, field.kind, f.Type.Kind())
		}
	}
}

// TestCollectorInterface tests if SystemCollector properly implements the Collector interface
func TestCollectorInterface(t *testing.T) {
	var _ m.Collector = m.NewSystemCollector()

}
