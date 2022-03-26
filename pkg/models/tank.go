package models

import (
	_ "embed"
	"fmt"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/runzhammer/gamedemo/pkg/engine"
	"github.com/runzhammer/gamedemo/pkg/core"
	r "github.com/runzhammer/gamedemo/resources"
)

type Tank struct {
	Name        string
	Sprite      *engine.Sprite
	Position    engine.Vec
	Steps       engine.Behavior
	PostSteps   engine.Behavior
}

func NewTank() Tank {

	t := Tank{}

	s := engine.Sprite{
		Tag:       t.Name,
		Pos:       engine.V(100, core.Config().Screen.Height/2-t.Sprite.Bounds().H()/2),
		Size:      engine.V(t.Sprite.Bounds().W(), t.Sprite.Bounds().H()),

		// Steps: engine.MakeBehaviors(
		// 	t.Sprite.behaviorBlueRotateOnButton,
		// ),
		// PostSteps: engine.MakeBehaviors(
		// 	t.Sprite.reflectInBounds,
		// 	t.Sprite.behaviorBlueHitsRedBullet,
		// ),
	}
	t.Sprite = engine.NewSprite(&s, r.TankSprite, r.TankSpec)

	return t
}

func (t *Tank) Drive() *ebiten.Image {
	// // fmt.Printf("tick: %v\nspeed: %v\nlen-frames: %v\n", int(*t.tick), uint64(t.Sprite.Animations.Move.Speed)+1, len(t.Sprite.Animations.Move.Frames))
	// frameNum := int(float64(*t.tick)/t.Sprite.SpriteSpec.Drive.Speed) % len(t.Sprite.SpriteSpec.Drive.Frames)
	// // fmt.Printf("frame: %v\n", frameNum)
	// f := t.Sprite.SpriteSpec.Drive.Frames[frameNum]
	// rect := image.Rect(f.X, f.Y, f.X+f.W, f.Y+f.H)
	// subImg := t.Sprite.Image.SubImage(rect).(*ebiten.Image)

	t.Steps = engine.Behavior(Movement)
}

func (t *Tank) Stand() *ebiten.Image {

	f := t.Sprite.SpriteSpec.Stand.Frames[0]
	rect := image.Rect(f.X, f.Y, f.X+f.W, f.Y+f.H)
	subImg := t.Sprite.Image.SubImage(rect).(*ebiten.Image)

	return subImg
}
