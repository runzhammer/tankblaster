APP_NAME ?= tankblaster
MODULE ?= github.com/runzhammer/gamedemo
BUILD_DIR ?= bin
DESKTOP_BIN ?= $(BUILD_DIR)/$(APP_NAME)
SERVER_BIN ?= $(BUILD_DIR)/$(APP_NAME)-server
WINDOWS_BIN ?= $(BUILD_DIR)/$(APP_NAME).exe
WINDOWS_ICON ?= resources/images/tankblaster.ico
WINDOWS_ICON_SYSO ?= tankblaster_windows.syso
ANDROID_SCRIPT ?= ./scripts/build-android.sh
GO ?= go
RSRC ?= $(GO) run github.com/akavel/rsrc@latest

.PHONY: all test run run-server build linux server windows android android-debug android-release android-env clean help

all: clean linux

help:
	@printf '%s\n' \
		'Targets:' \
		'  make              Build a fresh Linux binary' \
		'  make test         Run Go tests' \
		'  make run          Run the desktop game locally' \
		'  make run-server   Run the headless multiplayer server' \
		'  make build        Build the desktop binary' \
		'  make server       Build the headless multiplayer server' \
		'  make windows      Build a Windows EXE with app icon' \
		'  make android      Build Android debug APK' \
		'  make android-debug Build Android debug APK' \
		'  make android-release Build Android release APK' \
		'  make android-env  Print required Android build environment' \
		'  make clean        Remove build artifacts'

test:
	$(GO) test ./...

run:
	$(GO) run .

run-server:
	$(GO) run ./cmd/tankblaster-server -config config/server.example.yaml

build: linux

linux:
	mkdir -p $(BUILD_DIR)
	$(GO) build -o $(DESKTOP_BIN) .

server:
	mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 $(GO) build -o $(SERVER_BIN) ./cmd/tankblaster-server

windows:
	mkdir -p $(BUILD_DIR)
	$(RSRC) -ico $(WINDOWS_ICON) -o $(WINDOWS_ICON_SYSO)
	GOOS=windows GOARCH=amd64 $(GO) build -o $(WINDOWS_BIN) .
	rm -f $(WINDOWS_ICON_SYSO)

android: android-debug

android-debug:
	$(ANDROID_SCRIPT) debug

android-release:
	$(ANDROID_SCRIPT) release

android-env:
	@printf 'ANDROID_HOME=%s\n' "$${ANDROID_HOME:-}"
	@printf 'ANDROID_SDK_ROOT=%s\n' "$${ANDROID_SDK_ROOT:-}"
	@printf 'ANDROID_NDK_HOME=%s\n' "$${ANDROID_NDK_HOME:-}"
	@printf 'JAVA_HOME=%s\n' "$${JAVA_HOME:-}"
	@printf 'Debug APK: dist/tankblaster-debug.apk\n'
	@printf 'Release APK: dist/tankblaster-release.apk\n'

clean:
	rm -rf $(BUILD_DIR)
