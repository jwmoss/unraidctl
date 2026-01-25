# unraidctl

A command-line tool to interact with the [Unraid API](https://docs.unraid.net/API/).

## Installation

```bash
go install github.com/jwmoss/unraidctl@latest
```

Or build from source:

```bash
git clone https://github.com/jwmoss/unraidctl.git
cd unraidctl
go build -o unraidctl ./cmd/unraidctl
```

## Configuration

### Config file (recommended)

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
unraidctl array start --force
unraidctl array stop --force

# Docker containers
unraidctl docker list
unraidctl docker start <container>
unraidctl docker stop <container>
unraidctl docker restart <container>

# Virtual machines
unraidctl vm list
unraidctl vm start <name>
unraidctl vm stop <name>

# Shares
unraidctl share list

# Notifications
unraidctl notification list
unraidctl notification dismiss <id>

# JSON output for scripting
unraidctl info --json
unraidctl docker list --json | jq '.[] | .name'
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

## License

MIT
