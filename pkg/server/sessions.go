package server

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/hex"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/runzhammer/gamedemo/pkg/gamecore"
	"github.com/runzhammer/gamedemo/pkg/protocol"
)

type SessionType string
type SessionStatus string

const (
	SessionPublic      SessionType = "public"
	SessionPrivate     SessionType = "private"
	SessionMatchmaking SessionType = "matchmaking"

	SessionWaiting  SessionStatus = "waiting"
	SessionReady    SessionStatus = "ready"
	SessionInGame   SessionStatus = "in_game"
	SessionFinished SessionStatus = "finished"
)

type SessionPlayer struct {
	PlayerID    string
	DisplayName string
	Rating      int
	Ready       bool
	Client      *Client
}

type Session struct {
	ID          string
	Type        SessionType
	HostID      string
	Players     []SessionPlayer
	MaxPlayers  int
	CreatedAt   time.Time
	InviteToken string
	JoinCode    string
	InviteURL   string
	InviteUntil time.Time
	Status      SessionStatus
	Match       gamecore.MatchState
}

type Hub struct {
	mu       sync.Mutex
	cfg      Config
	store    *Store
	sessions map[string]*Session
	queue    []queuedPlayer
}

type queuedPlayer struct {
	Player   SessionPlayer
	QueuedAt time.Time
}

func NewHub(cfg Config, store *Store) *Hub {
	return &Hub{
		cfg:      cfg,
		store:    store,
		sessions: map[string]*Session{},
	}
}

func (h *Hub) CreateSession(kind SessionType, player SessionPlayer) (*Session, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	now := time.Now().UTC()
	s := &Session{
		ID:         randomID("ses", 12),
		Type:       kind,
		HostID:     player.PlayerID,
		Players:    []SessionPlayer{player},
		MaxPlayers: h.cfg.Sessions.MaxPlayers,
		CreatedAt:  now,
		Status:     SessionWaiting,
	}
	if kind == SessionPrivate {
		s.InviteToken = randomToken(h.cfg.Sessions.InviteTokenLength)
		s.JoinCode = strings.ToUpper(randomTokenBase32(h.cfg.Sessions.JoinCodeLength))
		s.InviteUntil = now.Add(h.cfg.Sessions.InviteTTL)
		s.InviteURL = strings.TrimRight(h.cfg.Server.PublicURL, "/") + "/join/" + url.PathEscape(s.InviteToken)
	}
	h.sessions[s.ID] = s
	return s, nil
}

func (h *Hub) JoinSession(req protocol.JoinSession, player SessionPlayer) (*Session, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	s := h.findSession(req)
	if s == nil {
		return nil, errProtocol("session_not_found", "session not found")
	}
	if s.Status != SessionWaiting && s.Status != SessionReady {
		return nil, errProtocol("session_closed", "session is not joinable")
	}
	if len(s.Players) >= s.MaxPlayers {
		return nil, errProtocol("session_full", "session is full")
	}
	for _, existing := range s.Players {
		if existing.PlayerID == player.PlayerID {
			return s, nil
		}
	}
	s.Players = append(s.Players, player)
	return s, nil
}

func (h *Hub) SetReady(sessionID, playerID string, ready bool) (*Session, bool, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	s := h.sessions[sessionID]
	if s == nil {
		return nil, false, errProtocol("session_not_found", "session not found")
	}
	for i := range s.Players {
		if s.Players[i].PlayerID == playerID {
			s.Players[i].Ready = ready
		}
	}
	allReady := len(s.Players) >= 2
	for _, p := range s.Players {
		allReady = allReady && p.Ready
	}
	if allReady && s.Status != SessionInGame {
		h.startMatchLocked(s)
		return s, true, nil
	}
	if allReady {
		s.Status = SessionReady
	}
	return s, false, nil
}

func (h *Hub) Leave(playerID string) []*Session {
	h.mu.Lock()
	defer h.mu.Unlock()
	var changed []*Session
	for id, s := range h.sessions {
		next := s.Players[:0]
		removed := false
		for _, p := range s.Players {
			if p.PlayerID == playerID {
				removed = true
				continue
			}
			next = append(next, p)
		}
		if !removed {
			continue
		}
		s.Players = next
		changed = append(changed, s)
		if len(s.Players) == 0 || s.HostID == playerID {
			delete(h.sessions, id)
			s.Status = SessionFinished
		}
	}
	return changed
}

func (h *Hub) Reconnect(matchID string, player SessionPlayer) (*Session, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, s := range h.sessions {
		if s.Match.MatchID != matchID {
			continue
		}
		for i := range s.Players {
			if s.Players[i].PlayerID == player.PlayerID {
				s.Players[i].Client = player.Client
				return s, nil
			}
		}
		return nil, errProtocol("not_in_match", "player is not in that match")
	}
	return nil, errProtocol("match_not_found", "match not found")
}

func (h *Hub) QuickMatch(player SessionPlayer) (*Session, bool, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if !h.cfg.Matchmaking.Enabled {
		return nil, false, errProtocol("matchmaking_disabled", "matchmaking is disabled")
	}
	now := time.Now().UTC()
	for i, queued := range h.queue {
		if queued.Player.PlayerID == player.PlayerID {
			continue
		}
		ratingRange := h.ratingRange(queued, now)
		diff := queued.Player.Rating - player.Rating
		if diff < 0 {
			diff = -diff
		}
		if diff <= ratingRange {
			h.queue = append(h.queue[:i], h.queue[i+1:]...)
			s := &Session{
				ID:         randomID("ses", 12),
				Type:       SessionMatchmaking,
				HostID:     queued.Player.PlayerID,
				Players:    []SessionPlayer{queued.Player, player},
				MaxPlayers: h.cfg.Sessions.MaxPlayers,
				CreatedAt:  now,
				Status:     SessionWaiting,
			}
			h.startMatchLocked(s)
			h.sessions[s.ID] = s
			return s, true, nil
		}
	}
	h.queue = append(h.queue, queuedPlayer{Player: player, QueuedAt: now})
	return nil, false, nil
}

func (h *Hub) CancelQuickMatch(playerID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	filtered := h.queue[:0]
	for _, queued := range h.queue {
		if queued.Player.PlayerID != playerID {
			filtered = append(filtered, queued)
		}
	}
	h.queue = filtered
}

func (h *Hub) PublicSessions(playerRating int) []protocol.SessionSummary {
	h.mu.Lock()
	defer h.mu.Unlock()
	out := make([]protocol.SessionSummary, 0, len(h.sessions))
	for _, s := range h.sessions {
		if s.Type != SessionPublic || (s.Status != SessionWaiting && s.Status != SessionReady) {
			continue
		}
		out = append(out, sessionSummary(s))
	}
	sortSessions(out, playerRating)
	return out
}

func (h *Hub) Fire(sessionID, playerID string, req protocol.FireCommand) (*Session, gamecore.ShotResult, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	s := h.sessions[sessionID]
	if s == nil {
		return nil, gamecore.ShotResult{}, errProtocol("session_not_found", "session not found")
	}
	result, err := gamecore.ApplyFire(&s.Match, gamecore.FireCommand{
		PlayerID: playerID,
		Weapon:   req.Weapon,
		Angle:    req.Angle,
		Power:    req.Power,
	})
	if err != nil {
		return nil, gamecore.ShotResult{}, errProtocol("invalid_fire", err.Error())
	}
	if s.Match.Status == gamecore.MatchFinished {
		s.Status = SessionFinished
	}
	return s, result, nil
}

func (h *Hub) CleanupExpired() {
	h.mu.Lock()
	defer h.mu.Unlock()
	now := time.Now().UTC()
	for id, s := range h.sessions {
		if now.Sub(s.CreatedAt) > h.cfg.Sessions.MaxSessionLifetime || (s.InviteUntil.After(time.Time{}) && now.After(s.InviteUntil)) {
			delete(h.sessions, id)
		}
	}
	filtered := h.queue[:0]
	for _, queued := range h.queue {
		if now.Sub(queued.QueuedAt) <= h.cfg.Matchmaking.QueueTimeout {
			filtered = append(filtered, queued)
		}
	}
	h.queue = filtered
}

func (h *Hub) startMatchLocked(s *Session) {
	players := make([]gamecore.Player, 0, len(s.Players))
	for _, p := range s.Players {
		players = append(players, gamecore.Player{ID: p.PlayerID, DisplayName: p.DisplayName, Rating: p.Rating})
	}
	s.Match = gamecore.NewEngine(time.Now().UnixNano()).NewMatch(randomID("mat", 12), players)
	s.Status = SessionInGame
	for i := range s.Players {
		s.Players[i].Ready = true
	}
}

func (h *Hub) findSession(req protocol.JoinSession) *Session {
	if req.SessionID != "" {
		return h.sessions[req.SessionID]
	}
	for _, s := range h.sessions {
		if s.Type != SessionPrivate {
			continue
		}
		if req.JoinCode != "" && strings.EqualFold(req.JoinCode, s.JoinCode) {
			return s
		}
		if req.InviteToken != "" && req.InviteToken == s.InviteToken && time.Now().UTC().Before(s.InviteUntil) {
			return s
		}
	}
	return nil
}

func (h *Hub) ratingRange(q queuedPlayer, now time.Time) int {
	elapsed := now.Sub(q.QueuedAt)
	steps := int(elapsed / h.cfg.Matchmaking.RatingRangeExpandInterval)
	value := h.cfg.Matchmaking.InitialRatingRange + steps*h.cfg.Matchmaking.RatingRangeStep
	if value > h.cfg.Matchmaking.MaximumRatingRange {
		return h.cfg.Matchmaking.MaximumRatingRange
	}
	return value
}

func sessionSummary(s *Session) protocol.SessionSummary {
	return protocol.SessionSummary{
		ID:            s.ID,
		Type:          string(s.Type),
		HostName:      s.hostName(),
		PlayerCount:   len(s.Players),
		MaxPlayers:    s.MaxPlayers,
		AverageRating: s.averageRating(),
		Status:        string(s.Status),
		CreatedAt:     s.CreatedAt,
	}
}

func (s *Session) hostName() string {
	for _, p := range s.Players {
		if p.PlayerID == s.HostID {
			return p.DisplayName
		}
	}
	return ""
}

func (s *Session) averageRating() int {
	if len(s.Players) == 0 {
		return 0
	}
	total := 0
	for _, p := range s.Players {
		total += p.Rating
	}
	return total / len(s.Players)
}

func randomID(prefix string, bytesLen int) string {
	return prefix + "_" + randomToken(bytesLen)
}

func randomToken(bytesLen int) string {
	if bytesLen <= 0 {
		bytesLen = 16
	}
	buf := make([]byte, bytesLen)
	if _, err := rand.Read(buf); err != nil {
		panic(err)
	}
	return hex.EncodeToString(buf)
}

func randomTokenBase32(length int) string {
	buf := make([]byte, length)
	if _, err := rand.Read(buf); err != nil {
		panic(err)
	}
	token := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(buf)
	if len(token) > length {
		return token[:length]
	}
	return token
}
