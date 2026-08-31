package tankblaster

import (
	"image"
	"image/color"
	"strconv"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/runzhammer/gamedemo/pkg/core"
	"github.com/runzhammer/gamedemo/pkg/protocol"
	"golang.org/x/image/colornames"
)

type onlineScene struct {
	g           *GameLoop
	client      *onlineClient
	displayName string
	joinInput   string
	status      string
	sessionID   string
	matchID     string
	inviteURL   string
	sessions    []protocol.SessionSummary
	leaders     []protocol.LeaderboardEntry
	inputRunes  []rune
}

func NewOnlineScene(game *GameLoop) (core.Scene, error) {
	id := loadOnlineIdentity()
	if id.DisplayName == "" {
		id.DisplayName = texts().GameDefaultPlayerName
	}
	s := &onlineScene{
		g:           game,
		displayName: id.DisplayName,
		status:      texts().OnlineConnecting,
	}
	s.client = newOnlineClient(s.displayName)
	return s, nil
}

func (s *onlineScene) Update() error {
	s.consumeNetwork()
	s.handleKeyboard()
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return s.back()
	}
	if !primaryPointerJustPressed() {
		return nil
	}
	x, y := primaryPointerPosition()
	p := image.Pt(x, y)
	if p.In(joinInputRect()) {
		s.copyJoinInput()
		return nil
	}
	if s.inviteURL != "" && p.In(inviteLinkRect()) {
		s.copyInviteURL()
		return nil
	}
	for _, b := range s.buttons() {
		if p.In(b.rect) {
			return b.action()
		}
	}
	for i, sess := range s.sessions {
		r := image.Rect(100, 410+i*42, 924, 444+i*42)
		if p.In(r) {
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
	drawTextFace(screen, t.OnlineDisplayName+": "+s.displayName, uiTextFace, 72, 106, colornames.Lightblue)
	drawTextFace(screen, core.Config().Online.ServerURL, dialogTextFace, 72, 130, colornames.Silver)

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
			s.client.Send(protocol.TypeCreatePublicSession, protocol.CreateSession{DisplayName: s.displayName})
			return nil
		}},
		{t.OnlineCreatePrivateSession, image.Rect(502, 232, 746, 264), func() error {
			s.client.Send(protocol.TypeCreatePrivateSession, protocol.CreateSession{DisplayName: s.displayName})
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
		}
	case protocol.TypeMatchmakingQueued:
		s.status = t.OnlineQueued
	case protocol.TypeSessionCreated:
		msg, err := protocol.Decode[protocol.SessionCreated](env)
		if err == nil {
			s.sessionID = msg.Session.ID
			s.status = t.OnlineSessionCreated + ": " + msg.Session.ID
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
		}
	case protocol.TypeMatchFound:
		msg, err := protocol.Decode[protocol.MatchFound](env)
		if err == nil {
			s.sessionID = msg.SessionID
			s.matchID = msg.MatchID
			s.status = t.OnlineMatchStarted + ": " + msg.MatchID
		}
	case protocol.TypeGameStart, protocol.TypeTurnStart, protocol.TypeStateUpdate:
		s.status = t.OnlineMatchStarted
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

func (s *onlineScene) handleKeyboard() {
	s.inputRunes = ebiten.AppendInputChars(s.inputRunes[:0])
	if len(s.inputRunes) > 0 {
		var b strings.Builder
		b.WriteString(s.joinInput)
		for _, r := range s.inputRunes {
			if len([]rune(b.String())) < 96 {
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
		line := sess.HostName + "    " + strconv.Itoa(sess.PlayerCount) + " / " + strconv.Itoa(sess.MaxPlayers) + "    Rating " + strconv.Itoa(sess.AverageRating)
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

func inviteLinkRect() image.Rectangle {
	return image.Rect(72, 704, 952, 736)
}

func (s *onlineScene) back() error {
	if s.client != nil {
		s.client.Close()
	}
	return s.g.SetNewScene(NewPlayerSelectionScene)
}
