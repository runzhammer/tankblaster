package models

import (
	_ "embed"
	"fmt"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

type Tank struct {
	Name        string
	Options     *ebiten.DrawImageOptions
	Sprite      Sprite
	tick        *uint64
	Position    Position
	Destination Position
}

func NewTank(tick *uint64, position Position) Tank {
	t := Tank{}
	t.Sprite.Init(tankSprite, tankSpec)
	t.tick = tick
	t.Options = &ebiten.DrawImageOptions{}
	t.Destination = position
	fmt.Printf("NewTank Destination: %v\n", t.Destination)
	return t
}

func (t *Tank) Drive() *ebiten.Image {

	// fmt.Printf("tick: %v\nspeed: %v\nlen-frames: %v\n", int(*t.tick), uint64(t.Sprite.Animations.Move.Speed)+1, len(t.Sprite.Animations.Move.Frames))
	frameNum := int(float64(*t.tick)/t.Sprite.SpriteSpec.Drive.Speed) % len(t.Sprite.SpriteSpec.Drive.Frames)
	// fmt.Printf("frame: %v\n", frameNum)
	f := t.Sprite.SpriteSpec.Drive.Frames[frameNum]
	rect := image.Rect(f.X, f.Y, f.X+f.W, f.Y+f.H)
	subImg := t.Sprite.Image.SubImage(rect).(*ebiten.Image)

	return subImg
}

func (t *Tank) Stand() *ebiten.Image {

	f := t.Sprite.SpriteSpec.Stand.Frames[0]
	rect := image.Rect(f.X, f.Y, f.X+f.W, f.Y+f.H)
	subImg := t.Sprite.Image.SubImage(rect).(*ebiten.Image)

	return subImg
}

func (t *Tank) Move(x float64, y float64) *ebiten.DrawImageOptions {

	var moveX, moveY float64

	if t.Position.X < x {
		moveX = 1
	}

	if t.Position.Y < y {
		moveY = 1
	}

	if t.Position.X > x {
		moveX = -1
	}

	if t.Position.Y > y {
		moveY = -1
	}

	t.Position.X = t.Position.X + moveX
	t.Position.Y = t.Position.Y + moveY

	fmt.Printf("%s: Move to  --  X: %v - Y: %v\n", t.Name, x, y)
	fmt.Printf("%s: Current  --  X: %v - Y: %v\n", t.Name, t.Position.X, t.Position.Y)
	t.Options.GeoM.Translate(moveX, moveY)

	return t.Options
}

func (t *Tank) Place(destination Position) (*ebiten.Image, *ebiten.DrawImageOptions) {

	if destination == (Position{}) && t.Destination == t.Position {
		return t.Stand(), t.Options
	} else if destination != (Position{}) {
		t.Destination = destination
	}

	fmt.Printf("Position %s at %v\n", t.Name, t.Position)
	fmt.Printf("Destination %s at %v\n", t.Name, t.Destination)

	var newX, newY float64

	if t.Destination.X > t.Position.X {
		newX = (t.Destination.X - t.Position.X)
	}
	if t.Position.X > t.Destination.X {
		newX = (t.Position.X - t.Destination.X) * (-1)
	}

	if t.Destination.Y > t.Position.Y {
		newY = (t.Destination.Y - t.Position.Y)
	}
	if t.Position.X > t.Destination.Y {
		newY = (t.Position.Y - t.Destination.Y) * (-1)
	}

	if t.Destination.X == t.Position.X {
		newX = 0
	}

	if t.Destination.Y == t.Position.Y {
		newY = 0
	}

	t.Options.GeoM.Translate(newX, newY)
	t.Position = t.Destination
	fmt.Printf("new pos, %s at %v, %v\n", t.Name, newX, newY)

	return t.Stand(), t.Options
}
