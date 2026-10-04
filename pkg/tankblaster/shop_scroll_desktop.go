//go:build (darwin || freebsd || linux || windows) && !android && !ios
// +build darwin freebsd linux windows
// +build !android
// +build !ios

package tankblaster

type shopTouchScrollState struct{}

func (s *GameScene) handleMobileShopListScroll(count int) bool {
	return false
}
