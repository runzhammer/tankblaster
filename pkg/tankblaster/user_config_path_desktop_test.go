//go:build (darwin || freebsd || linux || windows) && !android && !ios
// +build darwin freebsd linux windows
// +build !android
// +build !ios

package tankblaster

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestDesktopUserConfigPathUsesOSConfigDir(t *testing.T) {
	originalOverride := userConfigDirOverride
	userConfigDirOverride = ""
	t.Cleanup(func() {
		userConfigDirOverride = originalOverride
	})

	configRoot := t.TempDir()
	switch runtime.GOOS {
	case "windows":
		t.Setenv("AppData", configRoot)
	case "darwin":
		t.Setenv("HOME", configRoot)
	default:
		t.Setenv("XDG_CONFIG_HOME", configRoot)
	}

	path, err := userConfigPath()
	if err != nil {
		t.Fatalf("userConfigPath: %v", err)
	}

	var want string
	switch runtime.GOOS {
	case "darwin":
		want = filepath.Join(configRoot, "Library", "Application Support", userConfigAppDirName, userConfigFileName)
	default:
		want = filepath.Join(configRoot, userConfigAppDirName, userConfigFileName)
	}
	if path != want {
		t.Fatalf("userConfigPath = %q, want %q", path, want)
	}
}
