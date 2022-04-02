package models

import (
	_ "embed"
	"image"
	"image/color"
	"log"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/runzhammer/gamedemo/pkg/core"
	"github.com/runzhammer/gamedemo/pkg/engine"
)

type Ground struct {
	Name      string
	Sprite    *engine.Sprite
	Position  engine.Vec
	Size      engine.Vec
	MaxHeight float64
}

func NewGround() Ground {

	m := Ground{Name: "ground"}

	// m.Sprite = engine.NewSprite(r.GroundSprite, r.GroundSpec)
	m.Size = engine.Vec{X: core.Config().Screen.Width, Y: core.Config().Screen.Height / 2}

	groundImage := ebiten.NewImage(int(m.Size.X), int(m.Size.Y))
	groundImage.Fill(color.White)
	m.drawGround(groundImage)

	m.Sprite.Tag = m.Name
	m.Sprite.Pos = &m.Position
	m.Sprite.Size = &m.Size

	// log.Printf("%v", m.Sprite.Image)
	// log.Printf("%v", groundSprite)

	m.Position = engine.Vec{X: 0, Y: core.Config().Screen.Height - m.MaxHeight}

	return m
}

func (m *Ground) GetSprites() []*engine.Sprite {
	return []*engine.Sprite{m.Sprite}
}

type Number interface {
	float64 | int | float32
}

func intRand[T Number](maxY T) T {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	return T(r1.Intn(int(maxY)))
}

func float32Rand() float32 {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	return r1.Float32()
}

// returns highest Point of wave
func (m *Ground) drawGround(destinationImage *ebiten.Image) {

	emptyImage := ebiten.NewImage(3, 3)
	emptyImage.Fill(color.White)
	emptySubImage := emptyImage.SubImage(image.Rect(1, 1, 2, 2)).(*ebiten.Image)

	maxWidth := float32(destinationImage.Bounds().Dx())
	maxHeight := float32(destinationImage.Bounds().Dy()) // core.Config().Screen.Height / 2

	log.Printf("x: %v, y: %v", maxWidth, maxHeight)

	var path vector.Path

	numPoints := 8

	type npoint struct {
		x float32
		y float32
	}

	npoints := make(map[int]npoint)

	indexToPoint := func(i int) (float32, float32) {

		// log.Printf("maxCounter: %v", maxCounter(i))
		// x, y := maxWidth*float32(i)/float32(numPoints-1), intRand(maxHeight)
		var x, y float32

		if i == 0 {
			y = intRand(maxHeight)
			x = 0
		} else {

			oldPoint := npoints[i-1]

			// next max height
			if oldPoint.y > maxHeight-maxHeight/4 {
				y = oldPoint.y - intRand(maxHeight)
				if y < 0 {
					y = 0
				}
			} else {
				y = maxHeight / 4
			}

			// distance to last point
			x += oldPoint.x + intRand(float32(maxWidth))/float32(numPoints) - 1
		}

		// distance to last point
		x += float32(intRand(200))

		if x >= maxWidth {
			x = maxWidth
			return x, y
		}

		npoints[i] = npoint{x: x, y: y}

		return x, y
	}

	var highestPoint float32

	for i := 0; i <= numPoints; i++ {
		if i == 0 {
			path.MoveTo(indexToPoint(i))
			continue
		}

		oldPoint := npoints[i-1]

		// cpx0, cpy0 := oldPoint.x, oldPoint.y

		x, y := indexToPoint(i)
		// cpx1, cpy1 := x, y

		// cpx0 += 30
		// cpx1 -= 30

		log.Printf("old p: %v / %v", oldPoint.x, oldPoint.y)
		log.Printf("new p: %v / %v ", x, y)

		// path.CubicTo(cpx0, cpy0, cpx1, cpy1, x, y)
		path.LineTo(x, y)

		if y > highestPoint {
			highestPoint = y
		}
	}

	log.Printf("highestPoint: %v", maxHeight)

	path.LineTo(maxWidth, maxHeight)
	path.LineTo(0, maxHeight)

	op := &ebiten.DrawTrianglesOptions{
		FillRule: ebiten.EvenOdd,
	}
	vs, is := path.AppendVerticesAndIndicesForFilling(nil, nil)
	for i := range vs {
		vs[i].SrcX = 1
		vs[i].SrcY = 1
		vs[i].ColorR = 0x33 / float32(0xff)
		vs[i].ColorG = 0x66 / float32(0xff)
		vs[i].ColorB = 0xff / float32(0xff)
	}

	destinationImage.DrawTriangles(vs, is, emptySubImage, op)
	m.Sprite = engine.NewSpriteFromImage(destinationImage)
	m.MaxHeight = float64(highestPoint)
}

func maxCounter(index int) float32 {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	return r1.Float32()
}

// func maxCounter(index int) float64 {
// 	return float64(128 + (17*index+32)%64)
// }
