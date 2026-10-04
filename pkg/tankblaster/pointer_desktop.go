//go:build (darwin || freebsd || linux || windows) && !android && !ios
// +build darwin freebsd linux windows
// +build !android
// +build !ios

package tankblaster

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

func primaryPointerPosition() (int, int) {
	return ebiten.CursorPosition()
}

func primaryPointerPressed() bool {
	return ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
}

func primaryPointerJustPressed() bool {
	return inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
}

func primaryPointerJustReleased() bool {
	return inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft)
}

func primaryPointerPressedInRect(r image.Rectangle) bool {
	if !primaryPointerPressed() {
		return false
	}
	x, y := primaryPointerPosition()
	return image.Pt(x, y).In(r)
}

func primaryPointerIsTouch() bool {
	return false
}
