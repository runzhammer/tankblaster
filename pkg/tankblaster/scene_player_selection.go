package tankblaster

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/golang/freetype/truetype"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/runzhammer/tankblaster/pkg/buildinfo"
	"github.com/runzhammer/tankblaster/pkg/core"
	"github.com/runzhammer/tankblaster/pkg/engine/tinge"
	"github.com/runzhammer/tankblaster/pkg/gamecore"
	"github.com/runzhammer/tankblaster/pkg/protocol"
	"github.com/runzhammer/tankblaster/pkg/tankblaster/computerplayers"
	"github.com/runzhammer/tankblaster/pkg/tankblaster/soundpaths"
	r "github.com/runzhammer/tankblaster/resources"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
)

const (
	maxPlayerSlots = 10

	slotW = 150
	slotH = 232

	onlineAvailabilityCheckIntervalTicks = 300
	onlineUnavailableMessageTicks        = 180
)

var defaultPlayerColors = []color.RGBA{
	{R: 230, G: 34, B: 45, A: 255},
	{R: 11, G: 31, B: 255, A: 255},
	{R: 20, G: 150, B: 62, A: 255},
	{R: 255, G: 132, B: 0, A: 255},
	{R: 145, G: 235, B: 35, A: 255},
	{R: 240, G: 220, B: 20, A: 255},
	{R: 0, G: 170, B: 180, A: 255},
	{R: 185, G: 80, B: 25, A: 255},
	{R: 235, G: 85, B: 170, A: 255},
	{R: 40, G: 40, B: 40, A: 255},
}

var (
	uiTextFace      font.Face = loadUIFont(15)
	dialogTextFace  font.Face = loadUIFont(11)
	shopHeaderFace  font.Face = loadUIFont(16)
	overlayTextFace font.Face = loadUIFont(24)
	paintSplotch    *ebiten.Image
)

func loadUIFont(size float64) font.Face {
	f, err := truetype.Parse(r.DejaVuSansMono)
	if err != nil {
		return basicfont.Face7x13
	}
	return truetype.NewFace(f, &truetype.Options{
		Size:              size,
		GlyphCacheEntries: 128,
	})
}

var paletteColors = []color.RGBA{
	{R: 11, G: 31, B: 255, A: 255},
	{R: 230, G: 34, B: 45, A: 255},
	{R: 20, G: 150, B: 62, A: 255},
	{R: 255, G: 132, B: 0, A: 255},
	{R: 145, G: 235, B: 35, A: 255},
	{R: 240, G: 220, B: 20, A: 255},
	{R: 0, G: 170, B: 180, A: 255},
	{R: 235, G: 85, B: 170, A: 255},
	{R: 245, G: 245, B: 245, A: 255},
	{R: 40, G: 40, B: 40, A: 255},
}

type playerSelectionSlot struct {
	Kind       PlayerKind
	ComputerID computerplayers.ID
	Name       string
	Color      color.RGBA
	OwnerID    string
	PlayerID   string
}

type playerSelectionScene struct {
	g *GameLoop

	rounds                int
	tick                  int
	slots                 [maxPlayerSlots]playerSelectionSlot
	focusedName           int
	openPaletteFor        int
	inputRunes            []rune
	message               string
	optionsOpen           bool
	helpOpen              bool
	languageOpen          bool
	languageDraft         languageID
	optionsDraft          gameOptions
	transientMessage      string
	transientMessageUntil int

	onlineAvailable     bool
	onlineCheckInFlight bool
	onlineNextCheckTick int
	onlineCheckResults  chan bool
	onlineMode          bool
	onlineClient        *onlineClient
	onlineSessionID     string
	onlineJoinCode      string
	onlineHost          bool
	onlineConnected     bool
	onlineJoinOpen      bool
	onlineJoinInput     string
	onlineStatus        string
	onlineControlledIDs []string
	onlineLobbyRevision int64
	onlineStartState    *gamecore.MatchState
	onlineStartSlots    []protocol.LobbySlot
	onlinePendingLobby  *protocol.LobbyUpdate

	versionUpdateCheckInFlight bool
	versionUpdateResults       chan versionUpdateResult
	versionUpdate              *versionUpdateResult
	versionUpdateDismissed     bool

	baseImage         *ebiten.Image
	canvas            *ebiten.Image
	humanPortrait     *ebiten.Image
	computerPortraits map[computerplayers.ID]*ebiten.Image

	pressedDialogButton string
}

func NewPlayerSelectionScene(game *GameLoop) (core.Scene, error) {
	setPlayerNameInputActive(false)
	baseImage := mustImageFromPNG(r.PlayerSelectionBase)
	game.playSound(tankBlasterSounds.Events[soundEventPlayerSelectionStart])
	s := &playerSelectionScene{
		g:                    game,
		rounds:               game.rounds,
		focusedName:          -1,
		openPaletteFor:       -1,
		baseImage:            baseImage,
		canvas:               ebiten.NewImage(baseImage.Bounds().Dx(), baseImage.Bounds().Dy()),
		onlineCheckResults:   make(chan bool, 1),
		versionUpdateResults: make(chan versionUpdateResult, 1),
		humanPortrait:        mustImageFromPNG(r.PlayerHuman),
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
	s.queueOnlineAvailabilityCheck()
	s.queueVersionUpdateCheck()
	for i := range s.slots {
		s.slots[i].Color = defaultPlayerColors[i%len(defaultPlayerColors)]
	}
	return s, nil
}

func NewEmbeddedOnlinePlayerSelectionScene(game *GameLoop) (core.Scene, error) {
	scene, err := NewPlayerSelectionScene(game)
	if err != nil {
		return nil, err
	}
	s, ok := scene.(*playerSelectionScene)
	if !ok {
		return scene, nil
	}
	s.enterEmbeddedOnlineMode()
	if name := debugOnlineDisplayName(); name != "" {
		s.slots[0].Kind = PlayerHuman
		s.slots[0].Name = name
	}
	return s, nil
}

func (s *playerSelectionScene) Update() error {
	defer s.syncPlayerNameInputActive()
	s.tick++
	s.updateOnlineAvailability()
	s.updateVersionUpdateCheck()
	s.consumeEmbeddedOnline()
	if s.onlineStartState != nil && s.onlineClient != nil {
		state := *s.onlineStartState
		controlled := append([]string(nil), s.onlineControlledIDs...)
		slots := append([]protocol.LobbySlot(nil), s.onlineStartSlots...)
		s.onlineStartState = nil
		return s.g.SetNewScene(func(game *GameLoop) (core.Scene, error) {
			return NewOnlineGameSceneWithControl(game, s.onlineClient, state, controlled, slots, false)
		})
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyF1) {
		s.openHelpDialog()
		return nil
	}
	if s.optionsOpen || s.helpOpen || s.languageOpen {
		s.handleDialogKeyboard()
		if !s.optionsOpen && !s.helpOpen && !s.languageOpen {
			s.pressedDialogButton = ""
			return nil
		}
		if primaryPointerJustPressed() {
			x, y := primaryPointerPosition()
			x, y = s.toSelectionCoords(x, y)
			if button := s.dialogButtonAt(x, y); button != "" {
				s.pressedDialogButton = button
				return nil
			}
			if s.languageOpen {
				s.handleLanguageDialogClick(x, y)
			} else if s.optionsOpen {
				s.handleOptionsDialogClick(x, y)
			} else {
				s.handleHelpDialogClick(x, y)
			}
		}
		if primaryPointerJustReleased() {
			x, y := primaryPointerPosition()
			x, y = s.toSelectionCoords(x, y)
			s.releaseDialogButton(x, y)
		}
		return nil
	}

	if handled, err := s.handleSelectionShortcuts(); handled {
		return err
	}

	s.handleKeyboard()

	if !primaryPointerJustPressed() {
		return nil
	}

	x, y := primaryPointerPosition()
	x, y = s.toSelectionCoords(x, y)
	if s.handleVersionUpdateClick(x, y) {
		return nil
	}
	if s.handleEmbeddedOnlineJoinOverlayClick(x, y) {
		return nil
	}
	if handled := s.handleEmbeddedOnlinePanelClick(x, y); handled {
		return nil
	}
	if s.handleOptionsClick(x, y) {
		return nil
	}
	if s.handleRoundsClick(x, y) {
		return nil
	}
	if s.handleStartClick(x, y) {
		return s.startGame()
	}
	if handled, err := s.handleOnlineClick(x, y); handled {
		return err
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
	target := screen
	if s.canvas != nil {
		s.canvas.Clear()
		target = s.canvas
	}
	target.DrawImage(s.baseImage, nil)

	s.drawRounds(target)
	for i := range s.slots {
		s.drawSlot(target, i)
	}
	s.drawFooter(target)
	s.drawEmbeddedOnlinePanel(target)
	s.drawVersion(target)
	s.drawVersionUpdateNotice(target)
	s.drawStartState(target)
	if s.openPaletteFor >= 0 {
		s.drawPalette(target, s.openPaletteFor)
	}
	if showPlayerNameInputOverlay() {
		s.drawFocusedNameOverlay(target)
	}
	if s.optionsOpen {
		s.drawOptionsDialog(target)
	}
	if s.helpOpen {
		s.drawHelpDialog(target)
	}
	if s.languageOpen {
		s.drawLanguageDialog(target)
	}
	if s.onlineJoinOpen {
		s.drawEmbeddedOnlineJoinOverlay(target)
	}
	if s.canvas != nil {
		drawScaledImage(screen, s.canvas, screen.Bounds())
	}
}

func (s *playerSelectionScene) toSelectionCoords(x, y int) (int, int) {
	if s.baseImage == nil {
		return x, y
	}
	base := s.baseImage.Bounds()
	screenW := int(core.Config().Screen.Width)
	screenH := int(core.Config().Screen.Height)
	if screenW <= 0 || screenH <= 0 {
		return x, y
	}
	return x * base.Dx() / screenW, y * base.Dy() / screenH
}

func (s *playerSelectionScene) handleDialogKeyboard() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		if s.languageOpen {
			s.languageOpen = false
			return
		}
		s.optionsOpen = false
		s.helpOpen = false
	}
	if s.languageOpen && (inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeyKPEnter)) {
		currentLanguage = s.languageDraft
		s.languageOpen = false
		return
	}
	if s.helpOpen && (inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeyKPEnter)) {
		s.helpOpen = false
	}
}

func (s *playerSelectionScene) handleSelectionShortcuts() (bool, error) {
	if s.focusedName >= 0 {
		return false, nil
	}
	switch {
	case inpututil.IsKeyJustPressed(ebiten.KeyEscape) && s.onlineJoinOpen:
		s.onlineJoinOpen = false
		s.onlineJoinInput = ""
		return true, nil
	case inpututil.IsKeyJustPressed(ebiten.KeyEscape) && s.onlineHost && s.onlineJoinCode != "":
		s.cancelEmbeddedOnlineSession()
		return true, nil
	case inpututil.IsKeyJustPressed(ebiten.KeyO):
		s.optionsOpen = true
		s.helpOpen = false
		s.languageOpen = false
		s.optionsDraft = s.g.options
		s.focusedName = -1
		s.openPaletteFor = -1
		return true, nil
	case inpututil.IsKeyJustPressed(ebiten.KeyEscape):
		if s.onlineMode && !s.onlineHost {
			return true, nil
		}
		return true, s.startGame()
	case inpututil.IsKeyJustPressed(ebiten.KeyEqual) || inpututil.IsKeyJustPressed(ebiten.KeyKPAdd):
		if s.canEditLobbySettings() && s.rounds < 99 {
			s.rounds++
			s.playRoundCountChangeSound(1)
			s.saveUserConfig()
			s.syncEmbeddedOnlineLobby()
		}
		return true, nil
	case inpututil.IsKeyJustPressed(ebiten.KeyMinus) || inpututil.IsKeyJustPressed(ebiten.KeyKPSubtract):
		if s.canEditLobbySettings() && s.rounds > 1 {
			s.rounds--
			s.playRoundCountChangeSound(-1)
			s.saveUserConfig()
			s.syncEmbeddedOnlineLobby()
		}
		return true, nil
	default:
		return false, nil
	}
}

func (s *playerSelectionScene) handleKeyboard() {
	if s.onlineJoinOpen {
		s.handleEmbeddedOnlineJoinKeyboard()
		return
	}
	if s.focusedName < 0 || s.focusedName >= len(s.slots) {
		drainPlayerNameInputCommands()
		return
	}
	slot := &s.slots[s.focusedName]
	if slot.Kind == PlayerNone {
		drainPlayerNameInputCommands()
		return
	}

	if s.handleQueuedPlayerNameInput(slot) {
		s.syncEmbeddedOnlineLobby()
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
		s.syncEmbeddedOnlineLobby()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		name := []rune(slot.Name)
		if len(name) > 0 {
			slot.Name = string(name[:len(name)-1])
			s.syncEmbeddedOnlineLobby()
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeyKPEnter) {
		s.focusedName = -1
	}
}

func (s *playerSelectionScene) handleQueuedPlayerNameInput(slot *playerSelectionSlot) bool {
	commands := drainPlayerNameInputCommands()
	if len(commands) == 0 {
		return false
	}
	for _, command := range commands {
		switch {
		case command.finish:
			s.focusedName = -1
		case command.backspace:
			name := []rune(slot.Name)
			if len(name) > 0 {
				slot.Name = string(name[:len(name)-1])
			}
		case command.replace:
			name := []rune(command.text)
			if len(name) > 16 {
				name = name[:16]
			}
			slot.Name = string(name)
		case command.text != "":
			name := []rune(slot.Name)
			for _, r := range []rune(command.text) {
				if len(name) < 16 {
					name = append(name, r)
				}
			}
			slot.Name = string(name)
		}
	}
	return true
}

func (s *playerSelectionScene) syncPlayerNameInputActive() {
	setPlayerNameInputActive(s.focusedName >= 0)
}

func (s *playerSelectionScene) handleRoundsClick(x, y int) bool {
	minus := image.Rect(434, 72, 466, 103)
	plus := image.Rect(474, 72, 506, 103)
	p := image.Pt(x, y)
	switch {
	case p.In(minus):
		if s.canEditLobbySettings() && s.rounds > 1 {
			s.rounds--
			s.playRoundCountChangeSound(-1)
			s.saveUserConfig()
			s.syncEmbeddedOnlineLobby()
		}
		return true
	case p.In(plus):
		if s.canEditLobbySettings() && s.rounds < 99 {
			s.rounds++
			s.playRoundCountChangeSound(1)
			s.saveUserConfig()
			s.syncEmbeddedOnlineLobby()
		}
		return true
	default:
		return false
	}
}

func (s *playerSelectionScene) playRoundCountChangeSound(delta int) {
	if s == nil || s.g == nil {
		return
	}
	if delta > 0 {
		s.g.playSound(soundpaths.SoundUp)
		return
	}
	if delta < 0 {
		s.g.playSound(soundpaths.SoundDown)
	}
}

func (s *playerSelectionScene) handleStartClick(x, y int) bool {
	return image.Pt(x, y).In(image.Rect(780, 673, 922, 707))
}

func (s *playerSelectionScene) handleOnlineClick(x, y int) (bool, error) {
	if !core.Config().Online.Enabled {
		return false, nil
	}
	if !image.Pt(x, y).In(onlineSelectionButtonRect()) {
		return false, nil
	}
	if s.onlineMode {
		s.leaveEmbeddedOnlineMode()
		return true, nil
	}
	if !s.onlineAvailable {
		s.showOnlineUnavailableMessage()
		return true, nil
	}
	s.enterEmbeddedOnlineMode()
	return true, nil
}

func (s *playerSelectionScene) handleOptionsClick(x, y int) bool {
	if !image.Pt(x, y).In(image.Rect(624, 673, 756, 706)) {
		return false
	}
	s.optionsOpen = true
	s.helpOpen = false
	s.languageOpen = false
	s.optionsDraft = s.g.options
	s.focusedName = -1
	s.openPaletteFor = -1
	return true
}

func (s *playerSelectionScene) openHelpDialog() {
	s.helpOpen = true
	s.optionsOpen = false
	s.languageOpen = false
	s.focusedName = -1
	s.openPaletteFor = -1
}

func (s *playerSelectionScene) handleHelpDialogClick(x, y int) {
	p := image.Pt(x, y)
	r := helpDialogRect()
	if p.In(image.Rect(r.Max.X-75, r.Max.Y-63, r.Max.X-12, r.Max.Y-39)) {
		s.helpOpen = false
	}
}

func (s *playerSelectionScene) dialogButtonAt(x, y int) string {
	p := image.Pt(x, y)
	if s.languageOpen {
		r := languageDialogRect()
		if p.In(dialogCloseRect(r)) {
			return "language_close"
		}
		if p.In(image.Rect(r.Min.X+88, r.Min.Y+120, r.Min.X+162, r.Min.Y+141)) {
			return "language_ok"
		}
		return ""
	}
	if s.optionsOpen {
		r := optionsDialogRect()
		switch {
		case p.In(dialogCloseRect(r)):
			return "options_close"
		case p.In(image.Rect(r.Max.X-88, r.Min.Y+56, r.Max.X-16, r.Min.Y+77)):
			return "options_ok"
		case p.In(image.Rect(r.Max.X-88, r.Min.Y+86, r.Max.X-16, r.Min.Y+107)):
			return "options_cancel"
		case p.In(image.Rect(r.Max.X-88, r.Min.Y+166, r.Max.X-16, r.Min.Y+187)):
			return "options_language"
		}
	}
	if s.helpOpen {
		r := helpDialogRect()
		if p.In(dialogCloseRect(r)) {
			return "selection_help_close"
		}
		if p.In(image.Rect(r.Max.X-75, r.Max.Y-63, r.Max.X-12, r.Max.Y-39)) {
			return "selection_help_ok"
		}
	}
	return ""
}

func (s *playerSelectionScene) releaseDialogButton(x, y int) {
	if s.pressedDialogButton == "" {
		return
	}
	button := s.pressedDialogButton
	s.pressedDialogButton = ""
	if s.dialogButtonAt(x, y) != button {
		return
	}
	switch button {
	case "options_ok":
		if s.canEditLobbySettings() {
			s.g.options = s.optionsDraft
			s.saveUserConfig()
			s.syncEmbeddedOnlineLobby()
		}
		s.optionsOpen = false
	case "options_cancel":
		s.optionsOpen = false
	case "options_close":
		s.optionsOpen = false
	case "options_language":
		s.languageDraft = currentLanguage
		s.languageOpen = true
	case "language_ok":
		currentLanguage = s.languageDraft
		s.languageOpen = false
		s.saveUserConfig()
	case "language_close":
		s.languageOpen = false
	case "selection_help_ok":
		s.helpOpen = false
	case "selection_help_close":
		s.helpOpen = false
	}
}

func (s *playerSelectionScene) handleOptionsDialogClick(x, y int) {
	p := image.Pt(x, y)
	r := optionsDialogRect()
	if p.In(image.Rect(r.Max.X-88, r.Min.Y+56, r.Max.X-16, r.Min.Y+77)) {
		if s.canEditLobbySettings() {
			s.g.options = s.optionsDraft
			s.saveUserConfig()
			s.syncEmbeddedOnlineLobby()
		}
		s.optionsOpen = false
		return
	}
	if p.In(image.Rect(r.Max.X-88, r.Min.Y+86, r.Max.X-16, r.Min.Y+107)) {
		s.optionsOpen = false
		return
	}
	if p.In(image.Rect(r.Max.X-88, r.Min.Y+166, r.Max.X-16, r.Min.Y+187)) {
		s.languageDraft = currentLanguage
		s.languageOpen = true
		return
	}
	if !s.canEditLobbySettings() {
		return
	}
	if p.In(image.Rect(r.Min.X+120, r.Min.Y+236, r.Min.X+143, r.Min.Y+258)) {
		s.optionsDraft.cloudAggression = maxInt(0, s.optionsDraft.cloudAggression-5)
		return
	}
	if p.In(image.Rect(r.Min.X+150, r.Min.Y+236, r.Min.X+173, r.Min.Y+258)) {
		s.optionsDraft.cloudAggression = minInt(100, s.optionsDraft.cloudAggression+5)
		return
	}
	if p.In(image.Rect(r.Min.X+31, r.Min.Y+192, r.Min.X+271, r.Min.Y+214)) {
		relative := x - (r.Min.X + 31)
		s.optionsDraft.cloudAggression = maxInt(0, minInt(100, relative*100/240))
		return
	}

	reentryYs := []int{80, 104, 128}
	for i, y0 := range reentryYs {
		if p.In(image.Rect(r.Min.X+30, r.Min.Y+y0-8, r.Min.X+145, r.Min.Y+y0+8)) {
			s.optionsDraft.projectileReentry = i
			return
		}
	}
	palmValues := []int{0, 1, 2, -1}
	palmYs := []int{88, 112, 136, 160}
	for i, y0 := range palmYs {
		if p.In(image.Rect(r.Min.X+312, r.Min.Y+y0-12, r.Min.X+424, r.Min.Y+y0+12)) {
			s.optionsDraft.palmCount = palmValues[i]
			return
		}
	}
	if p.In(image.Rect(r.Min.X+314, r.Min.Y+235, r.Min.X+500, r.Min.Y+249)) {
		s.optionsDraft.quickRoundStart = !s.optionsDraft.quickRoundStart
	}
}

func (s *playerSelectionScene) handleLanguageDialogClick(x, y int) {
	p := image.Pt(x, y)
	r := languageDialogRect()
	switch {
	case p.In(image.Rect(r.Min.X+74, r.Min.Y+61, r.Min.X+145, r.Min.Y+76)):
		s.languageDraft = languageGerman
	case p.In(image.Rect(r.Min.X+74, r.Min.Y+94, r.Min.X+145, r.Min.Y+109)):
		s.languageDraft = languageEnglish
	case p.In(image.Rect(r.Min.X+88, r.Min.Y+120, r.Min.X+162, r.Min.Y+141)):
		currentLanguage = s.languageDraft
		s.languageOpen = false
		s.saveUserConfig()
	}
}

func (s *playerSelectionScene) saveUserConfig() {
	if s == nil || s.g == nil {
		return
	}
	s.g.rounds = s.rounds
	if err := s.g.saveUserConfig(); err != nil {
		log.Printf("save %s: %v", userConfigFileName, err)
	}
}

func (s *playerSelectionScene) handlePaletteClick(x, y int) bool {
	if s.openPaletteFor < 0 {
		return false
	}

	x0, y0 := palettePosForSlot(s.openPaletteFor)
	for i, c := range paletteColors {
		rect := image.Rect(x0+i*26, y0, x0+i*26+22, y0+22)
		if image.Pt(x, y).In(rect) {
			if !s.colorAvailableForSlot(s.openPaletteFor, c) {
				return true
			}
			s.slots[s.openPaletteFor].Color = c
			s.openPaletteFor = -1
			s.syncEmbeddedOnlineLobby()
			return true
		}
	}
	return false
}

func (s *playerSelectionScene) colorAvailableForSlot(slotIndex int, c color.RGBA) bool {
	for i := range s.slots {
		if i == slotIndex || s.slots[i].Kind == PlayerNone {
			continue
		}
		if sameColor(s.slots[i].Color, c) {
			return false
		}
	}
	return true
}

func sameColor(a, b color.RGBA) bool {
	return a.R == b.R && a.G == b.G && a.B == b.B && a.A == b.A
}

func (s *playerSelectionScene) handleSlotClick(x, y int) bool {
	p := image.Pt(x, y)
	for i := range s.slots {
		r := slotRect(i)
		titleRect := image.Rect(r.Min.X, r.Min.Y, r.Max.X, r.Min.Y+38)
		nameRect := nameInputRectForSlot(i, s.slots[i].Kind)
		swatchRect := colorSwatchRectForSlot(i)
		portraitRect := portraitRectForSlot(i, s.slots[i].Kind)

		switch {
		case p.In(titleRect):
			if !s.canEditSlotKind(i) {
				return true
			}
			s.cycleSlotKind(i)
			s.syncEmbeddedOnlineLobby()
			s.focusedName = -1
			s.openPaletteFor = -1
			return true
		case s.slots[i].Kind != PlayerNone && p.In(nameRect):
			if !s.canEditSlotDetails(i) {
				return true
			}
			s.focusedName = i
			s.slots[i].Name = ""
			SetPlayerNameText("")
			s.syncEmbeddedOnlineLobby()
			s.openPaletteFor = -1
			return true
		case s.slots[i].Kind != PlayerNone && p.In(swatchRect):
			if !s.canEditSlotDetails(i) {
				return true
			}
			s.openPaletteFor = i
			s.focusedName = -1
			return true
		case s.slots[i].Kind == PlayerComputer && p.In(portraitRect):
			if !s.canEditSlotKind(i) {
				return true
			}
			s.cycleComputerPlayer(i)
			s.syncEmbeddedOnlineLobby()
			s.focusedName = -1
			s.openPaletteFor = -1
			return true
		}
	}
	return false
}

func (s *playerSelectionScene) canEditSlotKind(index int) bool {
	if !s.onlineMode {
		return true
	}
	if index < 0 || index >= len(s.slots) || s.onlineClient == nil {
		return false
	}
	slot := s.slots[index]
	return slot.Kind == PlayerNone || slot.OwnerID == s.onlineClient.id.PlayerID
}

func (s *playerSelectionScene) canEditSlotDetails(index int) bool {
	if !s.onlineMode {
		return true
	}
	return index >= 0 && index < len(s.slots) && s.slots[index].OwnerID == s.onlineClient.id.PlayerID
}

func (s *playerSelectionScene) cycleSlotKind(index int) {
	s.message = ""
	slot := &s.slots[index]
	if s.onlineMode {
		s.cycleOnlineSlotKind(index)
		return
	}
	switch slot.Kind {
	case PlayerNone:
		name := "Spieler " + strconv.Itoa(s.nextHumanNumber())
		slot.Kind = PlayerHuman
		slot.Name = name
		if s.onlineMode && s.onlineClient != nil {
			slot.OwnerID = s.onlineClient.id.PlayerID
			slot.PlayerID = localLobbyPlayerID(slot.OwnerID, index)
		}
	case PlayerHuman:
		slot.Kind = PlayerComputer
		slot.ComputerID = computerplayers.DoedelID
		slot.Name = computerplayers.Name(slot.ComputerID)
	case PlayerComputer:
		slot.Kind = PlayerNone
		slot.ComputerID = computerplayers.DoedelID
		slot.Name = ""
		slot.OwnerID = ""
		slot.PlayerID = ""
	}
}

func (s *playerSelectionScene) cycleOnlineSlotKind(index int) {
	if index < 0 || index >= len(s.slots) || s.onlineClient == nil {
		return
	}
	slot := &s.slots[index]
	switch slot.Kind {
	case PlayerNone:
		slot.Kind = PlayerHuman
		slot.Name = s.localOnlineDisplayName()
		slot.OwnerID = s.onlineClient.id.PlayerID
		slot.PlayerID = localLobbyPlayerID(slot.OwnerID, index)
	case PlayerHuman:
		if slot.OwnerID != s.onlineClient.id.PlayerID {
			return
		}
		slot.Kind = PlayerComputer
		slot.ComputerID = computerplayers.DoedelID
		slot.Name = computerplayers.Name(slot.ComputerID)
		slot.OwnerID = s.onlineClient.id.PlayerID
		slot.PlayerID = localLobbyPlayerID(slot.OwnerID, index)
	case PlayerComputer:
		if slot.OwnerID != s.onlineClient.id.PlayerID {
			return
		}
		slot.Kind = PlayerNone
		slot.ComputerID = computerplayers.DoedelID
		slot.Name = ""
	}
}

func (s *playerSelectionScene) localOnlineDisplayName() string {
	if s.onlineClient != nil && strings.TrimSpace(s.onlineClient.id.DisplayName) != "" {
		return s.onlineClient.id.DisplayName
	}
	return s.firstHumanName()
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
	if s.onlineMode && !s.onlineHost {
		s.message = "Warte auf Session-Start"
		return nil
	}
	if s.onlineHost && s.onlineJoinCode != "" && !s.ensureEmbeddedOnlineHostPlayer() {
		return nil
	}
	if s.onlineHost && s.onlineSessionID != "" && s.onlineClient != nil {
		s.syncEmbeddedOnlineLobby()
		s.onlineClient.Send(protocol.TypeStartLobbyGame, protocol.StartLobbyGame{SessionID: s.onlineSessionID})
		return nil
	}
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
		s.message = texts().PlayerSelectionMinimumPlayers
		return nil
	}

	s.g.rounds = s.rounds
	s.g.players = players
	s.saveUserConfig()
	return s.g.SetNewScene(NewGameScene)
}

func (s *playerSelectionScene) updateOnlineAvailability() {
	for {
		select {
		case available := <-s.onlineCheckResults:
			s.onlineAvailable = available
			s.onlineCheckInFlight = false
			s.onlineNextCheckTick = s.tick + onlineAvailabilityCheckIntervalTicks
		default:
			if s.transientMessage != "" && s.tick >= s.transientMessageUntil {
				s.transientMessage = ""
			}
			if core.Config().Online.Enabled && !s.onlineCheckInFlight && s.tick >= s.onlineNextCheckTick {
				s.queueOnlineAvailabilityCheck()
			}
			return
		}
	}
}

func (s *playerSelectionScene) updateVersionUpdateCheck() {
	if s.versionUpdateResults == nil {
		return
	}
	for {
		select {
		case result := <-s.versionUpdateResults:
			s.versionUpdateCheckInFlight = false
			if result.available {
				next := result
				s.versionUpdate = &next
			}
		default:
			return
		}
	}
}

func (s *playerSelectionScene) queueVersionUpdateCheck() {
	if s.versionUpdateCheckInFlight || s.versionUpdateResults == nil {
		return
	}
	current := currentSemanticVersionString()
	if _, ok := parseSemanticVersion(current); !ok {
		return
	}
	s.versionUpdateCheckInFlight = true
	go func(results chan<- versionUpdateResult, current string) {
		ctx, cancel := context.WithTimeout(context.Background(), versionCheckTimeout)
		defer cancel()
		result := checkLatestVersion(ctx, versionCheckURL, current)
		select {
		case results <- result:
		default:
		}
	}(s.versionUpdateResults, current)
}

func (s *playerSelectionScene) queueOnlineAvailabilityCheck() {
	if !core.Config().Online.Enabled || s.onlineCheckInFlight || s.onlineCheckResults == nil {
		return
	}
	s.onlineCheckInFlight = true
	go func(results chan<- bool) {
		ctx, cancel := context.WithTimeout(context.Background(), 900*time.Millisecond)
		defer cancel()
		available := onlineServerAvailable(ctx)
		select {
		case results <- available:
		default:
		}
	}(s.onlineCheckResults)
}

func (s *playerSelectionScene) showOnlineUnavailableMessage() {
	s.transientMessage = texts().PlayerSelectionOnlineUnavailable
	s.transientMessageUntil = s.tick + onlineUnavailableMessageTicks
}

func (s *playerSelectionScene) canEditLobbySettings() bool {
	return !s.onlineMode || s.onlineHost
}

func (s *playerSelectionScene) enterEmbeddedOnlineMode() {
	s.onlineMode = true
	s.onlineStatus = ""
	if s.onlineClient == nil {
		s.onlineClient = newOnlineClient(s.firstHumanName())
	}
}

func (s *playerSelectionScene) leaveEmbeddedOnlineMode() {
	if s.onlineClient != nil {
		if s.onlineSessionID != "" {
			s.onlineClient.Send(protocol.TypeLeaveSession, struct{}{})
		}
		s.onlineClient.Close()
	}
	s.onlineClient = nil
	s.onlineMode = false
	s.onlineHost = false
	s.onlineConnected = false
	s.onlineSessionID = ""
	s.onlineJoinCode = ""
	s.onlineJoinOpen = false
	s.onlineJoinInput = ""
	s.onlineStatus = ""
}

func (s *playerSelectionScene) cancelEmbeddedOnlineSession() {
	if s.onlineClient != nil && s.onlineSessionID != "" {
		s.onlineClient.Send(protocol.TypeLeaveSession, struct{}{})
	}
	s.onlineHost = false
	s.onlineConnected = false
	s.onlineSessionID = ""
	s.onlineJoinCode = ""
	s.onlineStatus = ""
}

func (s *playerSelectionScene) consumeEmbeddedOnline() {
	if !s.onlineMode || s.onlineClient == nil {
		return
	}
	for {
		select {
		case env := <-s.onlineClient.recv:
			s.handleEmbeddedOnlineMessage(env)
		case err := <-s.onlineClient.errs:
			if err != nil {
				s.onlineStatus = err.Error()
			}
		default:
			return
		}
	}
}

func (s *playerSelectionScene) handleEmbeddedOnlineMessage(env protocol.Envelope) {
	t := texts()
	switch env.Type {
	case protocol.TypeHelloAck:
		msg, err := protocol.Decode[protocol.HelloAck](env)
		if err == nil {
			s.onlineClient.id.PlayerID = msg.PlayerID
			if msg.PlayerToken != "" {
				s.onlineClient.id.PlayerToken = msg.PlayerToken
			}
			s.onlineClient.id.DisplayName = msg.DisplayName
			saveOnlineIdentity(s.onlineClient.id)
			s.onlineStatus = t.OnlineConnected
		}
	case protocol.TypeSessionCreated:
		msg, err := protocol.Decode[protocol.SessionCreated](env)
		if err == nil {
			s.onlineSessionID = msg.Session.ID
			s.onlineConnected = true
			s.onlineHost = true
			s.onlineStatus = t.OnlineSessionCreated
			s.ensureEmbeddedOnlineHostPlayer()
			s.syncEmbeddedOnlineLobby()
		}
	case protocol.TypeInviteCreated:
		msg, err := protocol.Decode[protocol.InviteCreated](env)
		if err == nil {
			s.onlineJoinCode = msg.JoinCode
		}
	case protocol.TypeSessionJoined:
		msg, err := protocol.Decode[protocol.SessionJoined](env)
		if err == nil {
			s.onlineSessionID = msg.Session.ID
			s.onlineConnected = true
			if !s.onlineHost {
				s.onlineStatus = t.OnlineSessionJoined
			}
		}
	case protocol.TypeLobbyUpdate:
		msg, err := protocol.Decode[protocol.LobbyUpdate](env)
		if err == nil {
			if !s.acceptEmbeddedLobbyUpdate(msg) {
				return
			}
			s.onlineSessionID = msg.SessionID
			s.onlineControlledIDs = append([]string(nil), msg.ControlledPlayerIDs...)
			s.applyEmbeddedLobbyUpdate(msg)
		}
	case protocol.TypeGameStart, protocol.TypeTurnStart:
		msg, err := protocol.Decode[protocol.StateUpdate](env)
		if err == nil {
			s.onlineControlledIDs = append([]string(nil), msg.ControlledPlayerIDs...)
			s.onlineStartSlots = append([]protocol.LobbySlot(nil), msg.LobbySlots...)
			state := msg.State
			s.onlineStartState = &state
		}
	case protocol.TypeSessionClosed:
		s.cancelEmbeddedOnlineSession()
	case protocol.TypeError:
		msg, err := protocol.Decode[protocol.Error](env)
		if err == nil {
			s.onlineStatus = msg.Message
		}
	}
}

func (s *playerSelectionScene) acceptEmbeddedLobbyUpdate(update protocol.LobbyUpdate) bool {
	if update.Revision > 0 && update.Revision <= s.onlineLobbyRevision {
		return false
	}
	if s.onlinePendingLobby == nil {
		return true
	}
	if lobbyUpdateSatisfiesPending(update, *s.onlinePendingLobby, s.onlineHost, s.onlineClientID()) {
		s.onlinePendingLobby = nil
		return true
	}
	return false
}

func (s *playerSelectionScene) applyEmbeddedLobbyUpdate(update protocol.LobbyUpdate) {
	if update.Revision > s.onlineLobbyRevision {
		s.onlineLobbyRevision = update.Revision
	}
	for i := range s.slots {
		s.slots[i] = playerSelectionSlot{Color: defaultPlayerColors[i%len(defaultPlayerColors)]}
	}
	for _, slot := range update.Slots {
		if slot.Index < 0 || slot.Index >= len(s.slots) {
			continue
		}
		dst := &s.slots[slot.Index]
		dst.Name = slot.Name
		dst.Color = color.RGBA{R: slot.Color.R, G: slot.Color.G, B: slot.Color.B, A: slot.Color.A}
		dst.OwnerID = slot.OwnerID
		dst.PlayerID = slot.PlayerID
		dst.ComputerID = computerplayers.ID(slot.ComputerID)
		switch slot.Kind {
		case "human":
			dst.Kind = PlayerHuman
		case "computer":
			dst.Kind = PlayerComputer
		default:
			dst.Kind = PlayerNone
		}
	}
	if update.Rounds > 0 {
		s.rounds = normalizedOnlineRounds(update.Rounds)
	}
	if s.g != nil {
		s.g.options = gameOptionsFromLobbyOptions(update.Options, s.g.options)
		if s.optionsOpen && !s.canEditLobbySettings() {
			s.optionsDraft = s.g.options
		}
	}
}

func (s *playerSelectionScene) syncEmbeddedOnlineLobby() {
	if !s.onlineMode || s.onlineClient == nil || s.onlineSessionID == "" {
		return
	}
	update := protocol.LobbyUpdate{
		SessionID: s.onlineSessionID,
		Revision:  s.onlineLobbyRevision,
		Slots:     s.embeddedLobbySlots(),
		Rounds:    s.rounds,
		Options:   lobbyOptionsFromGameOptions(s.g.options),
	}
	s.onlinePendingLobby = cloneLobbyUpdate(update)
	s.onlineClient.Send(protocol.TypeLobbyUpdate, update)
}

func (s *playerSelectionScene) onlineClientID() string {
	if s.onlineClient == nil {
		return ""
	}
	return s.onlineClient.id.PlayerID
}

func cloneLobbyUpdate(update protocol.LobbyUpdate) *protocol.LobbyUpdate {
	clone := update
	clone.Slots = append([]protocol.LobbySlot(nil), update.Slots...)
	return &clone
}

func lobbyUpdateSatisfiesPending(update, pending protocol.LobbyUpdate, host bool, playerID string) bool {
	if host {
		return update.Rounds == pending.Rounds &&
			update.Options == pending.Options &&
			lobbySlotsContainAll(update.Slots, pending.Slots)
	}
	return lobbySlotsContainAll(update.Slots, lobbyOwnedSlots(pending.Slots, playerID))
}

func lobbyOwnedSlots(slots []protocol.LobbySlot, playerID string) []protocol.LobbySlot {
	out := make([]protocol.LobbySlot, 0, len(slots))
	for _, slot := range slots {
		if slot.OwnerID == playerID {
			out = append(out, slot)
		}
	}
	return out
}

func lobbySlotsContainAll(haystack, needles []protocol.LobbySlot) bool {
	for _, needle := range needles {
		if needle.Kind == "none" {
			if lobbySlotCleared(haystack, needle) {
				continue
			}
			return false
		}
		found := false
		for _, slot := range haystack {
			if lobbySlotMatches(slot, needle) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func lobbySlotCleared(slots []protocol.LobbySlot, needle protocol.LobbySlot) bool {
	for _, slot := range slots {
		if slot.Index == needle.Index && slot.OwnerID == needle.OwnerID {
			return false
		}
	}
	return true
}

func lobbySlotMatches(a, b protocol.LobbySlot) bool {
	return a.Index == b.Index &&
		a.Kind == b.Kind &&
		a.OwnerID == b.OwnerID &&
		a.PlayerID == b.PlayerID &&
		a.Name == b.Name &&
		a.ComputerID == b.ComputerID &&
		a.Color == b.Color
}

func lobbyOptionsFromGameOptions(options gameOptions) protocol.LobbyOptions {
	return protocol.LobbyOptions{
		ProjectileReentry: clampProjectileReentry(options.projectileReentry),
		PalmCount:         clampPalmCount(options.palmCount),
		CloudAggression:   clampCloudAggression(options.cloudAggression),
		QuickRoundStart:   options.quickRoundStart,
	}
}

func gameOptionsFromLobbyOptions(options protocol.LobbyOptions, fallback gameOptions) gameOptions {
	fallback.projectileReentry = clampProjectileReentry(options.ProjectileReentry)
	fallback.palmCount = clampPalmCount(options.PalmCount)
	fallback.cloudAggression = clampCloudAggression(options.CloudAggression)
	fallback.quickRoundStart = options.QuickRoundStart
	return fallback
}

func (s *playerSelectionScene) embeddedLobbySlots() []protocol.LobbySlot {
	slots := make([]protocol.LobbySlot, 0, len(s.slots))
	for i, slot := range s.slots {
		if slot.Kind == PlayerNone {
			if s.onlineClient == nil || slot.OwnerID != s.onlineClient.id.PlayerID {
				continue
			}
			slots = append(slots, protocol.LobbySlot{
				Index:   i,
				Kind:    "none",
				OwnerID: slot.OwnerID,
			})
			continue
		}
		kind := "human"
		ownerID := slot.OwnerID
		playerID := slot.PlayerID
		if slot.Kind == PlayerComputer {
			kind = "computer"
		}
		if ownerID == "" && s.onlineClient != nil {
			ownerID = s.onlineClient.id.PlayerID
			playerID = localLobbyPlayerID(ownerID, i)
		}
		slots = append(slots, protocol.LobbySlot{
			Index:      i,
			Kind:       kind,
			OwnerID:    ownerID,
			PlayerID:   playerID,
			Name:       slot.Name,
			ComputerID: int(slot.ComputerID),
			Color: protocol.RGBA{
				R: slot.Color.R,
				G: slot.Color.G,
				B: slot.Color.B,
				A: slot.Color.A,
			},
		})
	}
	return slots
}

func localLobbyPlayerID(ownerID string, index int) string {
	return ownerID + ":slot:" + strconv.Itoa(index)
}

func (s *playerSelectionScene) handleEmbeddedOnlinePanelClick(x, y int) bool {
	if !s.onlineMode || s.onlineJoinOpen {
		return false
	}
	p := image.Pt(x, y)
	switch {
	case p.In(embeddedOnlineStartRect()) && s.onlineJoinCode == "" && !s.onlineConnected:
		s.startEmbeddedOnlineSession()
		return true
	case p.In(embeddedOnlineJoinRect()) && s.onlineJoinCode == "" && !s.onlineConnected:
		s.onlineJoinOpen = true
		s.onlineJoinInput = ""
		s.focusedName = -1
		s.openPaletteFor = -1
		return true
	case p.In(embeddedOnlineCodeRect()) && s.onlineJoinCode != "":
		if err := copyTextToClipboard(s.onlineJoinCode); err != nil {
			s.onlineStatus = texts().OnlineClipboardUnavailable
		} else {
			s.onlineStatus = texts().OnlineCopiedCode
		}
		return true
	case p.In(embeddedOnlineCancelRect()) && s.onlineJoinCode != "":
		s.cancelEmbeddedOnlineSession()
		return true
	}
	return false
}

func (s *playerSelectionScene) startEmbeddedOnlineSession() {
	if s.onlineClient == nil {
		s.onlineClient = newOnlineClient(s.firstHumanName())
	}
	if !s.ensureEmbeddedOnlineHostPlayer() {
		return
	}
	s.onlineHost = true
	s.onlineStatus = texts().OnlineConnecting
	s.onlineClient.Send(protocol.TypeCreatePrivateSession, protocol.CreateSession{DisplayName: s.firstHumanName(), Rounds: s.rounds})
}

func (s *playerSelectionScene) ensureEmbeddedOnlineHostPlayer() bool {
	for i := range s.slots {
		if s.slots[i].Kind == PlayerHuman {
			if s.slots[i].OwnerID == "" && s.onlineClient != nil {
				s.slots[i].OwnerID = s.onlineClient.id.PlayerID
				s.slots[i].PlayerID = localLobbyPlayerID(s.slots[i].OwnerID, i)
			}
			return true
		}
	}
	for i := range s.slots {
		if s.slots[i].Kind == PlayerNone {
			s.slots[i].Kind = PlayerHuman
			s.slots[i].Name = s.firstHumanName()
			if s.onlineClient != nil {
				s.slots[i].OwnerID = s.onlineClient.id.PlayerID
				s.slots[i].PlayerID = localLobbyPlayerID(s.slots[i].OwnerID, i)
			}
			return true
		}
	}
	s.message = "Keine freien Slots für weitere Spieler"
	return false
}

func (s *playerSelectionScene) assignJoinedOnlineSlot() {
	name := texts().GameDefaultPlayerName
	if s.onlineClient != nil && s.onlineClient.id.DisplayName != "" {
		name = s.onlineClient.id.DisplayName
	}
	for i := range s.slots {
		if s.slots[i].Kind == PlayerNone {
			s.slots[i].Kind = PlayerHuman
			s.slots[i].Name = name
			return
		}
	}
}

func (s *playerSelectionScene) firstHumanName() string {
	for i := range s.slots {
		if s.slots[i].Kind == PlayerHuman && strings.TrimSpace(s.slots[i].Name) != "" {
			return s.slots[i].Name
		}
	}
	if name := debugOnlineDisplayName(); name != "" {
		return name
	}
	id := loadOnlineIdentity()
	if id.DisplayName != "" {
		return id.DisplayName
	}
	return texts().GameDefaultPlayerName
}

func debugOnlineDisplayName() string {
	return truncateRunes(strings.TrimSpace(os.Getenv("TANKBLASTER_ONLINE_DISPLAY_NAME")), 16)
}

func (s *playerSelectionScene) handleEmbeddedOnlineJoinKeyboard() {
	s.inputRunes = ebiten.AppendInputChars(s.inputRunes[:0])
	for _, r := range s.inputRunes {
		if len([]rune(s.onlineJoinInput)) < 10 && isJoinCodeRune(r) {
			s.onlineJoinInput += strings.ToUpper(string(r))
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		runes := []rune(s.onlineJoinInput)
		if len(runes) > 0 {
			s.onlineJoinInput = string(runes[:len(runes)-1])
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeyKPEnter) {
		s.submitEmbeddedOnlineJoin()
	}
}

func (s *playerSelectionScene) handleEmbeddedOnlineJoinOverlayClick(x, y int) bool {
	if !s.onlineJoinOpen {
		return false
	}
	p := image.Pt(x, y)
	if p.In(embeddedJoinPasteRect()) {
		value, err := readClipboardText()
		if err != nil {
			s.onlineStatus = texts().OnlineClipboardUnavailable
			return true
		}
		value = strings.TrimSpace(value)
		if validEmbeddedJoinCode(value) {
			s.onlineJoinInput = strings.ToUpper(value)
			s.submitEmbeddedOnlineJoin()
		}
		return true
	}
	if p.In(embeddedJoinSubmitRect()) {
		s.submitEmbeddedOnlineJoin()
		return true
	}
	if !p.In(embeddedJoinDialogRect()) {
		s.onlineJoinOpen = false
		s.onlineJoinInput = ""
		return true
	}
	return true
}

func (s *playerSelectionScene) submitEmbeddedOnlineJoin() {
	code := strings.ToUpper(strings.TrimSpace(s.onlineJoinInput))
	if !validEmbeddedJoinCode(code) {
		s.onlineStatus = "Ungültiger Join-Code"
		return
	}
	if s.onlineClient == nil {
		s.onlineClient = newOnlineClient(s.firstHumanName())
	}
	s.onlineHost = false
	s.onlineStatus = texts().OnlineConnecting
	s.onlineClient.Send(protocol.TypeJoinSession, protocol.JoinSession{JoinCode: code})
	s.onlineJoinOpen = false
}

func validEmbeddedJoinCode(value string) bool {
	runes := []rune(strings.TrimSpace(value))
	if len(runes) < 8 || len(runes) > 10 {
		return false
	}
	for _, r := range runes {
		if !isJoinCodeRune(r) {
			return false
		}
	}
	return true
}

func isJoinCodeRune(r rune) bool {
	return unicode.IsDigit(r) || (unicode.IsLetter(r) && r <= unicode.MaxASCII)
}

func onlineSelectionButtonRect() image.Rectangle {
	return image.Rect(205, 673, 348, 706)
}

func (s *playerSelectionScene) drawRounds(screen *ebiten.Image) {
	t := texts()
	drawFilledRect(screen, image.Rect(205, 76, 432, 99), colornames.White)
	drawText(screen, t.PlayerSelectionRounds+": "+strconv.Itoa(s.rounds), 218, 93, colornames.Black)
	drawButton(screen, image.Rect(434, 72, 466, 103), "-")
	drawButton(screen, image.Rect(474, 72, 506, 103), "+")
}

func (s *playerSelectionScene) drawStartState(screen *ebiten.Image) {
	if s.selectedPlayerCount() < 2 {
		drawFilledRect(screen, image.Rect(784, 677, 918, 702), color.RGBA{R: 180, G: 180, B: 180, A: 180})
	}
	message := ""
	if s.transientMessage != "" && s.tick < s.transientMessageUntil {
		message = s.transientMessage
	} else {
		message = s.message
	}
	if message != "" {
		drawFilledRect(screen, image.Rect(350, 650, 650, 671), colornames.Yellow)
		drawCenteredText(screen, message, image.Rect(350, 650, 650, 671), colornames.Black)
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
	t := texts()
	drawText(screen, t.PlayerSelectionTitle, 106, 62, color.RGBA{R: 0, G: 170, B: 180, A: 255})
	drawText(screen, t.PlayerSelectionTitle, 109, 65, colornames.Black)
	drawText(screen, t.PlayerSelectionTitle, 106, 59, colornames.White)

	drawFrame(screen, image.Rect(105, 70, 513, 106), colornames.White, colornames.Black)
	drawText(screen, t.PlayerSelectionRounds+": "+strconv.Itoa(s.rounds), 218, 93, colornames.Black)
	drawButton(screen, image.Rect(434, 73, 466, 102), "-")
	drawButton(screen, image.Rect(474, 73, 506, 102), "+")

	drawFrame(screen, image.Rect(575, 14, 877, 132), color.RGBA{R: 230, G: 236, B: 235, A: 255}, colornames.White)
	drawDesertPreview(screen, image.Rect(579, 18, 873, 128))
}

func (s *playerSelectionScene) drawSlot(screen *ebiten.Image, index int) {
	t := texts()
	slot := s.slots[index]
	r := slotRect(index)
	titleRect := image.Rect(r.Min.X, r.Min.Y, r.Max.X, r.Min.Y+38)
	bodyRect := image.Rect(r.Min.X, r.Min.Y+38, r.Max.X, r.Max.Y)

	bodyColor := color.RGBA{R: 190, G: 190, B: 190, A: 255}
	titleColor := color.RGBA{R: 224, G: 224, B: 224, A: 255}
	titleText := t.PlayerSelectionSlotNone
	textColor := colornames.Black
	if slot.Kind == PlayerHuman {
		bodyColor = colornames.White
		titleColor = color.RGBA{R: 224, G: 224, B: 255, A: 255}
		titleText = t.PlayerSelectionSlotHuman
		textColor = color.RGBA{R: 40, G: 55, B: 255, A: 255}
	} else if slot.Kind == PlayerComputer {
		bodyColor = colornames.White
		titleColor = color.RGBA{R: 255, G: 224, B: 224, A: 255}
		titleText = t.PlayerSelectionSlotComputer
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

	drawPaintSwatch(screen, colorSwatchRectForSlot(index), slot.Color)

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

func (s *playerSelectionScene) drawFocusedNameOverlay(screen *ebiten.Image) {
	if s.focusedName < 0 || s.focusedName >= len(s.slots) {
		return
	}
	slot := s.slots[s.focusedName]
	if slot.Kind == PlayerNone {
		return
	}
	r := image.Rect(184, 116, 840, 194)
	drawFrame(screen, r, color.RGBA{R: 255, G: 252, B: 218, A: 255}, colornames.Black)
	drawFrame(screen, insetRect(r, 5), colornames.White, color.RGBA{R: 80, G: 80, B: 80, A: 255})
	drawTextFace(screen, texts().OnlineDisplayName, uiTextFace, r.Min.X+20, r.Min.Y+27, colornames.Black)
	drawTextFace(screen, slot.Name, overlayTextFace, r.Min.X+20, r.Min.Y+64, colornames.Black)
	if s.tick%60 < 30 {
		cursorX := r.Min.X + 22 + text.BoundString(overlayTextFace, slot.Name).Dx()
		drawFilledRect(screen, image.Rect(cursorX, r.Min.Y+40, cursorX+3, r.Min.Y+68), colornames.Black)
	}
}

func portraitRectForSlot(index int, kind PlayerKind) image.Rectangle {
	r := slotRect(index)
	if kind == PlayerHuman || kind == PlayerComputer {
		return image.Rect(r.Min.X+4, r.Min.Y+40, r.Max.X-4, r.Max.Y-4)
	}
	return image.Rectangle{}
}

func colorSwatchRectForSlot(index int) image.Rectangle {
	r := slotRect(index)
	return image.Rect(r.Max.X-34, r.Min.Y+42, r.Max.X-12, r.Min.Y+66)
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

func (s *playerSelectionScene) drawEmbeddedOnlinePanel(screen *ebiten.Image) {
	if !s.onlineMode {
		return
	}
	r := embeddedOnlinePanelRect()
	drawFilledRect(screen, r, color.RGBA{A: 170})
	drawTextFace(screen, "Session", uiTextFace, r.Min.X+16, r.Min.Y+24, colornames.White)
	if s.onlineJoinCode == "" && !s.onlineConnected {
		drawButton(screen, embeddedOnlineStartRect(), "Start")
		drawButton(screen, embeddedOnlineJoinRect(), "Join")
	} else if s.onlineHost {
		drawTextFace(screen, "Join-Code", dialogTextFace, r.Min.X+16, r.Min.Y+55, colornames.Lightblue)
		codeRect := embeddedOnlineCodeRect()
		drawFrame(screen, codeRect, color.RGBA{R: 255, G: 255, B: 255, A: 230}, colornames.White)
		drawCenteredTextFace(screen, s.onlineJoinCode, codeRect, uiTextFace, colornames.Black)
		drawTextFace(screen, "⧉", uiTextFace, codeRect.Max.X+8, codeRect.Min.Y+22, colornames.White)
		drawButton(screen, embeddedOnlineCancelRect(), "Abbruch")
	} else if s.onlineConnected {
		drawTextFace(screen, "Verbunden!", uiTextFace, r.Min.X+16, r.Min.Y+58, colornames.Lightgreen)
		drawTextFace(screen, "Warte auf weitere Spieler...", dialogTextFace, r.Min.X+16, r.Min.Y+82, colornames.White)
	}
	if s.onlineStatus != "" {
		pos := embeddedOnlineStatusTextPosition(s.onlineHost && s.onlineJoinCode != "")
		drawTextFace(screen, s.onlineStatus, dialogTextFace, pos.X, pos.Y, colornames.Silver)
	}
}

func (s *playerSelectionScene) drawEmbeddedOnlineJoinOverlay(screen *ebiten.Image) {
	drawFilledRect(screen, screen.Bounds(), color.RGBA{A: 150})
	r := embeddedJoinDialogRect()
	drawFrame(screen, r, color.RGBA{R: 18, G: 20, B: 24, A: 235}, colornames.White)
	drawCenteredTextFace(screen, "Join-Code", image.Rect(r.Min.X, r.Min.Y+20, r.Max.X, r.Min.Y+48), uiTextFace, colornames.White)
	input := embeddedJoinInputRect()
	drawFrame(screen, input, colornames.White, colornames.Black)
	drawTextFace(screen, s.onlineJoinInput, uiTextFace, input.Min.X+10, input.Min.Y+24, colornames.Black)
	if s.tick%60 < 30 {
		cursorX := input.Min.X + 12 + text.BoundString(uiTextFace, s.onlineJoinInput).Dx()
		drawFilledRect(screen, image.Rect(cursorX, input.Min.Y+6, cursorX+2, input.Max.Y-6), colornames.Black)
	}
	drawButton(screen, embeddedJoinPasteRect(), "⧉")
	drawButton(screen, embeddedJoinSubmitRect(), "Join")
}

func embeddedOnlinePanelRect() image.Rectangle {
	return image.Rect(650, 24, 936, 150)
}

func embeddedOnlineStartRect() image.Rectangle {
	return image.Rect(666, 66, 782, 98)
}

func embeddedOnlineJoinRect() image.Rectangle {
	return image.Rect(802, 66, 918, 98)
}

func embeddedOnlineCodeRect() image.Rectangle {
	return image.Rect(748, 58, 884, 90)
}

func embeddedOnlineCancelRect() image.Rectangle {
	return image.Rect(666, 105, 782, 136)
}

func embeddedOnlineStatusTextPosition(hostWithCode bool) image.Point {
	if hostWithCode {
		cancel := embeddedOnlineCancelRect()
		return image.Pt(cancel.Max.X+20, cancel.Min.Y+21)
	}
	r := embeddedOnlinePanelRect()
	return image.Pt(r.Min.X+16, r.Max.Y-15)
}

func embeddedJoinDialogRect() image.Rectangle {
	return image.Rect(300, 245, 660, 390)
}

func embeddedJoinInputRect() image.Rectangle {
	return image.Rect(340, 306, 548, 340)
}

func embeddedJoinPasteRect() image.Rectangle {
	return image.Rect(562, 306, 604, 340)
}

func embeddedJoinSubmitRect() image.Rectangle {
	return image.Rect(435, 354, 525, 382)
}

func (s *playerSelectionScene) drawFooter(screen *ebiten.Image) {
	t := texts()
	drawFilledRect(screen, image.Rect(372, 676, 586, 695), colornames.Yellow)
	drawCenteredText(screen, t.PlayerSelectionHelpHint, image.Rect(372, 676, 586, 695), colornames.Black)
	if core.Config().Online.Enabled {
		r := onlineSelectionButtonRect()
		label := t.PlayerSelectionOnlineButton
		if s.onlineMode {
			label = "Single Player"
		}
		if s.onlineAvailable || s.onlineMode {
			drawButton(screen, r, label)
		} else {
			drawDisabledButton(screen, r, label)
			if s.onlineButtonHovered() && !primaryPointerIsTouch() {
				drawTooltip(screen, t.PlayerSelectionOnlineUnavailable, image.Pt(r.Min.X, r.Min.Y-10))
			}
		}
	}
	drawButton(screen, image.Rect(624, 673, 756, 706), t.PlayerSelectionOptionsButton)
	if s.onlineMode && !s.onlineHost {
		drawDisabledButton(screen, image.Rect(780, 673, 922, 706), t.PlayerSelectionStartButton)
	} else {
		drawButton(screen, image.Rect(780, 673, 922, 706), t.PlayerSelectionStartButton)
	}
}

func (s *playerSelectionScene) drawVersion(screen *ebiten.Image) {
	bounds := screen.Bounds()
	drawTextFace(screen, playerSelectionVersionLabel(), dialogTextFace, bounds.Min.X+14, bounds.Max.Y-14, color.RGBA{R: 45, G: 45, B: 45, A: 155})
}

func (s *playerSelectionScene) drawVersionUpdateNotice(screen *ebiten.Image) {
	if s.versionUpdate == nil || !s.versionUpdate.available || s.versionUpdateDismissed {
		return
	}
	t := texts()
	r := versionUpdateNoticeRect(screen.Bounds())
	drawFrame(screen, r, color.RGBA{R: 255, G: 250, B: 210, A: 242}, colornames.Black)
	drawTextFace(screen, fmt.Sprintf(t.PlayerSelectionUpdateAvailable, s.versionUpdate.latest), dialogTextFace, r.Min.X+10, r.Min.Y+18, colornames.Black)
	link := versionUpdateLinkRect(screen.Bounds())
	drawFrame(screen, link, color.RGBA{R: 220, G: 236, B: 255, A: 255}, colornames.Black)
	drawCenteredTextFace(screen, t.PlayerSelectionUpdateOpen, link, dialogTextFace, color.RGBA{R: 0, G: 38, B: 180, A: 255})
	close := versionUpdateCloseRect(screen.Bounds())
	drawFrame(screen, close, color.RGBA{R: 235, G: 235, B: 220, A: 255}, colornames.Black)
	drawCenteredTextFace(screen, "x", close, dialogTextFace, colornames.Black)
}

func (s *playerSelectionScene) handleVersionUpdateClick(x, y int) bool {
	if s.versionUpdate == nil || !s.versionUpdate.available || s.versionUpdateDismissed {
		return false
	}
	bounds := s.selectionBounds()
	p := image.Pt(x, y)
	if p.In(versionUpdateCloseRect(bounds)) {
		s.versionUpdateDismissed = true
		return true
	}
	if !p.In(versionUpdateLinkRect(bounds)) {
		return false
	}
	if err := openDefaultBrowser(versionDownloadURL); err != nil {
		s.transientMessage = texts().PlayerSelectionUpdateOpenFailed
		s.transientMessageUntil = s.tick + onlineUnavailableMessageTicks
	}
	return true
}

func (s *playerSelectionScene) selectionBounds() image.Rectangle {
	if s != nil && s.baseImage != nil {
		return s.baseImage.Bounds()
	}
	return image.Rect(0, 0, int(core.Config().Screen.Width), int(core.Config().Screen.Height))
}

func versionUpdateNoticeRect(bounds image.Rectangle) image.Rectangle {
	return image.Rect(bounds.Min.X+14, bounds.Max.Y-94, bounds.Min.X+374, bounds.Max.Y-58)
}

func versionUpdateLinkRect(bounds image.Rectangle) image.Rectangle {
	notice := versionUpdateNoticeRect(bounds)
	return image.Rect(notice.Max.X-150, notice.Min.Y+8, notice.Max.X-28, notice.Max.Y-8)
}

func versionUpdateCloseRect(bounds image.Rectangle) image.Rectangle {
	notice := versionUpdateNoticeRect(bounds)
	return image.Rect(notice.Max.X-24, notice.Min.Y+4, notice.Max.X-6, notice.Min.Y+22)
}

func playerSelectionVersionLabel() string {
	version := strings.TrimSpace(buildinfo.Version)
	if version == "" {
		version = "dev"
	}
	if strings.EqualFold(version, "dev") {
		if data, err := os.ReadFile("VERSION"); err == nil {
			if fileVersion := strings.TrimSpace(string(data)); fileVersion != "" {
				version = fileVersion + " (dev)"
			}
		}
	}
	if strings.HasPrefix(strings.ToLower(version), "v") {
		return version
	}
	return "v" + version
}

func (s *playerSelectionScene) onlineButtonHovered() bool {
	if !core.Config().Online.Enabled {
		return false
	}
	x, y := primaryPointerPosition()
	x, y = s.toSelectionCoords(x, y)
	return image.Pt(x, y).In(onlineSelectionButtonRect())
}

func (s *playerSelectionScene) drawOptionsDialog(screen *ebiten.Image) {
	t := texts()
	r := optionsDialogRect()
	drawDialogWindow(screen, r, t.OptionsTitle)

	drawGroupBox(screen, image.Rect(r.Min.X+16, r.Min.Y+53, r.Min.X+285, r.Min.Y+153), t.OptionsProjectileReentry)
	reentryLabels := []string{t.OptionsProjectileReentryOff, t.OptionsProjectileReentryAlways, t.OptionsProjectileReentryRandom}
	for i, label := range reentryLabels {
		y := r.Min.Y + 81 + i*24
		drawRadio(screen, r.Min.X+37, y, s.optionsDraft.projectileReentry == i)
		drawTextFace(screen, label, dialogTextFace, r.Min.X+52, y+4, colornames.Black)
	}

	drawGroupBox(screen, image.Rect(r.Min.X+16, r.Min.Y+174, r.Min.X+285, r.Min.Y+266), t.OptionsCloudAggression)
	drawSlider(screen, image.Rect(r.Min.X+31, r.Min.Y+192, r.Min.X+271, r.Min.Y+214), s.optionsDraft.cloudAggression)
	drawTextFace(screen, "0", dialogTextFace, r.Min.X+31, r.Min.Y+230, colornames.Black)
	drawTextFace(screen, "50", dialogTextFace, r.Min.X+143, r.Min.Y+230, colornames.Black)
	drawTextFace(screen, "100", dialogTextFace, r.Min.X+256, r.Min.Y+230, colornames.Black)
	drawDialogButton(screen, image.Rect(r.Min.X+120, r.Min.Y+236, r.Min.X+143, r.Min.Y+258), "<")
	drawDialogButton(screen, image.Rect(r.Min.X+150, r.Min.Y+236, r.Min.X+173, r.Min.Y+258), ">")

	drawGroupBox(screen, image.Rect(r.Min.X+301, r.Min.Y+53, r.Min.X+432, r.Min.Y+185), t.OptionsPalmCount)
	palmLabels := []string{"0", "1", "2", t.OptionsProjectileReentryRandom}
	palmValues := []int{0, 1, 2, -1}
	for i, label := range palmLabels {
		y := r.Min.Y + 88 + i*24
		drawRadio(screen, r.Min.X+326, y, s.optionsDraft.palmCount == palmValues[i])
		drawTextFace(screen, label, dialogTextFace, r.Min.X+343, y+4, colornames.Black)
	}

	drawGroupBox(screen, image.Rect(r.Min.X+301, r.Min.Y+208, r.Min.X+518, r.Min.Y+266), t.OptionsRoundStart)
	drawCheckbox(screen, r.Min.X+321, r.Min.Y+242, s.optionsDraft.quickRoundStart)
	drawTextFace(screen, t.OptionsQuickRoundStart, dialogTextFace, r.Min.X+338, r.Min.Y+247, colornames.Black)

	drawDialogButton(screen, image.Rect(r.Max.X-88, r.Min.Y+56, r.Max.X-16, r.Min.Y+77), t.DialogOK)
	drawDialogButton(screen, image.Rect(r.Max.X-88, r.Min.Y+86, r.Max.X-16, r.Min.Y+107), t.DialogCancel)
	drawDialogButton(screen, image.Rect(r.Max.X-88, r.Min.Y+166, r.Max.X-16, r.Min.Y+187), t.OptionsLanguageButton)
}

func optionsDialogRect() image.Rectangle {
	return image.Rect(0, 0, 530, 284).Add(image.Pt((960-530)/2, (721-284)/2))
}

func languageDialogRect() image.Rectangle {
	return image.Rect(0, 0, 248, 161).Add(image.Pt((960-248)/2, (721-161)/2))
}

func helpDialogRect() image.Rectangle {
	return image.Rect(0, 0, 620, 315).Add(image.Pt((960-620)/2, (721-315)/2))
}

func (s *playerSelectionScene) drawHelpDialog(screen *ebiten.Image) {
	t := texts()
	r := helpDialogRect()
	drawDialogWindow(screen, r, t.PlayerSelectionHelpTitle)

	leftCard := image.Rect(r.Min.X+14, r.Min.Y+49, r.Min.X+113, r.Min.Y+201)
	humanCard := image.Rect(r.Min.X+90, r.Min.Y+81, r.Min.X+190, r.Min.Y+235)
	computerCard := image.Rect(r.Min.X+214, r.Min.Y+81, r.Min.X+314, r.Min.Y+235)
	keysBox := image.Rect(r.Min.X+374, r.Min.Y+50, r.Max.X-12, r.Min.Y+222)

	drawFrame(screen, leftCard, color.RGBA{R: 190, G: 190, B: 190, A: 255}, colornames.Black)
	drawFilledRect(screen, image.Rect(leftCard.Min.X, leftCard.Min.Y, leftCard.Max.X, leftCard.Min.Y+24), color.RGBA{R: 220, G: 220, B: 220, A: 255})
	drawCenteredTextFace(screen, t.PlayerSelectionHelpSlotNone, image.Rect(leftCard.Min.X, leftCard.Min.Y, leftCard.Max.X, leftCard.Min.Y+24), dialogTextFace, colornames.Black)

	drawHelpPlayerCard(screen, humanCard, s.humanPortrait, t.PlayerSelectionHelpHuman, t.PlayerSelectionHelpPlayerName, color.RGBA{R: 35, G: 65, B: 255, A: 255})
	drawHelpPlayerCard(screen, computerCard, s.computerPortraitFor(computerplayers.DoedelID), t.PlayerSelectionSlotComputer, "D. Dödel", color.RGBA{R: 255, G: 45, B: 45, A: 255})

	drawHelpArrow(screen, r.Min.X+145, r.Min.Y+58, humanCard.Min.X+58, humanCard.Min.Y, t.PlayerSelectionHelpPlayerType)
	drawHelpArrow(screen, r.Min.X+296, r.Min.Y+58, computerCard.Max.X-15, computerCard.Min.Y+10, t.PlayerSelectionHelpColor)
	drawHelpArrow(screen, r.Min.X+105, r.Min.Y+281, humanCard.Min.X+50, humanCard.Max.Y-14, t.PlayerSelectionHelpName)
	drawHelpArrow(screen, r.Min.X+260, r.Min.Y+264, computerCard.Min.X+85, computerCard.Max.Y-18, t.PlayerSelectionHelpComputerType)

	drawGroupBox(screen, keysBox, t.PlayerSelectionHelpKeys)
	helpKeys := []string{
		t.PlayerSelectionHelpKeyIncreaseRounds,
		t.PlayerSelectionHelpKeyDecreaseRounds,
		t.PlayerSelectionHelpKeyOptions,
		t.PlayerSelectionHelpKeyStart,
	}
	for i, line := range helpKeys {
		drawTextFace(screen, line, dialogTextFace, keysBox.Min.X+16, keysBox.Min.Y+31+i*24, colornames.Black)
	}
	drawDialogButton(screen, image.Rect(r.Max.X-75, r.Max.Y-63, r.Max.X-12, r.Max.Y-39), t.DialogOK)
}

func (s *playerSelectionScene) drawLanguageDialog(screen *ebiten.Image) {
	t := texts()
	r := languageDialogRect()
	drawDialogWindow(screen, r, t.LanguageTitle)
	drawGermanFlag(screen, image.Rect(r.Min.X+22, r.Min.Y+53, r.Min.X+54, r.Min.Y+72))
	drawBritishFlag(screen, image.Rect(r.Min.X+22, r.Min.Y+86, r.Min.X+54, r.Min.Y+105))
	drawRadio(screen, r.Min.X+82, r.Min.Y+62, s.languageDraft == languageGerman)
	drawTextFace(screen, t.LanguageGerman, dialogTextFace, r.Min.X+94, r.Min.Y+66, colornames.Black)
	drawRadio(screen, r.Min.X+82, r.Min.Y+95, s.languageDraft == languageEnglish)
	drawTextFace(screen, t.LanguageEnglish, dialogTextFace, r.Min.X+94, r.Min.Y+99, colornames.Black)
	drawDialogButton(screen, image.Rect(r.Min.X+88, r.Min.Y+120, r.Min.X+162, r.Min.Y+141), t.DialogOK)
}

func drawGermanFlag(screen *ebiten.Image, r image.Rectangle) {
	drawFilledRect(screen, image.Rect(r.Min.X, r.Min.Y, r.Max.X, r.Min.Y+r.Dy()/3), colornames.Black)
	drawFilledRect(screen, image.Rect(r.Min.X, r.Min.Y+r.Dy()/3, r.Max.X, r.Min.Y+(r.Dy()*2)/3), colornames.Red)
	drawFilledRect(screen, image.Rect(r.Min.X, r.Min.Y+(r.Dy()*2)/3, r.Max.X, r.Max.Y), colornames.Yellow)
}

func drawBritishFlag(screen *ebiten.Image, r image.Rectangle) {
	drawFilledRect(screen, r, color.RGBA{R: 15, G: 45, B: 160, A: 255})
	ebitenutil.DrawLine(screen, float64(r.Min.X), float64(r.Min.Y), float64(r.Max.X), float64(r.Max.Y), colornames.White)
	ebitenutil.DrawLine(screen, float64(r.Min.X), float64(r.Max.Y), float64(r.Max.X), float64(r.Min.Y), colornames.White)
	drawFilledRect(screen, image.Rect(r.Min.X, r.Min.Y+r.Dy()/2-2, r.Max.X, r.Min.Y+r.Dy()/2+2), colornames.White)
	drawFilledRect(screen, image.Rect(r.Min.X+r.Dx()/2-2, r.Min.Y, r.Min.X+r.Dx()/2+2, r.Max.Y), colornames.White)
	drawFilledRect(screen, image.Rect(r.Min.X, r.Min.Y+r.Dy()/2-1, r.Max.X, r.Min.Y+r.Dy()/2+1), colornames.Red)
	drawFilledRect(screen, image.Rect(r.Min.X+r.Dx()/2-1, r.Min.Y, r.Min.X+r.Dx()/2+1, r.Max.Y), colornames.Red)
}

func drawHelpPlayerCard(screen *ebiten.Image, r image.Rectangle, portrait *ebiten.Image, title, name string, titleColor color.Color) {
	drawFrame(screen, r, colornames.White, colornames.Black)
	drawFilledRect(screen, image.Rect(r.Min.X+1, r.Min.Y+1, r.Max.X-1, r.Min.Y+26), color.RGBA{R: 225, G: 225, B: 255, A: 255})
	drawCenteredTextFace(screen, title, image.Rect(r.Min.X, r.Min.Y, r.Max.X, r.Min.Y+26), dialogTextFace, titleColor)
	if portrait != nil {
		drawScaledImage(screen, portrait, image.Rect(r.Min.X+8, r.Min.Y+29, r.Max.X-8, r.Max.Y-30))
	}
	drawFrame(screen, image.Rect(r.Min.X+4, r.Max.Y-28, r.Max.X-4, r.Max.Y-4), color.RGBA{R: 255, G: 255, B: 220, A: 255}, colornames.Black)
	drawCenteredTextFace(screen, name, image.Rect(r.Min.X+4, r.Max.Y-28, r.Max.X-4, r.Max.Y-4), dialogTextFace, colornames.Black)
	drawPaintSwatch(screen, image.Rect(r.Max.X-27, r.Min.Y+41, r.Max.X-6, r.Min.Y+64), color.RGBA{R: 35, G: 35, B: 255, A: 255})
}

func drawHelpArrow(screen *ebiten.Image, labelX, labelY, targetX, targetY int, label string) {
	drawTextFace(screen, label, dialogTextFace, labelX, labelY, colornames.Black)
	startX := labelX + text.BoundString(dialogTextFace, label).Dx()/2
	startY := labelY + 5
	drawRedArrow(screen, startX, startY, targetX, targetY)
}

func drawRedArrow(screen *ebiten.Image, x1, y1, x2, y2 int) {
	ebitenutil.DrawLine(screen, float64(x1), float64(y1), float64(x2), float64(y2), colornames.Red)
	ebitenutil.DrawLine(screen, float64(x2), float64(y2), float64(x2-5), float64(y2-3), colornames.Red)
	ebitenutil.DrawLine(screen, float64(x2), float64(y2), float64(x2+3), float64(y2-5), colornames.Red)
}

func drawDialogWindow(screen *ebiten.Image, r image.Rectangle, title string) {
	drawFilledRect(screen, r, color.RGBA{R: 240, G: 240, B: 240, A: 255})
	drawFilledRect(screen, image.Rect(r.Min.X, r.Min.Y, r.Max.X, r.Min.Y+28), color.RGBA{R: 34, G: 39, B: 49, A: 255})
	drawFilledRect(screen, image.Rect(r.Min.X, r.Min.Y, r.Min.X+1, r.Max.Y), color.RGBA{R: 85, G: 85, B: 85, A: 255})
	drawFilledRect(screen, image.Rect(r.Max.X-1, r.Min.Y, r.Max.X, r.Max.Y), colornames.White)
	drawFilledRect(screen, image.Rect(r.Min.X, r.Max.Y-1, r.Max.X, r.Max.Y), colornames.White)
	drawFrame(screen, image.Rect(r.Min.X+5, r.Min.Y+5, r.Min.X+23, r.Min.Y+23), colornames.White, color.RGBA{R: 115, G: 125, B: 140, A: 255})
	drawCenteredTextFace(screen, title, image.Rect(r.Min.X, r.Min.Y, r.Max.X, r.Min.Y+28), dialogTextFace, colornames.White)
	drawTextFace(screen, "v", dialogTextFace, r.Max.X-45, r.Min.Y+18, colornames.White)
	drawTextFace(screen, "x", dialogTextFace, r.Max.X-18, r.Min.Y+18, colornames.White)
}

func drawGroupBox(screen *ebiten.Image, r image.Rectangle, label string) {
	drawFilledRect(screen, image.Rect(r.Min.X, r.Min.Y, r.Max.X, r.Min.Y+1), color.RGBA{R: 170, G: 170, B: 170, A: 255})
	drawFilledRect(screen, image.Rect(r.Min.X, r.Max.Y-1, r.Max.X, r.Max.Y), color.RGBA{R: 170, G: 170, B: 170, A: 255})
	drawFilledRect(screen, image.Rect(r.Min.X, r.Min.Y, r.Min.X+1, r.Max.Y), color.RGBA{R: 170, G: 170, B: 170, A: 255})
	drawFilledRect(screen, image.Rect(r.Max.X-1, r.Min.Y, r.Max.X, r.Max.Y), color.RGBA{R: 170, G: 170, B: 170, A: 255})
	drawFilledRect(screen, image.Rect(r.Min.X+7, r.Min.Y-8, r.Min.X+10+text.BoundString(dialogTextFace, label).Dx(), r.Min.Y+5), color.RGBA{R: 240, G: 240, B: 240, A: 255})
	drawTextFace(screen, label, dialogTextFace, r.Min.X+9, r.Min.Y+2, colornames.Black)
}

func drawRadio(screen *ebiten.Image, x, y int, selected bool) {
	ebitenutil.DrawCircle(screen, float64(x), float64(y), 6, colornames.Black)
	ebitenutil.DrawCircle(screen, float64(x), float64(y), 5, colornames.White)
	if selected {
		ebitenutil.DrawCircle(screen, float64(x), float64(y), 3, colornames.Black)
	}
}

func drawCheckbox(screen *ebiten.Image, x, y int, checked bool) {
	drawFrame(screen, image.Rect(x, y, x+12, y+12), colornames.White, colornames.Black)
	if checked {
		ebitenutil.DrawLine(screen, float64(x+2), float64(y+6), float64(x+5), float64(y+9), colornames.Black)
		ebitenutil.DrawLine(screen, float64(x+5), float64(y+9), float64(x+10), float64(y+2), colornames.Black)
	}
}

func drawSlider(screen *ebiten.Image, r image.Rectangle, value int) {
	drawFrame(screen, r, colornames.White, color.RGBA{R: 220, G: 220, B: 220, A: 255})
	fillW := r.Dx() * maxInt(0, minInt(100, value)) / 100
	drawFilledRect(screen, image.Rect(r.Min.X, r.Min.Y, r.Min.X+fillW, r.Max.Y), color.RGBA{R: 64, G: 160, B: 235, A: 255})
}

func (s *playerSelectionScene) drawPalette(screen *ebiten.Image, slotIndex int) {
	x0, y0 := palettePosForSlot(slotIndex)
	frame := image.Rect(x0-6, y0-6, x0+len(paletteColors)*26+4, y0+28)
	drawFrame(screen, frame, colornames.White, colornames.Black)
	for i, c := range paletteColors {
		cell := image.Rect(x0+i*26, y0, x0+i*26+22, y0+22)
		drawFrame(screen, cell, colornames.White, colornames.Black)
		drawPaintSwatch(screen, insetRect(cell, 1), c)
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
	pressed := buttonPressed(screen, r)
	drawFrame(screen, r, color.RGBA{R: 210, G: 210, B: 210, A: 255}, colornames.Black)
	if pressed {
		drawPressedButtonEdges(screen, r, 3)
		drawCenteredText(screen, label, r.Add(image.Pt(1, 1)), colornames.Black)
		return
	}
	drawRaisedButtonEdges(screen, r, 3)
	drawCenteredText(screen, label, r, colornames.Black)
}

func drawDisabledButton(screen *ebiten.Image, r image.Rectangle, label string) {
	drawFrame(screen, r, color.RGBA{R: 175, G: 175, B: 175, A: 255}, color.RGBA{R: 95, G: 95, B: 95, A: 255})
	drawFilledRect(screen, image.Rect(r.Min.X+2, r.Min.Y+2, r.Max.X-2, r.Min.Y+5), color.RGBA{R: 205, G: 205, B: 205, A: 255})
	drawFilledRect(screen, image.Rect(r.Min.X+2, r.Min.Y+2, r.Min.X+5, r.Max.Y-2), color.RGBA{R: 205, G: 205, B: 205, A: 255})
	drawCenteredText(screen, label, r.Add(image.Pt(1, 1)), color.RGBA{R: 115, G: 115, B: 115, A: 255})
	drawCenteredText(screen, label, r, color.RGBA{R: 65, G: 65, B: 65, A: 255})
}

func drawTooltip(screen *ebiten.Image, label string, anchor image.Point) {
	paddingX := 8
	paddingY := 5
	b := text.BoundString(dialogTextFace, label)
	w := b.Dx() + paddingX*2
	h := b.Dy() + paddingY*2
	r := image.Rect(anchor.X, anchor.Y-h, anchor.X+w, anchor.Y)
	if r.Min.X < 4 {
		r = r.Add(image.Pt(4-r.Min.X, 0))
	}
	if r.Max.X > int(core.Config().Screen.Width)-4 {
		r = r.Add(image.Pt(int(core.Config().Screen.Width)-4-r.Max.X, 0))
	}
	drawFrame(screen, r, color.RGBA{R: 255, G: 252, B: 218, A: 255}, colornames.Black)
	drawTextFace(screen, label, dialogTextFace, r.Min.X+paddingX, r.Min.Y+paddingY+b.Dy(), colornames.Black)
}

func drawDialogButton(screen *ebiten.Image, r image.Rectangle, label string) {
	pressed := buttonPressed(screen, r)
	drawFrame(screen, r, color.RGBA{R: 214, G: 214, B: 214, A: 255}, colornames.Black)
	if pressed {
		drawPressedButtonEdges(screen, r, 2)
		drawCenteredTextFace(screen, label, r.Add(image.Pt(1, 1)), dialogTextFace, colornames.Black)
		return
	}
	drawRaisedButtonEdges(screen, r, 2)
	drawCenteredTextFace(screen, label, r, dialogTextFace, colornames.Black)
}

func buttonPressed(screen *ebiten.Image, r image.Rectangle) bool {
	if !primaryPointerPressed() {
		return false
	}
	x, y := primaryPointerPosition()
	return buttonCursorPoint(screen, x, y).In(r)
}

func buttonCursorPoint(screen *ebiten.Image, x, y int) image.Point {
	bounds := screen.Bounds()
	screenW := int(core.Config().Screen.Width)
	screenH := int(core.Config().Screen.Height)
	if screenW <= 0 || screenH <= 0 {
		return image.Pt(x, y)
	}
	return image.Pt(bounds.Min.X+x*bounds.Dx()/screenW, bounds.Min.Y+y*bounds.Dy()/screenH)
}

func drawRaisedButtonEdges(screen *ebiten.Image, r image.Rectangle, depth int) {
	light := colornames.White
	shadow := color.RGBA{R: 115, G: 115, B: 115, A: 255}
	drawFilledRect(screen, image.Rect(r.Min.X+2, r.Min.Y+2, r.Max.X-2, r.Min.Y+2+depth), light)
	drawFilledRect(screen, image.Rect(r.Min.X+2, r.Min.Y+2, r.Min.X+2+depth, r.Max.Y-2), light)
	drawFilledRect(screen, image.Rect(r.Min.X+2, r.Max.Y-2-depth, r.Max.X-2, r.Max.Y-2), shadow)
	drawFilledRect(screen, image.Rect(r.Max.X-2-depth, r.Min.Y+2, r.Max.X-2, r.Max.Y-2), shadow)
}

func drawPressedButtonEdges(screen *ebiten.Image, r image.Rectangle, depth int) {
	light := colornames.White
	shadow := color.RGBA{R: 105, G: 105, B: 105, A: 255}
	drawFilledRect(screen, image.Rect(r.Min.X+2, r.Min.Y+2, r.Max.X-2, r.Min.Y+2+depth), shadow)
	drawFilledRect(screen, image.Rect(r.Min.X+2, r.Min.Y+2, r.Min.X+2+depth, r.Max.Y-2), shadow)
	drawFilledRect(screen, image.Rect(r.Min.X+2, r.Max.Y-2-depth, r.Max.X-2, r.Max.Y-2), light)
	drawFilledRect(screen, image.Rect(r.Max.X-2-depth, r.Min.Y+2, r.Max.X-2, r.Max.Y-2), light)
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
	drawTextFace(screen, value, uiTextFace, x, y, c)
}

func drawTextFace(screen *ebiten.Image, value string, face font.Face, x, y int, c color.Color) {
	text.Draw(screen, value, face, x, y, c)
}

func drawCenteredText(screen *ebiten.Image, value string, r image.Rectangle, c color.Color) {
	drawCenteredTextFace(screen, value, r, uiTextFace, c)
}

func drawCenteredTextFace(screen *ebiten.Image, value string, r image.Rectangle, face font.Face, c color.Color) {
	b := text.BoundString(face, value)
	x := r.Min.X + (r.Dx()-b.Dx())/2
	y := r.Min.Y + (r.Dy()+b.Dy())/2
	drawTextFace(screen, value, face, x, y, c)
}

func drawCenteredBoldTextFace(screen *ebiten.Image, value string, r image.Rectangle, face font.Face, c color.Color) {
	b := text.BoundString(face, value)
	x := r.Min.X + (r.Dx()-b.Dx()-1)/2
	y := r.Min.Y + (r.Dy()+b.Dy())/2
	drawTextFace(screen, value, face, x+2, y+2, color.RGBA{A: 150})
	drawTextFace(screen, value, face, x+3, y+2, color.RGBA{A: 150})
	drawTextFace(screen, value, face, x, y, c)
	drawTextFace(screen, value, face, x+1, y, c)
}

func drawPaintSwatch(screen *ebiten.Image, rect image.Rectangle, c color.Color) {
	if paintSplotch == nil {
		paintSplotch = mustImageFromPNG(r.PaintSplotchPNG)
	}
	bounds := paintSplotch.Bounds()
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(float64(rect.Dx())/float64(bounds.Dx()), float64(rect.Dy())/float64(bounds.Dy()))
	op.GeoM.Translate(float64(rect.Min.X), float64(rect.Min.Y))
	op.ColorScale.ScaleWithColor(c)
	screen.DrawImage(paintSplotch, op)
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
