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
)

type introScene struct {
	g          *GameLoop
	tick       int
	background *ebiten.Image
	texts      []*ebiten.Image
	started    bool
	blastSound bool
	overlay    *ebiten.Image
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
	s.drawText(screen)
	s.drawColorWash(screen)
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
		s.drawIntroOverlay(screen, color.RGBA{R: 255, A: uint8(255 * colorFrame / introColorStageFrames)})
	case colorFrame < introColorStageFrames*2:
		local := colorFrame - introColorStageFrames
		s.drawIntroOverlay(screen, color.RGBA{R: 255, G: uint8(255 * local / introColorStageFrames), A: 255})
	case colorFrame < introColorStageFrames*3:
		local := colorFrame - introColorStageFrames*2
		value := uint8(255 - 255*local/introColorStageFrames)
		s.drawIntroOverlay(screen, color.RGBA{R: value, G: value, A: 255})
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
