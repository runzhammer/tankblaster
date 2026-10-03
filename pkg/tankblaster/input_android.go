package tankblaster

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/runzhammer/tankblaster/pkg/core"
)

func Begin() bool {
	for _, id := range ebiten.AppendTouchIDs(nil) {
		x, y := ebiten.TouchPosition(id)
		if imagePointInGameViewport(x, y) {
			return true
		}
	}
	return false
}

func MoveLeft() bool {
	return false
}

func MoveRight() bool {
	return false
}

func RotateLeft() bool {
	viewport := core.GameViewport()
	for _, id := range ebiten.AppendTouchIDs(nil) {
		x, y := ebiten.TouchPosition(id)
		if imagePointInGameViewport(x, y) && x < viewport.Min.X+ScreenWidth/2 {
			return true
		}
	}
	return false
}

func RotateRight() bool {
	viewport := core.GameViewport()
	for _, id := range ebiten.AppendTouchIDs(nil) {
		x, y := ebiten.TouchPosition(id)
		if imagePointInGameViewport(x, y) && x > viewport.Min.X+ScreenWidth/2 {
			return true
		}
	}
	return false
}

func imagePointInGameViewport(x, y int) bool {
	return image.Pt(x, y).In(core.GameViewport())
}
