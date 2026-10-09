//go:build !js

package tankblaster

import (
	"os/exec"
	"runtime"
)

func openDefaultBrowser(url string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", url).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "android":
		return exec.Command("am", "start", "-a", "android.intent.action.VIEW", "-d", url).Start()
	default:
		return exec.Command("xdg-open", url).Start()
	}
}
