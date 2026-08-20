package tankblaster

import (
	"github.com/hajimehoshi/ebiten/v2"
)

func Begin() bool {
	return len(ebiten.AppendTouchIDs(nil)) > 0
}

func MoveLeft() bool {
	return false
}

func MoveRight() bool {
	return false
}

func RotateLeft() bool {
	for _, id := range ebiten.AppendTouchIDs(nil) {
		x, _ := ebiten.TouchPosition(id)
		if x < ScreenWidth/2 {
			return true
		}
	}
	return false
}

func RotateRight() bool {
	for _, id := range ebiten.AppendTouchIDs(nil) {
		x, _ := ebiten.TouchPosition(id)
		if x > ScreenWidth/2 {
			return true
		}
	}
	return false
}
