# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.3.0] - 2026-09-20

### Added

- API health reports with raw UPS alarms, array devices, parity history, container status, and unread alert counts.
- UPS status, physical disk inventory, parity status/history, and CPU, memory, and temperature metrics.
- Deduplicated notification alerts and complete pagination for unread and archived notifications.
- Offline validation of every GraphQL operation against the released API 4.37.3 and 4.37.4 schemas.

### Fixed

- Replace the invalid notification `ALL` enum with separate unread and archive requests.
- Include parity, cache, and boot devices in array status.
- Correct decimal capacity conversion and binary unit labels; identify share capacity as storage capacity, not folder usage.
- Report exact Unraid, API, and kernel versions.
- Explain when VM Manager is unavailable without hiding other API errors.
- Parse GraphQL fields in API mocks instead of matching query substrings.
- Correct the Go installation path and document the implemented exit codes.

### Changed

- Retire direct disk removal. The command now directs users to the Unraid WebGUI without a server request.
- Health reports exit with status 1 for alerts, warnings, or incomplete data. JSON still contains the available results.
- Ignore the upstream hard-coded UPS battery health value. Preserve raw UPS status and explain diagnostic limits.

## [1.2.0] - 2026-08-19

### Added

- Network interface metrics from Unraid API 4.35 and later.
- Docker container restart support from Unraid API 4.36 and later.

### Changed

- Documented API version checks and the Unraid Connect plugin update path.

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

[1.3.0]: https://github.com/jwmoss/unraidctl/compare/v1.2.0...v1.3.0
[1.2.0]: https://github.com/jwmoss/unraidctl/compare/v1.1.0...v1.2.0
[1.1.0]: https://github.com/jwmoss/unraidctl/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/jwmoss/unraidctl/releases/tag/v1.0.0
