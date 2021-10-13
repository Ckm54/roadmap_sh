BINARY_NAME=task-cli
SRC_DIR=./cmd/task-cli
BUILD_DIR=./bin
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS=-s -w -X main.version=$(VERSION)

.PHONY: all build run install test cover vet fmt lint release-snapshot clean help

# Default target
all: fmt vet test build

## build: Compile the binary to ./bin
build:
	@echo "Building $(BINARY_NAME) $(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) $(SRC_DIR)

## run: Build and run the app, e.g. make run ARGS="list done"
run: build
	@$(BUILD_DIR)/$(BINARY_NAME) $(ARGS)

## install: Install the binary into $GOBIN (or $GOPATH/bin)
install:
	go install -ldflags "$(LDFLAGS)" $(SRC_DIR)

## test: Run all tests with the race detector
test:
	@echo "Running tests..."
	go test -race ./...

## cover: Run tests and open an HTML coverage report
cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

## vet: Run go vet
vet:
	go vet ./...

## fmt: Format all Go files
fmt:
	gofmt -s -w .

## lint: Run golangci-lint (https://golangci-lint.run)
lint:
	golangci-lint run ./...

## release-snapshot: Build release archives locally with goreleaser, without publishing
release-snapshot:
	goreleaser release --snapshot --clean

## clean: Remove build artifacts
clean:
	@echo "Cleaning up..."
	@rm -rf $(BUILD_DIR) dist coverage.out

## help: Show this help message
help:
	@echo "Usage:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'
