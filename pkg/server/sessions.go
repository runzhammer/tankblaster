package server

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/binary"
	"encoding/hex"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/runzhammer/tankblaster/pkg/gamecore"
	"github.com/runzhammer/tankblaster/pkg/protocol"
)

const sessionEventHistoryLimit = 512

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
	ID            string
	Type          SessionType
	HostID        string
	Rounds        int
	Players       []SessionPlayer
	MaxPlayers    int
	CreatedAt     time.Time
	InviteToken   string
	JoinCode      string
	InviteURL     string
	InviteUntil   time.Time
	Status        SessionStatus
	Match         gamecore.MatchState
	EventSeq      int64
	Events        []protocol.Envelope
	LobbyRevision int64
	LobbySlots    []protocol.LobbySlot
	LobbyOptions  protocol.LobbyOptions
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

func (h *Hub) CreateSession(kind SessionType, player SessionPlayer, rounds int) (*Session, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if len(h.sessions) >= h.cfg.Server.MaxSessions {
		return nil, errProtocol("server_busy", "too many open sessions")
	}
	now := time.Now().UTC()
	rounds = normalizedRounds(rounds)
	s := &Session{
		ID:         randomID("ses", 12),
		Type:       kind,
		HostID:     player.PlayerID,
		Rounds:     rounds,
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
	for i := range s.Players {
		if s.Players[i].PlayerID == player.PlayerID {
			s.Players[i].Client = player.Client
			return s, nil
		}
	}
	if len(s.Players) >= s.MaxPlayers {
		return nil, errProtocol("session_full", "session is full")
	}
	s.Players = append(s.Players, player)
	if !ensureLobbySlotForPlayer(s, player) {
		s.Players = s.Players[:len(s.Players)-1]
		return nil, errProtocol("session_full", "no free player slots")
	}
	s.LobbyRevision++
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
		s.LobbySlots = removeLobbySlotsForOwner(s.LobbySlots, playerID)
		s.LobbyRevision++
		changed = append(changed, s)
		if len(s.Players) == 0 || s.HostID == playerID {
			delete(h.sessions, id)
			s.Status = SessionFinished
		}
	}
	return changed
}

func removeLobbySlotsForOwner(slots []protocol.LobbySlot, ownerID string) []protocol.LobbySlot {
	next := slots[:0]
	for _, slot := range slots {
		if slot.OwnerID == ownerID {
			continue
		}
		next = append(next, slot)
	}
	return next
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

func (h *Hub) AttachClient(playerID string, client *Client) []*Session {
	h.mu.Lock()
	defer h.mu.Unlock()
	updated := make([]*Session, 0)
	for _, s := range h.sessions {
		changed := false
		for i := range s.Players {
			if s.Players[i].PlayerID == playerID {
				s.Players[i].Client = client
				changed = true
			}
		}
		if changed {
			updated = append(updated, s)
		}
	}
	return updated
}

func (h *Hub) UpdateLobby(sessionID, playerID string, slots []protocol.LobbySlot, rounds int, options protocol.LobbyOptions) (*Session, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	s := h.sessions[sessionID]
	if s == nil {
		return nil, errProtocol("session_not_found", "session not found")
	}
	if s.HostID == playerID {
		s.Rounds = normalizedRounds(rounds)
		s.LobbyOptions = normalizedLobbyOptions(options)
	}
	s.LobbySlots = mergePlayerLobbySlots(s, playerID, slots)
	s.LobbyRevision++
	return s, nil
}

func (h *Hub) StartLobbyGame(sessionID, playerID string) (*Session, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	s := h.sessions[sessionID]
	if s == nil {
		return nil, errProtocol("session_not_found", "session not found")
	}
	if s.HostID != playerID {
		return nil, errProtocol("not_host", "only the session host can start the game")
	}
	players := lobbyGamePlayers(s)
	if len(players) < 2 {
		return nil, errProtocol("not_enough_players", "at least two players are required")
	}
	s.Match = gamecore.NewEngine(time.Now().UnixNano()).NewMatch(randomID("mat", 12), players)
	s.Match.TotalRounds = normalizedRounds(s.Rounds)
	if len(s.Match.Players) > 0 {
		s.Match.CurrentPlayerIndex = randomIndex(len(s.Match.Players))
	}
	s.Status = SessionInGame
	return s, nil
}

func (h *Hub) QuickMatch(player SessionPlayer) (*Session, bool, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if !h.cfg.Matchmaking.Enabled {
		return nil, false, errProtocol("matchmaking_disabled", "matchmaking is disabled")
	}
	if len(h.sessions) >= h.cfg.Server.MaxSessions {
		return nil, false, errProtocol("server_busy", "too many open sessions")
	}
	now := time.Now().UTC()
	for i, queued := range h.queue {
		if queued.Player.PlayerID == player.PlayerID {
			return nil, false, nil
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
				Rounds:     normalizedRounds(0),
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
	if len(h.queue) >= h.cfg.Server.MaxQueueLength {
		return nil, false, errProtocol("server_busy", "matchmaking queue is full")
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

func (h *Hub) CompleteMatch(matchID, reporterID string, scores map[string]int) (*Session, string, string, int, int, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	var sess *Session
	for _, s := range h.sessions {
		if s.Match.MatchID == matchID {
			sess = s
			break
		}
	}
	if sess == nil {
		return nil, "", "", 0, 0, errProtocol("match_not_found", "match not found")
	}
	if sess.Status == SessionFinished {
		return sess, "", "", 0, 0, nil
	}
	if len(sess.Players) != 2 {
		return nil, "", "", 0, 0, errProtocol("unsupported_match", "only two-player matches can be rated")
	}
	left := sess.Players[0].PlayerID
	right := sess.Players[1].PlayerID
	leftScore, leftOK := scores[left]
	rightScore, rightOK := scores[right]
	if !leftOK || !rightOK {
		return nil, "", "", 0, 0, errProtocol("bad_message", "match scores are incomplete")
	}
	if leftScore == rightScore {
		return nil, "", "", 0, 0, errProtocol("draw", "draws are not rated")
	}
	winnerID, loserID := left, right
	winnerScore, loserScore := leftScore, rightScore
	if rightScore > leftScore {
		winnerID, loserID = right, left
		winnerScore, loserScore = rightScore, leftScore
	}
	if reporterID != winnerID {
		return nil, "", "", 0, 0, errProtocol("not_winner", "only the winner can report the match result")
	}
	sess.Status = SessionFinished
	sess.Match.Status = gamecore.MatchFinished
	sess.Match.WinnerID = winnerID
	return sess, winnerID, loserID, winnerScore, loserScore, nil
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
	s.Match.TotalRounds = normalizedRounds(s.Rounds)
	s.Status = SessionInGame
	for i := range s.Players {
		s.Players[i].Ready = true
	}
}

func ensureLobbySlotForPlayer(s *Session, player SessionPlayer) bool {
	if s == nil || player.PlayerID == "" {
		return false
	}
	for _, slot := range s.LobbySlots {
		if slot.Kind == "human" && slot.OwnerID == player.PlayerID {
			return true
		}
	}
	for i := 0; i < s.MaxPlayers; i++ {
		occupied := false
		for _, slot := range s.LobbySlots {
			if slot.Index == i && slot.Kind != "" && slot.Kind != "none" {
				occupied = true
				break
			}
		}
		if occupied {
			continue
		}
		s.LobbySlots = append(s.LobbySlots, protocol.LobbySlot{
			Index:    i,
			Kind:     "human",
			OwnerID:  player.PlayerID,
			PlayerID: lobbySlotPlayerID(player.PlayerID, i),
			Name:     player.DisplayName,
			Color:    defaultLobbyColor(i),
		})
		return true
	}
	return false
}

func normalizedLobbyOptions(options protocol.LobbyOptions) protocol.LobbyOptions {
	options.ProjectileReentry = clampInt(options.ProjectileReentry, 0, 2)
	options.PalmCount = clampInt(options.PalmCount, -1, 2)
	options.CloudAggression = clampInt(options.CloudAggression, 0, 100)
	return options
}

func clampInt(value, minValue, maxValue int) int {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}

func mergePlayerLobbySlots(s *Session, playerID string, slots []protocol.LobbySlot) []protocol.LobbySlot {
	incomingByIndex := map[int]protocol.LobbySlot{}
	for _, slot := range slots {
		if slot.Index < 0 || slot.Index >= s.MaxPlayers || (slot.Kind != "human" && slot.Kind != "computer" && slot.Kind != "none") {
			continue
		}
		if slot.OwnerID != "" && slot.OwnerID != playerID {
			continue
		}
		slot.OwnerID = playerID
		if slot.Kind == "none" {
			slot.PlayerID = ""
		} else {
			slot.PlayerID = lobbySlotPlayerID(playerID, slot.Index)
		}
		incomingByIndex[slot.Index] = slot
	}

	next := make([]protocol.LobbySlot, 0, len(s.LobbySlots)+len(incomingByIndex))
	occupied := map[int]bool{}
	for _, slot := range s.LobbySlots {
		if slot.OwnerID == playerID {
			if incoming, ok := incomingByIndex[slot.Index]; ok {
				if incoming.Kind != "none" {
					next = append(next, incoming)
					occupied[slot.Index] = true
				}
				delete(incomingByIndex, slot.Index)
			}
			continue
		}
		next = append(next, slot)
		occupied[slot.Index] = true
	}
	for _, slot := range incomingByIndex {
		if slot.Kind == "none" {
			continue
		}
		if occupied[slot.Index] {
			continue
		}
		next = append(next, slot)
		occupied[slot.Index] = true
	}
	return next
}

func lobbyGamePlayers(s *Session) []gamecore.Player {
	slots := append([]protocol.LobbySlot(nil), s.LobbySlots...)
	sort.Slice(slots, func(i, j int) bool {
		return slots[i].Index < slots[j].Index
	})
	players := make([]gamecore.Player, 0, len(slots))
	for _, slot := range slots {
		if slot.Kind != "human" && slot.Kind != "computer" {
			continue
		}
		name := slot.Name
		if name == "" {
			name = "Player"
		}
		players = append(players, gamecore.Player{ID: slot.PlayerID, DisplayName: name})
	}
	return players
}

func lobbySlotPlayerID(ownerID string, index int) string {
	return ownerID + ":slot:" + strconv.Itoa(index)
}

func defaultLobbyColor(index int) protocol.RGBA {
	colors := []protocol.RGBA{
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
	return colors[index%len(colors)]
}

func controlledLobbyPlayerIDs(s *Session, playerID string) []string {
	ids := make([]string, 0)
	for _, slot := range s.LobbySlots {
		if (slot.Kind == "human" || slot.Kind == "computer") && slot.OwnerID == playerID && slot.PlayerID != "" {
			ids = append(ids, slot.PlayerID)
		}
	}
	return ids
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
		Rounds:        normalizedRounds(s.Rounds),
		PlayerCount:   len(s.Players),
		MaxPlayers:    s.MaxPlayers,
		AverageRating: s.averageRating(),
		Status:        string(s.Status),
		CreatedAt:     s.CreatedAt,
	}
}

func normalizedRounds(rounds int) int {
	if rounds < 1 {
		return 1
	}
	if rounds > 99 {
		return 99
	}
	return rounds
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

func randomIndex(count int) int {
	if count <= 1 {
		return 0
	}
	var buf [8]byte
	if _, err := rand.Read(buf[:]); err == nil {
		return int(binary.LittleEndian.Uint64(buf[:]) % uint64(count))
	}
	return int(time.Now().UnixNano() % int64(count))
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
