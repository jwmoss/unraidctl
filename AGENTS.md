# Repository Guidelines

## Project Structure & Module Organization
- `cmd/unraidctl` – Cobra-based CLI entrypoint; wire commands and flags.
- `pkg/client` – Reusable API client for consumers outside the CLI.
- `internal/api` – Request/response models and API helpers.
- `internal/config` – Config loading (flags > env vars > `~/.config/unraidctl/config.yaml`).
- `internal/output` – Human/JSON renderers.
- Tests live alongside code in `*_test.go`; CI workflows in `.github/workflows/`.

## Build, Test, and Development Commands
- `make build` – Compile CLI to `./unraidctl` with `main.version` injected from `VERSION` (default `0.1.0`).
- `make install` – Install binary into your Go bin path with version metadata.
- `make test` or `go test -v ./...` – Run unit/integration tests.
- `make lint` – Run `golangci-lint`; matches CI expectations.
- `make build-all` – Cross-compile artifacts into `dist/` for darwin/linux/windows (amd64 + arm64).

## Coding Style & Naming Conventions
- Target Go 1.22; run `gofmt` (tabs, 4-space visual indent) before commits.
- Keep packages short, lowercase; exported identifiers use PascalCase; unexported use camelCase.
- Prefer table-driven tests and subtests; keep command names verb-first (e.g., `array status`, `docker list`).
- Lint locally with `make lint` to catch `errcheck` and style violations.

## Testing Guidelines
- Write `_test.go` files next to the code under test; use `t.Run` for sub-cases.
- Tests should not require a live Unraid server; mock responses where possible. Mark any network-dependent cases clearly and guard with env flags.
- Aim for coverage on command parsing, API client error paths, and JSON output formatting.

## Commit & Pull Request Guidelines
- Use Conventional Commit prefixes seen in history (`fix:`, `docs:`, `ci:`, etc.); keep subjects ≤72 chars.
- For PRs: include a short summary, linked issue (if any), commands/tests run, and note config or API impacts.
- Ensure CI parity: code must pass `go build ./...`, `go test -v ./...`, and `golangci-lint` before requesting review.

## Security & Configuration Tips
- Never commit API keys; prefer `UNRAID_SERVER` and `UNRAID_API_KEY` env vars or `~/.config/unraidctl/config.yaml` (excluded from repo).
- Validate `--server` URLs and avoid embedding credentials in flags when sharing command examples.
