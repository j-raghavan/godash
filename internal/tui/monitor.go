package tui

import (
	"context"
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/j-raghavan/godash/internal/metrics"
)

// UI represents the terminal user interface
type UI struct {
	app           *tview.Application
	metricsView   *tview.TextView
	statusBar     *tview.TextView
	collector     metrics.Collector
	metricsChan   chan metrics.Metric
	showGoRuntime bool
	ctx           context.Context
	cancel        context.CancelFunc
}

// NewUI initializes a new UI instance
func NewUI(collector metrics.Collector, showGoRuntime bool) *UI {
	ctx, cancel := context.WithCancel(context.Background())

	return &UI{
		app:           tview.NewApplication(),
		metricsView:   tview.NewTextView().SetDynamicColors(true),
		statusBar:     tview.NewTextView().SetDynamicColors(true),
		collector:     collector,
		metricsChan:   make(chan metrics.Metric, 10),
		showGoRuntime: showGoRuntime,
		ctx:           ctx,
		cancel:        cancel,
	}
}

// Start initializes and starts the UI
func (ui *UI) Start(refreshInterval time.Duration) error {
	// Set up the UI layout
	ui.metricsView.SetBorder(true).SetTitle(" GoDash Monitor ")
	ui.statusBar.SetText("[yellow]Press 'q' to quit, 'g' to toggle Go runtime stats[white]")

	// Create the layout
	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(ui.metricsView, 0, 1, false).
		AddItem(ui.statusBar, 1, 0, false)

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

	// Start metrics collection with a fixed 1000ms interval for smoother updates
	ui.collector.Start(1000*time.Millisecond, ui.metricsChan)

	// Start the UI update routine
	go ui.update()

	// Run the application
	return ui.app.SetRoot(flex, true).Run()
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

// formatBytes converts bytes to a human-readable format
func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// renderMetrics updates the UI with the provided metrics
func (ui *UI) renderMetrics(metric metrics.Metric) {
	ui.app.QueueUpdateDraw(func() {
		ui.metricsView.Clear()

		// Header
		_, _ = fmt.Fprintf(ui.metricsView, "[green]GoDash - Metrics at %s[white]\n", metric.Timestamp.Format("15:04:05"))
		_, _ = fmt.Fprintf(ui.metricsView, "---------------------------------------\n\n")

		// CPU Section
		_, _ = fmt.Fprintf(ui.metricsView, "[blue]CPU Usage[white]\n")
		if len(metric.CPU) > 0 {
			// Overall CPU usage
			_, _ = fmt.Fprintf(ui.metricsView, "  Overall: %.1f%%\n", metric.CPU[0])

			// Per-core usage with progress bars in three columns
			if len(metric.CPU) > 1 {
				_, _ = fmt.Fprintf(ui.metricsView, "  Per Core:\n")

				// Calculate number of rows needed (ceiling division)
				numCores := len(metric.CPU[1:])
				rows := (numCores + 2) / 3 // Round up division

				for row := 0; row < rows; row++ {
					// Print up to 3 cores per row
					for col := 0; col < 3; col++ {
						coreIndex := row*3 + col
						if coreIndex < numCores {
							cpu := metric.CPU[coreIndex+1]
							bar := createProgressBar(cpu, 15)
							_, _ = fmt.Fprintf(ui.metricsView, "    Core %2d: [%s] %5.1f%%  ",
								coreIndex, bar, cpu)
						}
					}
					_, _ = fmt.Fprintf(ui.metricsView, "\n")
				}
			}
		}

		// Memory Section
		_, _ = fmt.Fprintf(ui.metricsView, "\n[blue]Memory Usage[white]\n")
		memBar := createProgressBar(metric.Memory.UsedPercentage, 20)
		_, _ = fmt.Fprintf(ui.metricsView, "  [%s] %.1f%%\n", memBar, metric.Memory.UsedPercentage)
		_, _ = fmt.Fprintf(ui.metricsView, "  Used: %s / %s\n",
			formatBytes(metric.Memory.Used),
			formatBytes(metric.Memory.Total))

		// Go Runtime Section (if enabled)
		if ui.showGoRuntime {
			_, _ = fmt.Fprintf(ui.metricsView, "\n[blue]Go Runtime[white]\n")
			_, _ = fmt.Fprintf(ui.metricsView, "  Goroutines: %d\n", metric.GoRuntime.NumGoroutine)
			_, _ = fmt.Fprintf(ui.metricsView, "  Memory Alloc: %s\n", formatBytes(metric.GoRuntime.MemAlloc))
			_, _ = fmt.Fprintf(ui.metricsView, "  Memory Sys: %s\n", formatBytes(metric.GoRuntime.MemSys))
			_, _ = fmt.Fprintf(ui.metricsView, "  GC Count: %d\n", metric.GoRuntime.NumGC)
			_, _ = fmt.Fprintf(ui.metricsView, "  GC Pause: %s\n", time.Duration(metric.GoRuntime.PauseTotalNs))
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
