# unraidctl

[![CI](https://github.com/jwmoss/unraidctl/actions/workflows/ci.yml/badge.svg)](https://github.com/jwmoss/unraidctl/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/jwmoss/unraidctl)](https://goreportcard.com/report/github.com/jwmoss/unraidctl)
[![Go Reference](https://pkg.go.dev/badge/github.com/jwmoss/unraidctl.svg)](https://pkg.go.dev/github.com/jwmoss/unraidctl)
[![Release](https://img.shields.io/github/v/release/jwmoss/unraidctl)](https://github.com/jwmoss/unraidctl/releases/latest)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A command-line tool to interact with the [Unraid API](https://docs.unraid.net/API/).

## Features

- **System Information** - View exact Unraid, API, and kernel versions, CPU, and uptime
- **Health** - Check array devices, parity history, raw UPS alarms, containers, sensors, and unread alerts
- **UPS and Disks** - View UPS status, physical disks, and basic SMART status
- **Parity** - View current parity status and complete check history
- **Array Management** - Check array status, capacity, and disk health
- **Array Operations** - Start/stop the array and manage disk assignment/mount state
- **Docker Containers** - List, inspect, start, stop, restart, pause, update, remove, and view logs
- **System Metrics** - View CPU, memory, temperature, network traffic, errors, drops, and link utilization
- **Shares** - List user shares and their available storage capacity
- **Notifications** - List every unread/archive notification and deduplicated warnings/alerts
- **Virtual Machines** - List VMs (when enabled)
- **API Keys** - List, create, update, delete, and manage roles/permissions
- **Logs** - List and read Unraid API log files
- **Settings and SSO** - Inspect/update API settings and view OIDC provider configuration
- **JSON Output** - Machine-readable output for scripting

## Requirements

- Unraid 7.2+ (API built-in) or Unraid with the [Unraid Connect](https://docs.unraid.net/unraid-connect/overview-and-setup/) plugin
- An API key (create at **Settings → Management Access → API Keys**)

## API compatibility

Each Unraid OS release includes a specific API version. The Unraid Connect plugin can provide
newer API features before they become part of an Unraid OS release.

Run this command after an OS or plugin update:

```bash
unraidctl settings show
```

The `API version` row shows the active server API version.

| Command | Minimum API version |
|---------|---------------------|
| `unraidctl metrics network` | 4.35 |
| `unraidctl docker restart <container-id>` | 4.36 |

The contract tests cover API **4.37.3** and **4.37.4**. Older APIs can require a Connect plugin update.
The health and reporting commands require read access to their resources. On API 4.37.4, UPS access requires `CONFIG:READ_ANY`.
Feature flags can restrict Docker fields even when the generated schema includes them.

API 4.37.4 removes direct disk removal. `array remove-disk` now returns guidance without a server request.
Use the Unraid WebGUI storage workflow for disk removal.

## Health and capacity interpretation

`health` returns available results even if one API resource fails. An incomplete report has status `INCOMPLETE` and exits 1.
A report with alarms or warnings has status `ATTENTION` and also exits 1. JSON output remains valid in both cases.

UPS output uses the raw status. It does not trust API 4.37.x `battery.health`, which is hard-coded to `Good`.
The API can supply defaults for absent UPS measurements. An online status does not independently verify battery condition.

API health checks do not include detailed SMART attributes, Btrfs error counters, scrub results, or kernel I/O errors.
Use separate host diagnostics for those checks. A basic SMART pass does not establish full disk health.

Array and share capacity use decimal KB from the API and display decimal TB/GB.
Raw array-device sizes use KiB and display TiB. Byte-based memory, disk, and network values use explicit binary units.
Share capacity describes its storage backing, not measured folder contents. Do not sum share rows.

`notification list` paginates all unread items. `--all` adds all archived items. Neither command silently truncates results.
`notification alerts` uses the API's deduplicated warnings and alerts. These commands do not dismiss notifications.

## Installation

### From Release (recommended)

Download the latest binary from [Releases](https://github.com/jwmoss/unraidctl/releases/latest).

### Using Go

```bash
go install github.com/jwmoss/unraidctl/cmd/unraidctl@latest
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
# Show exact component versions and health
unraidctl info
unraidctl health
unraidctl health --json
unraidctl ups status
unraidctl disk list
unraidctl parity status
unraidctl parity history

# Array management
unraidctl array status
unraidctl array start
unraidctl array stop
unraidctl array add-disk <disk-id> --slot 1
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
unraidctl docker restart <container-id> # Requires Unraid API 4.36+
unraidctl docker pause <container-id>
unraidctl docker unpause <container-id>
unraidctl docker update <container-id>
unraidctl docker update-all
unraidctl docker autostart <container-id> --enable --wait 10
unraidctl docker remove <container-id> --with-image

# System metrics
unraidctl metrics cpu
unraidctl metrics memory
unraidctl metrics temperature
unraidctl metrics network

# Virtual machines
unraidctl vm list

# Shares
unraidctl share list

# Notifications
unraidctl notification list
unraidctl notification list --all
unraidctl notification alerts

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
| 1 | Command error, or health report with alerts, warnings, or incomplete data |

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

The CLI uses the MIT license. See [LICENSE](LICENSE).
The unmodified upstream test schemas retain their separate GPL-2.0-or-later license. See [fixture notice](internal/api/testdata/README.md).
