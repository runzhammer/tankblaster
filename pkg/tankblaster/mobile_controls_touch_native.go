//go:build android || ios

package tankblaster

import "github.com/hajimehoshi/ebiten/v2"

func mobileControlTouchPosition(id ebiten.TouchID) (int, int) {
	return ebiten.TouchPosition(id)
}
