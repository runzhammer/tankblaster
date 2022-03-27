package models

import (
	_ "embed"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/runzhammer/gamedemo/pkg/engine"
	r "github.com/runzhammer/gamedemo/resources"
)

type Tank struct {
	Name      string
	Speed     int
	Sprite    *engine.Sprite
	PreSteps  engine.Behavior
	Steps     engine.Behavior
	PostSteps engine.Behavior
}

func NewTank(name string) Tank {

	t := Tank{Name: name}
	t.Speed = 215
	t.Sprite = engine.NewSprite(r.TankSprite, r.TankSpec)

	t.Sprite.Tag = name
	t.Sprite.Pos = engine.Vec{X: 200, Y: 600} //core.Config().Screen.Height/2 - t.Sprite.Bounds().H()/2}
	t.Sprite.Size = engine.Vec{X: t.Sprite.Drawable.Bounds().W() / 2, Y: t.Sprite.Drawable.Bounds().H() / 2}

	//
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
