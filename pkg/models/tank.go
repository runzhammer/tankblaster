package models

import (
	_ "embed"

	"github.com/runzhammer/gamedemo/pkg/engine"
	r "github.com/runzhammer/gamedemo/resources"
)

type Tank struct {
	Name      string
	Sprites   *engine.Sprites
	Position  *engine.Vec
	Size      *engine.Vec
	PreSteps  engine.Behavior
	Steps     engine.Behavior
	PostSteps engine.Behavior
}

func NewTank(name string) Tank {

	scaleFactor := float64(1)

	m := Tank{Name: name}
	m.Sprites = engine.NewSprites()

	tankSprite := engine.NewSprite(r.TankSprite, r.TankSpec)
	m.Position = &engine.Vec{X: 200, Y: 600} //core.Config().Screen.Height/2 - t.Sprite.Bounds().H()/2}
	m.Size = &engine.Vec{X: tankSprite.Drawable.Bounds().W() * scaleFactor, Y: tankSprite.Drawable.Bounds().H() * scaleFactor}

	tankSprite.MovementSpeed = 2
	tankSprite.RotationSpeed = 1

	tankSprite.Tag = name
	tankSprite.Pos = m.Position
	tankSprite.Size = m.Size

	m.Sprites.Add(tankSprite)

	cannonSprite := engine.NewSprite(r.CannonSprite, r.CannonSpec)
	cannonSprite.MovementSpeed = 1
	cannonSprite.RotationSpeed = 1

	cannonSprite.Tag = "cannon"
	cannonSprite.Pos = &engine.Vec{X: m.Position.X + m.Size.X/2, Y: m.Position.Y + 5}
	cannonSprite.Size = &engine.Vec{X: cannonSprite.Drawable.Bounds().W() * scaleFactor, Y: cannonSprite.Drawable.Bounds().H() * scaleFactor}

	m.Sprites.Add(cannonSprite)

	//
	// PostSteps: engine.MakeBehaviors(
	// 	t.Sprite.reflectInBounds,
	// 	t.Sprite.behaviorBlueHitsRedBullet,
	// ),

	return m
}

func (m Tank) Body() *engine.Sprite {
	iter := m.Sprites.Tagged(m.Name).Iterator()
	body, _ := iter()
	return body
}

func (m Tank) Cannon() *engine.Sprite {
	iter := m.Sprites.Tagged("cannon").Iterator()
	cannon, _ := iter()
	return cannon
}
