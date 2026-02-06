package models

import (
	"bytes"
	_ "embed"
	"image"
	"image/color"
	"log"
	"math/rand"
	"time"

	"github.com/aquilax/go-perlin"

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

func NewGround(background *Background) Ground {

	m := Ground{Name: "ground"}

	m.Size = engine.Vec{X: core.Config().Screen.Width, Y: core.Config().Screen.Height / 2}

	emptyImage := ebiten.NewImage(int(core.Config().Screen.Width), int(core.Config().Screen.Height/2))
	emptyImage.Fill(color.Transparent)
	emptyGroundImage := emptyImage.SubImage(image.Rect(0, 0, int(core.Config().Screen.Width), int(core.Config().Screen.Height/2))).(*ebiten.Image)

	rawGroundImage := engine.NewSprite(r.GroundSprite, nil)
	rawGroundImage.Size = &m.Size
	rawGroundImage.Pos = &engine.Vec{X: 0, Y: 0}
	rawGroundImage.Draw(nil, emptyGroundImage)

	m.drawGroundAlpha(emptyGroundImage)

	m.Sprite.Tag = m.Name
	m.Sprite.Pos = &m.Position
	m.Sprite.Size = &m.Size

	// log.Printf("%v", m.Sprite.Image)
	// log.Printf("%v", groundSprite)

	m.Position = engine.Vec{X: 0, Y: background.Size.Y / 2}

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

// sets Coords, Sprite and MaxHeight of wave
func (m *Ground) drawGroundAlpha(destinationImage *ebiten.Image) {
	const (
		numPoints   = 64   // feinere Kurve
		amplitude   = 40.0 // maximale Höhenabweichung
		frequency   = 0.1  // wie oft Berge/Täler wechseln
		offset      = 10.0 // Kontrolle über Bezier-Krümmung
		perlinAlpha = 2.0
		perlinBeta  = 2.0
		perlinN     = 3
	)

	imgDecoded, _, err := image.Decode(bytes.NewReader(r.GroundSprite))
	if err != nil {
		log.Fatal("could not decode r.GroundSprite:", err)
	}
	groundImage := ebiten.NewImageFromImage(imgDecoded)

	width := float64(destinationImage.Bounds().Dx())
	height := float64(destinationImage.Bounds().Dy())
	baseHeight := height * 0.7

	p := perlin.NewPerlin(perlinAlpha, perlinBeta, perlinN, time.Now().UnixNano())

	points := make([]engine.Point, numPoints)
	for i := 0; i < numPoints; i++ {
		x := (width / float64(numPoints-1)) * float64(i)
		noise := p.Noise1D(float64(i) * frequency)
		y := baseHeight + noise*amplitude
		points[i] = engine.Point{X: float32(x), Y: float32(y)}
	}

	var path vector.Path
	path.MoveTo(points[0].X, points[0].Y)

	for i := 1; i < len(points); i++ {
		prev := points[i-1]
		cur := points[i]

		cpx0 := prev.X + float32(offset)
		cpy0 := prev.Y
		cpx1 := cur.X - float32(offset)
		cpy1 := cur.Y

		path.CubicTo(cpx0, cpy0, cpx1, cpy1, cur.X, cur.Y)
	}

	// Fläche nach unten schließen (rechte Seite, Boden, linke Seite)
	path.LineTo(float32(width), float32(height))
	path.LineTo(0, float32(height))
	path.LineTo(points[0].X, points[0].Y) // explizit zum Start zurück

	// Zeichnen auf Zielbild
	vs, is := path.AppendVerticesAndIndicesForFilling(nil, nil)
	for i := range vs {
		// Wiederhole die Textur
		vs[i].SrcX = float32(int(vs[i].DstX) % groundImage.Bounds().Dx())
		vs[i].SrcY = float32(int(vs[i].DstY) % groundImage.Bounds().Dy())
		vs[i].ColorR = 1
		vs[i].ColorG = 1
		vs[i].ColorB = 1
		vs[i].ColorA = 1
	}

	op := &ebiten.DrawTrianglesOptions{
		FillRule:      ebiten.EvenOdd,
		CompositeMode: ebiten.CompositeModeSourceOver,
	}

	destinationImage.DrawTriangles(vs, is, groundImage, op)

	m.Coords = vs
	m.Sprite = engine.NewSpriteFromImage(destinationImage)
	m.MaxHeight = float64(height)
}
