package models

import (
	_ "embed"

	"github.com/runzhammer/gamedemo/pkg/engine"
)

type Tank struct {
	Name        string
	TankChassis *TankChassis
	TankCannon  *TankCannon
	Position    *engine.Vec
	Size        *engine.Vec
	PreSteps    engine.Behaviors
	Steps       engine.Behaviors
	PostSteps   engine.Behaviors
	Ground      *Ground
}

func NewTank(name string, ground *Ground) *Tank { // , ground *Ground

	t := Tank{
		Name:        name,
		TankChassis: NewTankChassis(),
		TankCannon:  NewTankCannon(),
		Ground:      ground,
	}

	t.Size = &t.TankChassis.Size

	return &t
}

func (t *Tank) SetPosition(vec engine.Vec) { // xPos int

	leftWheel := engine.Vec{X: 13.5, Y: t.Size.Y}
	rightWheel := engine.Vec{X: t.Size.X - 13.5, Y: t.Size.Y}

	t.Position = &vec

	// Position of tankchassis
	t.TankChassis.SetPosition(vec)

	// Position of tankCannon
	t.TankCannon.SetPosition(engine.Vec{X: vec.X + t.TankChassis.Sprite.Size.X/2, Y: vec.Y + 5})

	t.UpdateSprites()
}

func (t *Tank) GetSprites() []*engine.Sprite {
	return []*engine.Sprite{t.TankChassis.Sprite, t.TankCannon.Sprite}
}

func (t *Tank) UpdateSprites() {
	t.TankChassis.Sprite.Pos = &t.TankChassis.Position
	t.TankCannon.Sprite.Pos = &t.TankCannon.Position
}
