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
