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
	r "github.com/runzhammer/gamedemo/resources"
)

type Ground struct {
	Name      string
	Sprite    *engine.Sprite
	Position  engine.Vec
	Size      engine.Vec
	MaxHeight float64
	Coords    []ebiten.Vertex
}

func NewGround() Ground {

	m := Ground{Name: "ground"}

	m.Size = engine.Vec{X: core.Config().Screen.Width, Y: core.Config().Screen.Height / 2}

	emptyImage := ebiten.NewImage(int(core.Config().Screen.Width), int(core.Config().Screen.Height/2))
	emptyImage.Fill(color.Transparent)
	emptyGroundImage := emptyImage.SubImage(image.Rect(0, 0, int(core.Config().Screen.Width), int(core.Config().Screen.Height/2))).(*ebiten.Image)

	rawGroundImage := engine.NewSprite(r.GroundSprite, r.GroundSpec)
	rawGroundImage.Size = &m.Size
	rawGroundImage.Pos = &engine.Vec{X: 0, Y: 0}
	rawGroundImage.Draw(nil, emptyGroundImage)

	m.drawGroundAlpha(emptyGroundImage)

	m.Sprite.Tag = m.Name
	m.Sprite.Pos = &m.Position
	m.Sprite.Size = &m.Size

	// log.Printf("%v", m.Sprite.Image)
	// log.Printf("%v", groundSprite)

	m.Position = engine.Vec{X: 0, Y: core.Config().Screen.Height / 2}

	return m
}

func (m *Ground) GetGroundY(x float64) float64 {
	LowX := float32(0)
	for _, v := range m.Coords {
		if v.DstX == float32(x) {
			return float64(v.DstY + float32(m.Position.Y))
		}
		if v.DstX < float32(x) {
			LowX = v.DstY
			continue
		}
		if v.DstX > float32(x) {
			return float64(v.DstY - LowX + v.DstY + float32(m.Position.Y))
		}
	}
	return float64(0)
}

func (m *Ground) GetSprites() []*engine.Sprite {
	return []*engine.Sprite{m.Sprite}
}

func float32Rand() float32 {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	return r1.Float32()
}

// returns highest Point of wave
func (m *Ground) drawGroundAlpha(destinationImage *ebiten.Image) {

	numPoints := 8

	emptyImage := ebiten.NewImage(3, 3)
	emptyImage.Fill(color.White)
	emptySubImage := emptyImage.SubImage(image.Rect(1, 1, 2, 2)).(*ebiten.Image)

	maxWidth := float32(destinationImage.Bounds().Dx())
	maxSegmentWidth := float32(destinationImage.Bounds().Dx() / numPoints)
	maxHeight := float32(destinationImage.Bounds().Dy())
	minHeight := float32(destinationImage.Bounds().Dy()) - (1 / 8 * float32(destinationImage.Bounds().Dy()))

	// log.Printf("x: %v, y: %v", maxWidth, maxHeight)

	var path vector.Path

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
			y = engine.IntRand(maxHeight)
			x = 0
		} else {

			oldPoint := npoints[i-1]

			y = maxHeight * float32Rand()

			// distance to last point
			x = oldPoint.x + engine.IntRand(maxSegmentWidth)
		}

		if x >= maxWidth {
			x = maxWidth
			return x, y
		}

		if y > minHeight {
			y = minHeight
		}

		npoints[i] = npoint{x: x, y: y}

		return x, y
	}

	var highestPoint float32
	var X, Y float32

	for i := 0; X < maxWidth; i++ {
		if i == 0 {
			path.MoveTo(indexToPoint(i))
			continue
		}

		oldPoint := npoints[i-1]

		cpx0, cpy0 := oldPoint.x, oldPoint.y

		X, Y = indexToPoint(i)
		cpx1, cpy1 := X, Y

		cpx0 += 30
		cpx1 -= 30

		// log.Printf("old p: %v / %v", oldPoint.x, oldPoint.y)
		// log.Printf("new p: %v / %v ", X, Y)

		path.CubicTo(cpx0, cpy0, cpx1, cpy1, X, Y)
		//path.LineTo(X, Y)

		if Y < highestPoint {
			highestPoint = Y
		}
	}

	highestPoint = maxHeight - highestPoint

	log.Printf("highestPoint: %v", highestPoint)

	// path.LineTo(maxWidth, maxHeight)
	// path.LineTo(0, maxHeight)

	path.LineTo(maxWidth, 0)
	path.LineTo(0, 0)

	op := &ebiten.DrawTrianglesOptions{
		FillRule: ebiten.EvenOdd,
	}
	vs, is := path.AppendVerticesAndIndicesForFilling(nil, nil)
	for i := range vs {
		vs[i].SrcX = 1
		vs[i].SrcY = 1
		vs[i].ColorA = 0
		log.Printf("%v, %v", vs[i].DstX, vs[i].DstY)
	}

	op.CompositeMode = ebiten.CompositeModeCopy

	destinationImage.DrawTriangles(vs, is, emptySubImage, op)

	// ground points
	m.Coords = vs

	// for k, v := range m.Coords {
	// 	log.Printf("%v -> %v", k, v)
	// }

	m.Sprite = engine.NewSpriteFromImage(destinationImage)
	m.MaxHeight = float64(highestPoint)
}
