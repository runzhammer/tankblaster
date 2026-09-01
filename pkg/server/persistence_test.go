package server

import (
	"context"
	"path/filepath"
	"testing"
)

func TestMatchEloScalesWithRoundsAndScoreMargin(t *testing.T) {
	shortWinner, shortLoser := matchElo(1000, 1000, 32, 0, 1, 5, 4)
	longWinner, longLoser := matchElo(1000, 1000, 32, 0, 9, 30, 0)

	shortDelta := shortWinner - 1000
	longDelta := longWinner - 1000
	if longDelta <= shortDelta {
		t.Fatalf("long match delta = %d, want > short match delta %d", longDelta, shortDelta)
	}
	if got, want := 1000-shortLoser, shortDelta; got != want {
		t.Fatalf("short loser delta = %d, want matching winner delta %d", got, want)
	}
	if got, want := 1000-longLoser, longDelta; got != want {
		t.Fatalf("long loser delta = %d, want matching winner delta %d", got, want)
	}
}

func TestRecordMatchResultUsesActualScoresAndAdjustedRating(t *testing.T) {
	ctx := context.Background()
	cfg := DefaultConfig()
	cfg.Database.Path = filepath.Join(t.TempDir(), "tankblaster-test.db")
	store, err := OpenStore(cfg)
	if err != nil {
		t.Fatalf("OpenStore() error = %v", err)
	}
	defer store.Close()

	winner, _, err := store.AuthenticateOrCreate(ctx, "", "", "Winner")
	if err != nil {
		t.Fatalf("create winner: %v", err)
	}
	loser, _, err := store.AuthenticateOrCreate(ctx, "", "", "Loser")
	if err != nil {
		t.Fatalf("create loser: %v", err)
	}

	if err := store.RecordMatchResult(ctx, winner.PlayerID, loser.PlayerID, 17, 5, 4); err != nil {
		t.Fatalf("RecordMatchResult() error = %v", err)
	}

	updatedWinner, err := store.playerByID(ctx, winner.PlayerID)
	if err != nil {
		t.Fatalf("load winner: %v", err)
	}
	updatedLoser, err := store.playerByID(ctx, loser.PlayerID)
	if err != nil {
		t.Fatalf("load loser: %v", err)
	}
	if got, want := updatedWinner.Score, 17; got != want {
		t.Fatalf("winner score = %d, want %d", got, want)
	}
	if got, want := updatedLoser.Score, 5; got != want {
		t.Fatalf("loser score = %d, want %d", got, want)
	}
	if updatedWinner.Rating <= winner.Rating {
		t.Fatalf("winner rating = %d, want > %d", updatedWinner.Rating, winner.Rating)
	}
	if updatedLoser.Rating >= loser.Rating {
		t.Fatalf("loser rating = %d, want < %d", updatedLoser.Rating, loser.Rating)
	}
}
