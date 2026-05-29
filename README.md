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
- **Array Operations** - Start/stop the array and manage disk assignment/mount state
- **Docker Containers** - List, inspect, start, stop, pause, update, remove, and view logs
- **Shares** - List user shares with usage statistics
- **Notifications** - View unread and all notifications
- **Virtual Machines** - List VMs (when enabled)
- **API Keys** - List, create, update, delete, and manage roles/permissions
- **Logs** - List and read Unraid API log files
- **Settings and SSO** - Inspect/update API settings and view OIDC provider configuration
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
unraidctl array start
unraidctl array stop
unraidctl array add-disk <disk-id> --slot 1
unraidctl array remove-disk <disk-id>
unraidctl array mount-disk <disk-id>
unraidctl array unmount-disk <disk-id>
unraidctl array clear-stats <disk-id>

# Docker containers
unraidctl docker list
unraidctl docker list --wide
unraidctl docker inspect <container-id>
unraidctl docker logs <container-id> --tail 200
unraidctl docker start <container-id>
unraidctl docker stop <container-id>
unraidctl docker pause <container-id>
unraidctl docker unpause <container-id>
unraidctl docker update <container-id>
unraidctl docker update-all
unraidctl docker autostart <container-id> --enable --wait 10
unraidctl docker remove <container-id> --with-image

# Virtual machines
unraidctl vm list

# Shares
unraidctl share list

# Notifications
unraidctl notification list
unraidctl notification list --all

# API keys
unraidctl apikey list
unraidctl apikey create --name automation --role VIEWER
unraidctl apikey update <api-key-id> --name automation-v2
unraidctl apikey add-role <api-key-id> --role ADMIN
unraidctl apikey remove-role <api-key-id> --role VIEWER
unraidctl apikey delete <api-key-id>
unraidctl apikey roles
unraidctl apikey permissions

# Logs
unraidctl log list
unraidctl log show /var/log/graphql-api.log --lines 100

# Settings and SSO/OIDC
unraidctl settings show
unraidctl settings show --values
unraidctl settings update --file settings.json
unraidctl sso status
unraidctl sso providers
unraidctl sso public-providers
unraidctl sso config
unraidctl sso validate-token <token>

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
