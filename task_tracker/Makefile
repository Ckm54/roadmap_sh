BINARY_NAME=task-cli
MODULE_NAME=$(shell go list -m)
SRC_DIR=./cmd/task-cli
BUILD_DIR=./bin

.PHONY: all build clean run test help

# Default target
all: test build

## build: Compile the binary
build:
			@echo "Building $(BINARY_NAME)..."
			@mkdir -p $(BUILD_DIR)
			go build -o $(BUILD_DIR)/$(BINARY_NAME) $(SRC_DIR)/main.go

## run: build and run the app
run: build
			@$(BUILD_DIR)/$(BINARY_NAME)

## test: run all the tests in the project
test:
			@echo "Running tests..."
			go test -v ./...

## clean: remove the build artifacts
clean:
			@echo "Cleaning up..."
			@rm -rf $(BUILD_DIR)

## help: show the help message
help:
			@echo "Usage:"
			@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'