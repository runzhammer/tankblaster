package engine

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// sprite is a game sprite that has basic physics, optional
// graphics, and associated Behaviors. It can be used standalone
// or managed (Updated and Drawn) by sprites.
type sprite struct {
	// Tag is an optional identifier for this type of sprite.
	// It can be retrieved as an spriteSet from an sprites by
	// this tag along with other sprites with the same tag.
	Tag string

	// Pos is the position of the sprite. The Drawable, if any,
	// will be drawn with this as the origin.
	Pos Vec
	// Size is the size of the sprite. The Drawable, if any,
	// will be scaled to fit.
	Size Vec
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

// Bounds gets the hitbox for this sprite. Any Drawable will
// scaled and translated to fit this box. Collision detection
// can be performed using this Rect.
func (o *sprite) Bounds() Rect {
	return R(o.Pos.X, o.Pos.Y, o.Pos.X+o.Size.X, o.Pos.Y+o.Size.Y)
}

// HitTest performs a hit test for the given point.
func (o *sprite) HitTest(v Vec) bool {
	if o == nil {
		return false
	}
	return o.Pos.X <= v.X &&
		o.Pos.X+o.Size.X >= v.X &&
		o.Pos.Y <= v.Y &&
		o.Pos.Y+o.Size.Y >= v.Y
}

// Draw will render this sprite on a target if a Drawable is associated with
// this sprite. The sprite's Drawable will be scaled and translated to fit
// this sprite's Bounds. It will also be rotated by Rot radians to
// This function does nothing if this sprite has no Drawable.
//
// The camera transformation is applied to draw, if it is not nil.
func (o *sprite) Draw(camera *ebiten.GeoM, image *ebiten.Image) {
	if o.Drawable == nil {
		return
	}
	bounds := o.Bounds()
	mat := FitRotated(o.Rot+o.RotNormal, o.Drawable.Bounds(), bounds)
	if camera != nil {
		mat.Concat(*camera)
	}
	o.Drawable.DrawAbsolute(image, mat)
}
