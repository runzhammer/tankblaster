//go:build (darwin || freebsd || linux || windows) && !android && !ios
// +build darwin freebsd linux windows
// +build !android
// +build !ios

package tankblaster

import (
	"github.com/hajimehoshi/ebiten/v2"
)

func Begin() bool {
	return ebiten.IsKeyPressed(ebiten.KeySpace)
}

func MoveLeft() bool {
	return ebiten.IsKeyPressed(ebiten.KeyArrowUp)
}

func MoveRight() bool {
	return ebiten.IsKeyPressed(ebiten.KeyArrowDown)
}

func RotateLeft() bool {
	return ebiten.IsKeyPressed(ebiten.KeyArrowLeft)
}

func RotateRight() bool {
	return ebiten.IsKeyPressed(ebiten.KeyArrowRight)
}
