//go:build js && !android && !ios
// +build js,!android,!ios

package tankblaster

import (
	"os"
	"path/filepath"
)

func userConfigPath() (string, error) {
	if userConfigDirOverride != "" {
		return filepath.Join(userConfigDirOverride, userConfigFileName), nil
	}

	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, userConfigAppDirName, userConfigFileName), nil
}
