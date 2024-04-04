# Define variables
LIBRARY_NAME=mio.so
LIBRARY_PATH=./pkg
BUILD_PATH=./build

# Default target
all: build

# Target for building the shared library
build:
	@mkdir -p $(BUILD_PATH)
	go build -o $(BUILD_PATH)/$(LIBRARY_NAME) -buildmode=c-shared $(LIBRARY_PATH)/main.go

# Target for cleaning up the build artifacts
clean:
	@rm -rf $(BUILD_PATH)

.PHONY: all build clean
