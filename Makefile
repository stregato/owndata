# Detect the operating system
ifeq ($(OS),Windows_NT)
    uname := windows
else
    uname := $(shell uname | tr '[:upper:]' '[:lower:]')
endif

# Default target
all: $(uname)

# Define paths
LIBRARY_PATH=./lib
CLI_PATH=./cli
BUILD_PATH=./build


# Target-specific variables and rules
linux: export GOOS=linux
linux: export GOARCH=amd64
linux: export LIBRARY_NAME=libstash.so
linux: export BUILD_MODE=-buildmode=c-shared
linux: build_targets

web: export GOOS=js
web: export GOARCH=wasm
web: export LIBRARY_NAME=stash.wasm
web: export BUILD_MODE=
web: build_targets

android: export GOOS=android
android: export GOARCH=arm
android: export LIBRARY_NAME=libstash.so
android: export BUILD_MODE=-buildmode=c-shared
android: build_targets

darwin: export GOOS=darwin
darwin: export GOARCH=arm64
darwin: export LIBRARY_NAME=libstash.dylib
darwin: export BUILD_MODE=-buildmode=c-shared
darwin: build_targets

windows: export GOOS=windows
windows: export GOARCH=amd64
windows: export LIBRARY_NAME=stash.dll
windows: export BUILD_MODE=-buildmode=c-shared
windows: build_targets

# Pattern rule to combine lib, cli, and py for each platform
build_targets: lib cli py java dart

# Target for building the shared library
lib:
	@mkdir -p $(BUILD_PATH)
	cd $(LIBRARY_PATH) && CGO_ENABLED=1 go build -o ../$(BUILD_PATH)/$(LIBRARY_NAME) $(BUILD_MODE) export.go main.go
	cp $(LIBRARY_PATH)/cfunc.h $(BUILD_PATH)

cli:
	@mkdir -p $(BUILD_PATH)
	cd $(CLI_PATH) && CGO_ENABLED=1 go build -o ../$(BUILD_PATH)/stash main.go

py:
	@mkdir -p ./py/lib
	cp $(BUILD_PATH)/$(LIBRARY_NAME) ./py/lib
	cd py && python setup.py build_ext --inplace

java:
	@mkdir -p ./java/lib
	cp $(BUILD_PATH)/$(LIBRARY_NAME) ./java/lib

dart:
	@mkdir -p ./dart/lib
	cp $(BUILD_PATH)/$(LIBRARY_NAME) ./dart/lib

# Target for cleaning up the build artifacts
clean:
	@rm -rf $(BUILD_PATH)

.PHONY: all lib cli py java dart clean linux android darwin windows
