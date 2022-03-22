package models

import (
	_ "embed"
	"fmt"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

type Ground struct {
	Name        string
	Options     *ebiten.DrawImageOptions
	Sprite      Sprite
	tick        *uint64
	Position    Position
	Destination Position
}

func NewGround(tick *uint64, position Position) Ground {
	gr := Ground{}
	gr.Sprite.Init(groundSprite, groundSpec)
	gr.tick = tick
	gr.Options = &ebiten.DrawImageOptions{}

	// fill screen
	gr.Options.GeoM.Scale(1024, 1)

	gr.Destination = position
	return gr
}

func (gr *Ground) Stand() *ebiten.Image {

	f := gr.Sprite.SpriteSpec.Stand.Frames[0]
	rect := image.Rect(f.X, f.Y, f.X+f.W, f.Y+f.H)
	subImg := gr.Sprite.Image.SubImage(rect).(*ebiten.Image)

	return subImg
}

func (gr *Ground) Place(destination Position) (*ebiten.Image, *ebiten.DrawImageOptions) {

	if destination == (Position{}) && gr.Destination == gr.Position {
		return gr.Stand(), gr.Options
	} else if destination != (Position{}) {
		gr.Destination = destination
	}

	fmt.Printf("Position %s at %v\n", gr.Name, gr.Position)
	fmt.Printf("Destination %s at %v\n", gr.Name, gr.Destination)

	var newX, newY float64

	if gr.Destination.X > gr.Position.X {
		newX = (gr.Destination.X - gr.Position.X)
	}
	if gr.Position.X > gr.Destination.X {
		newX = (gr.Position.X - gr.Destination.X) * (-1)
	}

	if gr.Destination.Y > gr.Position.Y {
		newY = (gr.Destination.Y - gr.Position.Y)
	}
	if gr.Position.X > gr.Destination.Y {
		newY = (gr.Position.Y - gr.Destination.Y) * (-1)
	}

	if gr.Destination.X == gr.Position.X {
		newX = 0
	}

	if gr.Destination.Y == gr.Position.Y {
		newY = 0
	}

	gr.Options.GeoM.Translate(newX, newY)
	gr.Position = gr.Destination
	fmt.Printf("new pos, %s at %v, %v\n", gr.Name, newX, newY)

	return gr.Stand(), gr.Options
}
