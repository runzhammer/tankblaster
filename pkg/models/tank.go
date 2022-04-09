package models

import (
	_ "embed"
	"math"

	"github.com/runzhammer/gamedemo/pkg/engine"
)

type Tank struct {
	Name        string
	TankChassis *TankChassis
	TankCannon  *TankCannon
	Position    *engine.Vec
	Size        *engine.Vec
	Rot         float64
	Velocity    *engine.Vec
	PreSteps    Behaviors
	Steps       Behaviors
	PostSteps   Behaviors
	Ground      *Ground
}

func NewTank(name string, ground *Ground) *Tank {

	t := Tank{
		Name:        name,
		TankChassis: NewTankChassis(name),
		TankCannon:  NewTankCannon(name),
		Ground:      ground,
	}

	t.Size = &t.TankChassis.Size

	return &t
}

func (t *Tank) SetPosition(vec *engine.Vec) {

	leftWheelPos := engine.Vec{X: t.TankChassis.Sprite.Pos.X + t.TankChassis.LeftWheel.X, Y: t.TankChassis.Sprite.Pos.Y}
	rightWheelPos := engine.Vec{X: t.TankChassis.Sprite.Pos.X + t.TankChassis.RightWheel.X, Y: t.TankChassis.Sprite.Pos.Y}

	// angle relative to the ground
	// left wheel is the origin
	// dstLeftWheel := engine.Vec{X: leftWheelPos.X, Y: t.Ground.Coords[int(leftWheelPos.X)]}
	// dstRightWheel := engine.Vec{X: rightWheelPos.X, Y: t.Ground.Coords[int(rightWheelPos.X)]}

	// calculate rotation
	angle := math.Atan2(t.Ground.Coords[int(rightWheelPos.X)], rightWheelPos.X-leftWheelPos.X)

	t.TankChassis.Sprite.Rot = angle
	t.TankCannon.Sprite.Rot = angle

	t.Position = vec

	// Position of tankchassis
	t.TankChassis.SetPosition(*vec)

	// Position of tankCannon
	t.TankCannon.ChassisOffset = engine.Vec{X: t.TankChassis.Sprite.Size.X / 2, Y: 5}
	t.TankCannon.SetPosition(*vec)

	t.UpdateSprites()
}

func (t *Tank) GetSprites() []*engine.Sprite {
	return []*engine.Sprite{t.TankChassis.Sprite, t.TankCannon.Sprite}
}

func (t *Tank) UpdateSprites() {
	t.TankChassis.Sprite.Pos = &t.TankChassis.Position
	t.TankCannon.Sprite.Pos = &t.TankCannon.Position
}
