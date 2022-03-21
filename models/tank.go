package models

import (
	_ "embed"
	"fmt"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed resources/tank.png
var sprite []byte

//go:embed resources/tank.yaml
var animations []byte

type Tank struct {
	Name     string
	Options  *ebiten.DrawImageOptions
	Sprite   Sprite
	tick     *uint64
	Position Position
}

type Position struct {
	X float64
	Y float64
}

func NewTank(tick *uint64, position Position) Tank {
	t := Tank{}
	t.Sprite.Init(sprite, animations)
	t.tick = tick
	return t
}

func (t *Tank) Move(x float64, y float64) (*ebiten.Image, *ebiten.DrawImageOptions) {
	if t.Options == nil {
		t.Options = &ebiten.DrawImageOptions{}
	}
	// fmt.Printf("tick: %v\nspeed: %v\nlen-frames: %v\n", int(*t.tick), uint64(t.Sprite.Animations.Move.Speed)+1, len(t.Sprite.Animations.Move.Frames))
	frameNum := int(float64(*t.tick)/t.Sprite.Animations.Move.Speed) % len(t.Sprite.Animations.Move.Frames)
	// fmt.Printf("frame: %v\n", frameNum)
	f := t.Sprite.Animations.Move.Frames[frameNum]
	rect := image.Rect(f.X, f.Y, f.X+f.W, f.Y+f.H)
	subImg := t.Sprite.Image.SubImage(rect).(*ebiten.Image)

	moveX := float64(0)
	moveY := float64(0)

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

	return subImg, t.Options
}
