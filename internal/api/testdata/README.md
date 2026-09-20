# Unraid schema fixtures

These unmodified schemas come from the official Unraid API releases:

- https://github.com/unraid/api/blob/v4.37.3/api/generated-schema.graphql
- https://github.com/unraid/api/blob/v4.37.4/api/generated-schema.graphql

The upstream license appears in `UNRAID-LICENSE.txt`.

`go test ./internal/api` validates every operation in `queries.go` against both schemas.
These fixtures describe the full schema. Feature flags and key permissions can restrict a live server.
