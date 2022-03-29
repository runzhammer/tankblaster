package models

import (
	_ "embed"

	"github.com/runzhammer/gamedemo/pkg/engine"
	r "github.com/runzhammer/gamedemo/resources"
)

type TankCannon struct {
	Sprite    *engine.Sprite
	Position  engine.Vec
	Size      engine.Vec
	PreSteps  engine.Behaviors
	Steps     engine.Behaviors
	PostSteps engine.Behaviors
}

func NewTankCannon() *TankCannon {

	scaleFactor := float64(1)

	m := TankCannon{}

	m.Sprite = engine.NewSprite(r.CannonSprite, r.CannonSpec)
	// m.Position = engine.Vec{X: m.Position.X + m.Size.X/2, Y: m.Position.Y + 5}
	m.Size = engine.Vec{X: m.Sprite.Drawable.Bounds().W() * scaleFactor, Y: m.Sprite.Drawable.Bounds().H() * scaleFactor}

	m.Sprite.MovementSpeed = 2
	m.Sprite.RotationSpeedPerSecond = 0.1

	m.Sprite.Tag = "cannon"
	m.Sprite.Pos = &m.Position
	m.Sprite.Size = &m.Size
	m.Sprite.MaxRange = []float64{-180, 0}
	m.Sprite.RotCenter = false

	return &m
}

func (t *TankCannon) SetPosition(vec engine.Vec) {
	t.Position = vec
}
