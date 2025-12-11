# Mihosh

English | [简体中文](README_CN.md)

A full-featured terminal UI (TUI) management tool for mihomo, allowing you to elegantly manage proxy nodes, monitor connections, and view logs directly from your terminal.

## Tech Stack

![Go](https://img.shields.io/badge/Go-00ADD8?style=flat-square&logo=go&logoColor=white)
![Bubbletea](https://img.shields.io/badge/Bubbletea-FF69B4?style=flat-square&logo=go&logoColor=white)
![Lipgloss](https://img.shields.io/badge/Lipgloss-9B59B6?style=flat-square&logo=go&logoColor=white)
![Cobra](https://img.shields.io/badge/Cobra-2ECC71?style=flat-square&logo=go&logoColor=white)
![Viper](https://img.shields.io/badge/Viper-E74C3C?style=flat-square&logo=go&logoColor=white)
![WebSocket](https://img.shields.io/badge/WebSocket-010101?style=flat-square&logo=socket.io&logoColor=white)

## Features

| Page | Description |
|------|-------------|
| 🎯 **Nodes** | Switch proxy nodes quickly, single/batch latency testing |
| 📊 **Connections** | Real-time active connections, traffic/memory charts, close connections |
| 📝 **Logs** | Live log streaming with level filtering and keyword search |
| 📋 **Rules** | View proxy rules with multi-keyword search |
| ⚙️ **Settings** | Modify configuration directly in the UI |
| ❓ **Help** | Built-in keyboard shortcuts reference |

## Installation

### Binary Release

Download the executable for your platform from [Releases](https://github.com/aimony/mihosh/releases).

### Build from Source

```bash
git clone https://github.com/aimony/mihosh.git
cd mihosh && go build -o mihosh.exe
```

## Quick Start

### 1. Initialize Configuration

```bash
mihosh config init
```

Enter your Mihomo API address and secret when prompted. Config is saved to `~/.mihosh/config.yaml`

### 2. Launch

```bash
mihosh
```

This opens the interactive TUI. Press `5` or `Tab` to switch to the Help page for keyboard shortcuts.

## Configuration

Config file located at `~/.mihosh/config.yaml`:

```yaml
api_address: http://127.0.0.1:9090
secret: your-secret-here
test_url: http://www.gstatic.com/generate_204
timeout: 5000
```

## CLI Mode (Optional)

In addition to the TUI, command-line operations are also supported:

```bash
mihosh list                          # List proxy groups
mihosh select <group> <node>         # Switch node
mihosh test <node>                   # Test node latency
mihosh connections                   # View connections
```

## FAQ

| Issue | Solution |
|-------|----------|
| Connection failed | Check if Mihomo is running, verify API address and secret |
| Nodes not found | Ensure proxy groups are configured in mihomo config |
| Test timeout | Increase `timeout` value or change `test_url` |

## Development

```bash
go mod tidy      # Install dependencies
go test ./...    # Run tests
go build         # Build
```

## License

MIT License
