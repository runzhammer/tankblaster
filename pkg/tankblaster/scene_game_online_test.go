package tankblaster

import (
	"math"
	"testing"

	"github.com/runzhammer/tankblaster/pkg/engine"
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

func TestApplyOnlineAimIgnoresLocalEcho(t *testing.T) {
	s := newOnlineTurnTestScene()
	s.tanks[0].body = &engine.Sprite{Rot: 0}
	s.tanks[0].cannon = &engine.Sprite{Rot: -math.Pi / 4}

	s.applyOnlineAimCommand(protocol.OnlineGameCommand{
		MatchID:         "match-1",
		PlayerID:        "p1",
		PlayerIndex:     0,
		Kind:            "aim",
		TurnSequence:    0,
		CommandSequence: 1,
		CannonRotation:  -math.Pi,
		ShotStrength:    80,
	})

	if got, want := s.tanks[0].cannon.Rot, -math.Pi/4; got != want {
		t.Fatalf("local echo changed cannon rotation = %f, want %f", got, want)
	}
	if got, want := s.tanks[0].shotStrength, 0; got != want {
		t.Fatalf("local echo changed shot strength = %d, want %d", got, want)
	}
}

func TestApplyOnlineAimIgnoresStaleRemoteSequence(t *testing.T) {
	s := newOnlineTurnTestScene()
	s.tanks[1].body = &engine.Sprite{Rot: 0}
	s.tanks[1].cannon = &engine.Sprite{Rot: -math.Pi / 2}
	s.g.online.receivedCommandSequences = map[string]int{"p2:aim": 3}

	s.applyOnlineAimCommand(protocol.OnlineGameCommand{
		MatchID:         "match-1",
		PlayerID:        "p2",
		PlayerIndex:     1,
		Kind:            "aim",
		TurnSequence:    0,
		CommandSequence: 2,
		CannonRotation:  -math.Pi,
		ShotStrength:    80,
	})

	if got, want := s.tanks[1].cannon.Rot, -math.Pi/2; got != want {
		t.Fatalf("stale aim changed cannon rotation = %f, want %f", got, want)
	}
}

func TestSyncOnlineShopContinueSendsOriginalShopPlayer(t *testing.T) {
	client := &onlineClient{send: make(chan protocol.Envelope, 1)}
	s := newOnlineTurnTestScene()
	s.g.online.client = client
	s.phase = phaseShop
	s.shopPlayerOrder = []int{0, 1}
	s.shopPlayerCursor = 0

	s.syncOnlineShopContinue()

	if got, want := s.shopPlayerCursor, 1; got != want {
		t.Fatalf("shop cursor = %d, want %d", got, want)
	}
	select {
	case env := <-client.send:
		cmd, err := protocol.Decode[protocol.OnlineGameCommand](env)
		if err != nil {
			t.Fatalf("decode online command: %v", err)
		}
		if got, want := cmd.Kind, "shop_continue"; got != want {
			t.Fatalf("command kind = %q, want %q", got, want)
		}
		if got, want := cmd.PlayerIndex, 0; got != want {
			t.Fatalf("command player index = %d, want original shop player %d", got, want)
		}
	default:
		t.Fatal("no shop continue command sent")
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
