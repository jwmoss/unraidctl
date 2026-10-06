.PHONY: build clean test install lint fmt-check tidy-check vet check release-tool-check release-check release-snapshot

BINARY_NAME=unraidctl
VERSION?=1.3.1
GORELEASER_VERSION := $(shell cat .goreleaser-version)
GORELEASER := goreleaser
VERSION_PACKAGE=github.com/jwmoss/unraidctl/cmd/unraidctl/cmd
LDFLAGS=-ldflags "-X $(VERSION_PACKAGE).version=$(VERSION)"

build:
	go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/unraidctl

install:
	go install $(LDFLAGS) ./cmd/unraidctl

clean:
	rm -f $(BINARY_NAME)
	rm -rf dist/

test:
	go test -v ./...

lint:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run

fmt-check:
	@test -z "$$(gofmt -l cmd internal)"

tidy-check:
	go mod tidy -diff

vet:
	go vet ./...

check: fmt-check tidy-check vet test build

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

# Fixture-only compiled CLI tests; no Unraid server credentials are required.
.PHONY: test-e2e
test-e2e: build
	npm exec --no -- e2e run
