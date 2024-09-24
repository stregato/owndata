.PHONY: all init clean lib py java package

BUILD_PATH=./build
LIBRARY_PATH=./lib
CLI_PATH=./cli

GO_FILES := $(wildcard $(LIBRARY_PATH)/*.go)
CLI_GO_FILES := $(wildcard $(CLI_PATH)/*.go)

all: lib py java dart

init:
	echo "Running init"
	mkdir -p $(BUILD_PATH)

$(BUILD_PATH)/linux_amd64/libstash.so: init $(GO_FILES)
	echo "Building for Linux AMD64"
	GOOS=linux GOARCH=amd64 CC=x86_64-linux-gnu-gcc CXX=x86_64-linux-gnu-g++ $(MAKE) build_targets GOOS=linux GOARCH=amd64 LIBRARY_NAME=libstash.so CLI_NAME=stash

$(BUILD_PATH)/darwin_amd64/libstash.dylib: init $(GO_FILES)
	echo "Building for Darwin AMD64"
	GOOS=darwin GOARCH=amd64 $(MAKE) build_targets GOOS=darwin GOARCH=amd64 LIBRARY_NAME=libstash.dylib CLI_NAME=stash

$(BUILD_PATH)/darwin_arm64/libstash.dylib: init $(GO_FILES)
	echo "Building for Darwin ARM64"
	GOOS=darwin GOARCH=arm64 $(MAKE) build_targets GOOS=darwin GOARCH=arm64 LIBRARY_NAME=libstash.dylib CLI_NAME=stash

$(BUILD_PATH)/windows_amd64/stash.dll: init $(GO_FILES)
	echo "Building for Windows AMD64"
	GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc $(MAKE) build_targets GOOS=windows GOARCH=amd64 LIBRARY_NAME=stash.dll CLI_NAME=stash.exe

NDK_PATH := $(shell ls -d $(HOME)/Library/Android/sdk/ndk/* | sort -V | tail -n 1)
$(BUILD_PATH)/android_arm64/libstash.so: init $(GO_FILES)
	echo "Building for Android ARM64"
	CC=$(NDK_PATH)/toolchains/llvm/prebuilt/darwin-x86_64/bin/aarch64-linux-android21-clang \
	GOOS=android GOARCH=arm64 $(MAKE) build_targets GOOS=android GOARCH=arm64 LIBRARY_NAME=libstash.so CLI_NAME=stash

build_targets:
	echo "GOOS=$(GOOS), GOARCH=$(GOARCH), LIBRARY_NAME=$(LIBRARY_NAME), CLI_NAME=$(CLI_NAME), CC=${CC}, CGO_CFLAGS=$(CGO_CFLAGS), CGO_LDFLAGS=$(CGO_LDFLAGS)"
	mkdir -p $(BUILD_PATH)/$(GOOS)_$(GOARCH)
	cd $(LIBRARY_PATH) && GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=1 CC=$(CC) go build -o ../$(BUILD_PATH)/$(GOOS)_$(GOARCH)/$(LIBRARY_NAME) -buildmode=c-shared export.go main.go
	cd $(CLI_PATH) && GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=1 CC=$(CC) go build -o ../$(BUILD_PATH)/$(GOOS)_$(GOARCH)/$(CLI_NAME) main.go

lib: $(BUILD_PATH)/linux_amd64/libstash.so \
     $(BUILD_PATH)/darwin_amd64/libstash.dylib \
     $(BUILD_PATH)/darwin_arm64/libstash.dylib \
     $(BUILD_PATH)/windows_amd64/stash.dll \
     $(BUILD_PATH)/android_arm64/libstash.so

py: lib
	echo "Building Python bindings"
	cd ./py && ./build.sh

java: lib
	echo "Building Java bindings"
	cd ./java && mvn clean install

dart: lib
	echo "Building Dart bindings"
	cd ./dart && ./build.sh

clean:
	echo "Cleaning up"
	rm -rf $(BUILD_PATH)
	rm -rf py/stash/_libs

package:
	echo "Packaging all directories in $(BUILD_PATH)"
	@for dir in $(shell find $(BUILD_PATH) -mindepth 1 -maxdepth 1 -type d); do \
		dirname=$$(basename $$dir); \
		zip -r $(BUILD_PATH)/$$dirname.zip $$dir; \
	done