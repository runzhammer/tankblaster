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

var SmallTankCannonMount = engine.V(8, 0)

func NewTank(name string, tankColor color.RGBA) Tank {

	scaleFactor := float64(1.0)
	cannonLength := 11
	cannonAnchor := engine.V(8, 1)
	cannonWidth := float64(int(cannonAnchor.X) + cannonLength)
	cannonHeight := 3.0

	m := Tank{Name: name}
	m.Sprites = engine.NewSprites()

	tankSprite := engine.NewSprite(r.TankSmallSprite, r.TankSmallSpec)
	tankSprite.Image = colorizedSpriteImage(r.TankSmallSprite, tankColor)
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
	cannonSprite.Pos = &engine.Vec{
		X: m.Position.X + SmallTankCannonMount.X - cannonAnchor.X,
		Y: m.Position.Y + SmallTankCannonMount.Y - cannonAnchor.Y,
	}
	cannonSprite.Size = &engine.Vec{X: cannonWidth, Y: cannonHeight}
	cannonSprite.RotAnchor = &cannonAnchor
	cannonSprite.Rot = engine.DegToRad(-55)
	cannonSprite.Drawable = engine.NewImageDrawable(generateCannonImage(int(cannonWidth), int(cannonHeight), int(cannonAnchor.X), int(cannonAnchor.Y), cannonLength, tankColor))

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
	body.Image = colorizedSpriteImage(r.TankSmallSprite, tankColor)
	body.Drawable = engine.NewImageDrawableFrames(body.Image, engine.R(0, 0, bounds.W(), bounds.H()))
}

func RecolorCannon(cannon *engine.Sprite, tankColor color.RGBA) {
	if cannon == nil || cannon.Drawable == nil {
		return
	}
	bounds := cannon.Drawable.Bounds()
	anchorX := int(bounds.W() / 2)
	anchorY := int(bounds.H() / 2)
	if cannon.RotAnchor != nil {
		anchorX = int(cannon.RotAnchor.X)
		anchorY = int(cannon.RotAnchor.Y)
	}
	length := int(bounds.W()) - anchorX
	cannon.Image = generateCannonImage(int(bounds.W()), int(bounds.H()), anchorX, anchorY, length, tankColor)
	cannon.Drawable = engine.NewImageDrawable(cannon.Image)
}

func generateCannonImage(width, height, anchorX, anchorY, length int, tankColor color.RGBA) *ebiten.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	for x := anchorX; x < anchorX+length && x < width; x++ {
		if anchorY >= 0 && anchorY < height {
			img.SetRGBA(x, anchorY, tankColor)
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
