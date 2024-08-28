.PHONY: all init linux_amd64 linux_arm64 darwin_amd64 darwin_arm64 windows_amd64 android_arm64 build_targets clean

BUILD_PATH=./build
LIBRARY_PATH=./lib
CLI_PATH=./cli

all: lib py 

init:
	echo "Running init"
	mkdir -p $(BUILD_PATH)

linux_amd64: init
	echo "Building for Linux AMD64"
	GOOS=linux GOARCH=amd64 CC=x86_64-linux-gnu-gcc CXX=x86_64-linux-gnu-g++ $(MAKE) build_targets LIBRARY_NAME=libstash.so CLI_NAME=stash

darwin_amd64: init
	echo "Building for Darwin AMD64"
	GOOS=darwin GOARCH=amd64 $(MAKE) build_targets LIBRARY_NAME=libstash.dylib CLI_NAME=stash

darwin_arm64: init
	echo "Building for Darwin ARM64"
	GOOS=darwin GOARCH=arm64 $(MAKE) build_targets LIBRARY_NAME=libstash.dylib CLI_NAME=stash

windows_amd64: init
	echo "Building for Windows AMD64"
	GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc $(MAKE) build_targets LIBRARY_NAME=stash.dll CLI_NAME=stash.exe

NDK_PATH := $(shell ls -d $(HOME)/Library/Android/sdk/ndk/* | sort -V | tail -n 1)
android_arm64: init
	echo "Building for Android ARM64"
	CC=$(NDK_PATH)/toolchains/llvm/prebuilt/darwin-x86_64/bin/aarch64-linux-android21-clang \
	GOOS=android GOARCH=arm64 $(MAKE) build_targets LIBRARY_NAME=libstash.so CLI_NAME=stash

build_targets:
	echo "GOOS=$(GOOS), GOARCH=$(GOARCH), LIBRARY_NAME=$(LIBRARY_NAME), CLI_NAME=$(CLI_NAME), CC=${CC}, CGO_CFLAGS=$(CGO_CFLAGS), CGO_LDFLAGS=$(CGO_LDFLAGS)"
	mkdir -p $(BUILD_PATH)/$(GOOS)_$(GOARCH)
	cd $(LIBRARY_PATH) && GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=1 CC=$(CC) go build -o ../$(BUILD_PATH)/$(GOOS)_$(GOARCH)/$(LIBRARY_NAME) -buildmode=c-shared export.go main.go
	cd $(CLI_PATH) && GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=1 CC=$(CC) go build -o ../$(BUILD_PATH)/$(GOOS)_$(GOARCH)/$(CLI_NAME) main.go

lib: linux_amd64 linux_arm64 darwin_amd64 darwin_arm64 windows_amd64 android_arm64

py: lib
	echo "Building Python bindings"
	cd ./py && ./build.sh

java: lib
	echo "Building Java bindings"
	mkdir -p java/stash/src/main/resources
	cp -R build/* java/lib
	cd ./java/stash && mvn clean install

clean:
	echo "Cleaning up"
	rm -rf $(BUILD_PATH)
	rm -rf py/stash/_libs

.PHONY: py lib clean linux_amd64 linux_arm64 darwin_amd64 darwin_arm64 windows_amd64 android_arm64