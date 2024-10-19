# Makefile for building and running the Go snapshot service

# Directories
BINDIR := ./bin
SRCDIR := ./cmd/snapshot_creator
CONFIGDIR := ./config

# Build binary for snapshot_creator
snapshot-creator:
	@echo "Building snapshot_creator..."
	@mkdir -p $(BINDIR)
	@go build -o $(BINDIR)/snapshot_creator $(SRCDIR)

# Build binary for file_server
file-server:
	@echo "Building file_server..."
	@go build -o $(BINDIR)/file_server ./cmd/file_server

# Run snapshot_creator with a specific config path
run-snapshot-creator:
	@echo "Running snapshot_creator..."
	@CONFIG_PATH=$(CONFIGDIR) $(BINDIR)/snapshot_creator

# Clean up binaries
clean:
	@echo "Cleaning up..."
	@rm -rf $(BINDIR)

# Default target
all: snapshot-creator file-server
