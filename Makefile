.PHONY: build clean test install lint

BINARY_NAME=unraidctl
VERSION?=1.1.0
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
	golangci-lint run

# Cross-compilation
build-all: clean
	mkdir -p dist
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 ./cmd/unraidctl
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 ./cmd/unraidctl
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 ./cmd/unraidctl
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-arm64 ./cmd/unraidctl
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe ./cmd/unraidctl
