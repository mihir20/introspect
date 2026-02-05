.PHONY: build run clean help fmt deps

# Package to target (override with: make run PKG=linear)
PKG ?= linear

# Build output directory
BIN_DIR=bin

# Build a package
build:
	@echo "Building $(PKG)..."
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_DIR)/$(PKG) ./$(PKG)/
	@echo "Build complete: $(BIN_DIR)/$(PKG)"

# Run a package
run:
	@go run ./$(PKG)/

# Build and run a package
build-run: build
	@./$(BIN_DIR)/$(PKG)

# Build all packages
build-all:
	@for dir in $(shell find . -maxdepth 2 -name '*.go' -exec dirname {} \; | sort -u | grep -v '^\./$$'); do \
		pkg=$$(basename $$dir); \
		echo "Building $$pkg..."; \
		mkdir -p $(BIN_DIR); \
		go build -o $(BIN_DIR)/$$pkg ./$$pkg/ || exit 1; \
	done
	@echo "All packages built!"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BIN_DIR)
	@rm -f linear_completed_tickets.json
	@rm -f linear_completed_tickets.csv
	@echo "Cleaned!"

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "Formatted!"

# Install dependencies
deps:
	@go mod tidy

# Display help
help:
	@echo "Available commands:"
	@echo "  make build  PKG=<pkg>  - Build a specific package (default: linear)"
	@echo "  make run    PKG=<pkg>  - Run a specific package (default: linear)"
	@echo "  make build-run PKG=<pkg> - Build and run a package"
	@echo "  make build-all         - Build all packages"
	@echo "  make clean             - Remove build artifacts and output files"
	@echo "  make fmt               - Format all code"
	@echo "  make deps              - Tidy go modules"
	@echo "  make help              - Show this help message"
