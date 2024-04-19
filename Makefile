# Define variables
LIBRARY_NAME=mio.so
LIBRARY_PATH=./lib
CLI_PATH=./cli
BUILD_PATH=./build

# Default target
all: lib cli

# Target for building the shared library
lib:
	@mkdir -p $(BUILD_PATH)
	cd $(LIBRARY_PATH) && go build -o ../$(BUILD_PATH)/$(LIBRARY_NAME) -buildmode=c-shared main.go

cli:
	@mkdir -p $(BUILD_PATH)
	cd $(CLI_PATH) && go build -o ../$(BUILD_PATH)/mio main.go

# Target for cleaning up the build artifacts
clean:
	@rm -rf $(BUILD_PATH)

.PHONY: all lib cli clean
