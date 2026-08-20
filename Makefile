APP_NAME ?= tankblaster
MODULE ?= github.com/runzhammer/gamedemo
BUILD_DIR ?= bin
DESKTOP_BIN ?= $(BUILD_DIR)/$(APP_NAME)
ANDROID_DIR ?= $(BUILD_DIR)/android
ANDROID_AAR ?= $(ANDROID_DIR)/$(APP_NAME).aar
ANDROID_TARGET ?= android
ANDROID_API ?= 21
ANDROID_JAVAPKG ?= com.runzhammer.tankblaster
MOBILE_PKG ?= ./mobile
GO ?= go
EBITENMOBILE ?= $(GO) run github.com/hajimehoshi/ebiten/v2/cmd/ebitenmobile

.PHONY: all test run build linux android android-aar android-env clean help

all: clean linux

help:
	@printf '%s\n' \
		'Targets:' \
		'  make              Build a fresh Linux binary' \
		'  make test         Run Go tests' \
		'  make run          Run the desktop game locally' \
		'  make build        Build the desktop binary' \
		'  make android      Build Android AAR via ebitenmobile' \
		'  make android-env  Print required Android build environment' \
		'  make clean        Remove build artifacts'

test:
	$(GO) test ./...

run:
	$(GO) run .

build: linux

linux:
	mkdir -p $(BUILD_DIR)
	$(GO) build -o $(DESKTOP_BIN) .

android: android-aar

android-aar:
	mkdir -p $(ANDROID_DIR)
	$(EBITENMOBILE) bind \
		-target $(ANDROID_TARGET) \
		-androidapi $(ANDROID_API) \
		-javapkg $(ANDROID_JAVAPKG) \
		-o $(ANDROID_AAR) \
		$(MOBILE_PKG)

android-env:
	@printf 'ANDROID_HOME=%s\n' "$${ANDROID_HOME:-}"
	@printf 'ANDROID_SDK_ROOT=%s\n' "$${ANDROID_SDK_ROOT:-}"
	@printf 'ANDROID_NDK_HOME=%s\n' "$${ANDROID_NDK_HOME:-}"
	@printf 'JAVA_HOME=%s\n' "$${JAVA_HOME:-}"
	@printf 'Output AAR: %s\n' "$(ANDROID_AAR)"

clean:
	rm -rf $(BUILD_DIR)
