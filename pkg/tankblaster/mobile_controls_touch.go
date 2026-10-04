//go:build android || ios || js

package tankblaster

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/runzhammer/tankblaster/pkg/core"
	"golang.org/x/image/colornames"
)

type mobileControlID string

const (
	mobileControlAngleMinus    mobileControlID = "angle_minus"
	mobileControlAnglePlus     mobileControlID = "angle_plus"
	mobileControlStrengthMinus mobileControlID = "strength_minus"
	mobileControlStrengthPlus  mobileControlID = "strength_plus"
	mobileControlWeaponPrev    mobileControlID = "weapon_prev"
	mobileControlWeaponNext    mobileControlID = "weapon_next"
	mobileControlFire          mobileControlID = "fire"
	mobileControlIgnition      mobileControlID = "ignition"
	mobileControlDriveLeft     mobileControlID = "drive_left"
	mobileControlDriveStop     mobileControlID = "drive_stop"
	mobileControlDriveRight    mobileControlID = "drive_right"
	mobileControlMotorOff      mobileControlID = "motor_off"
)

type mobileControlButton struct {
	id    mobileControlID
	rect  image.Rectangle
	label string
}

func (s *GameScene) handleMobileSideControls(tank *battleTank, strengthStep int) {
	if !mobileControlsEnabled() {
		return
	}
	if tank == nil {
		return
	}
	for _, button := range s.mobileControlButtons() {
		if !s.mobileControlAction(button.rect) {
			continue
		}
		switch button.id {
		case mobileControlAngleMinus:
			s.adjustTankCannon(tank, -s.humanCannonStep())
		case mobileControlAnglePlus:
			s.adjustTankCannon(tank, s.humanCannonStep())
		case mobileControlStrengthMinus:
			s.adjustShotStrength(tank, -strengthStep)
		case mobileControlStrengthPlus:
			s.adjustShotStrength(tank, strengthStep)
		case mobileControlWeaponPrev:
			s.selectPreviousWeapon(tank)
		case mobileControlWeaponNext:
			s.selectNextWeapon(tank)
		case mobileControlFire:
			s.requestFireActiveWeapon()
		case mobileControlIgnition:
			if mobileControlJustPressed(button.rect) && s.playerHasXMV12(tank.playerIndex) && s.dieselForPlayer(tank.playerIndex) > 0 {
				s.requestXMV12Start(tank)
			}
		}
	}
}

func (s *GameScene) handleMobileXMV12SideControls(tank *battleTank) {
	if !mobileControlsEnabled() {
		return
	}
	if tank == nil {
		return
	}
	for _, button := range s.mobileControlButtons() {
		if !s.mobileControlAction(button.rect) {
			continue
		}
		switch button.id {
		case mobileControlDriveLeft:
			s.requestXMV12Direction(tank, -1)
		case mobileControlDriveStop:
			s.requestXMV12Direction(tank, 0)
		case mobileControlDriveRight:
			s.requestXMV12Direction(tank, 1)
		case mobileControlMotorOff:
			if !mobileControlJustPressed(button.rect) {
				continue
			}
			s.requestXMV12MotorOff(tank)
		}
	}
}

func (s *GameScene) drawMobileSideControls(screen *ebiten.Image, viewport image.Rectangle) {
	if !mobileControlsEnabled() {
		return
	}
	if s.phase != phaseBattle || s.activeTank() == nil || s.activeTank().player.Kind == PlayerComputer {
		return
	}
	for _, button := range s.mobileControlButtons() {
		drawMobileControlButton(screen, button)
	}
}

func (s *GameScene) mobileControlAt(p image.Point) (mobileControlButton, bool) {
	for _, button := range s.mobileControlButtons() {
		if p.In(button.rect) {
			return button, true
		}
	}
	return mobileControlButton{}, false
}

func (s *GameScene) mobileControlButtons() []mobileControlButton {
	return s.mobileControlButtonsForViewport(core.GameViewport())
}

func (s *GameScene) mobileControlButtonsForViewport(viewport image.Rectangle) []mobileControlButton {
	tank := s.activeTank()
	if tank == nil {
		return nil
	}
	screenH := viewport.Dy()
	sideW := viewport.Min.X
	if sideW < 72 {
		return nil
	}
	gap := 8
	leftY := 36
	rightY := 76
	buttonH := minInt(58, (screenH-leftY-gap*7)/6)
	if buttonH < 42 {
		buttonH = 42
	}
	buttonW := sideW - gap*2
	if buttonW > 150 {
		buttonW = 150
	}
	leftX := viewport.Min.X - gap - buttonW
	rightX := viewport.Max.X + gap
	step := buttonH + gap

	if s.xmV12DriveMode {
		return []mobileControlButton{
			{mobileControlDriveRight, image.Rect(rightX, rightY, rightX+buttonW, rightY+buttonH), ">"},
			{mobileControlDriveStop, image.Rect(rightX, rightY+step, rightX+buttonW, rightY+step+buttonH), "STOP"},
			{mobileControlDriveLeft, image.Rect(rightX, rightY+step*2, rightX+buttonW, rightY+step*2+buttonH), "<"},
			{mobileControlMotorOff, image.Rect(rightX, rightY+step*3+20, rightX+buttonW, rightY+step*3+20+buttonH), texts().XMV12MotorOff},
		}
	}

	buttons := []mobileControlButton{
		{mobileControlAnglePlus, image.Rect(leftX, leftY, leftX+buttonW, leftY+buttonH), "Winkel +"},
		{mobileControlAngleMinus, image.Rect(leftX, leftY+step, leftX+buttonW, leftY+step+buttonH), "Winkel -"},
		{mobileControlStrengthPlus, image.Rect(leftX, leftY+step*2, leftX+buttonW, leftY+step*2+buttonH), "Power +"},
		{mobileControlStrengthMinus, image.Rect(leftX, leftY+step*3, leftX+buttonW, leftY+step*3+buttonH), "Power -"},
		{mobileControlWeaponNext, image.Rect(leftX, leftY+step*4, leftX+buttonW, leftY+step*4+buttonH), "Waffe >"},
		{mobileControlWeaponPrev, image.Rect(leftX, leftY+step*5, leftX+buttonW, minInt(screenH-gap, leftY+step*5+buttonH)), "Waffe <"},
		{mobileControlFire, image.Rect(rightX, rightY, rightX+buttonW, rightY+buttonH+18), "FEUER"},
	}
	if s.playerHasXMV12(tank.playerIndex) && s.dieselForPlayer(tank.playerIndex) > 0 {
		buttons = append(buttons, mobileControlButton{
			mobileControlIgnition,
			image.Rect(rightX, rightY+step+28, rightX+buttonW, rightY+step+28+buttonH),
			texts().GameHUDIgnition,
		})
	}
	return buttons
}

func (s *GameScene) mobileControlAction(r image.Rectangle) bool {
	if mobileControlJustPressed(r) {
		return true
	}
	return mobileControlPressed(r) && int(s.time)%humanCannonRepeatFrames == 0
}

func drawMobileControlButton(screen *ebiten.Image, button mobileControlButton) {
	if button.rect.Empty() {
		return
	}
	fill := colorRGBA(42, 48, 54, 235)
	if mobileControlPressed(button.rect) {
		fill = colorRGBA(82, 90, 99, 245)
	}
	drawFilledRect(screen, button.rect, fill)
	drawFrame(screen, button.rect, fill, colornames.White)
	ebitenutil.DebugPrintAt(screen, button.label, button.rect.Min.X+10, button.rect.Min.Y+button.rect.Dy()/2-7)
}

func mobileControlJustPressed(r image.Rectangle) bool {
	for _, id := range inpututil.AppendJustPressedTouchIDs(nil) {
		x, y := mobileControlTouchPosition(id)
		if image.Pt(x, y).In(r) {
			return true
		}
	}
	return false
}

func mobileControlPressed(r image.Rectangle) bool {
	for _, id := range ebiten.AppendTouchIDs(nil) {
		x, y := mobileControlTouchPosition(id)
		if image.Pt(x, y).In(r) {
			return true
		}
	}
	return false
}

func colorRGBA(r, g, b, a uint8) color.RGBA {
	return color.RGBA{R: r, G: g, B: b, A: a}
}
