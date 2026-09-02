#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ANDROID_DIR="$ROOT_DIR/android"
BUILD_TYPE="${1:-debug}"
ANDROID_HOME="${ANDROID_HOME:-/home/reznor/Android/Sdk}"
ANDROID_SDK_ROOT="${ANDROID_SDK_ROOT:-$ANDROID_HOME}"
JAVA_HOME="${JAVA_HOME:-/home/reznor/.local/share/tankblaster-android/jdk-21.0.12.1+1}"
if [ ! -x "$JAVA_HOME/bin/java" ]; then
	JAVA_HOME="/home/reznor/.local/share/tankblaster-android/jdk-21.0.12.1+1"
fi
GOBIN="$(go env GOBIN)"
GOPATH="$(go env GOPATH)"
GO_CGO_LDFLAGS="$(go env CGO_LDFLAGS)"

if [ -z "$GOBIN" ]; then
    GOBIN="$GOPATH/bin"
fi

export ANDROID_HOME
export ANDROID_SDK_ROOT
export JAVA_HOME
export CGO_LDFLAGS="${CGO_LDFLAGS:-$GO_CGO_LDFLAGS -Wl,-z,max-page-size=16384 -Wl,-z,common-page-size=16384}"
export PATH="$JAVA_HOME/bin:$ANDROID_HOME/platform-tools:$ANDROID_HOME/cmdline-tools/latest/bin:$GOBIN:/home/reznor/.local/share/vscodium-distrobox-bin:$PATH"

EBITEN_VERSION="$(go list -m -f '{{.Version}}' github.com/hajimehoshi/ebiten/v2)"
VERSION="$(sed -n '1p' "$ROOT_DIR/VERSION" 2>/dev/null || printf dev)"
VERSION="$(printf '%s' "$VERSION" | tr -d '[:space:]')"
if [ -z "$VERSION" ]; then
	VERSION=dev
fi
COMMIT="$(git -C "$ROOT_DIR" rev-parse --short HEAD 2>/dev/null || printf unknown)"
DIRTY=
if ! git -C "$ROOT_DIR" diff --quiet 2>/dev/null; then
	DIRTY=-dirty
fi
BUILD_TIME="$(date -u '+%Y-%m-%dT%H:%M:%SZ')"
GO_BUILD_LDFLAGS="-X github.com/runzhammer/gamedemo/pkg/buildinfo.Version=$VERSION -X github.com/runzhammer/gamedemo/pkg/buildinfo.Commit=$COMMIT$DIRTY -X github.com/runzhammer/gamedemo/pkg/buildinfo.BuildTime=$BUILD_TIME"
if ! command -v ebitenmobile >/dev/null 2>&1; then
	go install "github.com/hajimehoshi/ebiten/v2/cmd/ebitenmobile@${EBITEN_VERSION}"
fi

mkdir -p "$ANDROID_DIR/app/libs" "$ROOT_DIR/dist"
mkdir -p "$ROOT_DIR/tmp_tools"
export TMPDIR="${TMPDIR:-$ROOT_DIR/tmp_tools}"

if [ -z "${CONTAINER_ID:-}" ] && [ -x /app/bin/host-spawn ]; then
	/app/bin/host-spawn -no-pty \
		distrobox enter --no-tty -n golang-nvidia -- \
		env \
		JAVA_HOME="$JAVA_HOME" \
		ANDROID_HOME="$ANDROID_HOME" \
		ANDROID_SDK_ROOT="$ANDROID_SDK_ROOT" \
		CGO_LDFLAGS="$CGO_LDFLAGS" \
		GO_BUILD_LDFLAGS="$GO_BUILD_LDFLAGS" \
		PATH="$JAVA_HOME/bin:$ANDROID_HOME/platform-tools:$ANDROID_HOME/cmdline-tools/latest/bin:$GOBIN:/usr/bin:/usr/sbin:/bin:/sbin:/usr/local/bin:/usr/local/sbin" \
		sh -c 'cd "$1" && ebitenmobile bind -target android -javapkg com.runzhammer.tankblaster -ldflags "$GO_BUILD_LDFLAGS" -o "$2" ./mobile' \
		sh "$ROOT_DIR" "$ANDROID_DIR/app/libs/tankblaster.aar"
else
	cd "$ROOT_DIR"
	ebitenmobile bind \
		-target android \
		-javapkg com.runzhammer.tankblaster \
		-ldflags "$GO_BUILD_LDFLAGS" \
		-o "$ANDROID_DIR/app/libs/tankblaster.aar" \
		./mobile
fi

case "$BUILD_TYPE" in
	debug)
		GRADLE_TASK=assembleDebug
		APK_SOURCE="$ANDROID_DIR/app/build/outputs/apk/debug/app-debug.apk"
		APK_TARGET="$ROOT_DIR/dist/tankblaster-debug.apk"
		;;
	release|normal)
		GRADLE_TASK=assembleRelease
		APK_SOURCE="$ANDROID_DIR/app/build/outputs/apk/release/app-release.apk"
		APK_TARGET="$ROOT_DIR/dist/tankblaster-release.apk"
		;;
	*)
		printf 'Unknown Android build type: %s\n' "$BUILD_TYPE" >&2
		printf 'Use: %s [debug|release]\n' "$0" >&2
		exit 2
		;;
esac

"$ANDROID_DIR/gradlew" -p "$ANDROID_DIR" "$GRADLE_TASK"

cp "$APK_SOURCE" "$APK_TARGET"
printf 'APK: %s\n' "$APK_TARGET"
