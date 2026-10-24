.PHONY: complexity schema test-e2e build clean test install lint fmt-check tidy-check vet check release-tool-check release-check release-snapshot

BINARY_NAME=unraidctl
VERSION?=1.5.1
GORELEASER_VERSION := $(shell cat .goreleaser-version)
GORELEASER := goreleaser
VERSION_PACKAGE=github.com/jwmoss/unraidctl/cmd/unraidctl/cmd
LDFLAGS=-ldflags "-X $(VERSION_PACKAGE).version=$(VERSION)"

build:
	go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/unraidctl

install:
	go install $(LDFLAGS) ./cmd/unraidctl

clean:
	trash $(BINARY_NAME)
	trash dist/

test test-e2e: build
	go test -tags=live -count=1 -timeout=2m ./tests

lint:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run

fmt-check:
	@test -z "$$(gofmt -l cmd internal pkg tests)"

tidy-check:
	go mod tidy -diff

vet:
	go vet -tags=live ./...

schema:
	go test -count=1 -timeout=2m ./internal/api -run TestOperationsMatchReleasedSchemas

complexity:
	@if command -v cccc >/dev/null; then cccc --top-cognitive 5 cmd internal pkg; else echo "cccc not installed; skipping (cargo install cccc-cli)"; fi

check: fmt-check tidy-check vet schema complexity build

release-tool-check:
	@$(GORELEASER) --version | grep -Eq "^GitVersion: +v?$(subst .,[.],$(GORELEASER_VERSION:v%=%))$$" || { echo "Install GoReleaser $(GORELEASER_VERSION)"; exit 1; }

release-check: release-tool-check
	$(GORELEASER) check

release-snapshot: release-tool-check
	$(GORELEASER) release --snapshot --clean

# Cross-compilation
build-all: clean
	mkdir -p dist
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 ./cmd/unraidctl
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 ./cmd/unraidctl
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 ./cmd/unraidctl
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-arm64 ./cmd/unraidctl
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe ./cmd/unraidctl
