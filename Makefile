# Define paths
LIBRARY_PATH=./lib
CLI_PATH=./cli
BUILD_PATH=./build

# Default target
all: linux android darwin windows

# Target-specific variables and rules
linux: export GOOS=linux
linux: export GOARCH=amd64
linux: export LIBRARY_NAME=libmio.so
linux: build_targets

web: export GOOS=js
web: export GOARCH=wasm
web: export LIBRARY_NAME=mio.wasm
web: build_targets

android: export GOOS=android
android: export GOARCH=arm
android: export LIBRARY_NAME=libmio.so
android: build_targets

darwin: export GOOS=darwin
darwin: export GOARCH=arm64
darwin: export LIBRARY_NAME=libmio.dylib
darwin: build_targets

windows: export GOOS=windows
windows: export GOARCH=amd64
windows: export LIBRARY_NAME=mio.dll
windows: build_targets

# Pattern rule to combine lib, cli, and py for each platform
build_targets: lib cli py dart

# Target for building the shared library
lib:
	@mkdir -p $(BUILD_PATH)
	cd $(LIBRARY_PATH) && CGO_ENABLED=1 go build -o ../$(BUILD_PATH)/$(LIBRARY_NAME) -buildmode=c-shared export.go main.go

cli:
	@mkdir -p $(BUILD_PATH)
	cd $(CLI_PATH) && CGO_ENABLED=1 go build -o ../$(BUILD_PATH)/mio main.go

py:
	@mkdir -p ./py/lib
	cp $(BUILD_PATH)/$(LIBRARY_NAME) ./py/lib
	cd py && python setup.py build_ext --inplace

dart:
	@mkdir -p ./dart/lib
	cp $(BUILD_PATH)/$(LIBRARY_NAME) ./dart/lib

# Target for cleaning up the build artifacts
clean:
	@rm -rf $(BUILD_PATH)

.PHONY: all lib cli py dart clean linux android darwin windows
