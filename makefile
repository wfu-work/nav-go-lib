# Define variables
LIB_NAME = libNavMerge
BUILD_DIR = build
MACOS_LIB = $(BUILD_DIR)/$(LIB_NAME).dylib
WINDOWS_LIB = $(BUILD_DIR)/$(LIB_NAME).dll
LINUX_LIB = $(BUILD_DIR)/$(LIB_NAME).so
DARWIN_LIB = $(BUILD_DIR)/$(LIB_NAME).dylib

# Get the current OS
CURRENT_OS := $(shell uname -s)

# Default target: Build for the current OS
.PHONY: all
all: build-current-platform

# Build for macOS
.PHONY: build-macos
build-macos:
	@echo "Building for macOS..."
	go build -buildmode=c-shared -o $(MACOS_LIB)

# Build for Windows
.PHONY: build-windows
build-windows:
	@echo "Building for Windows..."
	GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc CGO_ENABLED=1 go build -buildmode=c-shared -ldflags="-s -w" -buildvcs=false -o $(WINDOWS_LIB)

# Build for Linux
.PHONY: build-linux
build-linux:
	@echo "Building for Linux..."
	GOOS=linux GOARCH=amd64 CC=x86_64-linux-gnu-gcc CGO_ENABLED=1 go build -buildmode=c-shared -o $(LINUX_LIB)

# Build for Darwin (macOS)
.PHONY: build-darwin
build-darwin:
	@echo "Building for Darwin (macOS)..."
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 go build -buildmode=c-shared -o $(DARWIN_LIB)

# Build for the current platform based on the operating system
.PHONY: build-current-platform
build-current-platform:
	@if [ "$(CURRENT_OS)" == "Darwin" ]; then \
		echo "Building for macOS..."; \
		make build-macos; \
	elif [ "$(CURRENT_OS)" == "Linux" ]; then \
		echo "Building for Linux..."; \
		make build-linux; \
	elif [ "$(CURRENT_OS)" == "CYGWIN" ] || [ "$(CURRENT_OS)" == "MINGW" ]; then \
		echo "Building for Windows..."; \
		make build-windows; \
	else \
		echo "Unsupported OS: $(CURRENT_OS)"; \
		exit 1; \
	fi

# Clean the build directory
.PHONY: clean
clean:
	@echo "Cleaning build directory..."
	rm -rf $(BUILD_DIR)/