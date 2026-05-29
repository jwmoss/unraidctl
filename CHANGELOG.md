# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.1.0] - 2026-05-29

### Added

- **API key management** (`unraidctl apikey`)
  - List, create, update, and delete API keys
  - Add/remove roles and inspect available roles/permissions
- **Array operations** (`unraidctl array`)
  - Start and stop the array
  - Add/remove disks, mount/unmount disks, and clear disk statistics
- **Docker operations** (`unraidctl docker`)
  - Inspect rich container metadata
  - Read container logs
  - Start, stop, pause, unpause, update, update all, remove, and configure autostart
- **Log access** (`unraidctl log`)
  - List and read Unraid API log files
- **Settings and SSO/OIDC** (`unraidctl settings`, `unraidctl sso`)
  - Inspect/update API settings
  - List OIDC providers, public login providers, OIDC configuration, and validate OIDC session tokens

### Changed

- Expanded Docker list responses with newer Unraid API metadata fields.
- Updated CI/release GitHub Actions dependencies.
- Dropped the invalid Go 1.21 CI job because the module requires Go 1.22+.
- Release builds now inject version, commit, and build date into `unraidctl version`.

## [1.0.0] - 2025-01-25

### Added

- Initial release of unraidctl CLI
- **System Information** (`unraidctl info`)
  - Display hostname, OS version, platform, uptime
  - CPU information (manufacturer, brand, cores, speed)
  - JSON output support
- **Array Management** (`unraidctl array`)
  - View array status and state
  - Display capacity (total, used, free)
  - List all disks with device, type, size, status, and temperature
- **Docker Containers** (`unraidctl docker`)
  - List all containers with state, status, image, and autostart setting
- **Shares** (`unraidctl share`)
  - List user shares with used/free space and comments
- **Notifications** (`unraidctl notification`)
  - List unread notifications with importance and timestamp
  - Option to show all notifications (`--all`)
- **Virtual Machines** (`unraidctl vm`)
  - List VMs with name and state (when VM manager is enabled)
- **Configuration**
  - Config file support (`~/.config/unraidctl/config.yaml`)
  - Environment variable support (`UNRAID_SERVER`, `UNRAID_API_KEY`)
  - Command-line flags (`--server`, `--api-key`)
  - Interactive configuration wizard (`unraidctl configure`)
- **Output Formats**
  - Human-readable table output (default)
  - JSON output (`--json`) for scripting
  - Quiet mode (`--quiet`)
  - Color output with `--no-color` and `NO_COLOR` support

### Technical

- GraphQL client for Unraid API communication
- Tested against Unraid 7.2
- Cross-platform support (macOS, Linux, Windows)

[1.1.0]: https://github.com/jwmoss/unraidctl/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/jwmoss/unraidctl/releases/tag/v1.0.0
