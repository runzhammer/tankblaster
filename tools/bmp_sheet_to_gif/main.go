package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
)

type sheetSpec struct {
	source      string
	target      string
	frameWidth  int
	frameHeight int
	delay       int
	colorMode   colorMode
}

type colorMode uint8

const (
	colorModeBMP colorMode = iota
	colorModeAsset
)

func main() {
	defaults := flag.Bool("defaults", false, "convert the known original animation sheets")
	source := flag.String("src", "", "source BMP spritesheet")
	target := flag.String("dst", "", "target PNG spritesheet")
	frameWidth := flag.Int("frame-width", 0, "single frame width in pixels")
	frameHeight := flag.Int("frame-height", 0, "single frame height in pixels; defaults to the full sheet height")
	delay := flag.Int("delay", 6, "GIF frame delay in 1/100 seconds")
	assetColors := flag.Bool("asset-colors", false, "read channels like the original game assets that are not normal BMP BGR")
	flag.Parse()

	specs := []sheetSpec{}
	if *defaults {
		specs = append(specs, defaultZeroPowerSpecs()...)
	}
	if *source != "" || *target != "" || *frameWidth > 0 {
		if *source == "" || *target == "" || *frameWidth <= 0 {
			fail("custom conversion requires -src, -dst and -frame-width")
		}
		mode := colorModeBMP
		if *assetColors {
			mode = colorModeAsset
		}
		specs = append(specs, sheetSpec{source: *source, target: *target, frameWidth: *frameWidth, frameHeight: *frameHeight, delay: *delay, colorMode: mode})
	}
	if len(specs) == 0 {
		fail("nothing to do; use -defaults or provide -src, -dst and -frame-width")
	}

	for _, spec := range specs {
		img, err := readBMPAssetRGB(spec.source, spec.colorMode)
		if err != nil {
			fail(err.Error())
		}
		if err := writeSheetPNG(spec.target, img, spec.frameWidth, spec.frameHeight); err != nil {
			fail(err.Error())
		}
		frameHeight := spec.frameHeight
		if frameHeight <= 0 {
			frameHeight = img.Bounds().Dy()
		}
		fmt.Printf("%s -> %s (%dx%d, %d frames)\n", spec.source, spec.target, img.Bounds().Dx(), img.Bounds().Dy(), (img.Bounds().Dx()/spec.frameWidth)*(img.Bounds().Dy()/frameHeight))
	}
}

func defaultZeroPowerSpecs() []sheetSpec {
	return []sheetSpec{
		{source: "original_assets/BITMAP/IDB_DUSTEXPLO.bmp", target: "resources/zero_power_dust_explosion.png", frameWidth: 20, delay: 6, colorMode: colorModeAsset},
		{source: "original_assets/BITMAP/IDB_EXPLOSION.bmp", target: "resources/zero_power_explosion.png", frameWidth: 67, frameHeight: 64, delay: 6, colorMode: colorModeBMP},
		{source: "original_assets/BITMAP/IDB_EXPLOSION_PILZ.bmp", target: "resources/zero_power_mushroom_explosion.png", frameWidth: 51, frameHeight: 57, delay: 6, colorMode: colorModeAsset},
		{source: "original_assets/BITMAP/IDB_PLAYER_RAUCHEN.bmp", target: "resources/zero_power_player_smoke.png", frameWidth: 21, delay: 6, colorMode: colorModeBMP},
		{source: "original_assets/BITMAP/IDB_FIREBALL.bmp", target: "resources/fireball_impact.png", frameWidth: 36, delay: 8, colorMode: colorModeBMP},
		{source: "original_assets/BITMAP/IDB_PALME_FEUER.bmp", target: "resources/palm_fire.png", frameWidth: 86, delay: 8, colorMode: colorModeAsset},
		{source: "original_assets/BITMAP/IDB_PALME_GERIPPE.bmp", target: "resources/palm_skeleton.png", frameWidth: 121, delay: 8, colorMode: colorModeBMP},
		{source: "original_assets/BITMAP/IDB_PALME_RAUCH.bmp", target: "resources/palm_smoke.png", frameWidth: 73, delay: 8, colorMode: colorModeBMP},
		{source: "original_assets/BITMAP/IDB_PALME_BROESEL.bmp", target: "resources/palm_crumble.png", frameWidth: 121, frameHeight: 152, delay: 8, colorMode: colorModeAsset},
		{source: "original_assets/BITMAP/IDB_WATERTEXTURE.bmp", target: "resources/water_texture.png", frameWidth: 64, delay: 8, colorMode: colorModeAsset},
		{source: "original_assets/BITMAP/IDB_WATER_BLUBBER.bmp", target: "resources/water_blubber.png", frameWidth: 19, delay: 8, colorMode: colorModeBMP},
	}
}

func readBMPAssetRGB(path string, mode colorMode) (*image.RGBA, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(data) < 54 || string(data[:2]) != "BM" {
		return nil, fmt.Errorf("%s: not a BMP", path)
	}

	offset := int(binary.LittleEndian.Uint32(data[10:14]))
	dibSize := int(binary.LittleEndian.Uint32(data[14:18]))
	if dibSize < 40 {
		return nil, fmt.Errorf("%s: unsupported DIB header", path)
	}
	width := int(int32(binary.LittleEndian.Uint32(data[18:22])))
	heightValue := int(int32(binary.LittleEndian.Uint32(data[22:26])))
	planes := binary.LittleEndian.Uint16(data[26:28])
	bpp := int(binary.LittleEndian.Uint16(data[28:30]))
	compression := binary.LittleEndian.Uint32(data[30:34])
	if planes != 1 || compression != 0 {
		return nil, fmt.Errorf("%s: unsupported BMP planes=%d compression=%d", path, planes, compression)
	}

	topDown := heightValue < 0
	height := heightValue
	if height < 0 {
		height = -height
	}

	palette := readPalette(data, offset, dibSize, bpp, mode)
	stride := ((width*bpp + 31) / 32) * 4
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		sourceY := y
		if !topDown {
			sourceY = height - 1 - y
		}
		row := offset + sourceY*stride
		for x := 0; x < width; x++ {
			c, ok, err := bmpPixel(data, row, x, bpp, palette, mode)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", path, err)
			}
			if !ok {
				continue
			}
			img.SetRGBA(x, y, c)
		}
	}

	transparent := img.RGBAAt(0, 0)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			c := img.RGBAAt(x, y)
			if sameRGB(c, transparent) || isTransparentKey(c) {
				c.A = 0
				img.SetRGBA(x, y, c)
			}
		}
	}
	return img, nil
}

func readPalette(data []byte, offset, dibSize, bpp int, mode colorMode) []color.RGBA {
	if bpp > 8 {
		return nil
	}
	count := (offset - 14 - dibSize) / 4
	start := 14 + dibSize
	palette := make([]color.RGBA, 0, count)
	for i := 0; i < count; i++ {
		p := start + i*4
		if p+3 >= len(data) {
			break
		}
		palette = append(palette, sourceRGB(data[p], data[p+1], data[p+2], mode))
	}
	return palette
}

func bmpPixel(data []byte, row, x, bpp int, palette []color.RGBA, mode colorMode) (color.RGBA, bool, error) {
	switch bpp {
	case 24:
		p := row + x*3
		if p+2 >= len(data) {
			return color.RGBA{}, false, nil
		}
		return sourceRGB(data[p], data[p+1], data[p+2], mode), true, nil
	case 8:
		p := row + x
		if p >= len(data) || int(data[p]) >= len(palette) {
			return color.RGBA{}, false, nil
		}
		return palette[data[p]], true, nil
	case 4:
		p := row + x/2
		if p >= len(data) {
			return color.RGBA{}, false, nil
		}
		index := data[p] >> 4
		if x%2 == 1 {
			index = data[p] & 0x0f
		}
		if int(index) >= len(palette) {
			return color.RGBA{}, false, nil
		}
		return palette[index], true, nil
	default:
		return color.RGBA{}, false, fmt.Errorf("unsupported bpp=%d", bpp)
	}
}

func bmpRGB(blue, green, red byte) color.RGBA {
	return color.RGBA{R: red, G: green, B: blue, A: 255}
}

func assetRGB(first, second, third byte) color.RGBA {
	return color.RGBA{R: first, G: third, B: second, A: 255}
}

func sourceRGB(first, second, third byte, mode colorMode) color.RGBA {
	if mode == colorModeAsset {
		return assetRGB(first, second, third)
	}
	return bmpRGB(first, second, third)
}

func writeSheetPNG(path string, sheet *image.RGBA, frameWidth, frameHeight int) error {
	bounds := sheet.Bounds()
	if frameWidth <= 0 || bounds.Dx()%frameWidth != 0 {
		return fmt.Errorf("%s: width %d is not divisible by frame width %d", path, bounds.Dx(), frameWidth)
	}
	if frameHeight <= 0 {
		frameHeight = bounds.Dy()
	}
	if bounds.Dy()%frameHeight != 0 {
		return fmt.Errorf("%s: height %d is not divisible by frame height %d", path, bounds.Dy(), frameHeight)
	}

	out := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	cols := bounds.Dx() / frameWidth
	rows := bounds.Dy() / frameHeight
	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			src := image.Rect(
				bounds.Min.X+col*frameWidth,
				bounds.Min.Y+row*frameHeight,
				bounds.Min.X+(col+1)*frameWidth,
				bounds.Min.Y+(row+1)*frameHeight,
			)
			frame := image.NewRGBA(image.Rect(0, 0, frameWidth, frameHeight))
			draw.Draw(frame, frame.Bounds(), sheet, src.Min, draw.Src)
			makeFrameBackgroundTransparent(frame)
			draw.Draw(out, image.Rect(col*frameWidth, row*frameHeight, (col+1)*frameWidth, (row+1)*frameHeight), frame, image.Point{}, draw.Src)
		}
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	return png.Encode(file, out)
}

func makeFrameBackgroundTransparent(frame *image.RGBA) {
	bounds := frame.Bounds()
	transparent := frame.RGBAAt(bounds.Min.X, bounds.Min.Y)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := frame.RGBAAt(x, y)
			if c.A != 0 && sameRGB(c, transparent) {
				c.A = 0
				frame.SetRGBA(x, y, c)
			}
		}
	}
	removeConnectedBackground(frame)
	removeSolidEdgeColumns(frame)
}

func removeConnectedBackground(frame *image.RGBA) {
	bounds := frame.Bounds()
	seen := make([]bool, bounds.Dx()*bounds.Dy())
	queue := make([]image.Point, 0, bounds.Dx()*2+bounds.Dy()*2)
	enqueue := func(x, y int) {
		if x < bounds.Min.X || x >= bounds.Max.X || y < bounds.Min.Y || y >= bounds.Max.Y {
			return
		}
		index := (y-bounds.Min.Y)*bounds.Dx() + x - bounds.Min.X
		if seen[index] {
			return
		}
		c := frame.RGBAAt(x, y)
		if c.A != 0 && !isTransparentKey(c) && !isNearWhite(c) {
			return
		}
		seen[index] = true
		queue = append(queue, image.Pt(x, y))
	}
	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		enqueue(x, bounds.Min.Y)
		enqueue(x, bounds.Max.Y-1)
	}
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		enqueue(bounds.Min.X, y)
		enqueue(bounds.Max.X-1, y)
	}
	for len(queue) > 0 {
		p := queue[0]
		queue = queue[1:]
		c := frame.RGBAAt(p.X, p.Y)
		c.A = 0
		frame.SetRGBA(p.X, p.Y, c)
		enqueue(p.X+1, p.Y)
		enqueue(p.X-1, p.Y)
		enqueue(p.X, p.Y+1)
		enqueue(p.X, p.Y-1)
	}
}

func isNearWhite(c color.RGBA) bool {
	return c.R > 245 && c.G > 245 && c.B > 245
}

func removeSolidEdgeColumns(frame *image.RGBA) {
	bounds := frame.Bounds()
	for _, x := range []int{bounds.Min.X, bounds.Min.X + 1, bounds.Max.X - 2, bounds.Max.X - 1} {
		if x < bounds.Min.X || x >= bounds.Max.X {
			continue
		}
		visible := 0
		colors := map[color.RGBA]struct{}{}
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			c := frame.RGBAAt(x, y)
			if c.A == 0 {
				continue
			}
			visible++
			colors[c] = struct{}{}
		}
		if visible > 0 && visible <= 2 {
			for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
				c := frame.RGBAAt(x, y)
				c.A = 0
				frame.SetRGBA(x, y, c)
			}
			continue
		}
		if visible < (bounds.Dy()*3)/4 || len(colors) > 3 {
			continue
		}
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			c := frame.RGBAAt(x, y)
			c.A = 0
			frame.SetRGBA(x, y, c)
		}
	}
}

func sameRGB(a, b color.RGBA) bool {
	return a.R == b.R && a.G == b.G && a.B == b.B
}

func isTransparentKey(c color.RGBA) bool {
	if c.R == 0 && c.G == 0 && c.B == 0 {
		return true
	}
	if c.R > 240 && c.G < 20 && c.B > 240 {
		return true
	}
	if c.R > 80 && c.G < 24 && c.B > 180 {
		return true
	}
	if c.R < 20 && c.G < 20 && c.B > 240 {
		return true
	}
	if c.G < 48 && c.B > 160 && c.R < 220 {
		return true
	}
	if c.G > 240 && c.R < 20 && c.B < 20 {
		return true
	}
	return false
}

func fail(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
