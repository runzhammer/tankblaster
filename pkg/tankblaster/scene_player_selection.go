package tankblaster

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"log"
	"strconv"

	"github.com/golang/freetype/truetype"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/runzhammer/gamedemo/pkg/core"
	"github.com/runzhammer/gamedemo/pkg/engine/tinge"
	"github.com/runzhammer/gamedemo/pkg/tankblaster/computerplayers"
	r "github.com/runzhammer/gamedemo/resources"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
)

const (
	maxPlayerSlots = 10

	slotW = 150
	slotH = 232
)

var defaultPlayerColors = []color.RGBA{
	{R: 11, G: 31, B: 255, A: 255},
	{R: 230, G: 34, B: 45, A: 255},
	{R: 20, G: 150, B: 62, A: 255},
	{R: 255, G: 132, B: 0, A: 255},
	{R: 135, G: 55, B: 210, A: 255},
	{R: 240, G: 220, B: 20, A: 255},
	{R: 0, G: 170, B: 180, A: 255},
	{R: 185, G: 80, B: 25, A: 255},
	{R: 235, G: 85, B: 170, A: 255},
	{R: 40, G: 40, B: 40, A: 255},
}

var uiTextFace font.Face = loadUITextFace()

func loadUITextFace() font.Face {
	f, err := truetype.Parse(r.DejaVuSansMono)
	if err != nil {
		return basicfont.Face7x13
	}
	return truetype.NewFace(f, &truetype.Options{
		Size:              15,
		GlyphCacheEntries: 128,
	})
}

var paletteColors = []color.RGBA{
	{R: 11, G: 31, B: 255, A: 255},
	{R: 230, G: 34, B: 45, A: 255},
	{R: 20, G: 150, B: 62, A: 255},
	{R: 255, G: 132, B: 0, A: 255},
	{R: 135, G: 55, B: 210, A: 255},
	{R: 240, G: 220, B: 20, A: 255},
	{R: 0, G: 170, B: 180, A: 255},
	{R: 235, G: 85, B: 170, A: 255},
}

type playerSelectionSlot struct {
	Kind       PlayerKind
	ComputerID computerplayers.ID
	Name       string
	Color      color.RGBA
}

type playerSelectionScene struct {
	g *GameLoop

	rounds         int
	slots          [maxPlayerSlots]playerSelectionSlot
	focusedName    int
	openPaletteFor int
	inputRunes     []rune
	message        string

	baseImage         *ebiten.Image
	humanPortrait     *ebiten.Image
	computerPortraits map[computerplayers.ID]*ebiten.Image
}

func NewPlayerSelectionScene(game *GameLoop) (core.Scene, error) {
	s := &playerSelectionScene{
		g:              game,
		rounds:         game.rounds,
		focusedName:    -1,
		openPaletteFor: -1,
		baseImage:      mustImageFromPNG(r.PlayerSelectionBase),
		humanPortrait:  mustImageFromPNG(r.PlayerHuman),
		computerPortraits: map[computerplayers.ID]*ebiten.Image{
			computerplayers.DoedelID:   mustImageFromPNG(r.PlayerComputerDoedel),
			computerplayers.FrederikID: mustImageFromPNG(r.PlayerComputerFrederik),
			computerplayers.MisterXID:  mustImageFromPNG(r.PlayerComputerMisterX),
			computerplayers.DrNukeID:   mustImageFromPNG(r.PlayerComputerDrNuke),
			computerplayers.HaraldID:   mustImageFromPNG(r.PlayerComputerHarald),
		},
	}
	if s.rounds <= 0 {
		s.rounds = 10
	}
	for i := range s.slots {
		s.slots[i].Color = defaultPlayerColors[i%len(defaultPlayerColors)]
	}
	return s, nil
}

func (s *playerSelectionScene) Update() error {
	s.handleKeyboard()

	if !inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		return nil
	}

	x, y := ebiten.CursorPosition()
	if s.handleRoundsClick(x, y) {
		return nil
	}
	if s.handleStartClick(x, y) {
		return s.startGame()
	}
	if s.handlePaletteClick(x, y) {
		return nil
	}
	if s.handleSlotClick(x, y) {
		return nil
	}

	s.focusedName = -1
	s.openPaletteFor = -1
	return nil
}

func (s *playerSelectionScene) Draw(screen *ebiten.Image) {
	screen.DrawImage(s.baseImage, nil)

	s.drawRounds(screen)
	for i := range s.slots {
		s.drawSlot(screen, i)
	}
	s.drawStartState(screen)
	if s.openPaletteFor >= 0 {
		s.drawPalette(screen, s.openPaletteFor)
	}
}

func (s *playerSelectionScene) handleKeyboard() {
	if s.focusedName < 0 || s.focusedName >= len(s.slots) {
		return
	}
	slot := &s.slots[s.focusedName]
	if slot.Kind == PlayerNone {
		return
	}

	s.inputRunes = ebiten.AppendInputChars(s.inputRunes[:0])
	if len(s.inputRunes) > 0 {
		name := []rune(slot.Name)
		for _, r := range s.inputRunes {
			if len(name) < 16 {
				name = append(name, r)
			}
		}
		slot.Name = string(name)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		name := []rune(slot.Name)
		if len(name) > 0 {
			slot.Name = string(name[:len(name)-1])
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeyKPEnter) {
		s.focusedName = -1
	}
}

func (s *playerSelectionScene) handleRoundsClick(x, y int) bool {
	minus := image.Rect(434, 72, 466, 103)
	plus := image.Rect(474, 72, 506, 103)
	p := image.Pt(x, y)
	switch {
	case p.In(minus):
		if s.rounds > 1 {
			s.rounds--
		}
		return true
	case p.In(plus):
		if s.rounds < 99 {
			s.rounds++
		}
		return true
	default:
		return false
	}
}

func (s *playerSelectionScene) handleStartClick(x, y int) bool {
	return image.Pt(x, y).In(image.Rect(780, 673, 922, 707))
}

func (s *playerSelectionScene) handlePaletteClick(x, y int) bool {
	if s.openPaletteFor < 0 {
		return false
	}

	x0, y0 := palettePosForSlot(s.openPaletteFor)
	for i, c := range paletteColors {
		rect := image.Rect(x0+i*26, y0, x0+i*26+22, y0+22)
		if image.Pt(x, y).In(rect) {
			s.slots[s.openPaletteFor].Color = c
			s.openPaletteFor = -1
			return true
		}
	}
	return false
}

func (s *playerSelectionScene) handleSlotClick(x, y int) bool {
	p := image.Pt(x, y)
	for i := range s.slots {
		r := slotRect(i)
		titleRect := image.Rect(r.Min.X, r.Min.Y, r.Max.X, r.Min.Y+38)
		nameRect := nameInputRectForSlot(i, s.slots[i].Kind)
		swatchRect := image.Rect(r.Max.X-34, r.Min.Y+42, r.Max.X-12, r.Min.Y+66)
		portraitRect := portraitRectForSlot(i, s.slots[i].Kind)

		switch {
		case p.In(titleRect):
			s.cycleSlotKind(i)
			s.focusedName = -1
			s.openPaletteFor = -1
			return true
		case s.slots[i].Kind != PlayerNone && p.In(nameRect):
			s.focusedName = i
			s.openPaletteFor = -1
			return true
		case s.slots[i].Kind == PlayerComputer && p.In(portraitRect):
			s.cycleComputerPlayer(i)
			s.focusedName = -1
			s.openPaletteFor = -1
			return true
		case s.slots[i].Kind != PlayerNone && p.In(swatchRect):
			s.openPaletteFor = i
			s.focusedName = -1
			return true
		}
	}
	return false
}

func (s *playerSelectionScene) cycleSlotKind(index int) {
	s.message = ""
	slot := &s.slots[index]
	switch slot.Kind {
	case PlayerNone:
		slot.Kind = PlayerHuman
		slot.Name = "Spieler " + strconv.Itoa(s.nextHumanNumber())
	case PlayerHuman:
		slot.Kind = PlayerComputer
		slot.ComputerID = computerplayers.DoedelID
		slot.Name = computerplayers.Name(slot.ComputerID)
	case PlayerComputer:
		slot.Kind = PlayerNone
		slot.ComputerID = computerplayers.DoedelID
		slot.Name = ""
	}
}

func (s *playerSelectionScene) cycleComputerPlayer(index int) {
	slot := &s.slots[index]
	slot.ComputerID = computerplayers.NextID(slot.ComputerID)
	slot.Name = computerplayers.Name(slot.ComputerID)
}

func (s *playerSelectionScene) nextHumanNumber() int {
	count := 1
	for i := range s.slots {
		if s.slots[i].Kind == PlayerHuman {
			count++
		}
	}
	return count
}

func (s *playerSelectionScene) startGame() error {
	players := make([]PlayerConfig, 0, maxPlayerSlots)
	for _, slot := range s.slots {
		if slot.Kind == PlayerNone {
			continue
		}
		name := slot.Name
		if name == "" {
			name = "Spieler " + strconv.Itoa(len(players)+1)
			if slot.Kind == PlayerComputer {
				name = computerplayers.Name(slot.ComputerID)
			}
		}
		players = append(players, PlayerConfig{
			Kind:       slot.Kind,
			ComputerID: slot.ComputerID,
			Name:       name,
			Color:      slot.Color,
		})
	}
	if len(players) < 2 {
		s.message = "Mindestens zwei Spieler auswaehlen"
		return nil
	}

	s.g.rounds = s.rounds
	s.g.players = players
	return s.g.SetNewScene(NewGameScene)
}

func (s *playerSelectionScene) drawRounds(screen *ebiten.Image) {
	drawFilledRect(screen, image.Rect(205, 76, 432, 99), colornames.White)
	drawText(screen, "Anzahl Runden: "+strconv.Itoa(s.rounds), 218, 93, colornames.Black)
}

func (s *playerSelectionScene) drawStartState(screen *ebiten.Image) {
	if s.selectedPlayerCount() < 2 {
		drawFilledRect(screen, image.Rect(784, 677, 918, 702), color.RGBA{R: 180, G: 180, B: 180, A: 180})
	}
	if s.message != "" {
		drawFilledRect(screen, image.Rect(350, 650, 650, 671), colornames.Yellow)
		drawCenteredText(screen, s.message, image.Rect(350, 650, 650, 671), colornames.Black)
	}
}

func (s *playerSelectionScene) selectedPlayerCount() int {
	count := 0
	for _, slot := range s.slots {
		if slot.Kind != PlayerNone {
			count++
		}
	}
	return count
}

func (s *playerSelectionScene) drawHeader(screen *ebiten.Image) {
	drawText(screen, "TANK BLASTER", 106, 62, color.RGBA{R: 0, G: 170, B: 180, A: 255})
	drawText(screen, "TANK BLASTER", 109, 65, colornames.Black)
	drawText(screen, "TANK BLASTER", 106, 59, colornames.White)

	drawFrame(screen, image.Rect(105, 70, 513, 106), colornames.White, colornames.Black)
	drawText(screen, "Anzahl Runden: "+strconv.Itoa(s.rounds), 218, 93, colornames.Black)
	drawButton(screen, image.Rect(434, 73, 466, 102), "-")
	drawButton(screen, image.Rect(474, 73, 506, 102), "+")

	drawFrame(screen, image.Rect(575, 14, 877, 132), color.RGBA{R: 230, G: 236, B: 235, A: 255}, colornames.White)
	drawDesertPreview(screen, image.Rect(579, 18, 873, 128))
}

func (s *playerSelectionScene) drawSlot(screen *ebiten.Image, index int) {
	slot := s.slots[index]
	r := slotRect(index)
	titleRect := image.Rect(r.Min.X, r.Min.Y, r.Max.X, r.Min.Y+38)
	bodyRect := image.Rect(r.Min.X, r.Min.Y+38, r.Max.X, r.Max.Y)

	bodyColor := color.RGBA{R: 190, G: 190, B: 190, A: 255}
	titleColor := color.RGBA{R: 224, G: 224, B: 224, A: 255}
	titleText := "Keiner"
	textColor := colornames.Black
	if slot.Kind == PlayerHuman {
		bodyColor = colornames.White
		titleColor = color.RGBA{R: 224, G: 224, B: 255, A: 255}
		titleText = "Mensch"
		textColor = color.RGBA{R: 40, G: 55, B: 255, A: 255}
	} else if slot.Kind == PlayerComputer {
		bodyColor = colornames.White
		titleColor = color.RGBA{R: 255, G: 224, B: 224, A: 255}
		titleText = "Computer"
		textColor = color.RGBA{R: 255, G: 45, B: 45, A: 255}
	}

	drawFrame(screen, bodyRect, bodyColor, colornames.Black)
	drawFrame(screen, titleRect, titleColor, colornames.Black)
	drawCenteredText(screen, titleText, titleRect, textColor)

	if slot.Kind == PlayerNone {
		return
	}

	portrait := s.humanPortrait
	if slot.Kind == PlayerComputer {
		portrait = s.computerPortraitFor(slot.ComputerID)
	}
	portraitRect := portraitRectForSlot(index, slot.Kind)
	drawScaledImage(screen, portrait, portraitRect)

	swatch := image.Rect(r.Max.X-34, r.Min.Y+42, r.Max.X-12, r.Min.Y+66)
	drawPaintSwatch(screen, swatch, slot.Color)

	nameRect := nameInputRectForSlot(index, slot.Kind)
	drawText(screen, slot.Name, nameRect.Min.X+6, nameRect.Min.Y+20, colornames.Black)
	if s.focusedName == index {
		cursorX := nameRect.Min.X + 8 + text.BoundString(uiTextFace, slot.Name).Dx()
		if cursorX > nameRect.Max.X-5 {
			cursorX = nameRect.Max.X - 5
		}
		drawFilledRect(screen, image.Rect(cursorX, nameRect.Min.Y+5, cursorX+2, nameRect.Max.Y-5), colornames.Black)
	}
}

func (s *playerSelectionScene) computerPortraitFor(id computerplayers.ID) *ebiten.Image {
	if portrait := s.computerPortraits[id]; portrait != nil {
		return portrait
	}
	return s.computerPortraits[computerplayers.DoedelID]
}

func portraitRectForSlot(index int, kind PlayerKind) image.Rectangle {
	r := slotRect(index)
	if kind == PlayerHuman || kind == PlayerComputer {
		return image.Rect(r.Min.X+4, r.Min.Y+40, r.Max.X-4, r.Max.Y-4)
	}
	return image.Rectangle{}
}

func nameInputRectForSlot(index int, kind PlayerKind) image.Rectangle {
	portraitRect := portraitRectForSlot(index, kind)
	if portraitRect.Empty() {
		return image.Rectangle{}
	}
	height := portraitRect.Dy()
	return image.Rect(
		portraitRect.Min.X+4,
		portraitRect.Min.Y+int(float64(height)*0.83),
		portraitRect.Max.X-4,
		portraitRect.Min.Y+int(float64(height)*0.965),
	)
}

func (s *playerSelectionScene) drawFooter(screen *ebiten.Image) {
	drawFilledRect(screen, image.Rect(372, 676, 586, 695), colornames.Yellow)
	drawCenteredText(screen, "Druecken Sie F1 fuer Hilfe", image.Rect(372, 676, 586, 695), colornames.Black)
	drawButton(screen, image.Rect(624, 673, 756, 706), "Optionen")
	drawButton(screen, image.Rect(780, 673, 922, 706), "Start >>")
}

func (s *playerSelectionScene) drawPalette(screen *ebiten.Image, slotIndex int) {
	x0, y0 := palettePosForSlot(slotIndex)
	frame := image.Rect(x0-6, y0-6, x0+len(paletteColors)*26+4, y0+28)
	drawFrame(screen, frame, colornames.White, colornames.Black)
	for i, c := range paletteColors {
		drawFrame(screen, image.Rect(x0+i*26, y0, x0+i*26+22, y0+22), c, colornames.Black)
	}
}

func slotRect(index int) image.Rectangle {
	col := index % 5
	row := index / 5
	return image.Rect(35+col*185, 150+row*269, 35+col*185+slotW, 150+row*269+slotH)
}

func palettePosForSlot(index int) (int, int) {
	r := slotRect(index)
	x := r.Min.X
	y := r.Min.Y + 70
	if x+len(paletteColors)*26 > int(core.Config().Screen.Width)-8 {
		x = int(core.Config().Screen.Width) - len(paletteColors)*26 - 12
	}
	return x, y
}

func drawButton(screen *ebiten.Image, r image.Rectangle, label string) {
	drawFrame(screen, r, color.RGBA{R: 210, G: 210, B: 210, A: 255}, colornames.Black)
	drawFilledRect(screen, image.Rect(r.Min.X+2, r.Min.Y+2, r.Max.X-2, r.Min.Y+5), colornames.White)
	drawFilledRect(screen, image.Rect(r.Min.X+2, r.Max.Y-5, r.Max.X-2, r.Max.Y-2), color.RGBA{R: 115, G: 115, B: 115, A: 255})
	drawCenteredText(screen, label, r, colornames.Black)
}

func drawFrame(screen *ebiten.Image, r image.Rectangle, fill, border color.Color) {
	drawFilledRect(screen, r, border)
	drawFilledRect(screen, image.Rect(r.Min.X+2, r.Min.Y+2, r.Max.X-2, r.Max.Y-2), fill)
}

func drawFilledRect(screen *ebiten.Image, r image.Rectangle, c color.Color) {
	if r.Empty() {
		return
	}
	ebitenutil.DrawRect(screen, float64(r.Min.X), float64(r.Min.Y), float64(r.Dx()), float64(r.Dy()), c)
}

func drawText(screen *ebiten.Image, value string, x, y int, c color.Color) {
	text.Draw(screen, value, uiTextFace, x, y, c)
}

func drawCenteredText(screen *ebiten.Image, value string, r image.Rectangle, c color.Color) {
	b := text.BoundString(uiTextFace, value)
	x := r.Min.X + (r.Dx()-b.Dx())/2
	y := r.Min.Y + (r.Dy()+b.Dy())/2
	drawText(screen, value, x, y, c)
}

func drawPaintSwatch(screen *ebiten.Image, r image.Rectangle, c color.Color) {
	drawFilledRect(screen, image.Rect(r.Min.X+7, r.Min.Y, r.Min.X+15, r.Max.Y), c)
	drawFilledRect(screen, image.Rect(r.Min.X, r.Min.Y+8, r.Max.X, r.Min.Y+17), c)
	drawFilledRect(screen, image.Rect(r.Min.X+3, r.Min.Y+4, r.Max.X-3, r.Max.Y-4), c)
}

func mustImageFromPNG(data []byte) *ebiten.Image {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		log.Fatal(err)
	}
	return ebiten.NewImageFromImage(img)
}

func drawDesertPreview(screen *ebiten.Image, r image.Rectangle) {
	img := image.NewRGBA(image.Rect(0, 0, r.Dx(), r.Dy()))
	for y := 0; y < r.Dy(); y++ {
		for x := 0; x < r.Dx(); x++ {
			if y < r.Dy()/2 {
				img.SetRGBA(x, y, color.RGBA{R: 180, G: 220, B: 230, A: 255})
				continue
			}
			depth := float64(y-r.Dy()/2) / float64(r.Dy()/2)
			img.SetRGBA(x, y, color.RGBA{R: uint8(245 - depth*55), G: uint8(230 - depth*110), B: uint8(165 - depth*90), A: 255})
		}
	}
	tinge.DrawCircle(img, 82, 22, 26, color.RGBA{R: 255, G: 250, B: 220, A: 255})
	tinge.DrawLine(img, 190, 64, 263, 84, color.RGBA{R: 40, G: 30, B: 24, A: 255})
	tinge.DrawLine(img, 220, 55, 220, 35, color.RGBA{R: 40, G: 30, B: 24, A: 255})
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(r.Min.X), float64(r.Min.Y))
	screen.DrawImage(ebiten.NewImageFromImage(img), op)
}

func generateHumanPortrait() image.Image {
	img := image.NewRGBA(image.Rect(0, 0, 110, 150))
	draw.Draw(img, img.Bounds(), image.NewUniform(colornames.White), image.Point{}, draw.Src)
	face := color.RGBA{R: 230, G: 230, B: 230, A: 255}
	ink := colornames.Black
	tinge.DrawCircle(img, 52, 74, 32, ink)
	drawFilledImageRect(img, image.Rect(28, 44, 77, 104), face)
	for x := 18; x < 78; x += 6 {
		tinge.DrawLine(img, x, 35+(x%18), x+10, 52, ink)
	}
	tinge.DrawCircle(img, 40, 72, 5, ink)
	tinge.DrawCircle(img, 62, 72, 5, ink)
	tinge.DrawLine(img, 51, 77, 45, 90, ink)
	tinge.DrawLine(img, 43, 100, 62, 100, ink)
	return img
}

func generateComputerPortrait() image.Image {
	img := image.NewRGBA(image.Rect(0, 0, 110, 150))
	draw.Draw(img, img.Bounds(), image.NewUniform(colornames.White), image.Point{}, draw.Src)
	face := color.RGBA{R: 220, G: 220, B: 220, A: 255}
	ink := colornames.Black
	drawFilledImageRect(img, image.Rect(20, 86, 90, 150), ink)
	drawFilledImageRect(img, image.Rect(35, 42, 84, 108), face)
	tinge.DrawCircle(img, 58, 72, 32, ink)
	drawFilledImageRect(img, image.Rect(34, 63, 81, 77), ink)
	drawFilledImageRect(img, image.Rect(37, 66, 55, 76), colornames.White)
	drawFilledImageRect(img, image.Rect(59, 66, 78, 76), colornames.White)
	tinge.DrawLine(img, 45, 98, 67, 98, ink)
	tinge.DrawLine(img, 37, 43, 60, 35, ink)
	tinge.DrawLine(img, 60, 35, 83, 47, ink)
	return img
}

func drawFilledImageRect(img draw.Image, r image.Rectangle, c color.Color) {
	draw.Draw(img, r, image.NewUniform(c), image.Point{}, draw.Src)
}
