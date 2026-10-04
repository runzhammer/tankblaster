//go:build js

package tankblaster

import (
	"image"
	"math"
	"syscall/js"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/runzhammer/tankblaster/pkg/core"
)

var lastPrimaryTouchPosition image.Point

func primaryPointerPosition() (int, int) {
	if !mobileControlsEnabled() {
		return ebiten.CursorPosition()
	}
	if ids := ebiten.AppendTouchIDs(nil); len(ids) > 0 {
		x, y := webTouchPosition(ids[0])
		lastPrimaryTouchPosition = image.Pt(x, y)
		return gamePointerPosition(x, y)
	}
	if ids := inpututil.AppendJustReleasedTouchIDs(nil); len(ids) > 0 {
		x, y := webTouchPosition(ids[0])
		if x != 0 || y != 0 {
			lastPrimaryTouchPosition = image.Pt(x, y)
		}
	}
	return gamePointerPosition(lastPrimaryTouchPosition.X, lastPrimaryTouchPosition.Y)
}

func primaryPointerPressed() bool {
	if !mobileControlsEnabled() {
		return ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
	}
	return len(ebiten.AppendTouchIDs(nil)) > 0
}

func primaryPointerJustPressed() bool {
	if !mobileControlsEnabled() {
		return inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
	}
	ids := inpututil.AppendJustPressedTouchIDs(nil)
	if len(ids) == 0 {
		return false
	}
	x, y := webTouchPosition(ids[0])
	lastPrimaryTouchPosition = image.Pt(x, y)
	return true
}

func primaryPointerJustReleased() bool {
	if !mobileControlsEnabled() {
		return inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft)
	}
	ids := inpututil.AppendJustReleasedTouchIDs(nil)
	if len(ids) == 0 {
		return false
	}
	x, y := webTouchPosition(ids[0])
	if x != 0 || y != 0 {
		lastPrimaryTouchPosition = image.Pt(x, y)
	}
	return true
}

func primaryPointerPressedInRect(r image.Rectangle) bool {
	if !mobileControlsEnabled() {
		if !primaryPointerPressed() {
			return false
		}
		x, y := primaryPointerPosition()
		return image.Pt(x, y).In(r)
	}
	for _, id := range ebiten.AppendTouchIDs(nil) {
		x, y := webTouchPosition(id)
		gx, gy := gamePointerPosition(x, y)
		if image.Pt(gx, gy).In(r) {
			lastPrimaryTouchPosition = image.Pt(x, y)
			return true
		}
	}
	return false
}

func primaryPointerIsTouch() bool {
	return mobileControlsEnabled()
}

func webTouchPosition(id ebiten.TouchID) (int, int) {
	x, y := ebiten.TouchPosition(id)
	return webCanvasPosition(x, y)
}

func webCanvasPosition(x, y int) (int, int) {
	canvas := js.Global().Get("document").Call("querySelector", "canvas")
	if canvas.IsNull() || canvas.IsUndefined() {
		return x, y
	}
	rect := canvas.Call("getBoundingClientRect")
	left := rect.Get("left").Float()
	top := rect.Get("top").Float()
	return int(math.Round(float64(x) - left)), int(math.Round(float64(y) - top))
}

func gamePointerPosition(x, y int) (int, int) {
	viewport := core.GameViewport()
	if viewport.Empty() {
		return x, y
	}
	cfg := core.Config().Screen
	gx := float64(x-viewport.Min.X) * cfg.Width / float64(viewport.Dx())
	gy := float64(y-viewport.Min.Y) * cfg.Height / float64(viewport.Dy())
	return int(gx + 0.5), int(gy + 0.5)
}
