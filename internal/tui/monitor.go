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

	// Start metrics collection
	ui.collector.Start(refreshInterval, ui.metricsChan)

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
		case metric := <-ui.metricsChan:
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

		_, _ = fmt.Fprintf(ui.metricsView, "[green]GoDash - Metrics at %s[white]\n", metric.Timestamp.Format("15:04:05"))
		_, _ = fmt.Fprintf(ui.metricsView, "---------------------------------------\n\n")

		// CPU
		if len(metric.CPU) > 0 {
			_, _ = fmt.Fprintf(ui.metricsView, "[blue]CPU:[white] %.1f%%\n", metric.CPU[0])

			if len(metric.CPU) > 1 {
				_, _ = fmt.Fprintf(ui.metricsView, "  [blue]Per Core:[white] ")
				for i, cpu := range metric.CPU[1:] {
					_, _ = fmt.Fprintf(ui.metricsView, "core%d: %.1f%% ", i, cpu)
				}
				_, _ = fmt.Fprintf(ui.metricsView, "\n")
			}
		}

		// Memory
		_, _ = fmt.Fprintf(ui.metricsView, "\n[blue]Memory:[white] %.2f GB / %.2f GB (%.1f%%)\n",
			float64(metric.Memory.Used)/(1024*1024*1024),
			float64(metric.Memory.Total)/(1024*1024*1024),
			metric.Memory.UsedPercentage)

		// Disk
		if len(metric.Disk) > 0 {
			_, _ = fmt.Fprintf(ui.metricsView, "\n[blue]Disk:[white]\n")
			for _, disk := range metric.Disk {
				_, _ = fmt.Fprintf(ui.metricsView, "  %s: %.1f%% used (%.2f GB / %.2f GB)\n",
					disk.Path,
					disk.UsedPercentage,
					float64(disk.Used)/(1024*1024*1024),
					float64(disk.Total)/(1024*1024*1024))
			}
		}

		// Network
		if len(metric.Network) > 0 {
			_, _ = fmt.Fprintf(ui.metricsView, "\n[blue]Network:[white]\n")
			for _, net := range metric.Network {
				_, _ = fmt.Fprintf(ui.metricsView, "  %s: \u2193 %s \u2191 %s\n",
					net.Interface,
					formatBytes(net.RxBytes),
					formatBytes(net.TxBytes))
			}
		}

		// Go Runtime
		if ui.showGoRuntime {
			_, _ = fmt.Fprintf(ui.metricsView, "\n[blue]Go Runtime:[white]\n")
			_, _ = fmt.Fprintf(ui.metricsView, "  Goroutines: %d\n", metric.GoRuntime.NumGoroutine)
			_, _ = fmt.Fprintf(ui.metricsView, "  Memory: %.2f MB allocated, %.2f MB system\n",
				float64(metric.GoRuntime.MemAlloc)/(1024*1024),
				float64(metric.GoRuntime.MemSys)/(1024*1024))
			_, _ = fmt.Fprintf(ui.metricsView, "  GC: %d collections, %.2f ms total pause\n",
				metric.GoRuntime.NumGC,
				float64(metric.GoRuntime.PauseTotalNs)/1000000)
		}
	})
}
