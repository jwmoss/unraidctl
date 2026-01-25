# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

unraidctl is a Go CLI tool for interacting with the Unraid GraphQL API. It provides commands for managing system info, array status, Docker containers, VMs, shares, and notifications.

## Build Commands

```bash
make build          # Build binary to ./unraidctl
make test           # Run all tests
make lint           # Run golangci-lint
make install        # Install to $GOPATH/bin
```

Run a single test:
```bash
go test -v -run TestName ./internal/api/
```

## Architecture

### Three-Layer Design

1. **GraphQL Client** (`pkg/client/`) - Generic HTTP client that sends queries to `/graphql` with `x-api-key` header
2. **API Layer** (`internal/api/`) - GraphQL query strings (`queries.go`) and response types (`types.go`)
3. **Commands** (`cmd/unraidctl/cmd/`) - Cobra commands that call the client and format output

### Key Patterns

- **Configuration precedence**: CLI flags > Environment variables > Config file (`~/.config/unraidctl/config.yaml`)
- **Output formatting**: All output goes through `internal/output/Formatter` which handles JSON/table/quiet modes
- **Error wrapping**: Use `fmt.Errorf("context: %w", err)` consistently
- **Context timeouts**: All API calls use `context.WithTimeout(context.Background(), 30*time.Second)`

### Adding a New Command

1. Add GraphQL query to `internal/api/queries.go`
2. Add response type to `internal/api/types.go`
3. Create command file in `cmd/unraidctl/cmd/`
4. Register command in `root.go` init function
5. Add integration test in `internal/api/integration_test.go`

### Testing

- Mock server in `internal/api/mock_server_test.go` simulates the Unraid API
- Integration tests in `internal/api/integration_test.go` test queries against the mock
- Uses standard `testing` package only

## Dependencies

Minimal dependencies by design:
- `github.com/spf13/cobra` - CLI framework
- `gopkg.in/yaml.v3` - Config parsing
- Standard library for HTTP and JSON
