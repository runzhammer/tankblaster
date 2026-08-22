package models

import (
	"bytes"
	_ "embed"
	"image"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
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

func NewTank(name string, tankColor color.RGBA) Tank {

	scaleFactor := float64(0.40)
	cannonWidth := 30.0
	cannonHeight := 8.0

	m := Tank{Name: name}
	m.Sprites = engine.NewSprites()

	tankSprite := engine.NewSprite(r.TankSprite, r.TankSpec)
	tankSprite.Image = colorizedSpriteImage(r.TankSprite, tankColor)
	tankSprite.Drawable = engine.NewImageDrawableFrames(tankSprite.Image, engine.R(0, 0, tankSprite.Drawable.Bounds().W(), tankSprite.Drawable.Bounds().H()))
	m.Position = &engine.Vec{X: 200, Y: 600} //core.Config().Screen.Height/2 - t.Sprite.Bounds().H()/2}
	m.Size = &engine.Vec{X: tankSprite.Drawable.Bounds().W() * scaleFactor, Y: tankSprite.Drawable.Bounds().H() * scaleFactor}

	tankSprite.MovementSpeed = 2
	tankSprite.RotationSpeed = 1

	tankSprite.Tag = name
	tankSprite.Pos = m.Position
	tankSprite.Size = m.Size

	m.Sprites.Add(tankSprite)

	cannonSprite := &engine.Sprite{}
	cannonSprite.MovementSpeed = 1
	cannonSprite.RotationSpeed = 2

	cannonSprite.Tag = "cannon"
	cannonSprite.Pos = &engine.Vec{X: m.Position.X + m.Size.X/2 - cannonWidth/2, Y: m.Position.Y + 5}
	cannonSprite.Size = &engine.Vec{X: cannonWidth, Y: cannonHeight}
	cannonSprite.Rot = engine.DegToRad(-55)
	cannonSprite.Drawable = engine.NewImageDrawable(generateCannonImage(int(cannonWidth), int(cannonHeight), tankColor))

	m.Sprites.Add(cannonSprite)

	//
	// PostSteps: engine.MakeBehaviors(
	// 	t.Sprite.reflectInBounds,
	// 	t.Sprite.behaviorBlueHitsRedBullet,
	// ),

	return m
}

func colorizedSpriteImage(sprite []byte, tint color.RGBA) *ebiten.Image {
	src, _, err := image.Decode(bytes.NewReader(sprite))
	if err != nil {
		log.Fatal(err)
	}

	bounds := src.Bounds()
	img := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			_, _, _, alpha := src.At(x, y).RGBA()
			if alpha == 0 {
				continue
			}
			tintedAlpha := (alpha / 257) * uint32(tint.A) / 255
			img.SetRGBA(x, y, color.RGBA{R: tint.R, G: tint.G, B: tint.B, A: uint8(tintedAlpha)})
		}
	}

	return ebiten.NewImageFromImage(img)
}

func RecolorTankBody(body *engine.Sprite, tankColor color.RGBA) {
	if body == nil || body.Drawable == nil {
		return
	}
	bounds := body.Drawable.Bounds()
	body.Image = colorizedSpriteImage(r.TankSprite, tankColor)
	body.Drawable = engine.NewImageDrawableFrames(body.Image, engine.R(0, 0, bounds.W(), bounds.H()))
}

func RecolorCannon(cannon *engine.Sprite, tankColor color.RGBA) {
	if cannon == nil || cannon.Drawable == nil {
		return
	}
	bounds := cannon.Drawable.Bounds()
	cannon.Image = generateCannonImage(int(bounds.W()), int(bounds.H()), tankColor)
	cannon.Drawable = engine.NewImageDrawable(cannon.Image)
}

func generateCannonImage(width, height int, tankColor color.RGBA) *ebiten.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	pivotX := width / 2
	centerY := height / 2

	for x := pivotX; x < width-1; x++ {
		for y := centerY - 2; y <= centerY+1; y++ {
			if y >= 0 && y < height {
				img.SetRGBA(x, y, tankColor)
			}
		}
	}

	return ebiten.NewImageFromImage(img)
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
