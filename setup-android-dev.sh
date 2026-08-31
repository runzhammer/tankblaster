#!/usr/bin/env bash
set -euo pipefail

echo "=== Tank Blaster Android Development Setup ==="

# ------------------------------------------------------------
# Configuration
# ------------------------------------------------------------

ANDROID_HOME="${ANDROID_HOME:-$HOME/Android/Sdk}"

CMDLINE_TOOLS_VERSION="15859902"
CMDLINE_TOOLS_ZIP="commandlinetools-linux-${CMDLINE_TOOLS_VERSION}_latest.zip"
CMDLINE_TOOLS_URL="https://dl.google.com/android/repository/${CMDLINE_TOOLS_ZIP}"

# Known stable SDK versions.
ANDROID_PLATFORM="android-36"
ANDROID_BUILD_TOOLS="36.0.0"

# Android NDK r27d (LTS)
ANDROID_NDK="27.3.13750724"

echo
echo "Android SDK location:"
echo "  $ANDROID_HOME"
echo


# ------------------------------------------------------------
# 1. Fedora packages
# ------------------------------------------------------------

echo "=== Installing Fedora packages ==="

sudo dnf install -y \
    java-25-openjdk-devel \
    git \
    curl \
    wget \
    unzip \
    zip \
    tar \
    gzip \
    make \
    gcc \
    gcc-c++ \
    pkgconf-pkg-config


# ------------------------------------------------------------
# 2. JAVA_HOME
# ------------------------------------------------------------

echo
echo "=== Configuring Java ==="

JAVA_BIN="$(readlink -f "$(command -v javac)")"
JAVA_HOME="$(dirname "$(dirname "$JAVA_BIN")")"

export JAVA_HOME
export PATH="$JAVA_HOME/bin:$PATH"

echo "JAVA_HOME=$JAVA_HOME"

java -version
javac -version


# ------------------------------------------------------------
# 3. Android SDK Command Line Tools
# ------------------------------------------------------------

echo
echo "=== Installing Android command line tools ==="

mkdir -p "$ANDROID_HOME/cmdline-tools"

TMPDIR="$(mktemp -d)"
trap 'rm -rf "$TMPDIR"' EXIT

cd "$TMPDIR"

curl -fL \
    "$CMDLINE_TOOLS_URL" \
    -o "$CMDLINE_TOOLS_ZIP"

unzip -q "$CMDLINE_TOOLS_ZIP"

rm -rf "$ANDROID_HOME/cmdline-tools/latest"
mkdir -p "$ANDROID_HOME/cmdline-tools/latest"

mv cmdline-tools/* "$ANDROID_HOME/cmdline-tools/latest/"


# ------------------------------------------------------------
# 4. Environment
# ------------------------------------------------------------

export ANDROID_HOME
export ANDROID_SDK_ROOT="$ANDROID_HOME"

export PATH="$PATH:$ANDROID_HOME/cmdline-tools/latest/bin"
export PATH="$PATH:$ANDROID_HOME/platform-tools"

SDKMANAGER="$ANDROID_HOME/cmdline-tools/latest/bin/sdkmanager"

echo
echo "sdkmanager:"
"$SDKMANAGER" --version


# ------------------------------------------------------------
# 5. Accept Android licenses
# ------------------------------------------------------------

echo
echo "=== Accepting Android SDK licenses ==="

yes | "$SDKMANAGER" --licenses >/dev/null || true


# ------------------------------------------------------------
# 6. Android SDK / Build Tools / NDK
# ------------------------------------------------------------

echo
echo "=== Installing Android SDK components ==="

"$SDKMANAGER" \
    "platform-tools" \
    "platforms;${ANDROID_PLATFORM}" \
    "build-tools;${ANDROID_BUILD_TOOLS}" \
    "ndk;${ANDROID_NDK}"


# ------------------------------------------------------------
# 7. Persist environment variables
# ------------------------------------------------------------

echo
echo "=== Writing environment to ~/.bashrc ==="

BEGIN_MARKER="# --- TankBlaster Android SDK BEGIN ---"
END_MARKER="# --- TankBlaster Android SDK END ---"

# Remove an older block if this script was already executed.
sed -i \
    "/^${BEGIN_MARKER}$/,/^${END_MARKER}$/d" \
    "$HOME/.bashrc"

cat >> "$HOME/.bashrc" <<EOF

${BEGIN_MARKER}
export JAVA_HOME="${JAVA_HOME}"
export ANDROID_HOME="${ANDROID_HOME}"
export ANDROID_SDK_ROOT="\$ANDROID_HOME"
export PATH="\$PATH:\$ANDROID_HOME/cmdline-tools/latest/bin"
export PATH="\$PATH:\$ANDROID_HOME/platform-tools"
export PATH="\$PATH:\$HOME/go/bin"
${END_MARKER}
EOF


# ------------------------------------------------------------
# 8. Ebitengine mobile build tool
# ------------------------------------------------------------

echo
echo "=== Installing ebitenmobile ==="

# If this script is run from a Go project that already uses Ebitengine,
# install the matching ebitenmobile version.
EBITEN_VERSION=""

if [ -f "go.mod" ]; then
    EBITEN_VERSION="$(
        go list -m -f '{{.Version}}' \
            github.com/hajimehoshi/ebiten/v2 2>/dev/null || true
    )"
fi

if [ -n "$EBITEN_VERSION" ]; then
    echo "Project uses Ebitengine $EBITEN_VERSION"
    go install \
        "github.com/hajimehoshi/ebiten/v2/cmd/ebitenmobile@${EBITEN_VERSION}"
else
    echo "No Ebitengine version detected from current go.mod."
    echo "Installing latest ebitenmobile."
    go install \
        github.com/hajimehoshi/ebiten/v2/cmd/ebitenmobile@latest
fi


# ------------------------------------------------------------
# 9. Verification
# ------------------------------------------------------------

echo
echo "============================================================"
echo " Verification"
echo "============================================================"

echo
echo "--- Go ---"
go version

echo
echo "--- Java ---"
java -version 2>&1 | head -n 1

echo
echo "--- Android sdkmanager ---"
sdkmanager --version

echo
echo "--- adb ---"
adb version | head -n 2

echo
echo "--- ebitenmobile ---"
command -v ebitenmobile

echo
echo "--- Installed Android components ---"
sdkmanager --list_installed | grep -E \
    'platform-tools|build-tools|platforms;android|ndk;' || true


echo
echo "============================================================"
echo " Android development environment is ready."
echo
echo "ANDROID_HOME=$ANDROID_HOME"
echo "JAVA_HOME=$JAVA_HOME"
echo
echo "Open a NEW shell or run:"
echo
echo "  source ~/.bashrc"
echo
echo "Then verify:"
echo
echo "  go version"
echo "  java -version"
echo "  sdkmanager --version"
echo "  adb version"
echo "  ebitenmobile"
echo "============================================================"
