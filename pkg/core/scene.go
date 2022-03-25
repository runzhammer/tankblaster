package core

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
)

type Scene interface {
	Update(dt float64) error
	Draw(image *ebiten.Image)
}

type ChangeSceneError struct {
	Game string
}

func (c *ChangeSceneError) Error() string {
	return fmt.Sprintf("change game: %s", c.Game)
}
