package metrics

import (
	m "github.com/j-raghavan/godash/internal/metrics"
	"time"
)

// MockMetricProvider allows for controlled metric results in tests
type MockMetricProvider struct {
	CPUMetricsFunc       func() ([]float64, error)
	MemoryMetricsFunc    func() (m.MemoryStat, error)
	DiskMetricsFunc      func() ([]m.DiskStat, error)
	NetworkMetricsFunc   func() ([]m.NetworkStat, error)
	GoRuntimeMetricsFunc func() m.GoRuntimeStat
}

// NewMockMetricProvider creates a new MockMetricProvider with default functions
func NewMockMetricProvider() *MockMetricProvider {
	return &MockMetricProvider{
		CPUMetricsFunc: func() ([]float64, error) {
			return []float64{1.0, 2.0, 3.0, 4.0}, nil
		},
		MemoryMetricsFunc: func() (m.MemoryStat, error) {
			return m.MemoryStat{
				Total:          16 * 1024 * 1024 * 1024,
				Free:           8 * 1024 * 1024 * 1024,
				Used:           8 * 1024 * 1024 * 1024,
				UsedPercentage: 50.0,
			}, nil
		},
		DiskMetricsFunc: func() ([]m.DiskStat, error) {
			return []m.DiskStat{
				{
					Path:           "/",
					Total:          500 * 1024 * 1024 * 1024,
					Used:           250 * 1024 * 1024 * 1024,
					Free:           250 * 1024 * 1024 * 1024,
					UsedPercentage: 50.0,
				},
			}, nil
		},
		NetworkMetricsFunc: func() ([]m.NetworkStat, error) {
			return []m.NetworkStat{
				{
					Interface: "eth0",
					RxBytes:   1024 * 1024,
					TxBytes:   512 * 1024,
					RxPackets: 1000,
					TxPackets: 500,
				},
			}, nil
		},
		GoRuntimeMetricsFunc: func() m.GoRuntimeStat {
			return m.GoRuntimeStat{
				NumGoroutine: 10,
				MemAlloc:     1024 * 1024,
				MemSys:       2048 * 1024,
				NumGC:        5,
				PauseTotalNs: 1000000,
			}
		},
	}
}

// MockCollector implements the Collector interface for controlled testing
type MockCollector struct {
	MetricProvider *MockMetricProvider
	stopChan       chan struct{}
	running        bool
}

// NewMockCollector creates a new MockCollector with the provided MetricProvider
func NewMockCollector(provider *MockMetricProvider) *MockCollector {
	if provider == nil {
		provider = NewMockMetricProvider()
	}

	return &MockCollector{
		MetricProvider: provider,
		stopChan:       make(chan struct{}),
		running:        false,
	}
}

// Collect returns mock system metrics
func (c *MockCollector) Collect() (*m.Metric, error) {
	metric := &m.Metric{
		Timesteamp: time.Now(),
	}

	// Collect CPU metrics
	cpuPercent, err := c.MetricProvider.CPUMetricsFunc()
	if err != nil {
		return nil, err
	}
	metric.CPU = cpuPercent

	// Collect Memory metrics
	memoryStat, err := c.MetricProvider.MemoryMetricsFunc()
	if err != nil {
		return nil, err
	}
	metric.Memory = memoryStat

	// Collect Disk metrics
	diskStats, err := c.MetricProvider.DiskMetricsFunc()
	if err != nil {
		return nil, err
	}
	metric.Disk = diskStats

	// Collect Network metrics
	networkStats, err := c.MetricProvider.NetworkMetricsFunc()
	if err != nil {
		return nil, err
	}
	metric.Network = networkStats

	// Collect Go runtime metrics
	metric.GoRuntime = c.MetricProvider.GoRuntimeMetricsFunc()

	return metric, nil
}

// Start begins periodic collection of mock system metrics
func (c *MockCollector) Start(interval time.Duration, metricsChan chan<- m.Metric) {
	if c.running {
		return
	}
	c.running = true

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				metric, err := c.Collect()
				if err == nil && metric != nil {
					metricsChan <- *metric
				}
			case <-c.stopChan:
				return
			}
		}
	}()
}

// Stop stops the periodic collection of mock system metrics
func (c *MockCollector) Stop() {
	if !c.running {
		return
	}
	c.stopChan <- struct{}{}
	c.running = false
}
