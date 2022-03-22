package models

import (
	_ "embed"
	"fmt"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

type Background struct {
	Name        string
	Options     *ebiten.DrawImageOptions
	Sprite      Sprite
	tick        *uint64
	Position    Position
	Destination Position
}

func NewBackground(tick *uint64, position Position) Background {
	b := Background{}
	b.Sprite.Init(backgroundSprite, backgroundSpec)
	b.tick = tick
	b.Options = &ebiten.DrawImageOptions{}

	// fill screen
	b.Options.GeoM.Scale(1024, 1)

	b.Destination = position
	return b
}

func (b *Background) Stand() *ebiten.Image {

	f := b.Sprite.SpriteSpec.Stand.Frames[0]
	rect := image.Rect(f.X, f.Y, f.X+f.W, f.Y+f.H)
	subImg := b.Sprite.Image.SubImage(rect).(*ebiten.Image)

	return subImg
}

func (b *Background) Place(destination Position) (*ebiten.Image, *ebiten.DrawImageOptions) {

	if destination == (Position{}) && b.Destination == b.Position {
		return b.Stand(), b.Options
	} else if destination != (Position{}) {
		b.Destination = destination
	}

	fmt.Printf("Position %s at %v\n", b.Name, b.Position)
	fmt.Printf("Destination %s at %v\n", b.Name, b.Destination)

	var newX, newY float64

	if b.Destination.X > b.Position.X {
		newX = (b.Destination.X - b.Position.X)
	}
	if b.Position.X > b.Destination.X {
		newX = (b.Position.X - b.Destination.X) * (-1)
	}

	if b.Destination.Y > b.Position.Y {
		newY = (b.Destination.Y - b.Position.Y)
	}
	if b.Position.X > b.Destination.Y {
		newY = (b.Position.Y - b.Destination.Y) * (-1)
	}

	if b.Destination.X == b.Position.X {
		newX = 0
	}

	if b.Destination.Y == b.Position.Y {
		newY = 0
	}

	b.Options.GeoM.Translate(newX, newY)
	b.Position = b.Destination
	fmt.Printf("new pos, %s at %v, %v\n", b.Name, newX, newY)

	return b.Stand(), b.Options
}
