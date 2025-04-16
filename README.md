# GoDash 🖥️📊

**GoDash** is a cross-platform system monitoring tool written in Golang that provides both a rich CLI interface and a local web dashboard to visualize CPU, memory, disk, network, and Go runtime statistics in real-time.

> 📦 Single binary.  
> 🌐 Web UI + CLI.  
> ⚡ Lightweight.  
> 🧠 Built to learn & extend Go!

---

## ✨ Features

- 📈 Live CPU, memory, disk, and network stats
- 🧵 Go runtime metrics (goroutines, GC, heap)
- 🌐 Web dashboard served at `http://localhost:8080`
- 🖥️ Terminal dashboard with optional TUI
- 🐳 Optional: Docker container stats
- 📦 Portable: Works on Linux, macOS, Windows

---

## 📸 Screenshots

### Terminal UI


### Web Dashboard 
![Web Dashboard Dark](scratchpad/Screenshot_2025-04-15_17-26-31.png)
*Web interface with real-time charts (Dark Theme)*

---

## 🚀 Getting Started

### 🔧 Install

```bash
git clone https://github.com/j-raghavan/godash.git
cd godash
go build -o godash ./cmd/godash
```

Or install via:
```bash
go install github.com/j-raghavan/godash/cmd/godash@latest
```

## 🖥️ Run CLI Mode

```bash
godash monitor
```

## 🌐 Run Web Dashboard
```bash
godash serve --port 8080
```
Then open http://localhost:8080


## ⚙️ Configuration

Supports config via flags, .godash.toml and env vars.

```bash
# ~/.godash.toml
TBD
```


## 🔭 Roadmap

See [TODO.md](TODO.md) 


## 🤝 Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md)


## 📄 License

MIT (c) J-Raghavan

