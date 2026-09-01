package tankblaster

import (
	"image"
	"image/color"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/runzhammer/gamedemo/pkg/core"
	"github.com/runzhammer/gamedemo/pkg/gamecore"
	"github.com/runzhammer/gamedemo/pkg/protocol"
	"golang.org/x/image/colornames"
)

const onlineJoinInputMaxRunes = 512

type onlineScene struct {
	g           *GameLoop
	client      *onlineClient
	displayName string
	joinInput   string
	nameFocused bool
	status      string
	sessionID   string
	matchID     string
	inviteURL   string
	rounds      int
	sessions    []protocol.SessionSummary
	leaders     []protocol.LeaderboardEntry
	inputRunes  []rune
	matchState  gamecore.MatchState
	autoJoin    bool
	autoPlay    bool
	autoQueued  bool
	autoTurnKey string
	autoFireAt  int
	readySent   map[string]bool
	rng         *rand.Rand
	tick        int
}

func NewOnlineScene(game *GameLoop) (core.Scene, error) {
	id := loadOnlineIdentity()
	if id.DisplayName == "" {
		id.DisplayName = texts().GameDefaultPlayerName
	}
	if value := strings.TrimSpace(os.Getenv("TANKBLASTER_ONLINE_DISPLAY_NAME")); value != "" {
		id.DisplayName = truncateRunes(value, 16)
	}
	autoJoin := onlineEnvBool("TANKBLASTER_ONLINE_AUTO_JOIN")
	autoPlay := onlineEnvBool("TANKBLASTER_ONLINE_AUTO_PLAY")
	s := &onlineScene{
		g:           game,
		displayName: id.DisplayName,
		status:      texts().OnlineConnecting,
		rounds:      normalizedOnlineRounds(game.rounds),
		autoJoin:    autoJoin || autoPlay,
		autoPlay:    autoPlay,
		readySent:   map[string]bool{},
		rng:         rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	s.client = newOnlineClient(s.displayName)
	return s, nil
}

func (s *onlineScene) Update() error {
	defer s.syncOnlineTextInputActive()
	s.tick++
	s.consumeNetwork()
	s.updateAutopilot()
	if s.matchID != "" && s.matchState.Status == gamecore.MatchInGame {
		return s.g.SetNewScene(func(game *GameLoop) (core.Scene, error) {
			return NewOnlineGameScene(game, s.client, s.matchState, s.autoPlay)
		})
	}
	s.handleKeyboard()
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return s.back()
	}
	if !primaryPointerJustPressed() {
		return nil
	}
	x, y := primaryPointerPosition()
	p := image.Pt(x, y)
	if p.In(onlineNameInputRect()) {
		s.nameFocused = true
		s.displayName = ""
		SetPlayerNameText("")
		return nil
	}
	if p.In(joinInputRect()) {
		s.finishDisplayNameInput()
		s.pasteJoinInput()
		return nil
	}
	if s.handleRoundsClick(p) {
		s.finishDisplayNameInput()
		return nil
	}
	if s.inviteURL != "" && p.In(inviteLinkRect()) {
		s.finishDisplayNameInput()
		s.copyInviteURL()
		return nil
	}
	for _, b := range s.buttons() {
		if p.In(b.rect) {
			s.finishDisplayNameInput()
			return b.action()
		}
	}
	for i, sess := range s.sessions {
		r := image.Rect(100, 410+i*42, 924, 444+i*42)
		if p.In(r) {
			s.finishDisplayNameInput()
			s.sessionID = sess.ID
			s.client.Send(protocol.TypeJoinSession, protocol.JoinSession{SessionID: sess.ID})
			return nil
		}
	}
	return nil
}

func (s *onlineScene) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{R: 24, G: 27, B: 31, A: 255})
	t := texts()
	drawTextFace(screen, t.OnlineTitle, uiTextFace, 72, 70, colornames.White)
	drawTextFace(screen, t.OnlineDisplayName, dialogTextFace, 72, 102, colornames.Lightblue)
	drawFrame(screen, onlineNameInputRect(), colornames.White, colornames.Black)
	nameText := s.displayName
	if s.nameFocused && s.tick/24%2 == 0 {
		nameText += "|"
	}
	drawTextFace(screen, nameText, uiTextFace, onlineNameInputRect().Min.X+12, onlineNameInputRect().Min.Y+24, colornames.Black)
	drawTextFace(screen, core.Config().Online.ServerURL, dialogTextFace, 446, 111, colornames.Silver)
	s.drawRounds(screen)

	drawFrame(screen, image.Rect(72, 158, 952, 214), color.RGBA{R: 39, G: 45, B: 52, A: 255}, colornames.Gray)
	drawTextFace(screen, s.status, uiTextFace, 92, 192, colornames.White)

	for _, b := range s.buttons() {
		drawButton(screen, b.rect, b.label)
	}
	drawFrame(screen, joinInputRect(), colornames.White, colornames.Black)
	joinLabel := t.OnlineJoinPrompt + ": " + s.joinInput
	drawTextFace(screen, joinLabel, uiTextFace, 84, 299, colornames.Black)

	s.drawSessions(screen)
	s.drawLeaderboard(screen)
	if s.inviteURL != "" {
		drawFrame(screen, inviteLinkRect(), color.RGBA{R: 255, G: 245, B: 184, A: 255}, colornames.Black)
		drawTextFace(screen, t.OnlineInviteLink+": "+s.inviteURL, dialogTextFace, 84, 725, colornames.Black)
	}
	s.drawFocusedNameOverlay(screen)
}

type onlineButton struct {
	label  string
	rect   image.Rectangle
	action func() error
}

func (s *onlineScene) buttons() []onlineButton {
	t := texts()
	return []onlineButton{
		{t.OnlineQuickMatch, image.Rect(72, 232, 228, 264), func() error {
			s.client.Send(protocol.TypeQuickMatch, struct{}{})
			return nil
		}},
		{t.OnlineCreatePublicSession, image.Rect(244, 232, 486, 264), func() error {
			s.client.Send(protocol.TypeCreatePublicSession, protocol.CreateSession{DisplayName: s.displayName, Rounds: s.rounds})
			return nil
		}},
		{t.OnlineCreatePrivateSession, image.Rect(502, 232, 746, 264), func() error {
			s.client.Send(protocol.TypeCreatePrivateSession, protocol.CreateSession{DisplayName: s.displayName, Rounds: s.rounds})
			return nil
		}},
		{t.OnlineOpenSessions, image.Rect(762, 232, 952, 264), func() error {
			s.client.Send(protocol.TypeListSessions, struct{}{})
			return nil
		}},
		{t.OnlineJoinSession, image.Rect(72, 316, 228, 348), func() error {
			token := inviteTokenFromInput(s.joinInput)
			if token != "" {
				s.client.Send(protocol.TypeJoinSession, protocol.JoinSession{JoinCode: token, InviteToken: token})
			}
			return nil
		}},
		{t.OnlineLeaderboard, image.Rect(244, 316, 420, 348), func() error {
			s.client.Send(protocol.TypeGetLeaderboard, protocol.LeaderboardRequest{})
			return nil
		}},
		{t.OnlineReady, image.Rect(436, 316, 560, 348), func() error {
			if s.sessionID != "" {
				s.client.Send(protocol.TypeReady, protocol.Ready{SessionID: s.sessionID, Ready: true})
			}
			return nil
		}},
		{t.OnlineLeave, image.Rect(576, 316, 700, 348), func() error {
			s.client.Send(protocol.TypeLeaveSession, struct{}{})
			s.sessionID = ""
			s.matchID = ""
			return nil
		}},
		{t.OnlineBack, image.Rect(828, 316, 952, 348), s.back},
	}
}

func (s *onlineScene) consumeNetwork() {
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

func (s *onlineScene) handleMessage(env protocol.Envelope) {
	t := texts()
	switch env.Type {
	case protocol.TypeHelloAck:
		msg, err := protocol.Decode[protocol.HelloAck](env)
		if err == nil {
			s.client.id.PlayerID = msg.PlayerID
			if msg.PlayerToken != "" {
				s.client.id.PlayerToken = msg.PlayerToken
			}
			s.client.id.DisplayName = msg.DisplayName
			saveOnlineIdentity(s.client.id)
			s.status = t.OnlineConnected + " - Rating " + strconv.Itoa(msg.Rating)
			s.updateAutopilot()
		}
	case protocol.TypeMatchmakingQueued:
		s.status = t.OnlineQueued
	case protocol.TypeSessionCreated:
		msg, err := protocol.Decode[protocol.SessionCreated](env)
		if err == nil {
			s.sessionID = msg.Session.ID
			s.status = t.OnlineSessionCreated + ": " + msg.Session.ID
			s.ensureSessionReady(msg.Session)
		}
	case protocol.TypeInviteCreated:
		msg, err := protocol.Decode[protocol.InviteCreated](env)
		if err == nil {
			s.inviteURL = msg.InviteURL
			s.joinInput = msg.JoinCode
		}
	case protocol.TypeSessionList:
		msg, err := protocol.Decode[protocol.SessionList](env)
		if err == nil {
			s.sessions = msg.Sessions
			s.status = t.OnlineOpenSessions + " " + onlineNowString()
		}
	case protocol.TypeSessionJoined, protocol.TypeSessionUpdated:
		msg, err := protocol.Decode[protocol.SessionJoined](env)
		if err == nil {
			s.sessionID = msg.Session.ID
			s.status = t.OnlineSessionJoined + ": " + msg.Session.ID
			s.ensureSessionReady(msg.Session)
		}
	case protocol.TypeMatchFound:
		msg, err := protocol.Decode[protocol.MatchFound](env)
		if err == nil {
			s.sessionID = msg.SessionID
			s.matchID = msg.MatchID
			s.status = t.OnlineMatchStarted + ": " + msg.MatchID
		}
	case protocol.TypeGameStart, protocol.TypeTurnStart, protocol.TypeStateUpdate:
		msg, err := protocol.Decode[protocol.StateUpdate](env)
		if err == nil {
			s.matchState = msg.State
		}
		s.status = t.OnlineMatchStarted
	case protocol.TypeShotResult:
		msg, err := protocol.Decode[protocol.ShotResult](env)
		if err == nil {
			s.matchState = msg.State
		}
	case protocol.TypeLeaderboard:
		msg, err := protocol.Decode[protocol.Leaderboard](env)
		if err == nil {
			s.leaders = msg.Entries
			s.status = t.OnlineLeaderboard + " " + onlineNowString()
		}
	case protocol.TypeError:
		msg, err := protocol.Decode[protocol.Error](env)
		if err == nil {
			s.status = msg.Code + ": " + msg.Message
		}
	}
}

func (s *onlineScene) ensureSessionReady(sess protocol.SessionSummary) {
	if sess.ID == "" || s.readySent[sess.ID] {
		return
	}
	s.readySent[sess.ID] = true
	s.client.Send(protocol.TypeReady, protocol.Ready{SessionID: sess.ID, Ready: true})
}

func (s *onlineScene) handleRoundsClick(p image.Point) bool {
	if p.In(onlineRoundsMinusRect()) {
		if s.rounds > 1 {
			s.rounds--
		}
		return true
	}
	if p.In(onlineRoundsPlusRect()) {
		if s.rounds < 99 {
			s.rounds++
		}
		return true
	}
	return false
}

func (s *onlineScene) drawRounds(screen *ebiten.Image) {
	t := texts()
	drawFrame(screen, onlineRoundsRect(), colornames.White, colornames.Black)
	drawTextFace(screen, t.PlayerSelectionRounds+": "+strconv.Itoa(s.rounds), dialogTextFace, onlineRoundsRect().Min.X+10, onlineRoundsRect().Min.Y+22, colornames.Black)
	drawButton(screen, onlineRoundsMinusRect(), "-")
	drawButton(screen, onlineRoundsPlusRect(), "+")
}

func (s *onlineScene) updateAutopilot() {
	if s.client == nil {
		return
	}
	if s.autoJoin && !s.autoQueued && s.client.id.PlayerID != "" {
		s.autoQueued = true
		s.client.Send(protocol.TypeQuickMatch, struct{}{})
	}
	if !s.autoPlay || s.matchID == "" || s.matchState.Status != gamecore.MatchInGame {
		return
	}
	if s.tick < s.autoFireAt || s.matchState.CurrentPlayerIndex < 0 || s.matchState.CurrentPlayerIndex >= len(s.matchState.Players) {
		return
	}
	current := s.matchState.Players[s.matchState.CurrentPlayerIndex]
	if current.ID != s.client.id.PlayerID {
		return
	}
	turnKey := s.matchState.MatchID + ":" + strconv.Itoa(s.matchState.Round) + ":" + strconv.Itoa(s.matchState.CurrentPlayerIndex) + ":" + strconv.Itoa(len(s.matchState.Terrain))
	if turnKey == s.autoTurnKey {
		return
	}
	s.autoTurnKey = turnKey
	s.autoFireAt = s.tick + 45 + s.rng.Intn(45)
	cmd := s.autoFireCommand()
	s.client.Send(protocol.TypeFire, cmd)
}

func (s *onlineScene) autoFireCommand() protocol.FireCommand {
	angle := 45.0 + s.rng.Float64()*90
	power := 55.0 + s.rng.Float64()*35
	shooter, target := s.autoShooterAndTarget()
	if shooter != nil && target != nil {
		if target.X >= shooter.X {
			angle = 0
		} else {
			angle = 180
		}
		dx := math.Abs(target.X-shooter.X) - math.Abs(float64(s.matchState.Wind))*0.8
		power = math.Max(25, math.Min(100, dx/7))
		if power > 82 {
			power = 82 + s.rng.Float64()*18
		}
	}
	return protocol.FireCommand{
		MatchID: s.matchID,
		Weapon:  "atom_bomb",
		Angle:   angle,
		Power:   power,
	}
}

func (s *onlineScene) autoShooterAndTarget() (*gamecore.TankState, *gamecore.TankState) {
	var shooter *gamecore.TankState
	for i := range s.matchState.Tanks {
		if s.matchState.Tanks[i].PlayerID == s.client.id.PlayerID {
			shooter = &s.matchState.Tanks[i]
			break
		}
	}
	if shooter == nil {
		return nil, nil
	}
	var best *gamecore.TankState
	bestDistance := math.MaxFloat64
	for i := range s.matchState.Tanks {
		tank := &s.matchState.Tanks[i]
		if !tank.Alive || tank.PlayerID == shooter.PlayerID {
			continue
		}
		distance := math.Abs(tank.X - shooter.X)
		if distance < bestDistance {
			best = tank
			bestDistance = distance
		}
	}
	return shooter, best
}

func onlineEnvBool(name string) bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(name))) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func (s *onlineScene) handleKeyboard() {
	s.inputRunes = ebiten.AppendInputChars(s.inputRunes[:0])
	if s.nameFocused {
		s.handleDisplayNameKeyboard()
		return
	}
	drainPlayerNameInputCommands()
	if inpututil.IsKeyJustPressed(ebiten.KeyV) && (ebiten.IsKeyPressed(ebiten.KeyControl) || ebiten.IsKeyPressed(ebiten.KeyMeta)) {
		s.pasteJoinInput()
		return
	}
	if len(s.inputRunes) > 0 {
		var b strings.Builder
		b.WriteString(s.joinInput)
		for _, r := range s.inputRunes {
			if len([]rune(b.String())) < onlineJoinInputMaxRunes {
				b.WriteRune(r)
			}
		}
		s.joinInput = b.String()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		runes := []rune(s.joinInput)
		if len(runes) > 0 {
			s.joinInput = string(runes[:len(runes)-1])
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeyKPEnter) {
		token := inviteTokenFromInput(s.joinInput)
		if token != "" {
			s.client.Send(protocol.TypeJoinSession, protocol.JoinSession{JoinCode: token, InviteToken: token})
		}
	}
}

func (s *onlineScene) pasteJoinInput() {
	value, err := readClipboardText()
	if err != nil {
		s.status = texts().OnlineClipboardUnavailable + ": " + err.Error()
		return
	}
	value = strings.TrimSpace(value)
	if value == "" {
		s.copyJoinInput()
		return
	}
	s.joinInput = truncateRunes(value, onlineJoinInputMaxRunes)
}

func (s *onlineScene) handleDisplayNameKeyboard() {
	for _, command := range drainPlayerNameInputCommands() {
		switch {
		case command.finish:
			s.finishDisplayNameInput()
		case command.backspace:
			s.displayName = trimLastRune(s.displayName)
		case command.replace:
			s.displayName = truncateRunes(command.text, 16)
		case command.text != "":
			s.displayName = truncateRunes(s.displayName+command.text, 16)
		}
	}
	if len(s.inputRunes) > 0 {
		var b strings.Builder
		b.WriteString(s.displayName)
		for _, r := range s.inputRunes {
			if len([]rune(b.String())) < 16 {
				b.WriteRune(r)
			}
		}
		s.displayName = b.String()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		s.displayName = trimLastRune(s.displayName)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeyKPEnter) {
		s.finishDisplayNameInput()
	}
}

func (s *onlineScene) finishDisplayNameInput() {
	if !s.nameFocused {
		return
	}
	s.nameFocused = false
	name := strings.TrimSpace(s.displayName)
	if name == "" {
		name = texts().GameDefaultPlayerName
	}
	s.displayName = truncateRunes(name, 16)
	s.client.id.DisplayName = s.displayName
	saveOnlineIdentity(s.client.id)
	s.client.Send(protocol.TypeHello, protocol.Hello{
		ProtocolVersion: protocol.ProtocolVersion,
		PlayerID:        s.client.id.PlayerID,
		PlayerToken:     s.client.id.PlayerToken,
		DisplayName:     s.displayName,
	})
}

func (s *onlineScene) syncOnlineTextInputActive() {
	setPlayerNameInputActive(s.nameFocused)
}

func (s *onlineScene) drawFocusedNameOverlay(screen *ebiten.Image) {
	if !showPlayerNameInputOverlay() || !s.nameFocused {
		return
	}
	r := image.Rect(184, 116, 840, 194)
	drawFrame(screen, r, color.RGBA{R: 255, G: 255, B: 245, A: 255}, color.RGBA{R: 40, G: 50, B: 60, A: 255})
	drawTextFace(screen, texts().OnlineDisplayName, dialogTextFace, r.Min.X+20, r.Min.Y+27, colornames.Black)
	value := s.displayName
	if s.tick/24%2 == 0 {
		value += "|"
	}
	drawTextFace(screen, value, overlayTextFace, r.Min.X+20, r.Min.Y+68, color.RGBA{R: 20, G: 25, B: 30, A: 255})
}

func (s *onlineScene) copyJoinInput() {
	value := strings.TrimSpace(s.joinInput)
	if value == "" {
		return
	}
	if err := copyTextToClipboard(value); err != nil {
		s.status = texts().OnlineClipboardUnavailable + ": " + err.Error()
		return
	}
	s.status = texts().OnlineCopiedCode
}

func (s *onlineScene) copyInviteURL() {
	if s.inviteURL == "" {
		return
	}
	if err := copyTextToClipboard(s.inviteURL); err != nil {
		s.status = texts().OnlineClipboardUnavailable + ": " + err.Error()
		return
	}
	s.status = texts().OnlineCopiedInviteLink
}

func (s *onlineScene) drawSessions(screen *ebiten.Image) {
	t := texts()
	drawTextFace(screen, t.OnlineOpenSessions, uiTextFace, 72, 392, colornames.White)
	if len(s.sessions) == 0 {
		drawTextFace(screen, t.OnlineNoSessions, dialogTextFace, 100, 424, colornames.Silver)
		return
	}
	for i, sess := range s.sessions {
		if i >= 5 {
			break
		}
		r := image.Rect(100, 410+i*42, 924, 444+i*42)
		drawFrame(screen, r, color.RGBA{R: 235, G: 240, B: 245, A: 255}, colornames.Black)
		line := sess.HostName + "    " + strconv.Itoa(sess.PlayerCount) + " / " + strconv.Itoa(sess.MaxPlayers) + "    " + t.PlayerSelectionRounds + ": " + strconv.Itoa(normalizedOnlineRounds(sess.Rounds)) + "    Rating " + strconv.Itoa(sess.AverageRating)
		drawTextFace(screen, line, dialogTextFace, r.Min.X+12, r.Min.Y+22, colornames.Black)
	}
}

func (s *onlineScene) drawLeaderboard(screen *ebiten.Image) {
	t := texts()
	drawTextFace(screen, t.OnlineLeaderboard, uiTextFace, 72, 612, colornames.White)
	if len(s.leaders) == 0 {
		drawTextFace(screen, t.OnlineNoLeaderboard, dialogTextFace, 100, 644, colornames.Silver)
		return
	}
	for i, entry := range s.leaders {
		if i >= 4 {
			break
		}
		line := strconv.Itoa(entry.Rank) + ". " + entry.DisplayName + "  Score " + strconv.Itoa(entry.Score) + "  Rating " + strconv.Itoa(entry.Rating) + "  W/L " + strconv.Itoa(entry.Wins) + "/" + strconv.Itoa(entry.Losses)
		drawTextFace(screen, line, dialogTextFace, 100, 644+i*20, colornames.White)
	}
}

func joinInputRect() image.Rectangle {
	return image.Rect(72, 278, 952, 308)
}

func onlineNameInputRect() image.Rectangle {
	return image.Rect(180, 84, 420, 118)
}

func onlineRoundsRect() image.Rectangle {
	return image.Rect(72, 124, 294, 150)
}

func onlineRoundsMinusRect() image.Rectangle {
	return image.Rect(306, 121, 338, 153)
}

func onlineRoundsPlusRect() image.Rectangle {
	return image.Rect(346, 121, 378, 153)
}

func inviteLinkRect() image.Rectangle {
	return image.Rect(72, 704, 952, 736)
}

func (s *onlineScene) back() error {
	s.nameFocused = false
	s.syncOnlineTextInputActive()
	if s.client != nil {
		s.client.Close()
	}
	return s.g.SetNewScene(NewPlayerSelectionScene)
}

func truncateRunes(value string, max int) string {
	runes := []rune(value)
	if len(runes) <= max {
		return value
	}
	return string(runes[:max])
}

func trimLastRune(value string) string {
	runes := []rune(value)
	if len(runes) == 0 {
		return value
	}
	return string(runes[:len(runes)-1])
}

func normalizedOnlineRounds(rounds int) int {
	if rounds < 1 {
		return 1
	}
	if rounds > 99 {
		return 99
	}
	return rounds
}
