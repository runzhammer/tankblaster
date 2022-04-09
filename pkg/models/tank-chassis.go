package models

import (
	_ "embed"

	"github.com/runzhammer/gamedemo/pkg/engine"
	r "github.com/runzhammer/gamedemo/resources"
)

type TankChassis struct {
	Sprite     *engine.Sprite
	Name       string
	Position   engine.Vec
	LeftWheel  engine.Vec
	RightWheel engine.Vec
	Size       engine.Vec
	PreSteps   engine.Behaviors
	Steps      engine.Behaviors
	PostSteps  engine.Behaviors
}

func NewTankChassis(name string) *TankChassis {

	scaleFactor := float64(1)

	m := TankChassis{Name: name}

	m.Sprite = engine.NewSprite(r.TankSprite, r.TankSpec)
	// m.Position = engine.Vec{X: 200, Y: 700} //core.Config().Screen.Height/2 - t.Sprite.Bounds().H()/2}
	m.Size = engine.Vec{X: m.Sprite.Drawable.Bounds().W() * scaleFactor, Y: m.Sprite.Drawable.Bounds().H() * scaleFactor}

	m.LeftWheel = engine.Vec{X: 13.5, Y: m.Size.Y}
	m.RightWheel = engine.Vec{X: m.Size.X - 13.5, Y: m.Size.Y}

	m.Sprite.MovementSpeed = 2
	m.Sprite.RotationSpeedPerSecond = 1

	m.Sprite.Tag = "chassis"
	m.Sprite.Pos = &m.Position
	m.Sprite.Size = &m.Size

	return &m
}

func (t *TankChassis) SetPosition(vec engine.Vec) {
	t.Position = vec
}
