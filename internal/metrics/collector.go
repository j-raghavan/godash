package metrics

import (
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/net"
)

// Metric represents a snapthot of system metrics at a pont in time.
type Metric struct {
	Timestamp time.Time
	CPU       []float64
	Memory    MemoryStat
	Disk      []DiskStat
	Network   []NetworkStat
	GoRuntime GoRuntimeStat
}

// MemoryStat represents the memory usage of the system.
type MemoryStat struct {
	Total          uint64
	Free           uint64
	Used           uint64
	UsedPercentage float64
	// Available uint64
	// Buffers uint64
	// Cached uint64
	// SwapTotal uint64
	// SwapFree uint64
	// SwapUsed uint64
}

// DiskStat represents the disk usage of the system.
type DiskStat struct {
	Path           string
	Total          uint64
	Used           uint64
	Free           uint64
	UsedPercentage float64
}

// NetworkStat represents the network usage of the system.
type NetworkStat struct {
	Interface string
	RxBytes   uint64
	TxBytes   uint64
	RxPackets uint64
	TxPackets uint64
}

// GoRuntimeStat represents the Go runtime statistics.
type GoRuntimeStat struct {
	NumGoroutine int
	MemAlloc     uint64
	MemSys       uint64
	NumGC        uint32
	PauseTotalNs uint64
}

// Collector interface defines methods to collect system metrics.
type Collector interface {
	Collect() (*Metric, error)
	Start(interval time.Duration,
		metricsChan chan<- Metric)
	Stop()
}

// SystemCollector implements the Collector interface
type SystemCollector struct {
	stopChan chan struct{}
	running  bool
	// Store previous network stats to calculate rates
	prevNetStats map[string]net.IOCountersStat
	prevTime     time.Time
}

// NewSystemCollector creates a new SystemCollector
func NewSystemCollector() *SystemCollector {
	return &SystemCollector{
		stopChan:     make(chan struct{}),
		prevNetStats: make(map[string]net.IOCountersStat),
		prevTime:     time.Now(),
	}
}

// Collect returns the current system metrics
func (c *SystemCollector) Collect() (*Metric, error) {
	metric := &Metric{
		Timestamp: time.Now(),
	}
	// Collect CPU metrics
	cpuPercent, err := collectCPUMetrics()
	if err != nil {
		return nil, err
	}
	metric.CPU = cpuPercent

	// Collect Memory metrics
	memoryStat, err := collectMemoryMetrics()
	if err != nil {
		return nil, err
	}
	metric.Memory = memoryStat

	// Collect Disk metrics
	diskStats, err := collectDiskMetrics()
	if err != nil {
		return nil, err
	}
	metric.Disk = diskStats

	// Collect Network metrics
	networkStats, err := c.collectNetworkMetrics()
	if err != nil {
		return nil, err
	}
	metric.Network = networkStats

	// Collect Go runtime metrics
	metric.GoRuntime = collectGoRuntimeMetrics()
	return metric, nil
}

// Start begins periodic collection of system metrics
func (c *SystemCollector) Start(interval time.Duration,
	metricsChan chan<- Metric,
) {
	if c.running {
		return
	}
	if interval <= 0 {
		interval = 100 * time.Millisecond
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

// Stop stops the periodic collection of system metrics
func (c *SystemCollector) Stop() {
	if !c.running {
		return
	}
	c.stopChan <- struct{}{}
	c.running = false
	close(c.stopChan)
}

// collectCPUMetrics collects CPU usage metrics
func collectCPUMetrics() ([]float64, error) {
	cpuPercent := make([]float64, runtime.NumCPU())
	for i := 0; i < runtime.NumCPU(); i++ {
		cpuPercent[i] = float64(i) // Placeholder for actual CPU usage
	}
	return cpuPercent, nil
}

// collectMemoryMetrics collects memory usage metrics
func collectMemoryMetrics() (MemoryStat, error) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	memoryStat := MemoryStat{
		Total:          memStats.Sys,
		Free:           memStats.HeapIdle,
		Used:           memStats.HeapAlloc,
		UsedPercentage: float64(memStats.HeapAlloc) / float64(memStats.Sys) * 100,
	}
	return memoryStat, nil
}

// collectDiskMetrics collects disk usage metrics
func collectDiskMetrics() ([]DiskStat, error) {
	partitions, err := disk.Partitions(false)
	if err != nil {
		return nil, err
	}

	var diskStats []DiskStat
	for _, partition := range partitions {
		usage, err := disk.Usage(partition.Mountpoint)
		if err != nil {
			continue
		}

		diskStats = append(diskStats, DiskStat{
			Path:           partition.Mountpoint,
			Total:          usage.Total,
			Used:           usage.Used,
			Free:           usage.Free,
			UsedPercentage: usage.UsedPercent,
		})
	}

	return diskStats, nil
}

// collectNetworkMetrics collects network usage metrics
func (c *SystemCollector) collectNetworkMetrics() ([]NetworkStat, error) {
	counters, err := net.IOCounters(true)
	if err != nil {
		return nil, err
	}

	currentTime := time.Now()
	var networkStats []NetworkStat

	for _, counter := range counters {
		netStat := NetworkStat{
			Interface: counter.Name,
			RxBytes:   counter.BytesRecv,
			TxBytes:   counter.BytesSent,
			RxPackets: counter.PacketsRecv,
			TxPackets: counter.PacketsSent,
		}

		// Calculate rates if we have previous measurements
		if prev, ok := c.prevNetStats[counter.Name]; ok {
			timeDiff := currentTime.Sub(c.prevTime).Seconds()
			if timeDiff > 0 {
				netStat.RxBytes = uint64(float64(counter.BytesRecv-prev.BytesRecv) / timeDiff)
				netStat.TxBytes = uint64(float64(counter.BytesSent-prev.BytesSent) / timeDiff)
				netStat.RxPackets = uint64(float64(counter.PacketsRecv-prev.PacketsRecv) / timeDiff)
				netStat.TxPackets = uint64(float64(counter.PacketsSent-prev.PacketsSent) / timeDiff)
			}
		}

		networkStats = append(networkStats, netStat)
		c.prevNetStats[counter.Name] = counter
	}

	c.prevTime = currentTime
	return networkStats, nil
}

// collectGoRuntimeMetrics collects Go runtime metrics
func collectGoRuntimeMetrics() GoRuntimeStat {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	goRuntimeStat := GoRuntimeStat{
		NumGoroutine: runtime.NumGoroutine(),
		MemAlloc:     memStats.Alloc,
		MemSys:       memStats.Sys,
		NumGC:        memStats.NumGC,
		PauseTotalNs: memStats.PauseTotalNs,
	}
	return goRuntimeStat
}
