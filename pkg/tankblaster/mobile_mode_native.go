//go:build android || ios

package tankblaster

func mobileControlsEnabled() bool {
	return true
}
