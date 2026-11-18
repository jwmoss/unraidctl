# Unraid schema fixtures

These unmodified schemas come from the official Unraid API releases:

- https://github.com/unraid/api/blob/v4.37.3/api/generated-schema.graphql
- https://github.com/unraid/api/blob/v4.37.4/api/generated-schema.graphql
- https://github.com/unraid/api/blob/v4.37.5/api/generated-schema.graphql
- https://github.com/unraid/api/blob/v4.37.7/api/generated-schema.graphql
- https://github.com/unraid/api/blob/v4.37.12/api/generated-schema.graphql

Releases 4.37.8 through 4.37.12 have byte-identical schemas. The 4.37.12 fixture covers those five releases.

The 4.37.6 schema is identical to 4.37.5 and 4.37.7, so it has no separate fixture.

The upstream license appears in `UNRAID-LICENSE.txt`.

`go test ./internal/api` validates every operation in `queries.go` against every schema.
These fixtures describe the full schema. Feature flags and key permissions can restrict a live server.
