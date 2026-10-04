//go:build js

package tankblaster

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/runzhammer/tankblaster/pkg/core"
)

func Begin() bool {
	if !mobileControlsEnabled() {
		return ebiten.IsKeyPressed(ebiten.KeySpace)
	}
	for _, id := range ebiten.AppendTouchIDs(nil) {
		x, y := webTouchPosition(id)
		if imagePointInGameViewport(x, y) {
			return true
		}
	}
	return false
}

func MoveLeft() bool {
	if mobileControlsEnabled() {
		return false
	}
	return ebiten.IsKeyPressed(ebiten.KeyArrowUp)
}

func MoveRight() bool {
	if mobileControlsEnabled() {
		return false
	}
	return ebiten.IsKeyPressed(ebiten.KeyArrowDown)
}

func RotateLeft() bool {
	if !mobileControlsEnabled() {
		return ebiten.IsKeyPressed(ebiten.KeyArrowLeft)
	}
	viewport := core.GameViewport()
	for _, id := range ebiten.AppendTouchIDs(nil) {
		x, y := webTouchPosition(id)
		if imagePointInGameViewport(x, y) && x < viewport.Min.X+ScreenWidth/2 {
			return true
		}
	}
	return false
}

func RotateRight() bool {
	if !mobileControlsEnabled() {
		return ebiten.IsKeyPressed(ebiten.KeyArrowRight)
	}
	viewport := core.GameViewport()
	for _, id := range ebiten.AppendTouchIDs(nil) {
		x, y := webTouchPosition(id)
		if imagePointInGameViewport(x, y) && x > viewport.Min.X+ScreenWidth/2 {
			return true
		}
	}
	return false
}

func imagePointInGameViewport(x, y int) bool {
	return image.Pt(x, y).In(core.GameViewport())
}
