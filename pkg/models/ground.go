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
	Image    *ebiten.Image
	pixels   *image.RGBA
	surface  []float64
}

type SandFallPixel struct {
	X     int
	FromY int
	ToY   int
	Color color.RGBA
}

func NewGround() Ground {
	return NewGroundWithWidth(core.Config().Screen.Width)
}

func NewGroundWithWidth(width float64) Ground {
	return NewGroundWithSize(width, core.Config().Screen.Height)
}

func NewGroundWithSize(width, height float64) Ground {

	m := Ground{Name: "ground"}
	m.Sprites = engine.NewSprites()
	m.Position = &engine.Vec{X: 0, Y: 0}
	m.Size = &engine.Vec{X: width, Y: height}
	m.pixels = generateGroundImage(m.Size.X, m.Size.Y)
	m.surface = make([]float64, m.pixels.Bounds().Dx())
	m.refreshSurfaceRange(0, len(m.surface)-1)
	m.Image = ebiten.NewImageFromImage(m.pixels)

	groundSprite := &engine.Sprite{}
	groundSprite.Tag = m.Name
	groundSprite.Pos = m.Position
	groundSprite.Size = m.Size
	groundSprite.Drawable = engine.NewImageDrawable(m.Image)

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
	if len(g.surface) > 0 {
		column := int(math.Round(x))
		if column < 0 {
			column = 0
		}
		if column >= len(g.surface) {
			column = len(g.surface) - 1
		}
		return g.surface[column]
	}

	t := x / width
	base := g.Size.Y - 118
	y := base +
		math.Sin(t*math.Pi*2.4+0.25)*46 +
		math.Sin(t*math.Pi*5.7+1.1)*30 +
		math.Sin(t*math.Pi*11.0+2.6)*12

	y += cartoonHill(t, 0.18, 0.11, -68)
	y += cartoonHill(t, 0.36, 0.09, 58)
	y += cartoonHill(t, 0.54, 0.12, -54)
	y += cartoonHill(t, 0.76, 0.10, 64)
	y += cartoonHill(t, 0.91, 0.08, -42)

	minY := g.Size.Y - 245
	maxY := g.Size.Y - 34
	return math.Max(minY, math.Min(maxY, y))
}

func (g Ground) ApplyCrater(cx, cy, radius float64) []SandFallPixel {
	if g.pixels == nil || g.Image == nil || radius <= 0 {
		return nil
	}

	bounds := g.pixels.Bounds()
	minX := maxInt(0, int(math.Floor(cx-radius)))
	maxX := minInt(bounds.Dx()-1, int(math.Ceil(cx+radius)))
	minY := maxInt(0, int(math.Floor(cy-radius)))
	maxY := minInt(bounds.Dy()-1, int(math.Ceil(cy+radius)))
	r2 := radius * radius

	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			dx := float64(x) - cx
			dy := float64(y) - cy
			if dx*dx+dy*dy <= r2 {
				g.pixels.SetRGBA(x, y, color.RGBA{})
			}
		}
	}

	falls := g.settleSandRange(minX, maxX, maxY)
	g.refreshSurfaceRange(minX, maxX)
	g.Image.WritePixels(g.pixels.Pix)
	return falls
}

func (g Ground) settleSandRange(minX, maxX, maxY int) []SandFallPixel {
	if g.pixels == nil {
		return nil
	}

	bounds := g.pixels.Bounds()
	maxY = minInt(maxY, bounds.Dy()-1)
	falls := make([]SandFallPixel, 0)

	for x := minX; x <= maxX; x++ {
		writeY := maxY
		for y := maxY; y >= 0; y-- {
			c := g.pixels.RGBAAt(x, y)
			if c.A == 0 {
				continue
			}
			if y != writeY {
				g.pixels.SetRGBA(x, writeY, c)
				g.pixels.SetRGBA(x, y, color.RGBA{})
				falls = append(falls, SandFallPixel{X: x, FromY: y, ToY: writeY, Color: c})
			}
			writeY--
		}
	}

	return falls
}

func (g Ground) refreshSurfaceRange(minX, maxX int) {
	if g.pixels == nil || len(g.surface) == 0 {
		return
	}
	if minX < 0 {
		minX = 0
	}
	if maxX >= len(g.surface) {
		maxX = len(g.surface) - 1
	}

	bounds := g.pixels.Bounds()
	for x := minX; x <= maxX; x++ {
		g.surface[x] = float64(bounds.Dy())
		for y := 0; y < bounds.Dy(); y++ {
			if g.pixels.RGBAAt(x, y).A > 0 {
				g.surface[x] = float64(y)
				break
			}
		}
	}
}

func cartoonHill(t, center, width, height float64) float64 {
	d := math.Abs(t-center) / width
	if d >= 1 {
		return 0
	}
	return math.Cos(d*math.Pi/2) * height
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

func generateGroundImage(width, height float64) *image.RGBA {
	w := int(math.Ceil(width))
	h := int(math.Ceil(height))
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	ground := Ground{Size: &engine.Vec{X: width, Y: height}}
	for x := 0; x < w; x++ {
		surface := int(math.Round(ground.SurfaceY(float64(x))))
		for y := surface; y < h; y++ {
			depth := float64(y-surface) / math.Max(1, float64(h-surface))
			img.SetRGBA(x, y, desertSandColor(depth))
		}
		for y := surface - 2; y <= surface+1; y++ {
			if y >= 0 && y < h {
				img.SetRGBA(x, y, color.RGBA{R: 255, G: 250, B: 178, A: 255})
			}
		}
	}

	return img
}

func desertSandColor(depth float64) color.RGBA {
	depth = math.Max(0, math.Min(1, depth))

	top := color.RGBA{R: 255, G: 251, B: 185, A: 255}
	mid := color.RGBA{R: 226, G: 181, B: 111, A: 255}
	bottom := color.RGBA{R: 196, G: 126, B: 67, A: 255}

	if depth < 0.42 {
		return blendRGBA(top, mid, depth/0.42)
	}
	return blendRGBA(mid, bottom, (depth-0.42)/0.58)
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func blendRGBA(a, b color.RGBA, t float64) color.RGBA {
	t = math.Max(0, math.Min(1, t))

	return color.RGBA{
		R: uint8(float64(a.R) + (float64(b.R)-float64(a.R))*t),
		G: uint8(float64(a.G) + (float64(b.G)-float64(a.G))*t),
		B: uint8(float64(a.B) + (float64(b.B)-float64(a.B))*t),
		A: uint8(float64(a.A) + (float64(b.A)-float64(a.A))*t),
	}
}
