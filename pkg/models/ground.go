package models

import (
	_ "embed"
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/runzhammer/gamedemo/pkg/core"
	"github.com/runzhammer/gamedemo/pkg/engine"
)

type Ground struct {
	Name     string
	Sprites  *engine.Sprites
	Position *engine.Vec
	Size     *engine.Vec
}

func NewGround() Ground {

	m := Ground{Name: "ground"}
	m.Sprites = engine.NewSprites()
	m.Position = &engine.Vec{X: 0, Y: 0}
	m.Size = &engine.Vec{X: core.Config().Screen.Width, Y: core.Config().Screen.Height}

	groundSprite := &engine.Sprite{}
	groundSprite.Tag = m.Name
	groundSprite.Pos = m.Position
	groundSprite.Size = m.Size
	groundSprite.Drawable = engine.NewImageDrawable(generateGroundImage(m.Size.X, m.Size.Y))

	m.Sprites.Add(groundSprite)

	return m
}

func (g Ground) SurfaceY(x float64) float64 {
	width := g.Size.X
	if width <= 0 {
		return g.Position.Y
	}

	if x < 0 {
		x = 0
	}
	if x > width {
		x = width
	}

	t := x / width
	base := g.Size.Y - 96
	y := base +
		math.Sin(t*math.Pi*2.1+0.35)*28 +
		math.Sin(t*math.Pi*6.4+1.2)*12

	minY := g.Size.Y - 168
	maxY := g.Size.Y - 42
	return math.Max(minY, math.Min(maxY, y))
}

func (g Ground) AlignSpriteToSurface(source *engine.Sprite) {
	const treadContact = 0.34

	centerX := source.Pos.X + source.Size.X/2
	leftX := centerX - source.Size.X*treadContact
	rightX := centerX + source.Size.X*treadContact
	leftY := g.SurfaceY(leftX)
	rightY := g.SurfaceY(rightX)

	source.Rot = math.Atan2(rightY-leftY, rightX-leftX)

	centerY := (leftY+rightY)/2 - math.Cos(source.Rot)*source.Size.Y/2
	source.Pos = &engine.Vec{
		X: math.Max(0, math.Min(g.Size.X-source.Size.X, source.Pos.X)),
		Y: centerY - source.Size.Y/2,
	}
}

func generateGroundImage(width, height float64) *ebiten.Image {
	w := int(math.Ceil(width))
	h := int(math.Ceil(height))
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	ground := Ground{Size: &engine.Vec{X: width, Y: height}}
	for x := 0; x < w; x++ {
		surface := int(math.Round(ground.SurfaceY(float64(x))))
		for y := surface; y < h; y++ {
			shade := uint8(72 + math.Min(48, float64(y-surface)/2))
			img.SetRGBA(x, y, color.RGBA{R: 52, G: shade, B: 42, A: 255})
		}
		for y := surface - 2; y <= surface+1; y++ {
			if y >= 0 && y < h {
				img.SetRGBA(x, y, color.RGBA{R: 95, G: 132, B: 68, A: 255})
			}
		}
	}

	return ebiten.NewImageFromImage(img)
}
