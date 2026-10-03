package tankblaster

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/runzhammer/gamedemo/pkg/core"
	r "github.com/runzhammer/gamedemo/resources"
	"golang.org/x/image/bmp"
)

const (
	introMusicLoopKey = "intro_music"
	introMusicVolume  = 0.65

	introInitialDelayFrames  = 8 * 60
	introRevealFrames        = 1 * 60
	introTextHoldFrames      = 4 * 60
	introTextGapFrames       = 2 * 60
	introAfterLastTextFrames = 5 * 60
	introColorStageFrames    = 2 * 60
	introBlackHoldFrames     = 30
	introTextScale           = 1.8
	introTextYOffset         = 100
	introRemageMarginX       = 24
	introRemageMarginY       = 70
	introRemageAngle         = 40 * math.Pi / 180
)

type introScene struct {
	g          *GameLoop
	tick       int
	background *ebiten.Image
	texts      []*ebiten.Image
	started    bool
	blastSound bool
	overlay    *ebiten.Image
	remage     *ebiten.Image
}

func NewIntroScene(game *GameLoop) (core.Scene, error) {
	s := &introScene{
		g:          game,
		background: mustIntroImage("intro/Intro_II_v2.bmp"),
		texts: []*ebiten.Image{
			mustIntroImage("intro/Intro_Text_1.png"),
			mustIntroImage("intro/Intro_Text_2.png"),
			mustIntroImage("intro/Intro_Text_3.png"),
			mustIntroImage("intro/Intro_Text_4.png"),
			mustIntroImage("intro/Intro_Text_5.png"),
			mustIntroImage("intro/Intro_Text_6.png"),
		},
		remage: renderIntroRemageLogo(),
	}
	return s, nil
}

func (s *introScene) Update() error {
	if !s.started {
		s.started = true
		s.g.playSoundLoopWithOptions(introMusicLoopKey, "intro/TECHNOMN.MOD", soundOptions{
			Loop:               true,
			Volume:             introMusicVolume,
			KeepSilenceForLoop: true,
		})
	}
	if primaryPointerJustPressed() {
		return s.finish()
	}
	s.tick++
	if !s.blastSound && s.tick >= s.colorStartFrame() {
		s.blastSound = true
		s.g.playSoundWithOptions("intro/2+3_11kHz.wav", soundOptions{Volume: 0.9})
	}
	s.updateMusicFade()
	if s.tick >= s.endFrame() {
		return s.finish()
	}
	return nil
}

func (s *introScene) finish() error {
	if s.g == nil {
		return nil
	}
	s.g.stopSoundLoop(introMusicLoopKey)
	return s.g.SetNewScene(NewPlayerSelectionScene)
}

func (s *introScene) updateMusicFade() {
	if s.g == nil {
		return
	}
	fadeFrame := s.tick - s.colorStartFrame()
	if fadeFrame < 0 {
		return
	}
	fadeFrames := introColorStageFrames * 3
	if fadeFrame >= fadeFrames {
		s.g.stopSoundLoop(introMusicLoopKey)
		return
	}
	progress := float64(fadeFrame) / float64(fadeFrames)
	s.g.setSoundLoopVolume(introMusicLoopKey, introMusicVolume*(1-progress))
}

func (s *introScene) Draw(screen *ebiten.Image) {
	drawScaledImage(screen, s.background, image.Rect(0, 0, ScreenWidth, ScreenHeight))
	s.drawRemage(screen)
	s.drawText(screen)
	s.drawColorWash(screen)
}

func (s *introScene) drawRemage(screen *ebiten.Image) {
	if s.remage == nil {
		return
	}
	bounds := s.remage.Bounds()
	w := float64(bounds.Dx())
	h := float64(bounds.Dy())
	t := float64(s.tick)
	scale := 1.0 + math.Sin(t*0.12)*0.07
	x := float64(ScreenWidth-int(w*scale)-introRemageMarginX) + math.Sin(t*0.047)*7
	y := float64(introRemageMarginY) + math.Sin(t*0.073+0.8)*5

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-w/2, -h/2)
	op.GeoM.Scale(scale, scale)
	op.GeoM.Rotate(introRemageAngle + math.Sin(t*0.036)*0.035)
	op.GeoM.Translate(x+w*scale/2, y+h*scale/2)
	screen.DrawImage(s.remage, op)
}

func (s *introScene) drawText(screen *ebiten.Image) {
	index, local, ok := s.textState()
	if !ok || index < 0 || index >= len(s.texts) {
		return
	}
	img := s.texts[index]
	bounds := img.Bounds()
	visibleW := bounds.Dx()
	if local < introRevealFrames {
		progress := float64(local) / float64(introRevealFrames)
		visibleW = int(math.Ceil(float64(bounds.Dx()) * progress))
	}
	if visibleW <= 0 {
		return
	}
	src := image.Rect(bounds.Min.X, bounds.Min.Y, bounds.Min.X+visibleW, bounds.Max.Y)
	sub, ok := img.SubImage(src).(*ebiten.Image)
	if !ok {
		return
	}
	scaledW := int(math.Round(float64(bounds.Dx()) * introTextScale))
	scaledH := int(math.Round(float64(bounds.Dy()) * introTextScale))
	x := (ScreenWidth - scaledW) / 2
	y := (ScreenHeight-scaledH)/2 + introTextYOffset

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(introTextScale, introTextScale)
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(sub, op)
}

func (s *introScene) drawColorWash(screen *ebiten.Image) {
	colorFrame := s.tick - s.colorStartFrame()
	if colorFrame < 0 {
		return
	}
	switch {
	case colorFrame < introColorStageFrames:
		s.drawIntroOverlay(screen, color.RGBA{R: 255, A: uint8(150 * colorFrame / introColorStageFrames)})
	case colorFrame < introColorStageFrames*2:
		local := colorFrame - introColorStageFrames
		s.drawIntroOverlay(screen, color.RGBA{R: 255, G: uint8(255 * local / introColorStageFrames), A: 150})
	case colorFrame < introColorStageFrames*3:
		local := colorFrame - introColorStageFrames*2
		alpha := uint8(150 + 105*local/introColorStageFrames)
		s.drawIntroOverlay(screen, color.RGBA{A: alpha})
	default:
		s.drawIntroOverlay(screen, color.RGBA{A: 255})
	}
}

func (s *introScene) textState() (int, int, bool) {
	frame := s.tick - introInitialDelayFrames
	if frame < 0 {
		return 0, 0, false
	}
	textCycle := introRevealFrames + introTextHoldFrames + introTextGapFrames
	for i := range s.texts {
		local := frame - i*textCycle
		if local < 0 {
			return 0, 0, false
		}
		if local < introRevealFrames+introTextHoldFrames {
			return i, local, true
		}
	}
	return 0, 0, false
}

func (s *introScene) colorStartFrame() int {
	textCycle := introRevealFrames + introTextHoldFrames + introTextGapFrames
	lastTextStart := introInitialDelayFrames + (len(s.texts)-1)*textCycle
	lastTextEnd := lastTextStart + introRevealFrames + introTextHoldFrames
	return lastTextEnd + introAfterLastTextFrames
}

func (s *introScene) endFrame() int {
	return s.colorStartFrame() + introColorStageFrames*3 + introBlackHoldFrames
}

func mustIntroImage(path string) *ebiten.Image {
	data, err := r.IntroBytes(path)
	if err != nil {
		log.Fatal(err)
	}
	img, err := decodeIntroImage(path, data)
	if err != nil {
		log.Fatal(err)
	}
	return ebiten.NewImageFromImage(img)
}

func decodeIntroImage(path string, data []byte) (image.Image, error) {
	switch filepath.Ext(path) {
	case ".bmp":
		return bmp.Decode(bytes.NewReader(data))
	case ".png":
		return png.Decode(bytes.NewReader(data))
	default:
		img, _, err := image.Decode(bytes.NewReader(data))
		return img, err
	}
}

func renderIntroRemageLogo() *ebiten.Image {
	const label = "remake"
	face := loadUIFont(38)
	bounds := text.BoundString(face, label)
	img := ebiten.NewImage(bounds.Dx()+42, bounds.Dy()+36)
	x := 20 - bounds.Min.X
	y := 15 - bounds.Min.Y

	for depth := 9; depth >= 1; depth-- {
		shade := uint8(80 + depth*7)
		text.Draw(img, label, face, x+depth, y+depth, color.RGBA{R: shade, G: 0, B: 0, A: 255})
	}
	for ox := -3; ox <= 3; ox++ {
		for oy := -3; oy <= 3; oy++ {
			if ox*ox+oy*oy <= 10 {
				text.Draw(img, label, face, x+ox, y+oy, color.RGBA{R: 95, G: 0, B: 0, A: 255})
			}
		}
	}
	text.Draw(img, label, face, x+2, y+2, color.RGBA{R: 120, G: 0, B: 0, A: 255})
	text.Draw(img, label, face, x, y, color.RGBA{R: 235, G: 18, B: 24, A: 255})
	text.Draw(img, label, face, x-2, y-3, color.RGBA{R: 255, G: 170, B: 170, A: 210})
	text.Draw(img, label, face, x-4, y-6, color.RGBA{R: 255, G: 235, B: 235, A: 135})
	return img
}

func (s *introScene) drawIntroOverlay(screen *ebiten.Image, c color.RGBA) {
	if c.A == 0 {
		return
	}
	if s.overlay == nil {
		s.overlay = ebiten.NewImage(ScreenWidth, ScreenHeight)
	}
	s.overlay.Fill(c)
	screen.DrawImage(s.overlay, nil)
}
