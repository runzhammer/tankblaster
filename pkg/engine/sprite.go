package engine

import (
	"bytes"
	_ "embed"
	"image"
	_ "image/png"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"gopkg.in/yaml.v3"
)

type Sprite struct {
	// The actual image loaded from resources
	Image *ebiten.Image

	DrawImageOptions *ebiten.DrawImageOptions

	// Spec of sprite sheet
	SpriteSpec spriteSpec

	// Tag is an optional identifier for this type of sprite.
	// It can be retrieved as an spriteSet from an sprites by
	// this tag along with other sprites with the same tag.
	Tag string

	// Pos is the position of the sprite. The Drawable, if any,
	// will be drawn with this as the origin.
	Pos *Vec

	// Size is the size of the sprite. The Drawable, if any,
	// will be scaled to fit.
	Size *Vec

	// Movement Speed of Sprite
	MovementSpeed float64

	// Rotation Speed of Sprite
	RotationSpeedPerSecond float64

	// Rotate on Center of Sprite
	RotCenter bool

	// Velocity is the Vec describing the movement speed
	// and direction of this sprite.
	Velocity Vec

	// Drawable is an optional Drawable to use to draw this
	// sprite on a Target.
	Drawable Drawable

	// Rot is an amount in radians used to rotate the Drawable
	// where 0 degrees is right and 90 degrees is upwards.
	Rot float64

	// RotNormal is the amount that the drawable should be rotated
	// initially such that its default orientation is right-facing,
	// or 0 degrees.
	RotNormal float64

	// MaxRange
	// first value max positive radiant (left)
	// second value max negative radiant (right)
	// unit: degree
	MaxRange []float64

	// PreSteps is Behaviors to execute before Steps and
	// PostSteps during an Update performed by sprites.
	PreSteps Behaviors
	// Steps is Behaviors to execute before PostSteps and
	// after PreSteps during an Update performed by sprites.
	Steps Behaviors
	// PostSteps is Behaviors to execute after Steps during
	// an Update performed by sprites.
	PostSteps Behaviors

	// Meta is an arbitrary value used to hold miscellaneous
	// information about this sprite.
	// It is not used by the engine library.
	Meta interface{}
}

type spriteSpec struct {
	Still struct {
		Speed  float64 `yaml:"speed"`
		Frames []frameSpec
	} `yaml:"still"`
	Stand struct {
		Speed  float64 `yaml:"speed"`
		Frames []frameSpec
	} `yaml:"stand"`
	Drive struct {
		Speed  float64 `yaml:"speed"`
		Frames []frameSpec
	} `yaml:"drive"`
}

type frameSpec struct {
	X int `yaml:"x"`
	Y int `yaml:"y"`
	W int `yaml:"w"`
	H int `yaml:"h"`
}

func NewSprite(sprite []byte, animations []byte) *Sprite {

	var err error
	s := Sprite{}

	// load sprite
	spriteDecoded, _, err := image.Decode(bytes.NewReader(sprite))
	if err != nil {
		log.Fatal(err)
	}

	s.Image = ebiten.NewImageFromImage(spriteDecoded)

	s.DrawImageOptions = &ebiten.DrawImageOptions{}

	if err != nil {
		log.Fatal(err)
	}

	// load frames
	err = yaml.Unmarshal(animations, &s.SpriteSpec)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	s.MaxRange = []float64{-360, 360}

	s.Drawable = NewImageDrawableFrames(s.Image, R(0, 0, float64(s.SpriteSpec.Still.Frames[0].W), float64(s.SpriteSpec.Still.Frames[0].H)))

	return &s
}

// Bounds gets the hitbox for this sprite. Any Drawable will be
// scaled and translated to fit this box. Collision detection
// can be performed using this Rect.
func (s *Sprite) Bounds() Rect {
	return R(s.Pos.X, s.Pos.Y, s.Pos.X+s.Size.X, s.Pos.Y+s.Size.Y)
}

// HitTest performs a hit test for the given point.
func (s *Sprite) HitTest(v Vec) bool {
	if s == nil {
		return false
	}
	return s.Pos.X <= v.X &&
		s.Pos.X+s.Size.X >= v.X &&
		s.Pos.Y <= v.Y &&
		s.Pos.Y+s.Size.Y >= v.Y
}

// Draw will render this sprite on a target if a Drawable is associated with
// this sprite. The sprite's Drawable will be scaled and translated to fit
// this sprite's Bounds. It will also be rotated by Rot radians.
// This function does nothing if this sprite has no Drawable.
//
// The camera transformation is applied to draw, if it is not nil.
func (s *Sprite) Draw(camera *ebiten.GeoM, screen *ebiten.Image) {
	if s.Drawable == nil {
		return
	}
	bounds := s.Bounds()

	mat := FitRotated(s, bounds)
	if camera != nil {
		mat.Concat(*camera)
	}
	s.Drawable.DrawAbsolute(screen, mat)
}
