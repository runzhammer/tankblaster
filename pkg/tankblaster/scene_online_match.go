package tankblaster

import (
	"image"
	"image/color"
	"math"
	"math/rand"
	"strconv"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/runzhammer/tankblaster/pkg/core"
	"github.com/runzhammer/tankblaster/pkg/gamecore"
	"github.com/runzhammer/tankblaster/pkg/protocol"
	"golang.org/x/image/colornames"
)

type onlineMatchScene struct {
	g           *GameLoop
	client      *onlineClient
	state       gamecore.MatchState
	playerID    string
	matchID     string
	angle       float64
	power       float64
	status      string
	autoPlay    bool
	autoTurnKey string
	autoFireAt  int
	rng         *rand.Rand
	tick        int
}

func NewOnlineMatchScene(game *GameLoop, client *onlineClient, state gamecore.MatchState, autoPlay bool) core.Scene {
	s := &onlineMatchScene{
		g:        game,
		client:   client,
		state:    state,
		playerID: client.id.PlayerID,
		matchID:  state.MatchID,
		angle:    45,
		power:    80,
		status:   texts().OnlineMatchStarted,
		autoPlay: autoPlay,
		rng:      rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	s.aimAtOpponent()
	return s
}

func (s *onlineMatchScene) Update() error {
	s.tick++
	s.consumeNetwork()
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		if s.client != nil {
			s.client.Close()
		}
		return s.g.SetNewScene(NewOnlineScene)
	}
	if s.state.Status != gamecore.MatchInGame {
		return nil
	}
	if s.autoPlay {
		s.updateAutopilot()
		return nil
	}
	if !s.isMyTurn() {
		return nil
	}
	s.handleKeyboard()
	if primaryPointerJustPressed() {
		x, y := primaryPointerPosition()
		s.handlePointer(image.Pt(x, y))
	}
	return nil
}

func (s *onlineMatchScene) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{R: 94, G: 167, B: 220, A: 255})
	ebitenutil.DrawRect(screen, 0, 590, 1024, 178, color.RGBA{R: 92, G: 137, B: 62, A: 255})
	ebitenutil.DrawRect(screen, 0, 632, 1024, 136, color.RGBA{R: 115, G: 87, B: 52, A: 255})
	for _, crater := range s.state.Terrain {
		ebitenutil.DrawCircle(screen, crater.X, crater.Y, crater.Radius, color.RGBA{R: 62, G: 43, B: 31, A: 255})
	}
	for i, tank := range s.state.Tanks {
		s.drawTank(screen, i, tank)
	}
	s.drawHUD(screen)
}

func (s *onlineMatchScene) consumeNetwork() {
	if s.client == nil {
		return
	}
	for {
		select {
		case env := <-s.client.recv:
			s.handleMessage(env)
		case err := <-s.client.errs:
			s.status = err.Error()
		default:
			return
		}
	}
}

func (s *onlineMatchScene) handleMessage(env protocol.Envelope) {
	switch env.Type {
	case protocol.TypeGameStart, protocol.TypeTurnStart, protocol.TypeStateUpdate, protocol.TypeGameOver:
		msg, err := protocol.Decode[protocol.StateUpdate](env)
		if err == nil {
			s.state = msg.State
			s.matchID = msg.State.MatchID
			s.aimAtOpponent()
		}
	case protocol.TypeShotResult:
		msg, err := protocol.Decode[protocol.ShotResult](env)
		if err == nil {
			s.state = msg.State
			s.status = "Treffer: " + strconv.Itoa(len(msg.Result.Damage))
			s.aimAtOpponent()
		}
	case protocol.TypeError:
		msg, err := protocol.Decode[protocol.Error](env)
		if err == nil {
			s.status = msg.Code + ": " + msg.Message
		}
	}
}

func (s *onlineMatchScene) handleKeyboard() {
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		s.angle = math.Max(0, s.angle-1)
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		s.angle = math.Min(180, s.angle+1)
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		s.power = math.Max(0, s.power-1)
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		s.power = math.Min(100, s.power+1)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeyKPEnter) {
		s.fire()
	}
}

func (s *onlineMatchScene) handlePointer(p image.Point) {
	switch {
	case p.In(onlineMatchAngleMinusRect()):
		s.angle = math.Max(0, s.angle-5)
	case p.In(onlineMatchAnglePlusRect()):
		s.angle = math.Min(180, s.angle+5)
	case p.In(onlineMatchPowerMinusRect()):
		s.power = math.Max(0, s.power-5)
	case p.In(onlineMatchPowerPlusRect()):
		s.power = math.Min(100, s.power+5)
	case p.In(onlineMatchFireRect()):
		s.fire()
	}
}

func (s *onlineMatchScene) fire() {
	if s.client == nil || !s.isMyTurn() || s.matchID == "" {
		return
	}
	s.client.Send(protocol.TypeFire, protocol.FireCommand{
		MatchID: s.matchID,
		Weapon:  "atom_bomb",
		Angle:   s.angle,
		Power:   s.power,
	})
	s.status = "Schuss gesendet"
}

func (s *onlineMatchScene) updateAutopilot() {
	if !s.isMyTurn() || s.tick < s.autoFireAt {
		return
	}
	turnKey := s.state.MatchID + ":" + strconv.Itoa(s.state.CurrentPlayerIndex) + ":" + strconv.Itoa(len(s.state.Terrain))
	if turnKey == s.autoTurnKey {
		return
	}
	s.autoTurnKey = turnKey
	s.autoFireAt = s.tick + 45 + s.rng.Intn(45)
	s.aimAtOpponent()
	s.fire()
}

func (s *onlineMatchScene) aimAtOpponent() {
	shooter, target := s.shooterAndTarget()
	if shooter == nil || target == nil {
		return
	}
	if target.X >= shooter.X {
		s.angle = 0
	} else {
		s.angle = 180
	}
	dx := math.Abs(target.X-shooter.X) - math.Abs(float64(s.state.Wind))*0.8
	s.power = math.Max(25, math.Min(100, dx/7))
	if s.power > 82 {
		s.power = 82 + s.rng.Float64()*18
	}
}

func (s *onlineMatchScene) shooterAndTarget() (*gamecore.TankState, *gamecore.TankState) {
	var shooter *gamecore.TankState
	for i := range s.state.Tanks {
		if s.state.Tanks[i].PlayerID == s.playerID {
			shooter = &s.state.Tanks[i]
			break
		}
	}
	if shooter == nil {
		return nil, nil
	}
	var target *gamecore.TankState
	best := math.MaxFloat64
	for i := range s.state.Tanks {
		tank := &s.state.Tanks[i]
		if !tank.Alive || tank.PlayerID == shooter.PlayerID {
			continue
		}
		distance := math.Abs(tank.X - shooter.X)
		if distance < best {
			best = distance
			target = tank
		}
	}
	return shooter, target
}

func (s *onlineMatchScene) isMyTurn() bool {
	return s.state.CurrentPlayerIndex >= 0 &&
		s.state.CurrentPlayerIndex < len(s.state.Players) &&
		s.state.Players[s.state.CurrentPlayerIndex].ID == s.playerID
}

func (s *onlineMatchScene) drawTank(screen *ebiten.Image, index int, tank gamecore.TankState) {
	body := image.Rect(int(tank.X)-22, int(tank.Y)-18, int(tank.X)+22, int(tank.Y)+4)
	fill := defaultTankColor(index)
	if !tank.Alive {
		fill = color.RGBA{R: 60, G: 60, B: 60, A: 255}
	}
	drawFrame(screen, body, fill, colornames.Black)
	ebitenutil.DrawRect(screen, float64(body.Min.X+4), float64(body.Min.Y-6), 36, 6, colornames.Black)
	if index == s.state.CurrentPlayerIndex && s.state.Status == gamecore.MatchInGame {
		ebitenutil.DrawCircle(screen, tank.X, tank.Y-36, 9, colornames.Yellow)
	}
	name := s.playerName(tank.PlayerID)
	drawCenteredTextFace(screen, name, image.Rect(body.Min.X-80, body.Min.Y-35, body.Max.X+80, body.Min.Y-18), dialogTextFace, colornames.Black)
	hpRect := image.Rect(body.Min.X, body.Max.Y+6, body.Max.X, body.Max.Y+13)
	drawFrame(screen, hpRect, colornames.Darkred, colornames.Black)
	if tank.Health > 0 {
		w := hpRect.Dx() * tank.Health / 100
		if w > 2 {
			ebitenutil.DrawRect(screen, float64(hpRect.Min.X+1), float64(hpRect.Min.Y+1), float64(w-2), float64(hpRect.Dy()-2), colornames.Limegreen)
		}
	}
}

func (s *onlineMatchScene) drawHUD(screen *ebiten.Image) {
	hud := image.Rect(24, 24, 1000, 118)
	drawFrame(screen, hud, color.RGBA{R: 20, G: 24, B: 28, A: 230}, colornames.White)
	status := s.status
	if s.state.Status == gamecore.MatchFinished {
		status = "Spiel beendet - Gewinner: " + s.playerName(s.state.WinnerID)
	} else if s.isMyTurn() {
		status = "Du bist dran"
	} else if current := s.currentPlayerName(); current != "" {
		status = current + " ist dran"
	}
	drawTextFace(screen, status, uiTextFace, 44, 58, colornames.White)
	drawTextFace(screen, "Wind "+strconv.Itoa(s.state.Wind), dialogTextFace, 44, 88, colornames.Lightblue)
	drawTextFace(screen, "Winkel "+strconv.Itoa(int(math.Round(s.angle))), dialogTextFace, 210, 88, colornames.White)
	drawTextFace(screen, "Staerke "+strconv.Itoa(int(math.Round(s.power))), dialogTextFace, 360, 88, colornames.White)
	if s.isMyTurn() && !s.autoPlay && s.state.Status == gamecore.MatchInGame {
		drawButton(screen, onlineMatchAngleMinusRect(), "-")
		drawButton(screen, onlineMatchAnglePlusRect(), "+")
		drawButton(screen, onlineMatchPowerMinusRect(), "-")
		drawButton(screen, onlineMatchPowerPlusRect(), "+")
		drawButton(screen, onlineMatchFireRect(), "FEUER")
	}
}

func (s *onlineMatchScene) playerName(playerID string) string {
	for _, player := range s.state.Players {
		if player.ID == playerID {
			return player.DisplayName
		}
	}
	return ""
}

func (s *onlineMatchScene) currentPlayerName() string {
	if s.state.CurrentPlayerIndex < 0 || s.state.CurrentPlayerIndex >= len(s.state.Players) {
		return ""
	}
	return s.state.Players[s.state.CurrentPlayerIndex].DisplayName
}

func onlineMatchAngleMinusRect() image.Rectangle {
	return image.Rect(520, 66, 552, 98)
}

func onlineMatchAnglePlusRect() image.Rectangle {
	return image.Rect(558, 66, 590, 98)
}

func onlineMatchPowerMinusRect() image.Rectangle {
	return image.Rect(610, 66, 642, 98)
}

func onlineMatchPowerPlusRect() image.Rectangle {
	return image.Rect(648, 66, 680, 98)
}

func onlineMatchFireRect() image.Rectangle {
	return image.Rect(708, 66, 812, 98)
}
