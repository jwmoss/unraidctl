# unraidctl

[![CI](https://github.com/jwmoss/unraidctl/actions/workflows/ci.yml/badge.svg)](https://github.com/jwmoss/unraidctl/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/jwmoss/unraidctl)](https://goreportcard.com/report/github.com/jwmoss/unraidctl)
[![Go Reference](https://pkg.go.dev/badge/github.com/jwmoss/unraidctl.svg)](https://pkg.go.dev/github.com/jwmoss/unraidctl)
[![Release](https://img.shields.io/github/v/release/jwmoss/unraidctl)](https://github.com/jwmoss/unraidctl/releases/latest)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A command-line tool to interact with the [Unraid API](https://docs.unraid.net/API/).

## Features

- **System Information** - View hostname, OS, CPU, and uptime
- **Array Management** - Check array status, capacity, and disk health
- **Docker Containers** - List containers with state and status
- **Shares** - List user shares with usage statistics
- **Notifications** - View unread and all notifications
- **Virtual Machines** - List VMs (when enabled)
- **JSON Output** - Machine-readable output for scripting

## Requirements

- Unraid 7.2+ (API built-in) or Unraid with the [Unraid Connect](https://docs.unraid.net/unraid-connect/overview-and-setup/) plugin
- An API key (create at **Settings → Management Access → API Keys**)

## Installation

### From Release (recommended)

Download the latest binary from [Releases](https://github.com/jwmoss/unraidctl/releases/latest).

### Using Go

```bash
go install github.com/jwmoss/unraidctl@latest
```

### Build from source

```bash
git clone https://github.com/jwmoss/unraidctl.git
cd unraidctl
go build -o unraidctl ./cmd/unraidctl
```

## Configuration

### Interactive setup (recommended)

```bash
unraidctl configure
```

### Config file

Create `~/.config/unraidctl/config.yaml`:

```yaml
server: http://192.168.1.100
api_key: your-api-key-here
```

### Environment variables

```bash
export UNRAID_SERVER="http://192.168.1.100"
export UNRAID_API_KEY="your-api-key-here"
```

### Precedence

Flags > Environment variables > Config file

## Usage

```bash
# Show system information
unraidctl info

# Array management
unraidctl array status

# Docker containers
unraidctl docker list

# Virtual machines
unraidctl vm list

# Shares
unraidctl share list

# Notifications
unraidctl notification list
unraidctl notification list --all

# JSON output for scripting
unraidctl info --json
unraidctl docker list --json | jq '.[].names[0]'
```

## Global Flags

| Flag | Env Var | Description |
|------|---------|-------------|
| `--server` | `UNRAID_SERVER` | Unraid server URL |
| `--api-key` | `UNRAID_API_KEY` | API key for authentication |
| `--json` | - | Output in JSON format |
| `--quiet` | - | Suppress non-essential output |
| `--no-color` | `NO_COLOR` | Disable colored output |
| `--config` | - | Path to config file |

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Invalid usage / bad arguments |
| 3 | Authentication error |
| 4 | Connection error |

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT - see [LICENSE](LICENSE) for details.
