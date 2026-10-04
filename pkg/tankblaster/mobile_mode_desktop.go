//go:build (darwin || freebsd || linux || windows) && !android && !ios

package tankblaster

func mobileControlsEnabled() bool {
	return false
}
