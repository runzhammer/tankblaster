package models

import (
	_ "embed"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/runzhammer/gamedemo/pkg/core"
	"github.com/runzhammer/gamedemo/pkg/engine"
	r "github.com/runzhammer/gamedemo/resources"
)

type Tank struct {
	Name      string
	Speed     int
	Sprite    *engine.Sprite
	Position  engine.Vec
	Steps     engine.Behavior
	PostSteps engine.Behavior
}

func NewTank() Tank {

	t := Tank{}
	t.Speed = 215

	t.Sprite = engine.NewSprite(r.TankSprite, r.TankSpec)

	t.Sprite.Pos = engine.V(100, core.Config().Screen.Height/2-t.Sprite.Bounds().H()/2)
	t.Sprite.Size = engine.V(t.Sprite.Bounds().W(), t.Sprite.Bounds().H())
	t.Sprite.Steps = engine.MakeBehaviors(
		engine.Movement,
	)
	// PostSteps: engine.MakeBehaviors(
	// 	t.Sprite.reflectInBounds,
	// 	t.Sprite.behaviorBlueHitsRedBullet,
	// ),

	return t
}

func (t *Tank) Stand() *ebiten.Image {

	f := t.Sprite.SpriteSpec.Stand.Frames[0]
	rect := image.Rect(f.X, f.Y, f.X+f.W, f.Y+f.H)
	subImg := t.Sprite.Image.SubImage(rect).(*ebiten.Image)

	return subImg
}
