package models

import (
	_ "embed"
	"fmt"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/runzhammer/gamedemo/pkg/engine"
	"github.com/runzhammer/gamedemo/pkg/tankblaster"
	r "github.com/runzhammer/gamedemo/resources"
)

type Tank struct {
	Name        string
	Sprite      engine.Sprite
	tick        *uint64
	Position    Position
	Destination Position
	Steps       engine.Behavior
	PostSteps   engine.Behavior
}

func NewTank(tick *uint64, position Position) Tank {

	t := Tank{}
	s := engine.NewSprite(r.TankSprite, r.TankSpec)

	engine.NewImageDrawableFrames(tanksImage, engine.R(0, 0, 148, 333./2))

	s := &engine.Sprite{
		Tag:       t.Name,
		Pos:       engine.V(100, tankblaster.ScreenHeight/2-tankHeight/2),
		Size:      engine.V(tankWidth, tankHeight),
		Rot:       rotBlue,
		RotNormal: tankRotateOffset,

		Steps: engine.MakeBehaviors(
			s.behaviorBlueRotateOnButton,
		),
		PostSteps: engine.MakeBehaviors(
			s.reflectInBounds,
			s.behaviorBlueHitsRedBullet,
		),
	}
	t.Sprite = *s

	t.tick = tick
	t.Options = &ebiten.DrawImageOptions{}
	t.Destination = position
	fmt.Printf("NewTank Destination: %v\n", t.Destination)
	return t
}

func (t *Tank) Drive() *ebiten.Image {

	// fmt.Printf("tick: %v\nspeed: %v\nlen-frames: %v\n", int(*t.tick), uint64(t.Sprite.Animations.Move.Speed)+1, len(t.Sprite.Animations.Move.Frames))
	frameNum := int(float64(*t.tick)/t.Sprite.SpriteSpec.Drive.Speed) % len(t.Sprite.SpriteSpec.Drive.Frames)
	// fmt.Printf("frame: %v\n", frameNum)
	f := t.Sprite.SpriteSpec.Drive.Frames[frameNum]
	rect := image.Rect(f.X, f.Y, f.X+f.W, f.Y+f.H)
	subImg := t.Sprite.Image.SubImage(rect).(*ebiten.Image)

	return subImg
}

func (t *Tank) Stand() *ebiten.Image {

	f := t.Sprite.SpriteSpec.Stand.Frames[0]
	rect := image.Rect(f.X, f.Y, f.X+f.W, f.Y+f.H)
	subImg := t.Sprite.Image.SubImage(rect).(*ebiten.Image)

	return subImg
}
