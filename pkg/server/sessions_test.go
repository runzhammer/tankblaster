package server

import (
	"context"
	"testing"

	"github.com/runzhammer/tankblaster/pkg/protocol"
)

func TestCreateSessionStoresNormalizedRounds(t *testing.T) {
	h := NewHub(DefaultConfig(), nil)

	sess, err := h.CreateSession(SessionPublic, SessionPlayer{PlayerID: "p1", DisplayName: "Player 1"}, 7)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}

	if got, want := sess.Rounds, 7; got != want {
		t.Fatalf("session rounds = %d, want %d", got, want)
	}
	if got, want := sessionSummary(sess).Rounds, 7; got != want {
		t.Fatalf("summary rounds = %d, want %d", got, want)
	}
}

func TestStartMatchCopiesSessionRoundsToMatchState(t *testing.T) {
	h := NewHub(DefaultConfig(), nil)
	sess, err := h.CreateSession(SessionPrivate, SessionPlayer{PlayerID: "p1", DisplayName: "Player 1"}, 123)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	sess.Players = append(sess.Players, SessionPlayer{PlayerID: "p2", DisplayName: "Player 2"})

	h.startMatchLocked(sess)

	if got, want := sess.Rounds, 99; got != want {
		t.Fatalf("session rounds = %d, want clamped %d", got, want)
	}
	if got, want := sess.Match.TotalRounds, 99; got != want {
		t.Fatalf("match total rounds = %d, want %d", got, want)
	}
}

func TestCompleteMatchRequiresWinningReporter(t *testing.T) {
	h := NewHub(DefaultConfig(), nil)
	sess, err := h.CreateSession(SessionPrivate, SessionPlayer{PlayerID: "p1", DisplayName: "Player 1"}, 3)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	sess.Players = append(sess.Players, SessionPlayer{PlayerID: "p2", DisplayName: "Player 2"})
	h.startMatchLocked(sess)

	_, _, _, _, _, err = h.CompleteMatch(sess.Match.MatchID, "p2", map[string]int{"p1": 9, "p2": 3})
	if err == nil {
		t.Fatal("CompleteMatch() error = nil, want not_winner")
	}
	if sess.Status == SessionFinished {
		t.Fatal("session was finished by losing reporter")
	}
}

func TestCompleteMatchFinishesWithScoreWinner(t *testing.T) {
	h := NewHub(DefaultConfig(), nil)
	sess, err := h.CreateSession(SessionPrivate, SessionPlayer{PlayerID: "p1", DisplayName: "Player 1"}, 3)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	sess.Players = append(sess.Players, SessionPlayer{PlayerID: "p2", DisplayName: "Player 2"})
	h.startMatchLocked(sess)

	gotSess, winnerID, loserID, winnerScore, loserScore, err := h.CompleteMatch(sess.Match.MatchID, "p2", map[string]int{"p1": 4, "p2": 10})
	if err != nil {
		t.Fatalf("CompleteMatch() error = %v", err)
	}
	if gotSess != sess {
		t.Fatal("CompleteMatch() returned a different session")
	}
	if winnerID != "p2" || loserID != "p1" || winnerScore != 10 || loserScore != 4 {
		t.Fatalf("result = winner %q loser %q scores %d:%d, want p2 p1 10:4", winnerID, loserID, winnerScore, loserScore)
	}
	if sess.Status != SessionFinished {
		t.Fatalf("session status = %s, want %s", sess.Status, SessionFinished)
	}
}

func TestCreateSessionRespectsMaxSessions(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Server.MaxSessions = 1
	h := NewHub(cfg, nil)

	if _, err := h.CreateSession(SessionPublic, SessionPlayer{PlayerID: "p1", DisplayName: "Player 1"}, 1); err != nil {
		t.Fatalf("first CreateSession() error = %v", err)
	}
	if _, err := h.CreateSession(SessionPublic, SessionPlayer{PlayerID: "p2", DisplayName: "Player 2"}, 1); err == nil {
		t.Fatal("second CreateSession() error = nil, want server_busy")
	}
}

func TestQuickMatchRespectsMaxQueueLength(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Server.MaxQueueLength = 1
	cfg.Matchmaking.InitialRatingRange = 0
	cfg.Matchmaking.RatingRangeStep = 0
	cfg.Matchmaking.MaximumRatingRange = 0
	h := NewHub(cfg, nil)

	if _, matched, err := h.QuickMatch(SessionPlayer{PlayerID: "p1", DisplayName: "Player 1", Rating: 1000}); err != nil || matched {
		t.Fatalf("first QuickMatch() matched=%t error=%v, want queued", matched, err)
	}
	if _, matched, err := h.QuickMatch(SessionPlayer{PlayerID: "p1", DisplayName: "Player 1", Rating: 1000}); err != nil || matched {
		t.Fatalf("duplicate QuickMatch() matched=%t error=%v, want queued no-op", matched, err)
	}
	if _, _, err := h.QuickMatch(SessionPlayer{PlayerID: "p2", DisplayName: "Player 2", Rating: 2000}); err == nil {
		t.Fatal("full queue QuickMatch() error = nil, want server_busy")
	}
}

func TestBroadcastAssignsSessionSequences(t *testing.T) {
	h := NewHub(DefaultConfig(), nil)
	srv := &Server{hub: h}
	client := &Client{send: make(chan protocol.Envelope, 4)}
	sess, err := h.CreateSession(SessionPrivate, SessionPlayer{PlayerID: "p1", DisplayName: "Player 1", Client: client}, 1)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	sess.Players = append(sess.Players, SessionPlayer{PlayerID: "p2", DisplayName: "Player 2"})
	h.startMatchLocked(sess)

	if err := srv.broadcast(sess, protocol.TypeOnlineGameCommand, protocol.OnlineGameCommand{MatchID: sess.Match.MatchID, Kind: "shop_continue"}); err != nil {
		t.Fatalf("broadcast first: %v", err)
	}
	if err := srv.broadcast(sess, protocol.TypeOnlineGameCommand, protocol.OnlineGameCommand{MatchID: sess.Match.MatchID, Kind: "turn"}); err != nil {
		t.Fatalf("broadcast second: %v", err)
	}

	first := <-client.send
	second := <-client.send
	if first.Sequence != 1 || second.Sequence != 2 {
		t.Fatalf("sequences = %d,%d, want 1,2", first.Sequence, second.Sequence)
	}
	if first.MatchID != sess.Match.MatchID || second.MatchID != sess.Match.MatchID {
		t.Fatalf("envelope match ids = %q,%q, want %q", first.MatchID, second.MatchID, sess.Match.MatchID)
	}
}

func TestSendSessionEventsAfterReplaysMissedEvents(t *testing.T) {
	h := NewHub(DefaultConfig(), nil)
	srv := &Server{hub: h}
	sess, err := h.CreateSession(SessionPrivate, SessionPlayer{PlayerID: "p1", DisplayName: "Player 1"}, 1)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	sess.Players = append(sess.Players, SessionPlayer{PlayerID: "p2", DisplayName: "Player 2"})
	h.startMatchLocked(sess)
	if err := srv.broadcast(sess, protocol.TypeOnlineGameCommand, protocol.OnlineGameCommand{MatchID: sess.Match.MatchID, Kind: "shop_continue"}); err != nil {
		t.Fatalf("broadcast first: %v", err)
	}
	if err := srv.broadcast(sess, protocol.TypeOnlineGameCommand, protocol.OnlineGameCommand{MatchID: sess.Match.MatchID, Kind: "turn"}); err != nil {
		t.Fatalf("broadcast second: %v", err)
	}

	catchUp := &Client{send: make(chan protocol.Envelope, 2)}
	sent, err := srv.sendSessionEventsAfter(sess, catchUp, 1)
	if err != nil {
		t.Fatalf("sendSessionEventsAfter() error = %v", err)
	}
	if !sent {
		t.Fatal("sendSessionEventsAfter() sent = false, want true")
	}
	env := <-catchUp.send
	if got, want := env.Sequence, int64(2); got != want {
		t.Fatalf("replayed sequence = %d, want %d", got, want)
	}
	cmd, err := protocol.Decode[protocol.OnlineGameCommand](env)
	if err != nil {
		t.Fatalf("decode replayed command: %v", err)
	}
	if got, want := cmd.Kind, "turn"; got != want {
		t.Fatalf("replayed command = %q, want %q", got, want)
	}
}

func TestAttachClientReplacesSessionClientAfterReconnect(t *testing.T) {
	h := NewHub(DefaultConfig(), nil)
	oldClient := &Client{send: make(chan protocol.Envelope, 1), done: make(chan struct{})}
	sess, err := h.CreateSession(SessionPrivate, SessionPlayer{PlayerID: "p1", DisplayName: "Player 1", Client: oldClient}, 1)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	newClient := &Client{send: make(chan protocol.Envelope, 1), done: make(chan struct{})}

	updated := h.AttachClient("p1", newClient)

	if got, want := len(updated), 1; got != want {
		t.Fatalf("updated sessions = %d, want %d", got, want)
	}
	if sess.Players[0].Client != newClient {
		t.Fatal("session player still points to old client after reconnect")
	}
}

func TestLobbyUpdateAndStartCreatesMatchFromSlots(t *testing.T) {
	h := NewHub(DefaultConfig(), nil)
	sess, err := h.CreateSession(SessionPrivate, SessionPlayer{PlayerID: "host", DisplayName: "Host"}, 3)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}

	sess, err = h.UpdateLobby(sess.ID, "host", []protocol.LobbySlot{
		{Index: 0, Kind: "human", OwnerID: "host", Name: "Host", Color: protocol.RGBA{A: 255}},
		{Index: 1, Kind: "computer", Name: "CPU", ComputerID: 2, Color: protocol.RGBA{A: 255}},
	}, 3, protocol.LobbyOptions{ProjectileReentry: 2, PalmCount: -1, CloudAggression: 35, QuickRoundStart: true})
	if err != nil {
		t.Fatalf("UpdateLobby() error = %v", err)
	}
	if got, want := len(sess.LobbySlots), 2; got != want {
		t.Fatalf("lobby slots = %d, want %d", got, want)
	}

	sess, err = h.StartLobbyGame(sess.ID, "host")
	if err != nil {
		t.Fatalf("StartLobbyGame() error = %v", err)
	}
	if got, want := len(sess.Match.Players), 2; got != want {
		t.Fatalf("match players = %d, want %d", got, want)
	}
	if got, want := sess.Match.Players[1].DisplayName, "CPU"; got != want {
		t.Fatalf("second player name = %q, want %q", got, want)
	}
	if got := sess.Match.CurrentPlayerIndex; got < 0 || got >= len(sess.Match.Players) {
		t.Fatalf("current player index = %d, want within 0..%d", got, len(sess.Match.Players)-1)
	}
	if got, want := sess.LobbyOptions.CloudAggression, 35; got != want {
		t.Fatalf("lobby cloud aggression = %d, want %d", got, want)
	}
}

func TestLobbyGuestCanOnlyUpdateOwnSlotDetails(t *testing.T) {
	h := NewHub(DefaultConfig(), nil)
	sess, err := h.CreateSession(SessionPrivate, SessionPlayer{PlayerID: "host", DisplayName: "Host"}, 3)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	sess.Players = append(sess.Players, SessionPlayer{PlayerID: "guest", DisplayName: "Guest"})
	sess.LobbySlots = []protocol.LobbySlot{
		{Index: 0, Kind: "human", OwnerID: "host", Name: "Host", Color: protocol.RGBA{R: 1, A: 255}},
		{Index: 1, Kind: "human", OwnerID: "guest", Name: "Guest", Color: protocol.RGBA{R: 2, A: 255}},
	}

	_, err = h.UpdateLobby(sess.ID, "guest", []protocol.LobbySlot{
		{Index: 0, Kind: "human", OwnerID: "guest", Name: "Hacked", Color: protocol.RGBA{R: 9, A: 255}},
		{Index: 1, Kind: "human", OwnerID: "guest", Name: "Guest 2", Color: protocol.RGBA{R: 3, A: 255}},
	}, 3, protocol.LobbyOptions{CloudAggression: 99})
	if err != nil {
		t.Fatalf("guest UpdateLobby() error = %v", err)
	}
	if got, want := sess.LobbySlots[0].Name, "Host"; got != want {
		t.Fatalf("host slot name = %q, want %q", got, want)
	}
	if got, want := sess.LobbySlots[1].Name, "Guest 2"; got != want {
		t.Fatalf("guest slot name = %q, want %q", got, want)
	}
	if got, want := sess.LobbyOptions.CloudAggression, 0; got != want {
		t.Fatalf("guest changed lobby options to %d, want unchanged %d", got, want)
	}
}

func TestLobbyAllowsEmptyHumanNamesWhileEditing(t *testing.T) {
	h := NewHub(DefaultConfig(), nil)
	sess, err := h.CreateSession(SessionPrivate, SessionPlayer{PlayerID: "host", DisplayName: "Host"}, 3)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}

	_, err = h.UpdateLobby(sess.ID, "host", []protocol.LobbySlot{
		{Index: 0, Kind: "human", OwnerID: "host", Name: "", Color: protocol.RGBA{A: 255}},
	}, 3, protocol.LobbyOptions{})
	if err != nil {
		t.Fatalf("host UpdateLobby() error = %v", err)
	}
	if got := sess.LobbySlots[0].Name; got != "" {
		t.Fatalf("host slot name = %q, want empty", got)
	}
}

func TestLobbyGuestCanClearOwnHumanName(t *testing.T) {
	h := NewHub(DefaultConfig(), nil)
	sess, err := h.CreateSession(SessionPrivate, SessionPlayer{PlayerID: "host", DisplayName: "Host"}, 3)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	sess.Players = append(sess.Players, SessionPlayer{PlayerID: "guest", DisplayName: "Guest"})
	sess.LobbySlots = []protocol.LobbySlot{
		{Index: 0, Kind: "human", OwnerID: "host", Name: "Host", Color: protocol.RGBA{R: 1, A: 255}},
		{Index: 1, Kind: "human", OwnerID: "guest", Name: "Guest", Color: protocol.RGBA{R: 2, A: 255}},
	}

	_, err = h.UpdateLobby(sess.ID, "guest", []protocol.LobbySlot{
		{Index: 1, Kind: "human", OwnerID: "guest", Name: "", Color: protocol.RGBA{R: 2, A: 255}},
	}, 3, protocol.LobbyOptions{})
	if err != nil {
		t.Fatalf("guest UpdateLobby() error = %v", err)
	}
	if got := sess.LobbySlots[1].Name; got != "" {
		t.Fatalf("guest slot name = %q, want empty", got)
	}
}

func TestLobbyGuestCanAddOwnHumanSlot(t *testing.T) {
	h := NewHub(DefaultConfig(), nil)
	sess, err := h.CreateSession(SessionPrivate, SessionPlayer{PlayerID: "host", DisplayName: "Host"}, 3)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	sess.Players = append(sess.Players, SessionPlayer{PlayerID: "guest", DisplayName: "Guest"})
	sess.LobbySlots = []protocol.LobbySlot{
		{Index: 0, Kind: "human", OwnerID: "host", Name: "Host", Color: protocol.RGBA{R: 1, A: 255}},
		{Index: 1, Kind: "human", OwnerID: "guest", Name: "Guest", Color: protocol.RGBA{R: 2, A: 255}},
	}

	_, err = h.UpdateLobby(sess.ID, "guest", []protocol.LobbySlot{
		{Index: 1, Kind: "human", OwnerID: "guest", Name: "Guest", Color: protocol.RGBA{R: 2, A: 255}},
		{Index: 2, Kind: "human", OwnerID: "guest", Name: "Guest 2", Color: protocol.RGBA{R: 3, A: 255}},
	}, 3, protocol.LobbyOptions{})
	if err != nil {
		t.Fatalf("guest UpdateLobby() error = %v", err)
	}
	if !lobbySlotExists(sess.LobbySlots, 2, "human", "guest", "Guest 2") {
		t.Fatalf("guest-added slot missing: %+v", sess.LobbySlots)
	}
}

func TestLobbyGuestCanAddComputerButCannotStealOccupiedSlot(t *testing.T) {
	h := NewHub(DefaultConfig(), nil)
	sess, err := h.CreateSession(SessionPrivate, SessionPlayer{PlayerID: "host", DisplayName: "Host"}, 3)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	sess.Players = append(sess.Players, SessionPlayer{PlayerID: "guest", DisplayName: "Guest"})
	sess.LobbySlots = []protocol.LobbySlot{
		{Index: 0, Kind: "human", OwnerID: "host", Name: "Host", Color: protocol.RGBA{R: 1, A: 255}},
		{Index: 1, Kind: "human", OwnerID: "guest", Name: "Guest", Color: protocol.RGBA{R: 2, A: 255}},
	}

	_, err = h.UpdateLobby(sess.ID, "guest", []protocol.LobbySlot{
		{Index: 0, Kind: "human", OwnerID: "guest", Name: "Stolen", Color: protocol.RGBA{R: 9, A: 255}},
		{Index: 1, Kind: "human", OwnerID: "guest", Name: "Guest", Color: protocol.RGBA{R: 2, A: 255}},
		{Index: 2, Kind: "computer", Name: "CPU", Color: protocol.RGBA{R: 3, A: 255}},
	}, 3, protocol.LobbyOptions{})
	if err != nil {
		t.Fatalf("guest UpdateLobby() error = %v", err)
	}
	if !lobbySlotExists(sess.LobbySlots, 0, "human", "host", "Host") {
		t.Fatalf("host slot was changed: %+v", sess.LobbySlots)
	}
	if !lobbySlotExists(sess.LobbySlots, 2, "computer", "guest", "CPU") {
		t.Fatalf("guest-added computer slot missing: %+v", sess.LobbySlots)
	}
}

func TestLobbyHostCannotRenameGuestOwnedSlot(t *testing.T) {
	h := NewHub(DefaultConfig(), nil)
	sess, err := h.CreateSession(SessionPrivate, SessionPlayer{PlayerID: "host", DisplayName: "Host"}, 3)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	sess.Players = append(sess.Players, SessionPlayer{PlayerID: "guest", DisplayName: "Guest"})
	sess.LobbySlots = []protocol.LobbySlot{
		{Index: 0, Kind: "human", OwnerID: "host", Name: "Host", Color: protocol.RGBA{R: 1, A: 255}},
		{Index: 1, Kind: "human", OwnerID: "guest", Name: "Guest", Color: protocol.RGBA{R: 2, A: 255}},
	}

	_, err = h.UpdateLobby(sess.ID, "host", []protocol.LobbySlot{
		{Index: 0, Kind: "human", OwnerID: "host", Name: "Host", Color: protocol.RGBA{R: 1, A: 255}},
		{Index: 1, Kind: "human", OwnerID: "guest", Name: "Renamed", Color: protocol.RGBA{R: 9, A: 255}},
	}, 3, protocol.LobbyOptions{})
	if err != nil {
		t.Fatalf("host UpdateLobby() error = %v", err)
	}
	if !lobbySlotExists(sess.LobbySlots, 1, "human", "guest", "Guest") {
		t.Fatalf("guest slot was renamed by host: %+v", sess.LobbySlots)
	}
}

func TestLobbyOwnerCanClearSlotAndOthersCanReuseIt(t *testing.T) {
	h := NewHub(DefaultConfig(), nil)
	sess, err := h.CreateSession(SessionPrivate, SessionPlayer{PlayerID: "host", DisplayName: "Host"}, 3)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	sess.Players = append(sess.Players, SessionPlayer{PlayerID: "guest", DisplayName: "Guest"})
	sess.LobbySlots = []protocol.LobbySlot{
		{Index: 0, Kind: "human", OwnerID: "host", Name: "Host", Color: protocol.RGBA{R: 1, A: 255}},
		{Index: 1, Kind: "computer", OwnerID: "guest", PlayerID: "guest:slot:1", Name: "Guest CPU", Color: protocol.RGBA{R: 2, A: 255}},
	}

	_, err = h.UpdateLobby(sess.ID, "guest", []protocol.LobbySlot{
		{Index: 1, Kind: "none", OwnerID: "guest"},
	}, 3, protocol.LobbyOptions{})
	if err != nil {
		t.Fatalf("guest clear UpdateLobby() error = %v", err)
	}
	if lobbySlotExists(sess.LobbySlots, 1, "computer", "guest", "Guest CPU") {
		t.Fatalf("guest-cleared slot still present: %+v", sess.LobbySlots)
	}

	_, err = h.UpdateLobby(sess.ID, "host", []protocol.LobbySlot{
		{Index: 0, Kind: "human", OwnerID: "host", Name: "Host", Color: protocol.RGBA{R: 1, A: 255}},
		{Index: 1, Kind: "computer", OwnerID: "host", Name: "Host CPU", Color: protocol.RGBA{R: 3, A: 255}},
	}, 3, protocol.LobbyOptions{})
	if err != nil {
		t.Fatalf("host reuse UpdateLobby() error = %v", err)
	}
	if !lobbySlotExists(sess.LobbySlots, 1, "computer", "host", "Host CPU") {
		t.Fatalf("freed slot was not reusable by host: %+v", sess.LobbySlots)
	}
}

func TestLobbyUpdateBroadcastsMergedSlotsToAllClients(t *testing.T) {
	h := NewHub(DefaultConfig(), nil)
	srv := &Server{hub: h}
	hostClient := &Client{send: make(chan protocol.Envelope, 4), done: make(chan struct{}), player: PlayerRecord{PlayerID: "host"}}
	guestClient := &Client{send: make(chan protocol.Envelope, 4), done: make(chan struct{}), player: PlayerRecord{PlayerID: "guest"}}
	sess, err := h.CreateSession(SessionPrivate, SessionPlayer{PlayerID: "host", DisplayName: "Host", Client: hostClient}, 3)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	sess.Players = append(sess.Players, SessionPlayer{PlayerID: "guest", DisplayName: "Guest", Client: guestClient})

	env, err := protocol.Wrap(protocol.TypeLobbyUpdate, protocol.LobbyUpdate{
		SessionID: sess.ID,
		Slots: []protocol.LobbySlot{
			{Index: 0, Kind: "human", OwnerID: "host", Name: "Host", Color: protocol.RGBA{R: 1, A: 255}},
			{Index: 1, Kind: "computer", OwnerID: "host", Name: "Host CPU", Color: protocol.RGBA{R: 2, A: 255}},
		},
		Rounds: 3,
	})
	if err != nil {
		t.Fatalf("wrap host update: %v", err)
	}
	if err := srv.handleEnvelope(context.Background(), hostClient, env); err != nil {
		t.Fatalf("handle host update: %v", err)
	}
	hostUpdate := mustReceiveLobbyUpdate(t, hostClient)
	guestUpdate := mustReceiveLobbyUpdate(t, guestClient)
	if !lobbySlotExists(hostUpdate.Slots, 1, "computer", "host", "Host CPU") {
		t.Fatalf("host broadcast missing host CPU slot: %+v", hostUpdate.Slots)
	}
	if !lobbySlotExists(guestUpdate.Slots, 1, "computer", "host", "Host CPU") {
		t.Fatalf("guest broadcast missing host CPU slot: %+v", guestUpdate.Slots)
	}
	if hostUpdate.Revision == 0 || guestUpdate.Revision != hostUpdate.Revision {
		t.Fatalf("first broadcast revisions host=%d guest=%d, want same non-zero revision", hostUpdate.Revision, guestUpdate.Revision)
	}
	firstRevision := hostUpdate.Revision

	env, err = protocol.Wrap(protocol.TypeLobbyUpdate, protocol.LobbyUpdate{
		SessionID: sess.ID,
		Slots: []protocol.LobbySlot{
			{Index: 2, Kind: "human", OwnerID: "guest", Name: "Guest 2", Color: protocol.RGBA{R: 3, A: 255}},
		},
	})
	if err != nil {
		t.Fatalf("wrap guest update: %v", err)
	}
	if err := srv.handleEnvelope(context.Background(), guestClient, env); err != nil {
		t.Fatalf("handle guest update: %v", err)
	}
	hostUpdate = mustReceiveLobbyUpdate(t, hostClient)
	guestUpdate = mustReceiveLobbyUpdate(t, guestClient)
	if !lobbySlotExists(hostUpdate.Slots, 2, "human", "guest", "Guest 2") {
		t.Fatalf("host broadcast missing guest slot: %+v", hostUpdate.Slots)
	}
	if !lobbySlotExists(guestUpdate.Slots, 2, "human", "guest", "Guest 2") {
		t.Fatalf("guest broadcast missing guest slot: %+v", guestUpdate.Slots)
	}
	if hostUpdate.Revision <= firstRevision || guestUpdate.Revision != hostUpdate.Revision {
		t.Fatalf("second broadcast revisions host=%d guest=%d first=%d, want same newer revision", hostUpdate.Revision, guestUpdate.Revision, firstRevision)
	}
}

func TestOnlineGameCommandKeepsControlledLobbySlotPlayerID(t *testing.T) {
	h := NewHub(DefaultConfig(), nil)
	srv := &Server{hub: h}
	hostClient := &Client{send: make(chan protocol.Envelope, 4), done: make(chan struct{}), player: PlayerRecord{PlayerID: "host"}}
	guestClient := &Client{send: make(chan protocol.Envelope, 4), done: make(chan struct{}), player: PlayerRecord{PlayerID: "guest"}}
	sess, err := h.CreateSession(SessionPrivate, SessionPlayer{PlayerID: "host", DisplayName: "Host", Client: hostClient}, 3)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	sess.Players = append(sess.Players, SessionPlayer{PlayerID: "guest", DisplayName: "Guest", Client: guestClient})
	sess.LobbySlots = []protocol.LobbySlot{
		{Index: 0, Kind: "human", OwnerID: "host", PlayerID: "host:slot:0", Name: "Host"},
		{Index: 1, Kind: "computer", OwnerID: "host", PlayerID: "host:slot:1", Name: "CPU"},
		{Index: 2, Kind: "human", OwnerID: "guest", PlayerID: "guest:slot:2", Name: "Guest"},
	}
	sess.Match.MatchID = "match-1"
	sess.Status = SessionInGame

	env, err := protocol.Wrap(protocol.TypeOnlineGameCommand, protocol.OnlineGameCommand{
		MatchID:     "match-1",
		PlayerID:    "host:slot:1",
		PlayerIndex: 1,
		Kind:        "fire",
	})
	if err != nil {
		t.Fatalf("wrap online command: %v", err)
	}
	if err := srv.handleEnvelope(context.Background(), hostClient, env); err != nil {
		t.Fatalf("handle online command: %v", err)
	}

	received := mustReceiveOnlineGameCommand(t, guestClient)
	if got, want := received.PlayerID, "host:slot:1"; got != want {
		t.Fatalf("broadcast player id = %q, want slot id %q", got, want)
	}
}

func TestOnlineGameCommandRejectsForeignLobbySlotPlayerID(t *testing.T) {
	h := NewHub(DefaultConfig(), nil)
	srv := &Server{hub: h}
	hostClient := &Client{send: make(chan protocol.Envelope, 4), done: make(chan struct{}), player: PlayerRecord{PlayerID: "host"}}
	guestClient := &Client{send: make(chan protocol.Envelope, 4), done: make(chan struct{}), player: PlayerRecord{PlayerID: "guest"}}
	sess, err := h.CreateSession(SessionPrivate, SessionPlayer{PlayerID: "host", DisplayName: "Host", Client: hostClient}, 3)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	sess.Players = append(sess.Players, SessionPlayer{PlayerID: "guest", DisplayName: "Guest", Client: guestClient})
	sess.LobbySlots = []protocol.LobbySlot{
		{Index: 0, Kind: "human", OwnerID: "host", PlayerID: "host:slot:0", Name: "Host"},
		{Index: 1, Kind: "human", OwnerID: "guest", PlayerID: "guest:slot:1", Name: "Guest"},
	}
	sess.Match.MatchID = "match-1"
	sess.Status = SessionInGame

	env, err := protocol.Wrap(protocol.TypeOnlineGameCommand, protocol.OnlineGameCommand{
		MatchID:     "match-1",
		PlayerID:    "guest:slot:1",
		PlayerIndex: 1,
		Kind:        "fire",
	})
	if err != nil {
		t.Fatalf("wrap online command: %v", err)
	}
	if err := srv.handleEnvelope(context.Background(), hostClient, env); err == nil {
		t.Fatal("handle online command error = nil, want foreign slot rejection")
	}
}

func mustReceiveLobbyUpdate(t *testing.T, c *Client) protocol.LobbyUpdate {
	t.Helper()
	select {
	case env := <-c.send:
		if env.Type != protocol.TypeLobbyUpdate {
			t.Fatalf("envelope type = %q, want %q", env.Type, protocol.TypeLobbyUpdate)
		}
		update, err := protocol.Decode[protocol.LobbyUpdate](env)
		if err != nil {
			t.Fatalf("decode lobby update: %v", err)
		}
		return update
	default:
		t.Fatal("no lobby update sent")
		return protocol.LobbyUpdate{}
	}
}

func mustReceiveOnlineGameCommand(t *testing.T, c *Client) protocol.OnlineGameCommand {
	t.Helper()
	select {
	case env := <-c.send:
		if env.Type != protocol.TypeOnlineGameCommand {
			t.Fatalf("envelope type = %q, want %q", env.Type, protocol.TypeOnlineGameCommand)
		}
		cmd, err := protocol.Decode[protocol.OnlineGameCommand](env)
		if err != nil {
			t.Fatalf("decode online game command: %v", err)
		}
		return cmd
	default:
		t.Fatal("no online game command sent")
		return protocol.OnlineGameCommand{}
	}
}

func lobbySlotExists(slots []protocol.LobbySlot, index int, kind, ownerID, name string) bool {
	for _, slot := range slots {
		if slot.Index == index && slot.Kind == kind && slot.OwnerID == ownerID && slot.Name == name {
			return true
		}
	}
	return false
}

func TestJoinSessionFailsWhenLobbySlotsAreFullOfComputers(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Sessions.MaxPlayers = 2
	h := NewHub(cfg, nil)
	sess, err := h.CreateSession(SessionPrivate, SessionPlayer{PlayerID: "host", DisplayName: "Host"}, 1)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	sess.LobbySlots = []protocol.LobbySlot{
		{Index: 0, Kind: "computer", Name: "CPU 1"},
		{Index: 1, Kind: "computer", Name: "CPU 2"},
	}

	_, err = h.JoinSession(protocol.JoinSession{SessionID: sess.ID}, SessionPlayer{PlayerID: "guest", DisplayName: "Guest"})
	if err == nil {
		t.Fatal("JoinSession() error = nil, want no free player slots")
	}
}

func TestLeaveRemovesGuestLobbySlotsWithoutClosingHostSession(t *testing.T) {
	h := NewHub(DefaultConfig(), nil)
	sess, err := h.CreateSession(SessionPrivate, SessionPlayer{PlayerID: "host", DisplayName: "Host"}, 1)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	sess.Players = append(sess.Players, SessionPlayer{PlayerID: "guest", DisplayName: "Guest"})
	sess.LobbySlots = []protocol.LobbySlot{
		{Index: 0, Kind: "human", OwnerID: "host", Name: "Host"},
		{Index: 1, Kind: "human", OwnerID: "guest", Name: "Guest"},
		{Index: 2, Kind: "computer", OwnerID: "guest", Name: "Guest CPU"},
	}

	changed := h.Leave("guest")

	if got, want := len(changed), 1; got != want {
		t.Fatalf("changed sessions = %d, want %d", got, want)
	}
	if sess.Status == SessionFinished {
		t.Fatal("host session was closed when guest left")
	}
	if got, want := len(sess.LobbySlots), 1; got != want {
		t.Fatalf("lobby slots = %d, want %d", got, want)
	}
	if got, want := sess.LobbySlots[0].OwnerID, "host"; got != want {
		t.Fatalf("remaining owner = %q, want %q", got, want)
	}
}
