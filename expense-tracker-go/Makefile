BINARY_NAME=expense-tracker
SRC_DIR=./cmd/expense-tracker
BUILD_DIR=./bin

.PHONY: all build run install test cover vet fmt lint clean help

# Default
all: fmt vet test build

## build: Compile the binary to ./bin
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(SRC_DIR)

## run: Build and run the app, e.g. make run ARGS="summary --month 8"
run: build
	@$(BUILD_DIR)/$(BINARY_NAME) $(ARGS)

## install: Install the binary into $GOBIN (or $GOPATH/bin)
install:
	go install $(SRC_DIR)

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

## lint: Run golangci-lint
lint:
	golangci-lint run ./...

## clean: Remove build artifacts
clean:
	@echo "Cleaning up..."
	@rm -rf $(BUILD_DIR) coverage.out

## help: Show this help message
help:
	@echo "Usage:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'
