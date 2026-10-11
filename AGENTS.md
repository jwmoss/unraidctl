# Repository Guidelines

## Project Structure & Module Organization
- `cmd/unraidctl` – Cobra-based CLI entrypoint; wire commands and flags.
- `pkg/client` – Reusable API client for consumers outside the CLI.
- `internal/api` – Request/response models and API helpers.
- `internal/config` – Config loading (flags > env vars > `~/.config/unraidctl/config.yaml`).
- `internal/output` – Human/JSON renderers.
- Live checks use `tests/`; static schema checks use `internal/api/schema_test.go`.
- CI workflows live in `.github/workflows/`.

## Build, Test, and Development Commands
- `make build` – Compile CLI to `./unraidctl` with command package version metadata from `VERSION` (default `1.4.0`).
- `make install` – Install binary into your Go bin path with version metadata.
- `make test` – Run live read-only CLI checks against the configured server.
- `make check` – Run formatting, module, vet, schema, and build checks.
- `make lint` – Run `golangci-lint`; matches CI expectations.
- `make build-all` – Cross-compile artifacts into `dist/` for darwin/linux/windows (amd64 + arm64).

## Coding Style & Naming Conventions
- Use the Go version in `go.mod`; run `gofmt` before commits.
- Keep packages short, lowercase; exported identifiers use PascalCase; unexported use camelCase.
- Prefer table-driven tests and subtests; keep command names verb-first (e.g., `array status`, `docker list`).
- Lint locally with `make lint` to catch `errcheck` and style violations.

## Testing Guidelines
- Keep live read-only tests under `tests/` with the `live` build tag.
- Use the compiled CLI and the real configured server.
- Do not add unit tests, mocked transports, or fake-server subprocess tests.
- Keep static GraphQL schema checks against the released schemas.
- Run `make check` without credentials in CI.
- Run `make test` locally with existing credentials; missing access is a failure.
- Never log credentials or provider response bodies from live checks.

## Commit & Pull Request Guidelines
- Use Conventional Commit prefixes seen in history (`fix:`, `docs:`, `ci:`, etc.); keep subjects ≤72 chars.
- For PRs: include a short summary, linked issue (if any), commands/tests run, and note config or API impacts.
- Ensure CI parity: code must pass `make check` and `make lint` before requesting review.

## Security & Configuration Tips
- Never commit API keys; prefer `UNRAID_SERVER` and `UNRAID_API_KEY` env vars or `~/.config/unraidctl/config.yaml` (excluded from repo).
- Validate `--server` URLs and avoid embedding credentials in flags when sharing command examples.
