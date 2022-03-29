package models

import (
	_ "embed"
	"image"
	"log"
	"math"
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
}

func NewGround() Ground {

	m := Ground{Name: "ground"}

	m.Sprite = engine.NewSprite(r.GroundSprite, r.GroundSpec)
	m.MaxHeight = m.drawWave()

	m.Position = engine.Vec{X: 0, Y: core.Config().Screen.Height - m.MaxHeight}
	m.Size = engine.Vec{X: core.Config().Screen.Width, Y: m.MaxHeight}

	log.Printf("maxHeight: %v\n", m.MaxHeight)

	m.Sprite.Tag = m.Name
	m.Sprite.Pos = &m.Position
	m.Sprite.Size = &m.Size

	return m
}

func (m *Ground) GetSprites() []*engine.Sprite {
	return []*engine.Sprite{m.Sprite}
}

func floatRand() float64 {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	return r1.Float64()
}

func intRand() int {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	return r1.Intn(int(core.Config().Screen.Height * 2))
}

func maxCounter(index int) float64 {
	return float64(128 + (17*index+32)%64)
}

// returns highest Point of wave
func (m *Ground) drawWave() float64 {

	emptyImage := ebiten.NewImage(3, 3)
	emptySubImage := emptyImage.SubImage(image.Rect(1, 1, 2, 2)).(*ebiten.Image)

	maxWidth := core.Config().Screen.Width
	maxHeight := core.Config().Screen.Height / 2

	var path vector.Path

	const npoints = 8
	indexToPoint := func(i int) (float32, float32) {
		x, y := float32(float32(i)*float32(maxWidth/(npoints-1))), float32(floatRand()*maxHeight)
		// y += float32(30 * math.Sin(float64(counter)*2*math.Pi/maxCounter(i)))
		y += float32(30 * math.Sin(2*math.Pi))
		return x, y
	}

	var highestPoint float32

	for i := 0; i <= npoints; i++ {
		if i == 0 {
			path.MoveTo(indexToPoint(i))
			continue
		}
		cpx0, cpy0 := indexToPoint(i - 1)
		x, y := indexToPoint(i)
		cpx1, cpy1 := x, y
		cpx0 += float32(intRand())
		cpx1 -= float32(intRand())
		path.CubicTo(cpx0, cpy0, cpx1, cpy1, x, y)
		log.Printf("current maxHeight: %v\n", y)
		if y > highestPoint {
			highestPoint = y
		}
	}
	path.LineTo(float32(core.Config().Screen.Width), float32(core.Config().Screen.Height))
	path.LineTo(0, float32(core.Config().Screen.Height))

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
	m.Sprite.Image.DrawTriangles(vs, is, emptySubImage, op)

	return float64(highestPoint)
}
