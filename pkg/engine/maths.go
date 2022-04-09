package engine

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// DegToRad converts degrees to radians
func DegToRad(deg float64) (rad float64) {
	return deg * math.Pi / 180
}

// RadToDeg converts radians to degrees
func RadToDeg(rad float64) (deg float64) {
	return rad * 180 / math.Pi
}

// Fit returns the Matrix that will transform a source Rect
// into the dest Rect
func Fit(source, dest Rect) ebiten.GeoM {
	scaleX := dest.W() / source.W()
	scaleY := dest.H() / source.H()

	mat := ebiten.GeoM{}
	mat.Translate(-source.Min.X, -source.Min.Y)
	mat.Scale(scaleX, scaleY)
	mat.Translate(dest.Min.X, dest.Min.Y)

	return mat
}

// FitGeoM returns the Matrix that will transform a source Rect
// into the dest Rect
func FitRotated(source *Sprite, dest Rect) ebiten.GeoM {

	scaleX := dest.W() / source.Drawable.Bounds().W()
	scaleY := dest.H() / source.Drawable.Bounds().H()

	mat := ebiten.GeoM{}

	// rotate about center of source
	if source.RotCenter {
		mat.Translate(-source.Drawable.Bounds().W()/2, -source.Drawable.Bounds().H()/2)
	}

	mat.Rotate(source.Rot + source.RotNormal)

	if source.RotCenter {
		mat.Translate(source.Drawable.Bounds().W()/2, source.Drawable.Bounds().H()/2)
	}

	// scale
	mat.Scale(scaleX, scaleY)

	// move to destination
	mat.Translate(dest.Min.X, dest.Min.Y)

	return mat
}

// Collision returns if two rectangles intersect
func Collision(r1, r2 Rect) bool {
	if r1.Min.X > r2.Max.X || r2.Min.X > r1.Max.X {
		return false
	}
	if r1.Min.Y > r2.Max.Y || r2.Min.Y > r1.Max.Y {
		return false
	}
	return true
}

func RotationOnGround(source *Sprite) {
  float s = sin(angle);
  float c = cos(angle);

  // translate point back to origin:
  p.x -= cx;
  p.y -= cy;

  // rotate point
  float xnew = p.x * c - p.y * s;
  float ynew = p.x * s + p.y * c;

  // translate point back:
  p.x = xnew + cx;
  p.y = ynew + cy;
  return p;
}