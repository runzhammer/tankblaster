package models

import (
	_ "embed"
	"log"
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

	// set on surface of ground
	vec.Y -= t.Size.Y

	leftWheelPos := engine.Vec{X: vec.X + t.TankChassis.LeftWheel.X, Y: vec.Y + t.TankChassis.LeftWheel.Y}
	rightWheelPos := engine.Vec{X: vec.X + t.TankChassis.RightWheel.X, Y: vec.Y + t.TankChassis.RightWheel.Y}

	// angle relative to the ground
	// left wheel is the origin
	dstLeftWheel := engine.Vec{X: leftWheelPos.X, Y: t.Ground.GetGroundY(leftWheelPos.X)}
	dstRightWheel := engine.Vec{X: rightWheelPos.X, Y: t.Ground.GetGroundY(rightWheelPos.X)}

	// calculate rotation
	angle := math.Atan2(dstRightWheel.Y-dstLeftWheel.Y, dstRightWheel.X-dstLeftWheel.X)
	log.Printf("Angle: %v - %v = %v", dstRightWheel.Y-dstLeftWheel.Y, dstRightWheel.X-dstLeftWheel.X, angle)
	t.TankChassis.Sprite.Rot = angle
	// t.TankCannon.Sprite.Rot = angle

	log.Printf("Position: %v", vec)

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
