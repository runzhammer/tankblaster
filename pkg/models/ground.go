package models

import (
	_ "embed"
	"image"
	"image/color"
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

func (m *Ground) intRand() int {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	return r1.Intn(int(m.Sprite.Drawable.Bounds().Max.Y))
}

// returns highest Point of wave
func (m *Ground) drawWave() float64 {

	emptyImage := ebiten.NewImage(3, 3)
	emptyImage.Fill(color.White)
	emptySubImage := emptyImage.SubImage(image.Rect(1, 1, 2, 2)).(*ebiten.Image)

	maxWidth := float32(core.Config().Screen.Width)
	maxHeight := float32(m.Sprite.Drawable.Bounds().Max.Y) // core.Config().Screen.Height / 2

	var path vector.Path

	npoints := 8
	indexToPoint := func(i int) (float32, float32) {
		x, y := float32(float32(i)*float32(maxWidth)/float32(npoints-1)), float32Rand()*maxHeight
		// y += float32(30 * math.Sin(float64(counter)*2*math.Pi/maxCounter(i)))
		y += float32(30 * math.Sin(2*math.Pi/maxCounter(i)))
		return x, y
	}

	var highestPoint float32

	for i := 1; i <= npoints+1; i++ {
		cpx0, cpy0 := indexToPoint(i - 1)
		x, y := indexToPoint(i)
		cpx1, cpy1 := x, y
		cpx0 += float32(m.intRand())
		cpx1 -= float32(m.intRand())
		path.CubicTo(cpx0, cpy0, cpx1, cpy1, x, y)
		if y > highestPoint {
			highestPoint = y
		}
	}
	path.LineTo(maxWidth, maxHeight)
	path.LineTo(0, maxHeight)

	op := &ebiten.DrawTrianglesOptions{
		FillRule: ebiten.FillAll,
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

func float32Rand() float32 {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	return r1.Float32()
}

func maxCounter(index int) float64 {
	return float64(128 + (17*index+32)%64)
}
