# 🧾 Software Requirements Specification (SRS) – GoDash

---

## 📌 1. Project Overview

**GoDash** is a self-contained, cross-platform system monitoring tool written in Go. It provides real-time system resource metrics via both a Command-Line Interface (CLI) and a lightweight local web dashboard.

This tool is aimed at developers, DevOps engineers, and homelab enthusiasts who need:
- A portable and install-free performance monitor
- A single-binary alternative to `htop`, `btop`, or web-based dashboards like Netdata
- A simple way to introspect **Go runtime stats**, system load, and optionally container stats

---

## 🎯 2. Goals and Non-Goals

### ✅ Goals
- Display **real-time system resource metrics**
- Provide a **CLI mode** for terminal monitoring
- Provide a **local web UI** via HTTP/WebSocket
- Use **Go's standard library and minimal dependencies**
- Cross-platform (Linux/macOS/Windows)
- Extensible and modular code structure

### ❌ Non-Goals
- Persistent storage of metrics
- Cloud-hosted dashboards
- Remote multi-node system monitoring

---

## 🛠 3. Features & Functional Requirements

### 🧩 3.1 Core Components

#### 3.1.1 Metrics Collector
- Collect metrics every `X` seconds (configurable):
    - CPU usage (per-core + total)
    - Memory usage (used/free)
    - Disk usage
    - Network I/O (per interface)
    - Load average (if supported)
    - Number of processes
- Collect Go runtime metrics:
    - Goroutines
    - GC pause time
    - Memory allocations

##### Interfaces:
```go
type Metric struct {
    Timestamp time.Time
    CPU       []float64
    Memory    MemoryStat
    Disk      []DiskStat
    Network   []NetStat
    GoRuntime GoRuntimeStat
}
```

---

#### 3.1.2 CLI Interface
- Interactive CLI (`godash monitor`):
    - Auto-refreshes every second
    - Color-coded TUI using `tview` or fallback plain-text mode
- Basic controls:
    - Press `q` to quit
    - Toggle Go runtime stats

##### Example CLI Output:
```
GoDash - v0.1
---------------------------------------
CPU: 37.5% (core0: 30.1%, core1: 45.0%)
Memory: 2.1 GB / 8 GB
Disk: 14% used (/)
Network: ↓ 123 KB/s ↑ 45 KB/s
Goroutines: 17   GC Pause: 4ms
---------------------------------------
```

---

#### 3.1.3 Web Dashboard (HTTP + WebSocket)
- Served at `http://localhost:8080` by default
- Static HTML + JavaScript frontend
- Real-time updates via WebSocket (`/ws`)
- Components:
    - CPU chart
    - Memory gauge
    - Network line chart
    - Optional Go runtime pane

##### REST API:
| Endpoint     | Method | Description                    |
|--------------|--------|--------------------------------|
| `/metrics`   | GET    | JSON snapshot of latest stats  |
| `/ws`        | WS     | Stream real-time metrics       |
| `/healthz`   | GET    | Simple healthcheck endpoint    |

---

#### 3.1.4 Configuration
- Support environment variables and config file (YAML):
```yaml
refresh_interval: 2
web_port: 8080
enable_go_runtime: true
```

- CLI flags override config:
```bash
godash monitor --interval 1 --go-runtime
```

---

#### 3.1.5 Optional: Docker Stats
- Collect per-container CPU/mem/net via Docker API
- Show container table:
```bash
CONTAINER       CPU     MEM     NET IN    NET OUT
nginx           5.2%    45MB    120KB/s    90KB/s
```

---

## 📦 4. Non-Functional Requirements

### 4.1 Performance
- CLI should refresh under 200ms on typical systems
- Web UI should render with <500ms latency

### 4.2 Portability
- Should compile and work on:
    - Linux
    - macOS
    - Windows (partial, fallback for some syscalls)

### 4.3 Security
- Web dashboard accessible only on `localhost`
- No authentication or TLS (unless contributed later)

### 4.4 Usability
- Zero-config by default: `godash monitor` just works
- Docs + CLI help: `godash --help`

---

## 📁 5. Directory Structure

```bash
godash/
├── cmd/                  # CLI entrypoints
│   └── godash/
│       └── main.go
├── internal/
│   ├── metrics/          # System & Go runtime metrics
│   ├── web/              # HTTP, WebSocket handlers
│   ├── tui/              # Terminal UI
│   └── config/           # Config loader
├── static/               # JS/CSS for Web UI
├── templates/            # HTML pages
├── pkg/                  # Shared data models
├── scripts/              # Build or install scripts
├── go.mod
└── README.md
```

---

## 🧪 6. Test Plan

- Unit test for each collector (mock `/proc`, etc.)
- Integration test: run `godash monitor` with fake data
- WebSocket test: ensure live stream works as expected

---

## 🚀 7. Future Enhancements

| Idea                           | Benefit                                |
|--------------------------------|----------------------------------------|
| Export Prometheus `/metrics`  | Integrate with Grafana                 |
| Plugin system                  | Community-contributed collectors       |
| Dark/light theme toggle        | Improve UX                             |
| Docker container               | `docker run --net=host godash` ready   |
| GUI build (via Wails or Webview)| Native look for desktop users         |

---

## ✅ 8. Success Criteria

- Runs with `go run ./cmd/godash` with no external deps
- CLI and Web UI work independently and together
- GitHub repo has README, license, and build instructions
- At least 80% unit test coverage
- Tagged `v0.1.0` release ready for showcase

