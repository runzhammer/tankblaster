package server

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"time"

	"github.com/runzhammer/gamedemo/pkg/protocol"
	_ "modernc.org/sqlite"
)

type PlayerRecord struct {
	PlayerID      string
	DisplayName   string
	Token         string
	CreatedAt     time.Time
	LastSeen      time.Time
	Rating        int
	MatchesPlayed int
	Wins          int
	Losses        int
	Score         int
}

type Store struct {
	db  *sql.DB
	cfg Config
}

func OpenStore(cfg Config) (*Store, error) {
	db, err := sql.Open(cfg.Database.Driver, cfg.Database.Path)
	if err != nil {
		return nil, err
	}
	s := &Store{db: db, cfg: cfg}
	if err := s.migrate(context.Background()); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) migrate(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS players (
	player_id TEXT PRIMARY KEY,
	display_name TEXT NOT NULL,
	token TEXT NOT NULL,
	created_at TIMESTAMP NOT NULL,
	last_seen TIMESTAMP NOT NULL,
	rating INTEGER NOT NULL,
	matches_played INTEGER NOT NULL DEFAULT 0,
	wins INTEGER NOT NULL DEFAULT 0,
	losses INTEGER NOT NULL DEFAULT 0,
	score INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS players_score_idx ON players(score DESC, rating DESC);
`)
	return err
}

func (s *Store) AuthenticateOrCreate(ctx context.Context, playerID, token, displayName string) (PlayerRecord, bool, error) {
	now := time.Now().UTC()
	if playerID != "" && token != "" {
		player, err := s.playerByID(ctx, playerID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return PlayerRecord{}, false, err
		}
		if err == nil && player.Token == token {
			if displayName != "" {
				player.DisplayName = displayName
			}
			player.LastSeen = now
			_, err = s.db.ExecContext(ctx, `UPDATE players SET display_name = ?, last_seen = ? WHERE player_id = ?`, player.DisplayName, player.LastSeen, player.PlayerID)
			return player, false, err
		}
	}
	player := PlayerRecord{
		PlayerID:    randomID("plr", 18),
		DisplayName: displayName,
		Token:       randomToken(32),
		CreatedAt:   now,
		LastSeen:    now,
		Rating:      s.cfg.Rating.InitialRating,
	}
	if player.DisplayName == "" {
		player.DisplayName = "Player"
	}
	_, err := s.db.ExecContext(ctx, `
INSERT INTO players(player_id, display_name, token, created_at, last_seen, rating, matches_played, wins, losses, score)
VALUES (?, ?, ?, ?, ?, ?, 0, 0, 0, 0)`,
		player.PlayerID, player.DisplayName, player.Token, player.CreatedAt, player.LastSeen, player.Rating)
	return player, true, err
}

func (s *Store) playerByID(ctx context.Context, playerID string) (PlayerRecord, error) {
	var p PlayerRecord
	err := s.db.QueryRowContext(ctx, `
SELECT player_id, display_name, token, created_at, last_seen, rating, matches_played, wins, losses, score
FROM players WHERE player_id = ?`, playerID).Scan(
		&p.PlayerID, &p.DisplayName, &p.Token, &p.CreatedAt, &p.LastSeen, &p.Rating,
		&p.MatchesPlayed, &p.Wins, &p.Losses, &p.Score,
	)
	return p, err
}

func (s *Store) RecordWin(ctx context.Context, winnerID, loserID string) error {
	winner, err := s.playerByID(ctx, winnerID)
	if err != nil {
		return err
	}
	loser, err := s.playerByID(ctx, loserID)
	if err != nil {
		return err
	}
	winnerRating, loserRating := winner.Rating, loser.Rating
	if s.cfg.Rating.Enabled {
		winnerRating, loserRating = elo(winner.Rating, loser.Rating, s.cfg.Rating.KFactor, s.cfg.Rating.MinimumRating)
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx, `
UPDATE players
SET rating = ?, matches_played = matches_played + 1, wins = wins + 1, score = score + ?
WHERE player_id = ?`, winnerRating, s.cfg.Score.WinPoints, winnerID)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `
UPDATE players
SET rating = ?, matches_played = matches_played + 1, losses = losses + 1, score = score + ?
WHERE player_id = ?`, loserRating, s.cfg.Score.LossPoints, loserID)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) Leaderboard(ctx context.Context, limit int) ([]protocol.LeaderboardEntry, error) {
	if limit <= 0 {
		limit = s.cfg.Leaderboard.DefaultLimit
	}
	if limit > s.cfg.Leaderboard.MaximumLimit {
		limit = s.cfg.Leaderboard.MaximumLimit
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT player_id, display_name, score, rating, wins, losses, matches_played
FROM players ORDER BY score DESC, rating DESC, wins DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var entries []protocol.LeaderboardEntry
	rank := 1
	for rows.Next() {
		var e protocol.LeaderboardEntry
		e.Rank = rank
		if err := rows.Scan(&e.PlayerID, &e.DisplayName, &e.Score, &e.Rating, &e.Wins, &e.Losses, &e.MatchesPlayed); err != nil {
			return nil, err
		}
		entries = append(entries, e)
		rank++
	}
	return entries, rows.Err()
}

func elo(winnerRating, loserRating, k, minimum int) (int, int) {
	expectedWinner := 1 / (1 + math.Pow(10, float64(loserRating-winnerRating)/400))
	expectedLoser := 1 / (1 + math.Pow(10, float64(winnerRating-loserRating)/400))
	nextWinner := winnerRating + int(math.Round(float64(k)*(1-expectedWinner)))
	nextLoser := loserRating + int(math.Round(float64(k)*(0-expectedLoser)))
	if nextWinner < minimum {
		nextWinner = minimum
	}
	if nextLoser < minimum {
		nextLoser = minimum
	}
	return nextWinner, nextLoser
}
