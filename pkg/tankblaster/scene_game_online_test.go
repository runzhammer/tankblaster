package tankblaster

import (
	"testing"

	"github.com/runzhammer/tankblaster/pkg/gamecore"
	"github.com/runzhammer/tankblaster/pkg/protocol"
)

func newOnlineTurnTestScene() *GameScene {
	return &GameScene{
		g: &GameLoop{
			online: &onlineGameRuntime{
				playerID: "p1",
				state: gamecore.MatchState{
					MatchID: "match-1",
					Players: []gamecore.Player{
						{ID: "p1", DisplayName: "Player 1"},
						{ID: "p2", DisplayName: "Player 2"},
					},
				},
			},
		},
		tanks: []*battleTank{
			{playerIndex: 0, power: 100, landed: true, player: PlayerConfig{Name: "Player 1"}},
			{playerIndex: 1, power: 100, landed: true, player: PlayerConfig{Name: "Player 2"}},
		},
		activePlayerIndex: 0,
	}
}

func TestApplyOnlineTurnIgnoresStaleTurnSequence(t *testing.T) {
	s := newOnlineTurnTestScene()
	s.g.online.turnSequence = 1

	s.applyOnlineTurnCommand(onlineTurnCommand(1, 1))

	if got, want := s.activePlayerIndex, 0; got != want {
		t.Fatalf("active player = %d, want unchanged %d", got, want)
	}
	if got, want := s.g.online.turnSequence, 1; got != want {
		t.Fatalf("turn sequence = %d, want %d", got, want)
	}
}

func TestApplyOnlineTurnWaitsForLocalTurnAdvance(t *testing.T) {
	s := newOnlineTurnTestScene()
	s.projectiles = []*projectile{{}}

	s.applyOnlineTurnCommand(onlineTurnCommand(1, 1))

	if got, want := s.activePlayerIndex, 0; got != want {
		t.Fatalf("active player = %d, want unchanged while projectile active %d", got, want)
	}
	if got, want := s.g.online.turnSequence, 0; got != want {
		t.Fatalf("turn sequence = %d, want unchanged %d", got, want)
	}
}

func TestApplyOnlineTurnAcceptsNewIdleTurnSequence(t *testing.T) {
	s := newOnlineTurnTestScene()

	s.applyOnlineTurnCommand(onlineTurnCommand(1, 1))

	if got, want := s.activePlayerIndex, 1; got != want {
		t.Fatalf("active player = %d, want %d", got, want)
	}
	if got, want := s.g.online.turnSequence, 1; got != want {
		t.Fatalf("turn sequence = %d, want %d", got, want)
	}
}

func onlineTurnCommand(playerIndex, turnSequence int) protocol.OnlineGameCommand {
	return protocol.OnlineGameCommand{
		MatchID:      "match-1",
		PlayerID:     "p1",
		PlayerIndex:  playerIndex,
		Kind:         "turn",
		TurnSequence: turnSequence,
	}
}
