package metrics

import (
	m "github.com/j-raghavan/godash/internal/metrics"
	"testing"
	"time"
)

// SimpleMockCollector is a minimal mock implementation of the Collector interface
type SimpleMockCollector struct {
	CollectCalled  bool
	StartCalled    bool
	StopCalled     bool
	MetricToReturn *m.Metric
	ErrorToReturn  error
}

func NewSimpleMockCollector() *SimpleMockCollector {
	return &SimpleMockCollector{
		MetricToReturn: &m.Metric{
			Timestamp: time.Now(),
			CPU:       []float64{1.0, 2.0},
			Memory: m.MemoryStat{
				Total:          1000,
				Free:           500,
				Used:           500,
				UsedPercentage: 50.0,
			},
			Disk: []m.DiskStat{
				{
					Path:           "/",
					Total:          1000,
					Used:           500,
					Free:           500,
					UsedPercentage: 50.0,
				},
			},
			Network: []m.NetworkStat{
				{
					Interface: "eth0",
					RxBytes:   1000,
					TxBytes:   500,
					RxPackets: 100,
					TxPackets: 50,
				},
			},
			GoRuntime: m.GoRuntimeStat{
				NumGoroutine: 10,
				MemAlloc:     1000,
				MemSys:       2000,
				NumGC:        5,
				PauseTotalNs: 1000,
			},
		},
	}
}

func (mock *SimpleMockCollector) Collect() (*m.Metric, error) {
	mock.CollectCalled = true
	return mock.MetricToReturn, mock.ErrorToReturn
}

func (mock *SimpleMockCollector) Start(interval time.Duration, metricsChan chan<- m.Metric) {
	mock.StartCalled = true
	if mock.MetricToReturn != nil {
		metricsChan <- *mock.MetricToReturn
	}
}

func (mock *SimpleMockCollector) Stop() {
	mock.StopCalled = true
}

// TestSimpleMockCollector tests that the SimpleMockCollector implements the Collector interface
func TestSimpleMockCollector(t *testing.T) {
	var collector m.Collector = NewSimpleMockCollector()

	mockCollector := collector.(*SimpleMockCollector)

	// Test Collect
	metric, err := mockCollector.Collect()
	if !mockCollector.CollectCalled {
		t.Error("Collect method was not called")
	}
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if metric == nil {
		t.Error("Expected non-nil metric")
	}

	// Test Start
	metricsChan := make(chan m.Metric, 1)
	mockCollector.Start(100*time.Millisecond, metricsChan)
	if !mockCollector.StartCalled {
		t.Error("Start method was not called")
	}
	select {
	case <-metricsChan:
		// Success
	default:
		t.Error("Expected to receive a metric")
	}

	// Test Stop
	mockCollector.Stop()
	if !mockCollector.StopCalled {
		t.Error("Stop method was not called")
	}
}
