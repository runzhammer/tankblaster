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

const (
	TankBodyKindSmall = "small"
	TankBodyKindXMV12 = "xm-v12"
)

type TankBodyMeta struct {
	Kind             string
	Facing           int
	CannonMountRight engine.Vec
	CannonLength     int
}

func IsSmallTankBody(body *engine.Sprite) bool {
	if body == nil || body.Drawable == nil {
		return false
	}
	if meta, ok := body.Meta.(*TankBodyMeta); ok {
		return meta.Kind == TankBodyKindSmall
	}
	bounds := body.Drawable.Bounds()
	return int(bounds.W()) == 17 && int(bounds.H()) == 14
}

func NewTank(name string, tankColor color.RGBA) Tank {
	return newTankFromSprite(name, tankColor, r.TankSmallSprite, TankBodyKindSmall, engine.V(8, 0), 9)
}

func NewXMV12Tank(name string, tankColor color.RGBA) Tank {
	return newTankFromSprite(name, tankColor, r.XMV12TankSprite, TankBodyKindXMV12, engine.V(6, 10), 26)
}

func newTankFromSprite(name string, tankColor color.RGBA, sprite []byte, bodyKind string, cannonMount engine.Vec, cannonLength int) Tank {
	scaleFactor := float64(1.0)
	cannonAnchor := engine.V(8, 1)
	cannonWidth := float64(int(cannonAnchor.X) + cannonLength)
	cannonHeight := 3.0

	m := Tank{Name: name}
	m.Sprites = engine.NewSprites()

	tankSprite := newTankBodySprite(sprite, tankColor)
	tankSprite.Drawable = engine.NewImageDrawableFrames(tankSprite.Image, engine.R(0, 0, tankSprite.Drawable.Bounds().W(), tankSprite.Drawable.Bounds().H()))
	m.Position = &engine.Vec{X: 200, Y: 600} //core.Config().Screen.Height/2 - t.Sprite.Bounds().H()/2}
	m.Size = &engine.Vec{X: tankSprite.Drawable.Bounds().W() * scaleFactor, Y: tankSprite.Drawable.Bounds().H() * scaleFactor}

	tankSprite.MovementSpeed = 2
	tankSprite.RotationSpeed = 1

	tankSprite.Tag = name
	tankSprite.Pos = m.Position
	tankSprite.Size = m.Size
	tankSprite.Meta = &TankBodyMeta{
		Kind:             bodyKind,
		Facing:           1,
		CannonMountRight: cannonMount,
		CannonLength:     cannonLength,
	}

	m.Sprites.Add(tankSprite)

	cannonSprite := &engine.Sprite{}
	cannonSprite.MovementSpeed = 1
	cannonSprite.RotationSpeed = 2

	cannonSprite.Tag = "cannon"
	cannonSprite.Pos = &engine.Vec{
		X: m.Position.X + cannonMount.X - cannonAnchor.X,
		Y: m.Position.Y + cannonMount.Y - cannonAnchor.Y,
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

func newTankBodySprite(sprite []byte, tankColor color.RGBA) *engine.Sprite {
	img := colorizedSpriteImage(sprite, tankColor)
	bounds := img.Bounds()
	return &engine.Sprite{
		Image:    img,
		Drawable: engine.NewImageDrawableFrames(img, engine.R(0, 0, float64(bounds.Dx()), float64(bounds.Dy()))),
	}
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
	sprite := r.TankSmallSprite
	if meta, ok := body.Meta.(*TankBodyMeta); ok && meta.Kind == TankBodyKindXMV12 {
		sprite = r.XMV12TankSprite
	}
	body.Image = colorizedSpriteImage(sprite, tankColor)
	if meta, ok := body.Meta.(*TankBodyMeta); ok && meta.Kind == TankBodyKindXMV12 && meta.Facing < 0 {
		body.Image = mirroredImage(body.Image)
	}
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

func SetTankFacing(body *engine.Sprite, tankColor color.RGBA, facing int) {
	if body == nil {
		return
	}
	meta, ok := body.Meta.(*TankBodyMeta)
	if !ok || meta.Kind != TankBodyKindXMV12 {
		return
	}
	nextFacing := 1
	if facing < 0 {
		nextFacing = -1
	}
	if meta.Facing == nextFacing {
		return
	}
	meta.Facing = nextFacing
	RecolorTankBody(body, tankColor)
}

func TankCannonMount(body *engine.Sprite) engine.Vec {
	if body == nil {
		return SmallTankCannonMount
	}
	meta, ok := body.Meta.(*TankBodyMeta)
	if !ok {
		return SmallTankCannonMount
	}
	if meta.Kind != TankBodyKindXMV12 || meta.Facing >= 0 || body.Size == nil {
		return meta.CannonMountRight
	}
	return engine.V(body.Size.X-meta.CannonMountRight.X, meta.CannonMountRight.Y)
}

func mirroredImage(src *ebiten.Image) *ebiten.Image {
	if src == nil {
		return nil
	}
	bounds := src.Bounds()
	dst := ebiten.NewImage(bounds.Dx(), bounds.Dy())
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(-1, 1)
	op.GeoM.Translate(float64(bounds.Dx()), 0)
	dst.DrawImage(src, op)
	return dst
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
