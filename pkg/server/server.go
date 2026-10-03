package server

import (
	"context"
	"errors"
	"html"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/runzhammer/tankblaster/pkg/protocol"
)

type Server struct {
	cfg       Config
	store     *Store
	hub       *Hub
	connSlots chan struct{}
}

type protocolError struct {
	Code    string
	Message string
}

func errProtocol(code, message string) protocolError {
	return protocolError{Code: code, Message: message}
}

func (e protocolError) Error() string {
	return e.Message
}

type Client struct {
	player      PlayerRecord
	conn        *websocket.Conn
	send        chan protocol.Envelope
	done        chan struct{}
	sendTimeout time.Duration
}

func New(cfg Config, store *Store) *Server {
	return &Server{
		cfg:       cfg,
		store:     store,
		hub:       NewHub(cfg, store),
		connSlots: make(chan struct{}, cfg.Server.MaxConnections),
	}
}

func (s *Server) Run(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/game", s.handleGame)
	mux.HandleFunc("/join/", s.handleJoinPage)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	})
	srv := &http.Server{
		Addr:              s.cfg.Server.Address,
		Handler:           mux,
		ReadHeaderTimeout: s.cfg.Server.ReadHeaderTimeout,
		ReadTimeout:       s.cfg.Server.ReadTimeout,
		WriteTimeout:      s.cfg.Server.WriteTimeout,
		IdleTimeout:       s.cfg.Server.IdleTimeout,
		MaxHeaderBytes:    8 * 1024,
	}
	go s.cleanupLoop(ctx)
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()
	log.Printf("tankblaster server listening on %s", s.cfg.Server.Address)
	err := srv.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (s *Server) handleGame(w http.ResponseWriter, r *http.Request) {
	if !s.acquireConnectionSlot() {
		http.Error(w, "too many connections", http.StatusServiceUnavailable)
		return
	}
	defer s.releaseConnectionSlot()

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: s.cfg.Server.AllowedOrigins,
	})
	if err != nil {
		return
	}
	conn.SetReadLimit(s.cfg.Server.WebSocketReadLimit)
	client := &Client{conn: conn, send: make(chan protocol.Envelope, 32), done: make(chan struct{}), sendTimeout: s.cfg.Server.ErrorChannelTimeout}
	ctx := r.Context()
	go client.writeLoop(ctx)
	defer func() {
		close(client.done)
		_ = conn.Close(websocket.StatusNormalClosure, "")
		if client.player.PlayerID != "" {
			s.hub.CancelQuickMatch(client.player.PlayerID)
		}
	}()

	for {
		var env protocol.Envelope
		if err := wsjson.Read(ctx, conn, &env); err != nil {
			return
		}
		if err := s.handleEnvelope(ctx, client, env); err != nil {
			client.enqueueError(err)
		}
	}
}

func (s *Server) acquireConnectionSlot() bool {
	if s == nil || s.connSlots == nil {
		return true
	}
	select {
	case s.connSlots <- struct{}{}:
		return true
	default:
		return false
	}
}

func (s *Server) releaseConnectionSlot() {
	if s == nil || s.connSlots == nil {
		return
	}
	select {
	case <-s.connSlots:
	default:
	}
}

func (s *Server) handleEnvelope(ctx context.Context, c *Client, env protocol.Envelope) error {
	switch env.Type {
	case protocol.TypeHello:
		req, err := protocol.Decode[protocol.Hello](env)
		if err != nil {
			return errProtocol("bad_message", "invalid hello")
		}
		if req.ProtocolVersion != protocol.ProtocolVersion {
			return errProtocol("protocol_version", "incompatible protocol version")
		}
		name := validateName(req.DisplayName)
		player, created, err := s.store.AuthenticateOrCreate(ctx, req.PlayerID, req.PlayerToken, name)
		if err != nil {
			return err
		}
		c.player = player
		token := ""
		if created {
			token = player.Token
		}
		return c.sendMessage(protocol.TypeHelloAck, protocol.HelloAck{
			PlayerID: player.PlayerID, PlayerToken: token, DisplayName: player.DisplayName,
			Rating: player.Rating, Score: player.Score,
		})
	case protocol.TypeQuickMatch:
		if err := c.requireAuth(); err != nil {
			return err
		}
		sess, matched, err := s.hub.QuickMatch(c.sessionPlayer())
		if err != nil {
			return err
		}
		if !matched {
			return c.sendMessage(protocol.TypeMatchmakingQueued, struct{}{})
		}
		return s.broadcastSessionStart(sess)
	case protocol.TypeCancelQuickMatch:
		if err := c.requireAuth(); err != nil {
			return err
		}
		s.hub.CancelQuickMatch(c.player.PlayerID)
		return c.sendMessage(protocol.TypeMatchmakingCancelled, struct{}{})
	case protocol.TypeCreatePublicSession:
		if err := c.requireAuth(); err != nil {
			return err
		}
		req, err := protocol.Decode[protocol.CreateSession](env)
		if err != nil {
			return errProtocol("bad_message", "invalid create session request")
		}
		sess, err := s.hub.CreateSession(SessionPublic, c.sessionPlayer(), req.Rounds)
		if err != nil {
			return err
		}
		return c.sendMessage(protocol.TypeSessionCreated, protocol.SessionCreated{Session: sessionSummary(sess)})
	case protocol.TypeCreatePrivateSession:
		if err := c.requireAuth(); err != nil {
			return err
		}
		req, err := protocol.Decode[protocol.CreateSession](env)
		if err != nil {
			return errProtocol("bad_message", "invalid create session request")
		}
		sess, err := s.hub.CreateSession(SessionPrivate, c.sessionPlayer(), req.Rounds)
		if err != nil {
			return err
		}
		if err := c.sendMessage(protocol.TypeSessionCreated, protocol.SessionCreated{Session: sessionSummary(sess)}); err != nil {
			return err
		}
		return c.sendMessage(protocol.TypeInviteCreated, protocol.InviteCreated{
			SessionID: sess.ID, JoinCode: sess.JoinCode, InviteURL: sess.InviteURL, ExpiresAt: sess.InviteUntil,
		})
	case protocol.TypeListSessions:
		if err := c.requireAuth(); err != nil {
			return err
		}
		return c.sendMessage(protocol.TypeSessionList, protocol.SessionList{Sessions: s.hub.PublicSessions(c.player.Rating)})
	case protocol.TypeJoinSession:
		if err := c.requireAuth(); err != nil {
			return err
		}
		req, err := protocol.Decode[protocol.JoinSession](env)
		if err != nil {
			return errProtocol("bad_message", "invalid join request")
		}
		sess, err := s.hub.JoinSession(req, c.sessionPlayer())
		if err != nil {
			return err
		}
		return s.broadcast(sess, protocol.TypeSessionJoined, protocol.SessionJoined{Session: sessionSummary(sess)})
	case protocol.TypeLeaveSession:
		if err := c.requireAuth(); err != nil {
			return err
		}
		sessions := s.hub.Leave(c.player.PlayerID)
		for _, sess := range sessions {
			_ = s.broadcast(sess, protocol.TypeSessionClosed, protocol.SessionJoined{Session: sessionSummary(sess)})
		}
		return nil
	case protocol.TypeReady:
		if err := c.requireAuth(); err != nil {
			return err
		}
		req, err := protocol.Decode[protocol.Ready](env)
		if err != nil {
			return errProtocol("bad_message", "invalid ready request")
		}
		sess, started, err := s.hub.SetReady(req.SessionID, c.player.PlayerID, req.Ready)
		if err != nil {
			return err
		}
		if started {
			return s.broadcastSessionStart(sess)
		}
		return s.broadcast(sess, protocol.TypeSessionUpdated, protocol.SessionJoined{Session: sessionSummary(sess)})
	case protocol.TypeReconnect:
		if err := c.requireAuth(); err != nil {
			return err
		}
		req, err := protocol.Decode[protocol.Reconnect](env)
		if err != nil || req.MatchID == "" {
			return errProtocol("bad_message", "invalid reconnect request")
		}
		sess, err := s.hub.Reconnect(req.MatchID, c.sessionPlayer())
		if err != nil {
			return err
		}
		return c.sendMessage(protocol.TypeStateUpdate, protocol.StateUpdate{State: sess.Match})
	case protocol.TypeFire:
		if err := c.requireAuth(); err != nil {
			return err
		}
		req, err := protocol.Decode[protocol.FireCommand](env)
		if err != nil {
			return errProtocol("bad_message", "invalid fire request")
		}
		sess, result, err := s.hub.Fire(findSessionForMatch(s.hub, req.MatchID), c.player.PlayerID, req)
		if err != nil {
			return err
		}
		if err := s.broadcast(sess, protocol.TypeShotResult, protocol.ShotResult{Result: result, State: sess.Match}); err != nil {
			return err
		}
		if sess.Status == SessionFinished && sess.Match.WinnerID != "" && len(sess.Players) == 2 {
			loser := sess.Players[0].PlayerID
			if loser == sess.Match.WinnerID {
				loser = sess.Players[1].PlayerID
			}
			if err := s.store.RecordWin(ctx, sess.Match.WinnerID, loser); err != nil {
				log.Printf("record match result: %v", err)
			}
			return s.broadcast(sess, protocol.TypeGameOver, protocol.StateUpdate{State: sess.Match})
		}
		return nil
	case protocol.TypeOnlineGameCommand:
		if err := c.requireAuth(); err != nil {
			return err
		}
		req, err := protocol.Decode[protocol.OnlineGameCommand](env)
		if err != nil || req.MatchID == "" || req.Kind == "" {
			return errProtocol("bad_message", "invalid online game command")
		}
		sess := findSessionForMatchState(s.hub, req.MatchID)
		if sess == nil {
			return errProtocol("match_not_found", "match not found")
		}
		if !sessionHasPlayer(sess, c.player.PlayerID) {
			return errProtocol("not_in_match", "player is not in that match")
		}
		req.PlayerID = c.player.PlayerID
		return s.broadcast(sess, protocol.TypeOnlineGameCommand, req)
	case protocol.TypeMatchComplete:
		if err := c.requireAuth(); err != nil {
			return err
		}
		req, err := protocol.Decode[protocol.MatchComplete](env)
		if err != nil || req.MatchID == "" {
			return errProtocol("bad_message", "invalid match result")
		}
		sess, winnerID, loserID, winnerScore, loserScore, err := s.hub.CompleteMatch(req.MatchID, c.player.PlayerID, req.Scores)
		if err != nil {
			return err
		}
		if winnerID != "" && s.store != nil {
			if err := s.store.RecordMatchResult(ctx, winnerID, loserID, winnerScore, loserScore, normalizedRounds(sess.Match.TotalRounds)); err != nil {
				log.Printf("record match result: %v", err)
			}
		}
		return s.broadcast(sess, protocol.TypeGameOver, protocol.StateUpdate{State: sess.Match})
	case protocol.TypeGetLeaderboard:
		req, _ := protocol.Decode[protocol.LeaderboardRequest](env)
		entries, err := s.store.Leaderboard(ctx, req.Limit)
		if err != nil {
			return err
		}
		return c.sendMessage(protocol.TypeLeaderboard, protocol.Leaderboard{Entries: entries})
	case protocol.TypePing:
		return c.sendMessage(protocol.TypePong, struct{}{})
	default:
		return errProtocol("unknown_message", "unknown message type")
	}
}

func (s *Server) broadcastSessionStart(sess *Session) error {
	if err := s.broadcast(sess, protocol.TypeMatchFound, protocol.MatchFound{SessionID: sess.ID, MatchID: sess.Match.MatchID}); err != nil {
		return err
	}
	if err := s.broadcast(sess, protocol.TypeGameStart, protocol.StateUpdate{State: sess.Match}); err != nil {
		return err
	}
	return s.broadcast(sess, protocol.TypeTurnStart, protocol.StateUpdate{State: sess.Match})
}

func (s *Server) broadcast(sess *Session, typ protocol.MessageType, payload any) error {
	env, err := protocol.Wrap(typ, payload)
	if err != nil {
		return err
	}
	for _, p := range sess.Players {
		if p.Client == nil {
			continue
		}
		select {
		case p.Client.send <- env:
		case <-p.Client.done:
		default:
		}
	}
	return nil
}

func (c *Client) writeLoop(ctx context.Context) {
	for {
		select {
		case <-c.done:
			return
		case <-ctx.Done():
			return
		case env, ok := <-c.send:
			if !ok {
				return
			}
			if err := wsjson.Write(ctx, c.conn, env); err != nil {
				return
			}
		}
	}
}

func (c *Client) sendMessage(typ protocol.MessageType, payload any) error {
	env, err := protocol.Wrap(typ, payload)
	if err != nil {
		return err
	}
	if c.sendTimeout <= 0 {
		c.sendTimeout = 250 * time.Millisecond
	}
	select {
	case c.send <- env:
	case <-c.done:
		return errProtocol("client_closed", "client connection is closed")
	case <-time.After(c.sendTimeout):
		return errProtocol("client_slow", "client is not reading messages")
	}
	return nil
}

func (c *Client) enqueueError(err error) {
	code := "server_error"
	message := err.Error()
	var pe protocolError
	if errors.As(err, &pe) {
		code = pe.Code
		message = pe.Message
	}
	_ = c.sendMessage(protocol.TypeError, protocol.Error{Code: code, Message: message})
}

func (c *Client) requireAuth() error {
	if c.player.PlayerID == "" {
		return errProtocol("unauthenticated", "send hello first")
	}
	return nil
}

func (c *Client) sessionPlayer() SessionPlayer {
	return SessionPlayer{
		PlayerID:    c.player.PlayerID,
		DisplayName: c.player.DisplayName,
		Rating:      c.player.Rating,
		Client:      c,
	}
}

func validateName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "Player"
	}
	if len([]rune(name)) > 16 {
		return string([]rune(name)[:16])
	}
	return name
}

func findSessionForMatch(h *Hub, matchID string) string {
	if sess := findSessionForMatchState(h, matchID); sess != nil {
		return sess.ID
	}
	return ""
}

func findSessionForMatchState(h *Hub, matchID string) *Session {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, sess := range h.sessions {
		if sess.Match.MatchID == matchID {
			return sess
		}
	}
	return nil
}

func sessionHasPlayer(sess *Session, playerID string) bool {
	for _, player := range sess.Players {
		if player.PlayerID == playerID {
			return true
		}
	}
	return false
}

func (s *Server) cleanupLoop(ctx context.Context) {
	ticker := time.NewTicker(s.cfg.Sessions.CleanupInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.hub.CleanupExpired()
		}
	}
}

func (s *Server) handleJoinPage(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimPrefix(r.URL.Path, "/join/")
	tokenHTML := html.EscapeString(token)
	tokenURL := html.EscapeString(url.PathEscape(token))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(`<!doctype html><html><head><meta charset="utf-8"><title>Tank Blaster Invite</title></head><body><h1>Tank Blaster</h1><p>Open Tank Blaster and paste this invite token in Join Session.</p><pre>` + tokenHTML + `</pre><p><a href="tankblaster://join/` + tokenURL + `">Open in Tank Blaster</a></p></body></html>`))
}
