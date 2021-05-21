BINARY_NAME=gh-activity
SRC_DIR=./cmd/gh-activity
BUILD_DIR=./bin

.PHONY: all build run install test vet fmt clean help

# Default
all: fmt vet test build

## build: Compile the binary to ./bin
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(SRC_DIR)

## run: Build and run the app, e.g. make run ARGS="octocat"
run: build
	@$(BUILD_DIR)/$(BINARY_NAME) $(ARGS)

## install: Install the binary into $GOBIN (or $GOPATH/bin)
install:
	go install $(SRC_DIR)

## test: Run all tests with the race detector
test:
	go test -race ./...

## vet: Run go vet
vet:
	go vet ./...

## fmt: Format all Go files
fmt:
	gofmt -s -w .

## clean: Remove build artifacts
clean:
	@rm -rf $(BUILD_DIR)

## help: Show this help message
help:
	@echo "Usage:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'
