# Makefile

# Folder where HRA is built
HRA_DIR = hall_request_assigner

# Name of the compiled HRA binary
HRA_EXEC = hall_request_assigner

# Default port (can be overridden via `make run PORT=20003`)
PORT ?= 20003

# Build the HRA executable
hra:
	@echo "Building HRA executable..."
	cd $(HRA_DIR) && ./build.sh

# Run the Go project (depends on HRA being built)
run: hra
	go run . --port=$(PORT)

# Clean up (optional)
clean:
	rm -f $(HRA_DIR)/$(HRA_EXEC)

.DEFAULT_GOAL := run
