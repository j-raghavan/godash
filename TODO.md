# GoDash - Roadmap & TODO

## 🧩 MVP (v0.1.0)
- [x] Collect CPU/mem/disk/net stats via `gopsutil`
- [x] Print stats in CLI
- [ ] WebSocket server
- [ ] HTML + JS dashboard (Chart.js)
- [ ] Serve via `http://localhost:8080`

## 🖥 CLI (tview-based)
- [ ] Use `tview` or `termui`
- [ ] Keybindings: `q` to quit, `g` to toggle Go stats

## 🌐 Web UI
- [ ] Live updates via WebSocket
- [ ] Add memory/cpu gauges
- [ ] Add goroutines + GC chart

##  📦 Docker Support (optional)
- [ ] Docker client integration
- [ ] Container resource panel

## 🚀 Future Ideas
- [ ] Prometheus export mode
- [ ] Dark/light mode toggle
- [ ] Plugin architecture
