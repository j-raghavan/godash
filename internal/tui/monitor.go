package tui

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/j-raghavan/godash/internal/metrics"
)

// UI represents the terminal user interface
type UI struct {
	app                 *tview.Application
	grid                *tview.Grid
	cpuView             *tview.TextView
	memoryView          *tview.TextView
	diskView            *tview.TextView
	networkView         *tview.TextView
	statusBar           *tview.TextView
	collector           metrics.Collector
	metricsChan         chan metrics.Metric
	showGoRuntime       bool
	ctx                 context.Context
	cancel              context.CancelFunc
	lastNetworkUpdate   time.Time
	lastMemoryUpdate    time.Time
	topInterfaces       []string // Store top 3 interfaces
	lastInterfaceUpdate time.Time
}

// NewUI initializes a new UI instance
func NewUI(collector metrics.Collector, showGoRuntime bool) *UI {
	ctx, cancel := context.WithCancel(context.Background())

	// Create text views with proper type
	cpuView := tview.NewTextView()
	cpuView.SetDynamicColors(true).
		SetBorder(true).
		SetTitle("CPU Usage")

	memoryView := tview.NewTextView()
	memoryView.SetDynamicColors(true).
		SetBorder(true).
		SetTitle("Memory Usage (Updates every 5s)")

	diskView := tview.NewTextView()
	diskView.SetDynamicColors(true).
		SetBorder(true).
		SetTitle("Disk Usage")

	networkView := tview.NewTextView()
	networkView.SetDynamicColors(true).
		SetBorder(true).
		SetTitle("Network I/O (Updates every 5s)")

	statusBar := tview.NewTextView()
	statusBar.SetDynamicColors(true)

	// Create grid layout
	grid := tview.NewGrid().
		SetRows(10, 10, 10, 1). // Three main rows of height 10, and 1 row for status
		SetColumns(-1).         // Full width
		SetBorders(false)

	// Add items to grid
	grid.AddItem(cpuView, 0, 0, 1, 1, 0, 0, false).
		AddItem(tview.NewFlex().
			AddItem(diskView, 0, 1, false).
			AddItem(memoryView, 0, 1, false),
			1, 0, 1, 1, 0, 0, false).
		AddItem(networkView, 2, 0, 1, 1, 0, 0, false).
		AddItem(statusBar, 3, 0, 1, 1, 0, 0, false)

	return &UI{
		app:                 tview.NewApplication(),
		grid:                grid,
		cpuView:             cpuView,
		memoryView:          memoryView,
		diskView:            diskView,
		networkView:         networkView,
		statusBar:           statusBar,
		collector:           collector,
		metricsChan:         make(chan metrics.Metric, 10),
		showGoRuntime:       showGoRuntime,
		ctx:                 ctx,
		cancel:              cancel,
		lastNetworkUpdate:   time.Now().Add(-5 * time.Second),  // Force first update
		lastMemoryUpdate:    time.Now().Add(-5 * time.Second),  // Force first update
		lastInterfaceUpdate: time.Now().Add(-30 * time.Second), // Force first update
		topInterfaces:       make([]string, 0),
	}
}

// Start initializes and starts the UI
func (ui *UI) Start(refreshInterval time.Duration) error {
	// Set up status bar
	ui.statusBar.SetText("[yellow]Press 'q' to quit, 'g' to toggle Go runtime stats[white]")

	// Set up key handlers
	ui.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'q':
			ui.cancel()
			ui.app.Stop()
			return nil
		case 'g':
			ui.showGoRuntime = !ui.showGoRuntime
			return nil
		}
		return event
	})

	// Start metrics collection with a fixed 100ms interval for smoother updates
	ui.collector.Start(100*time.Millisecond, ui.metricsChan)

	// Start the UI update routine
	go ui.update()

	// Run the application
	return ui.app.SetRoot(ui.grid, true).Run()
}

// Stop shuts down the UI and metrics collection
func (ui *UI) Stop() {
	ui.cancel()
	ui.collector.Stop()
	close(ui.metricsChan)
}

// update refreshes the UI with the latest metrics
func (ui *UI) update() {
	for {
		select {
		case metric, ok := <-ui.metricsChan:
			if !ok {
				return
			}
			ui.renderMetrics(metric)
		case <-ui.ctx.Done():
			return
		}
	}
}

// renderMetrics updates the UI with the provided metrics
func (ui *UI) renderMetrics(metric metrics.Metric) {
	ui.app.QueueUpdateDraw(func() {
		// Update CPU View
		ui.cpuView.Clear()
		if len(metric.CPU) > 0 {
			_, _ = fmt.Fprintf(ui.cpuView, "Overall: %.1f%%\n\n", metric.CPU[0])

			// Display CPU cores in 4 columns
			if len(metric.CPU) > 1 {
				numCores := len(metric.CPU[1:])
				cols := 4
				rows := (numCores + cols - 1) / cols

				for row := 0; row < rows; row++ {
					for col := 0; col < cols; col++ {
						coreIndex := row*cols + col
						if coreIndex < numCores {
							cpu := metric.CPU[coreIndex+1]
							bar := createProgressBar(cpu, 12)
							_, _ = fmt.Fprintf(ui.cpuView, "Core %2d: [%s] %5.1f%%   ",
								coreIndex, bar, cpu)
						}
					}
					_, _ = fmt.Fprintf(ui.cpuView, "\n")
				}
			}
		}

		// Update Memory View every 5 seconds
		if time.Since(ui.lastMemoryUpdate) >= 5*time.Second {
			ui.memoryView.Clear()
			memBar := createProgressBar(metric.Memory.UsedPercentage, 20)
			_, _ = fmt.Fprintf(ui.memoryView, "[%s] %.1f%%\n", memBar, metric.Memory.UsedPercentage)
			_, _ = fmt.Fprintf(ui.memoryView, "Used: %s\nTotal: %s\n",
				formatBytes(metric.Memory.Used),
				formatBytes(metric.Memory.Total))
			if ui.showGoRuntime {
				_, _ = fmt.Fprintf(ui.memoryView, "\nGo Runtime:\n")
				_, _ = fmt.Fprintf(ui.memoryView, "Goroutines: %d\n", metric.GoRuntime.NumGoroutine)
				_, _ = fmt.Fprintf(ui.memoryView, "Alloc: %s\n", formatBytes(metric.GoRuntime.MemAlloc))
			}
			ui.lastMemoryUpdate = time.Now()
		}

		// Update Disk View
		ui.diskView.Clear()
		for _, disk := range metric.Disk {
			bar := createProgressBar(disk.UsedPercentage, 20)
			_, _ = fmt.Fprintf(ui.diskView, "%s\n[%s] %.1f%%\n",
				disk.Path, bar, disk.UsedPercentage)
			_, _ = fmt.Fprintf(ui.diskView, "Used: %s / %s\n\n",
				formatBytes(disk.Used),
				formatBytes(disk.Total))
		}

		// Update top interfaces list every 30 seconds
		if time.Since(ui.lastInterfaceUpdate) >= 30*time.Second {
			// Create a slice of interfaces with their total traffic
			type interfaceStats struct {
				name       string
				totalBytes uint64
			}
			var netStats []interfaceStats
			for _, net := range metric.Network {
				totalBytes := net.RxBytes + net.TxBytes
				netStats = append(netStats, interfaceStats{
					name:       net.Interface,
					totalBytes: totalBytes,
				})
			}

			// Sort interfaces by total traffic (descending)
			sort.Slice(netStats, func(i, j int) bool {
				return netStats[i].totalBytes > netStats[j].totalBytes
			})

			// Update top 3 interfaces
			ui.topInterfaces = make([]string, 0)
			for i := 0; i < len(netStats) && i < 3; i++ {
				ui.topInterfaces = append(ui.topInterfaces, netStats[i].name)
			}
			ui.lastInterfaceUpdate = time.Now()
		}

		// Update Network View every 5 seconds
		if time.Since(ui.lastNetworkUpdate) >= 5*time.Second {
			ui.networkView.Clear()

			// Create a map for quick lookup
			netMap := make(map[string]metrics.NetworkStat)
			for _, net := range metric.Network {
				netMap[net.Interface] = net
			}

			if len(ui.topInterfaces) > 0 {
				colWidth := 30 // Fixed width for each column

				// Print headers
				_, _ = fmt.Fprintf(ui.networkView, "Top 3 Interfaces by Traffic:\n\n")
				for _, iface := range ui.topInterfaces {
					paddingLen := colWidth - len(iface)
					if paddingLen < 0 {
						paddingLen = 0
					}
					padding := strings.Repeat(" ", paddingLen)
					_, _ = fmt.Fprintf(ui.networkView, "%.*s%s", colWidth, iface, padding)
				}
				_, _ = fmt.Fprintf(ui.networkView, "\n")

				// Print RX stats
				for _, iface := range ui.topInterfaces {
					if net, ok := netMap[iface]; ok {
						stats := fmt.Sprintf("↓ RX: %s/s (%d pkts/s)",
							formatBytes(net.RxBytes),
							net.RxPackets)
						paddingLen := colWidth - len(stats)
						if paddingLen < 0 {
							paddingLen = 0
						}
						padding := strings.Repeat(" ", paddingLen)
						_, _ = fmt.Fprintf(ui.networkView, "%.*s%s", colWidth, stats, padding)
					}
				}
				_, _ = fmt.Fprintf(ui.networkView, "\n")

				// Print TX stats
				for _, iface := range ui.topInterfaces {
					if net, ok := netMap[iface]; ok {
						stats := fmt.Sprintf("↑ TX: %s/s (%d pkts/s)",
							formatBytes(net.TxBytes),
							net.TxPackets)
						paddingLen := colWidth - len(stats)
						if paddingLen < 0 {
							paddingLen = 0
						}
						padding := strings.Repeat(" ", paddingLen)
						_, _ = fmt.Fprintf(ui.networkView, "%.*s%s", colWidth, stats, padding)
					}
				}

				// Print total traffic for each interface
				_, _ = fmt.Fprintf(ui.networkView, "\n")
				for _, iface := range ui.topInterfaces {
					if net, ok := netMap[iface]; ok {
						totalBytes := net.RxBytes + net.TxBytes
						stats := fmt.Sprintf("Total: %s/s",
							formatBytes(totalBytes))
						paddingLen := colWidth - len(stats)
						if paddingLen < 0 {
							paddingLen = 0
						}
						padding := strings.Repeat(" ", paddingLen)
						_, _ = fmt.Fprintf(ui.networkView, "%.*s%s", colWidth, stats, padding)
					}
				}
			}
			ui.lastNetworkUpdate = time.Now()
		}
	})
}

// createProgressBar creates a colored progress bar
func createProgressBar(percentage float64, width int) string {
	filled := int(percentage * float64(width) / 100)
	if filled > width {
		filled = width
	}
	empty := width - filled

	// Choose color based on percentage
	var color string
	switch {
	case percentage < 50:
		color = "green"
	case percentage < 80:
		color = "yellow"
	default:
		color = "red"
	}

	bar := ""
	for i := 0; i < filled; i++ {
		bar += "█"
	}
	for i := 0; i < empty; i++ {
		bar += "░"
	}
	return color + "]" + bar + "[white"
}

// formatBytes formats bytes to human readable format
func formatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}

// CPUView returns the CPU metrics view
func (ui *UI) CPUView() *tview.TextView {
	return ui.cpuView
}

// MemoryView returns the memory metrics view
func (ui *UI) MemoryView() *tview.TextView {
	return ui.memoryView
}

// DiskView returns the disk metrics view
func (ui *UI) DiskView() *tview.TextView {
	return ui.diskView
}

// NetworkView returns the network metrics view
func (ui *UI) NetworkView() *tview.TextView {
	return ui.networkView
}

// App returns the tview application
func (ui *UI) App() *tview.Application {
	return ui.app
}

// ShowGoRuntime returns whether Go runtime stats are shown
func (ui *UI) ShowGoRuntime() bool {
	return ui.showGoRuntime
}

// RenderMetrics renders the metrics in the UI
func (ui *UI) RenderMetrics(metric metrics.Metric) {
	ui.renderMetrics(metric)
}

// FormatBytes formats bytes into a human-readable string
func FormatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// CreateProgressBar creates a progress bar string
func CreateProgressBar(percent float64, width int) string {
	filled := int(percent * float64(width) / 100)
	if filled > width {
		filled = width
	}
	bar := strings.Repeat("█", filled)
	empty := strings.Repeat("░", width-filled)
	return fmt.Sprintf("[green]%s[white]%s", bar, empty)
}

// SetApp sets the tview application
func (ui *UI) SetApp(app *tview.Application) {
	ui.app = app
}

// ToggleGoRuntime toggles the display of Go runtime stats
func (ui *UI) ToggleGoRuntime() {
	ui.showGoRuntime = !ui.showGoRuntime
}
