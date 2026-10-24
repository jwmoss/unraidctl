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
- **UPS and Disks** - View UPS status and configuration, change UPS settings, and view physical disks and basic SMART status
- **Parity** - View parity status and history; start, pause, resume, and cancel parity checks
- **Array Management** - Check array status, capacity, and disk health
- **Array Operations** - Start/stop the array and manage disk assignment/mount state
- **Docker Containers** - List, inspect, start, stop, restart, pause, update, remove, view logs, check for updates, and stream live stats
- **System Metrics** - View or stream CPU, memory, temperature, network traffic, and CPU package power
- **System** - View services, server identity, network interfaces, plugins, and time settings; change time settings
- **Flash Backup** - Start an rclone backup of the USB flash device
- **Shares** - List user shares and their available storage capacity
- **Notifications** - List, archive, unarchive, delete, create, and stream notifications
- **Virtual Machines** - List, start, stop, pause, resume, force-stop, reboot, and reset VMs (when enabled)
- **API Keys** - List, create, update, delete, and manage roles/permissions
- **Logs** - List, read, and follow Unraid API log files
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

The contract tests cover API **4.37.3** through **4.37.7**. Older APIs can require a Connect plugin update.
The health and reporting commands require read access to their resources. On API 4.37.4, UPS access requires `CONFIG:READ_ANY`.
Feature flags can restrict Docker fields even when the generated schema includes them.

API 4.37.4 removes direct disk removal. `array remove-disk` now returns guidance without a server request.
Use the Unraid WebGUI storage workflow for disk removal.

`--watch`, `--follow`, `docker stats`, `metrics cpu-telemetry`, and `notification watch` use GraphQL subscriptions over WebSocket (`graphql-transport-ws` on `/graphql`).
API 4.37.5 never delivers array, parity, or live UPS subscription events, so those commands have no watch mode.

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

`unraidctl version` uses the module version from `go install` when the build has no explicit version.
Explicit build flags take precedence. Local source builds use `dev` when no version is available.

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

Use `--config PATH` to select a different file. `configure` replaces an existing regular file with a private file (POSIX mode `0600`). Windows uses inherited directory permissions.
It rejects destination symlinks and nonregular files. A failed save preserves the previous contents.
It creates missing config directories with mode `0700`.

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

A missing config file permits flags and environment variables. Other file read errors stop the command, even when credentials are available.

Use the final server URL. The client follows same-origin redirects only.
It blocks redirects to another origin to protect the API key.

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
unraidctl parity start            # read-only check; add --correct to write corrections
unraidctl parity pause | resume | cancel
unraidctl ups config
unraidctl ups set --battery-level 15 --minutes 10 --yes

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
unraidctl docker update <container-id> <container-id>
unraidctl docker check-updates
unraidctl docker stats --count 20

# System metrics
unraidctl metrics cpu
unraidctl metrics memory
unraidctl metrics temperature
unraidctl metrics network
unraidctl metrics cpu --watch      # stream until Ctrl-C; --count N stops after N updates
unraidctl metrics cpu-telemetry

# System
unraidctl system services
unraidctl system server
unraidctl system interfaces
unraidctl system plugins
unraidctl system time
unraidctl system timezones
unraidctl system set-time --timezone America/New_York --ntp-server time.nist.gov --yes

# Flash backup (needs a configured rclone remote)
unraidctl flash backup --remote <remote> --destination unraid/flash

# Virtual machines
unraidctl vm list
unraidctl vm start <vm-id>        # also stop, pause, resume, force-stop, reboot, reset

# Shares
unraidctl share list

# Notifications
unraidctl notification list
unraidctl notification list --all
unraidctl notification alerts
unraidctl notification archive <notification-id>...
unraidctl notification archive --all --importance INFO
unraidctl notification unarchive <notification-id>...
unraidctl notification delete <notification-id> --yes
unraidctl notification delete-archived --yes
unraidctl notification create --title Backup --subject Done --description "Nightly job finished" --unique
unraidctl notification recalculate
unraidctl notification watch

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
unraidctl log show /var/log/syslog --follow

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

## Development tests

Run `make check` for formatting, module, vet, released-schema, and build checks.
Run `make lint` for the pinned Go lint checks.
Run `make test` with existing credentials for live information and array-status checks.
These checks read the configured server and print no server data.
They do not verify mutations or failure paths.

Use `--dry-run` to block GraphQL mutations and local configuration writes.
Read commands remain available. Use `--timeout 10s` to set the request deadline.
Use `--trace-http` for method, path, status, and duration on stderr.
Use `--plain` for table rows without headers and color.
The version command supports `--json` and `--plain`.
Settings updates preserve large JSON integers.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

The CLI uses the MIT license. See [LICENSE](LICENSE).
The unmodified upstream test schemas retain their separate GPL-2.0-or-later license. See [fixture notice](internal/api/testdata/README.md).

CI runs static checks without service credentials. Unit tests and fake-server suites are intentionally absent.
