package server

import "testing"

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
