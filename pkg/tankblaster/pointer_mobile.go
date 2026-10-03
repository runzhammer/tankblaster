//go:build android || ios
// +build android ios

package tankblaster

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/runzhammer/tankblaster/pkg/core"
)

var lastPrimaryTouchPosition image.Point

func primaryPointerPosition() (int, int) {
	if ids := ebiten.AppendTouchIDs(nil); len(ids) > 0 {
		x, y := ebiten.TouchPosition(ids[0])
		lastPrimaryTouchPosition = image.Pt(x, y)
		return gamePointerPosition(x, y)
	}
	if ids := inpututil.AppendJustReleasedTouchIDs(nil); len(ids) > 0 {
		x, y := ebiten.TouchPosition(ids[0])
		if x != 0 || y != 0 {
			lastPrimaryTouchPosition = image.Pt(x, y)
		}
	}
	return gamePointerPosition(lastPrimaryTouchPosition.X, lastPrimaryTouchPosition.Y)
}

func primaryPointerPressed() bool {
	return len(ebiten.AppendTouchIDs(nil)) > 0
}

func primaryPointerJustPressed() bool {
	ids := inpututil.AppendJustPressedTouchIDs(nil)
	if len(ids) == 0 {
		return false
	}
	x, y := ebiten.TouchPosition(ids[0])
	lastPrimaryTouchPosition = image.Pt(x, y)
	return true
}

func primaryPointerJustReleased() bool {
	ids := inpututil.AppendJustReleasedTouchIDs(nil)
	if len(ids) == 0 {
		return false
	}
	x, y := ebiten.TouchPosition(ids[0])
	if x != 0 || y != 0 {
		lastPrimaryTouchPosition = image.Pt(x, y)
	}
	return true
}

func primaryPointerPressedInRect(r image.Rectangle) bool {
	for _, id := range ebiten.AppendTouchIDs(nil) {
		x, y := ebiten.TouchPosition(id)
		gx, gy := gamePointerPosition(x, y)
		if image.Pt(gx, gy).In(r) {
			lastPrimaryTouchPosition = image.Pt(x, y)
			return true
		}
	}
	return false
}

func primaryPointerIsTouch() bool {
	return true
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
