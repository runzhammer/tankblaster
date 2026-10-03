APP_NAME ?= tankblaster
MODULE ?= github.com/runzhammer/tankblaster
VERSION ?= $(shell sed -n '1p' VERSION 2>/dev/null || printf dev)
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || printf unknown)
DIRTY ?= $(shell git diff --quiet 2>/dev/null || printf '%s' -dirty)
BUILD_TIME ?= $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
BUILDINFO_PKG ?= $(MODULE)/pkg/buildinfo
LD_FLAGS ?= -X $(BUILDINFO_PKG).Version=$(VERSION) -X $(BUILDINFO_PKG).Commit=$(COMMIT)$(DIRTY) -X $(BUILDINFO_PKG).BuildTime=$(BUILD_TIME)
GO_BUILD_FLAGS ?= -trimpath -ldflags "$(LD_FLAGS)"
BUILD_DIR ?= bin
DESKTOP_BIN ?= $(BUILD_DIR)/$(APP_NAME)
SERVER_BIN ?= $(BUILD_DIR)/$(APP_NAME)-server
WINDOWS_BIN ?= $(BUILD_DIR)/$(APP_NAME).exe
WINDOWS_ICON ?= resources/images/tankblaster.ico
WINDOWS_ICON_SYSO ?= tankblaster_windows.syso
ANDROID_SCRIPT ?= ./scripts/build-android.sh
LIVE_TEST_DB ?= .live-test/tankblaster.db
LIVE_TEST_ROUNDS ?= 5
SERVER_DOCKER_IMAGE ?= tankblaster-server:latest
GO ?= go
RSRC ?= $(GO) run github.com/akavel/rsrc@latest

.PHONY: all test run run-server live-test live-test-auto live-test-cp build linux server server-docker windows android android-debug android-release android-env clean help

all: clean linux

help:
	@printf '%s\n' \
		'Targets:' \
		'  make              Build a fresh Linux binary' \
		'  make test         Run Go tests' \
		'  make run          Run the desktop game locally' \
		'  make run-server   Run the headless multiplayer server' \
		'  make live-test    Run server and two clients in the online screen' \
		'  make live-test-auto Run server and two autopiloted online clients' \
		'  make live-test-cp Run server, one human client, and one autopiloted client' \
		'  make build        Build the desktop binary' \
		'  make server       Build the headless multiplayer server' \
		'  make server-docker Build the headless multiplayer server image with Docker or Podman' \
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

live-test:
	@set -eu; \
	mkdir -p .live-test/client1 .live-test/client2; \
	server=; client1=; client2=; \
	trap 'kill $$server $$client1 $$client2 2>/dev/null || true; wait 2>/dev/null || true' INT TERM EXIT; \
	TANKBLASTER_SERVER_DB=$(LIVE_TEST_DB) $(GO) run ./cmd/tankblaster-server -config config/server.example.yaml & server=$$!; \
	sleep 1; \
	TANKBLASTER_DEBUG_ENABLED=1 TANKBLASTER_DEBUG_START_SCENE=online TANKBLASTER_DEBUG_ROUNDS=$(LIVE_TEST_ROUNDS) TANKBLASTER_ONLINE_DISPLAY_NAME='Live 1' XDG_CONFIG_HOME=$$(pwd)/.live-test/client1 $(GO) run . & client1=$$!; \
	TANKBLASTER_DEBUG_ENABLED=1 TANKBLASTER_DEBUG_START_SCENE=online TANKBLASTER_DEBUG_ROUNDS=$(LIVE_TEST_ROUNDS) TANKBLASTER_ONLINE_DISPLAY_NAME='Live 2' XDG_CONFIG_HOME=$$(pwd)/.live-test/client2 $(GO) run . & client2=$$!; \
	wait

live-test-auto:
	@set -eu; \
	mkdir -p .live-test/auto1 .live-test/auto2; \
	server=; client1=; client2=; \
	trap 'kill $$server $$client1 $$client2 2>/dev/null || true; wait 2>/dev/null || true' INT TERM EXIT; \
	TANKBLASTER_SERVER_DB=$(LIVE_TEST_DB) $(GO) run ./cmd/tankblaster-server -config config/server.example.yaml & server=$$!; \
	sleep 1; \
	TANKBLASTER_DEBUG_ENABLED=1 TANKBLASTER_DEBUG_START_SCENE=online TANKBLASTER_DEBUG_ROUNDS=$(LIVE_TEST_ROUNDS) TANKBLASTER_ONLINE_DISPLAY_NAME='CPU 1' TANKBLASTER_ONLINE_AUTO_PLAY=1 XDG_CONFIG_HOME=$$(pwd)/.live-test/auto1 $(GO) run . & client1=$$!; \
	TANKBLASTER_DEBUG_ENABLED=1 TANKBLASTER_DEBUG_START_SCENE=online TANKBLASTER_DEBUG_ROUNDS=$(LIVE_TEST_ROUNDS) TANKBLASTER_ONLINE_DISPLAY_NAME='CPU 2' TANKBLASTER_ONLINE_AUTO_PLAY=1 XDG_CONFIG_HOME=$$(pwd)/.live-test/auto2 $(GO) run . & client2=$$!; \
	wait

live-test-cp:
	@set -eu; \
	mkdir -p .live-test/human .live-test/cpu; \
	server=; client1=; client2=; \
	trap 'kill $$server $$client1 $$client2 2>/dev/null || true; wait 2>/dev/null || true' INT TERM EXIT; \
	TANKBLASTER_SERVER_DB=$(LIVE_TEST_DB) $(GO) run ./cmd/tankblaster-server -config config/server.example.yaml & server=$$!; \
	sleep 1; \
	TANKBLASTER_DEBUG_ENABLED=1 TANKBLASTER_DEBUG_START_SCENE=online TANKBLASTER_DEBUG_ROUNDS=$(LIVE_TEST_ROUNDS) TANKBLASTER_ONLINE_DISPLAY_NAME='Mensch' TANKBLASTER_ONLINE_AUTO_JOIN=1 XDG_CONFIG_HOME=$$(pwd)/.live-test/human $(GO) run . & client1=$$!; \
	TANKBLASTER_DEBUG_ENABLED=1 TANKBLASTER_DEBUG_START_SCENE=online TANKBLASTER_DEBUG_ROUNDS=$(LIVE_TEST_ROUNDS) TANKBLASTER_ONLINE_DISPLAY_NAME='Computer' TANKBLASTER_ONLINE_AUTO_PLAY=1 XDG_CONFIG_HOME=$$(pwd)/.live-test/cpu $(GO) run . & client2=$$!; \
	wait

build: linux

linux:
	mkdir -p $(BUILD_DIR)
	$(GO) build $(GO_BUILD_FLAGS) -o $(DESKTOP_BIN) .

server:
	mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 $(GO) build $(GO_BUILD_FLAGS) -o $(SERVER_BIN) ./cmd/tankblaster-server

server-docker:
	@if [ -n "$${DISTROBOX_ENTER_PATH:-}" ] && command -v distrobox-host-exec >/dev/null 2>&1; then \
		distrobox-host-exec podman build --build-arg VERSION=$(VERSION) --build-arg COMMIT=$(COMMIT)$(DIRTY) --build-arg BUILD_TIME=$(BUILD_TIME) -f docker/server.Dockerfile -t $(SERVER_DOCKER_IMAGE) .; \
	elif command -v docker >/dev/null 2>&1; then \
		docker build --build-arg VERSION=$(VERSION) --build-arg COMMIT=$(COMMIT)$(DIRTY) --build-arg BUILD_TIME=$(BUILD_TIME) -f docker/server.Dockerfile -t $(SERVER_DOCKER_IMAGE) .; \
	elif command -v podman >/dev/null 2>&1; then \
		podman build --build-arg VERSION=$(VERSION) --build-arg COMMIT=$(COMMIT)$(DIRTY) --build-arg BUILD_TIME=$(BUILD_TIME) -f docker/server.Dockerfile -t $(SERVER_DOCKER_IMAGE) .; \
	else \
		printf '%s\n' 'Neither docker nor podman was found in PATH.' >&2; \
		exit 1; \
	fi

windows:
	mkdir -p $(BUILD_DIR)
	$(RSRC) -ico $(WINDOWS_ICON) -o $(WINDOWS_ICON_SYSO)
	GOOS=windows GOARCH=amd64 $(GO) build $(GO_BUILD_FLAGS) -o $(WINDOWS_BIN) .
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
