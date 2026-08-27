package models

import (
	_ "embed"
	"image"
	"image/color"
	"math"
	"math/rand"

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
	terrain  terrainProfile
}

type SandFallPixel struct {
	X     int
	FromY int
	ToY   int
	Color color.RGBA
}

type terrainProfile struct {
	baseFromBottom float64
	waves          []terrainWave
	hills          []terrainHill
}

type terrainWave struct {
	frequency float64
	phase     float64
	amplitude float64
}

type terrainHill struct {
	center float64
	width  float64
	height float64
}

func NewGround() Ground {
	return NewGroundWithWidth(core.Config().Screen.Width)
}

func NewGroundWithWidth(width float64) Ground {
	return NewGroundWithSize(width, core.Config().Screen.Height)
}

func NewGroundWithSize(width, height float64) Ground {
	return newGroundWithProfile(width, height, defaultTerrainProfile())
}

func NewRandomGroundWithSize(width, height float64, seed int64) Ground {
	return newGroundWithProfile(width, height, randomTerrainProfile(seed))
}

func newGroundWithProfile(width, height float64, terrain terrainProfile) Ground {
	m := Ground{Name: "ground"}
	m.Sprites = engine.NewSprites()
	m.Position = &engine.Vec{X: 0, Y: 0}
	m.Size = &engine.Vec{X: width, Y: height}
	m.terrain = terrain
	m.pixels = generateGroundImage(m.Size.X, m.Size.Y, terrain)
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

	profile := g.terrain
	if len(profile.waves) == 0 && len(profile.hills) == 0 {
		profile = defaultTerrainProfile()
	}
	return normalizedTerrainSurfaceY(width, g.Size.Y, profile, x)
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

func (g Ground) ApplyRingCrater(cx, cy, radius, spacing, thickness float64) []SandFallPixel {
	if g.pixels == nil || g.Image == nil || radius <= 0 {
		return nil
	}
	if spacing <= 0 {
		spacing = 2
	}
	if thickness <= 0 {
		thickness = 1
	}

	bounds := g.pixels.Bounds()
	minX := maxInt(0, int(math.Floor(cx-radius-thickness)))
	maxX := minInt(bounds.Dx()-1, int(math.Ceil(cx+radius+thickness)))
	minY := maxInt(0, int(math.Floor(cy-radius-thickness)))
	maxY := minInt(bounds.Dy()-1, int(math.Ceil(cy+radius+thickness)))
	halfThickness := thickness / 2

	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			distance := math.Hypot(float64(x)-cx, float64(y)-cy)
			if distance > radius {
				continue
			}
			ring := math.Round(distance / spacing)
			if math.Abs(distance-ring*spacing) <= halfThickness {
				g.pixels.SetRGBA(x, y, color.RGBA{})
			}
		}
	}

	falls := g.settleSandRange(minX, maxX, maxY)
	g.refreshSurfaceRange(minX, maxX)
	g.Image.WritePixels(g.pixels.Pix)
	return falls
}

func (g Ground) ClearCircle(cx, cy, radius float64) image.Rectangle {
	if g.pixels == nil || g.Image == nil || radius <= 0 {
		return image.Rectangle{}
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

	g.refreshSurfaceRange(minX, maxX)
	g.Image.WritePixels(g.pixels.Pix)
	return image.Rect(minX, minY, maxX+1, maxY+1)
}

func (g Ground) ClearLine(x1, y1, x2, y2, thickness float64) image.Rectangle {
	if g.pixels == nil || g.Image == nil || thickness <= 0 {
		return image.Rectangle{}
	}

	bounds := g.pixels.Bounds()
	radius := thickness / 2
	minX := maxInt(0, int(math.Floor(math.Min(x1, x2)-radius)))
	maxX := minInt(bounds.Dx()-1, int(math.Ceil(math.Max(x1, x2)+radius)))
	minY := maxInt(0, int(math.Floor(math.Min(y1, y2)-radius)))
	maxY := minInt(bounds.Dy()-1, int(math.Ceil(math.Max(y1, y2)+radius)))

	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			if distancePointToSegment(float64(x), float64(y), x1, y1, x2, y2) <= radius {
				g.pixels.SetRGBA(x, y, color.RGBA{})
			}
		}
	}

	g.refreshSurfaceRange(minX, maxX)
	g.Image.WritePixels(g.pixels.Pix)
	return image.Rect(minX, minY, maxX+1, maxY+1)
}

func (g Ground) ClearRect(area image.Rectangle) image.Rectangle {
	if g.pixels == nil || g.Image == nil || area.Empty() {
		return image.Rectangle{}
	}

	bounds := g.pixels.Bounds()
	minX := maxInt(0, area.Min.X)
	maxX := minInt(bounds.Dx()-1, area.Max.X-1)
	minY := maxInt(0, area.Min.Y)
	maxY := minInt(bounds.Dy()-1, area.Max.Y-1)
	if minX > maxX || minY > maxY {
		return image.Rectangle{}
	}

	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			g.pixels.SetRGBA(x, y, color.RGBA{})
		}
	}

	g.refreshSurfaceRange(minX, maxX)
	g.Image.WritePixels(g.pixels.Pix)
	return image.Rect(minX, minY, maxX+1, maxY+1)
}

func (g Ground) ClearRects(areas []image.Rectangle) image.Rectangle {
	if g.pixels == nil || g.Image == nil || len(areas) == 0 {
		return image.Rectangle{}
	}

	bounds := g.pixels.Bounds()
	cleared := image.Rectangle{}
	for _, area := range areas {
		if area.Empty() {
			continue
		}
		minX := maxInt(0, area.Min.X)
		maxX := minInt(bounds.Dx()-1, area.Max.X-1)
		minY := maxInt(0, area.Min.Y)
		maxY := minInt(bounds.Dy()-1, area.Max.Y-1)
		if minX > maxX || minY > maxY {
			continue
		}
		for y := minY; y <= maxY; y++ {
			for x := minX; x <= maxX; x++ {
				g.pixels.SetRGBA(x, y, color.RGBA{})
			}
		}
		rect := image.Rect(minX, minY, maxX+1, maxY+1)
		if cleared.Empty() {
			cleared = rect
		} else {
			cleared = cleared.Union(rect)
		}
	}
	if cleared.Empty() {
		return image.Rectangle{}
	}

	g.refreshSurfaceRange(cleared.Min.X, cleared.Max.X-1)
	g.Image.WritePixels(g.pixels.Pix)
	return cleared
}

func (g Ground) SettleArea(area image.Rectangle) []SandFallPixel {
	if g.pixels == nil || area.Empty() {
		return nil
	}
	bounds := g.pixels.Bounds()
	minX := maxInt(0, area.Min.X)
	maxX := minInt(bounds.Dx()-1, area.Max.X-1)
	maxY := minInt(bounds.Dy()-1, area.Max.Y-1)
	if minX > maxX || maxY < 0 {
		return nil
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

func distancePointToSegment(px, py, x1, y1, x2, y2 float64) float64 {
	dx := x2 - x1
	dy := y2 - y1
	lengthSquared := dx*dx + dy*dy
	if lengthSquared == 0 {
		return math.Hypot(px-x1, py-y1)
	}
	t := ((px-x1)*dx + (py-y1)*dy) / lengthSquared
	t = math.Max(0, math.Min(1, t))
	closestX := x1 + dx*t
	closestY := y1 + dy*t
	return math.Hypot(px-closestX, py-closestY)
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

func (g Ground) AlignSpriteUprightToSurface(source *engine.Sprite) {
	if source == nil {
		return
	}

	source.Rot = 0
	source.Pos = &engine.Vec{
		X: math.Max(0, math.Min(g.Size.X-source.Size.X, source.Pos.X)),
		Y: source.Pos.Y,
	}

	minX := int(math.Floor(source.Pos.X))
	maxX := int(math.Ceil(source.Pos.X + source.Size.X))
	bottomY := g.SurfaceY(source.Pos.X + source.Size.X/2)
	for x := minX; x <= maxX; x++ {
		bottomY = math.Max(bottomY, g.SurfaceY(float64(x)))
	}

	source.Pos = &engine.Vec{
		X: source.Pos.X,
		Y: bottomY - source.Size.Y,
	}
	g.ClearRect(image.Rect(
		int(math.Floor(source.Pos.X)),
		int(math.Floor(source.Pos.Y)),
		int(math.Ceil(source.Pos.X+source.Size.X)),
		int(math.Ceil(source.Pos.Y+source.Size.Y)),
	))
}

func defaultTerrainProfile() terrainProfile {
	return terrainProfile{
		baseFromBottom: 118,
		waves: []terrainWave{
			{frequency: 2.4, phase: 0.25, amplitude: 46},
			{frequency: 5.7, phase: 1.1, amplitude: 30},
			{frequency: 11.0, phase: 2.6, amplitude: 12},
		},
		hills: []terrainHill{
			{center: 0.18, width: 0.11, height: -68},
			{center: 0.36, width: 0.09, height: 58},
			{center: 0.54, width: 0.12, height: -54},
			{center: 0.76, width: 0.10, height: 64},
			{center: 0.91, width: 0.08, height: -42},
		},
	}
}

func randomTerrainProfile(seed int64) terrainProfile {
	rng := rand.New(rand.NewSource(seed))
	profile := terrainProfile{
		baseFromBottom: 105 + rng.Float64()*58,
		waves: []terrainWave{
			{frequency: 1.5 + rng.Float64()*2.2, phase: rng.Float64() * math.Pi * 2, amplitude: 50 + rng.Float64()*58},
			{frequency: 4.0 + rng.Float64()*4.6, phase: rng.Float64() * math.Pi * 2, amplitude: 24 + rng.Float64()*42},
			{frequency: 8.5 + rng.Float64()*8.5, phase: rng.Float64() * math.Pi * 2, amplitude: 8 + rng.Float64()*26},
		},
	}
	if rng.Intn(2) == 0 {
		profile.waves = append(profile.waves, terrainWave{
			frequency: 2.5 + rng.Float64()*5.5,
			phase:     rng.Float64() * math.Pi * 2,
			amplitude: 18 + rng.Float64()*38,
		})
	}

	hillCount := 5 + rng.Intn(5)
	for i := 0; i < hillCount; i++ {
		height := 45 + rng.Float64()*95
		if rng.Intn(2) == 0 {
			height = -height
		}
		profile.hills = append(profile.hills, terrainHill{
			center: 0.04 + rng.Float64()*0.92,
			width:  0.045 + rng.Float64()*0.12,
			height: height,
		})
	}
	return profile
}

func generateGroundImage(width, height float64, terrain terrainProfile) *image.RGBA {
	w := int(math.Ceil(width))
	h := int(math.Ceil(height))
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	surfaces := generateTerrainSurfaces(width, height, terrain)

	for x := 0; x < w; x++ {
		surface := int(math.Round(surfaces[x]))
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

func normalizedTerrainSurfaceY(width, height float64, terrain terrainProfile, x float64) float64 {
	if width <= 0 {
		return height
	}
	surfaces := generateTerrainSurfaces(width, height, terrain)
	if len(surfaces) == 0 {
		return height
	}
	column := int(math.Round(x))
	if column < 0 {
		column = 0
	}
	if column >= len(surfaces) {
		column = len(surfaces) - 1
	}
	return surfaces[column]
}

func generateTerrainSurfaces(width, height float64, terrain terrainProfile) []float64 {
	w := int(math.Ceil(width))
	if w <= 0 {
		return nil
	}

	raw := make([]float64, w)
	minRaw := math.Inf(1)
	maxRaw := math.Inf(-1)
	for x := 0; x < w; x++ {
		y := rawTerrainSurfaceY(width, height, terrain, float64(x))
		raw[x] = y
		minRaw = math.Min(minRaw, y)
		maxRaw = math.Max(maxRaw, y)
	}

	minY := height - 330
	maxY := height - 56
	allowedRange := maxY - minY
	rawRange := maxRaw - minRaw

	surfaces := make([]float64, w)
	switch {
	case rawRange > allowedRange && rawRange > 0:
		scale := allowedRange / rawRange
		for x, y := range raw {
			surfaces[x] = minY + (y-minRaw)*scale
		}
	default:
		offset := 0.0
		if minRaw < minY {
			offset = minY - minRaw
		}
		if maxRaw+offset > maxY {
			offset = maxY - maxRaw
		}
		for x, y := range raw {
			surfaces[x] = y + offset
		}
	}
	return surfaces
}

func rawTerrainSurfaceY(width, height float64, terrain terrainProfile, x float64) float64 {
	t := 0.0
	if width > 0 {
		t = x / width
	}
	y := height - terrain.baseFromBottom
	for _, wave := range terrain.waves {
		y += math.Sin(t*math.Pi*wave.frequency+wave.phase) * wave.amplitude
	}
	for _, hill := range terrain.hills {
		y += cartoonHill(t, hill.center, hill.width, hill.height)
	}
	return y
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
