package tankblaster

import (
	"image/color"
	"math"
	"testing"

	"github.com/runzhammer/tankblaster/pkg/engine"
	"github.com/runzhammer/tankblaster/pkg/gamecore"
	"github.com/runzhammer/tankblaster/pkg/protocol"
	"github.com/runzhammer/tankblaster/pkg/tankblaster/computerplayers"
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

func TestHandleOnlineEnvelopeRequestsCatchUpOnSequenceGap(t *testing.T) {
	client := &onlineClient{send: make(chan protocol.Envelope, 1)}
	s := newOnlineTurnTestScene()
	s.g.online.client = client
	s.g.online.lastServerSequence = 2

	s.handleOnlineGameEnvelope(onlineCommandEnvelope(t, 4, onlineTurnCommand(1, 1)))

	if got, want := s.activePlayerIndex, 0; got != want {
		t.Fatalf("active player = %d, want unchanged %d", got, want)
	}
	if got, want := s.g.online.lastServerSequence, int64(2); got != want {
		t.Fatalf("last server sequence = %d, want %d", got, want)
	}
	if !s.g.online.awaitingCatchUp {
		t.Fatal("client is not awaiting catch-up")
	}
	select {
	case env := <-client.send:
		if env.Type != protocol.TypeReconnect {
			t.Fatalf("sent type = %s, want %s", env.Type, protocol.TypeReconnect)
		}
		req, err := protocol.Decode[protocol.Reconnect](env)
		if err != nil {
			t.Fatalf("decode reconnect: %v", err)
		}
		if got, want := req.AfterSequence, int64(2); got != want {
			t.Fatalf("after sequence = %d, want %d", got, want)
		}
	default:
		t.Fatal("no catch-up request sent")
	}
}

func TestHandleOnlineHelloAckRequestsCatchUpAfterReconnect(t *testing.T) {
	client := &onlineClient{send: make(chan protocol.Envelope, 1)}
	s := newOnlineTurnTestScene()
	s.g.online.client = client
	s.g.online.lastServerSequence = 7

	s.handleOnlineGameEnvelope(protocol.Envelope{Type: protocol.TypeHelloAck})

	select {
	case env := <-client.send:
		req, err := protocol.Decode[protocol.Reconnect](env)
		if err != nil {
			t.Fatalf("decode reconnect: %v", err)
		}
		if got, want := req.MatchID, "match-1"; got != want {
			t.Fatalf("reconnect match id = %q, want %q", got, want)
		}
		if got, want := req.AfterSequence, int64(7); got != want {
			t.Fatalf("reconnect after sequence = %d, want %d", got, want)
		}
	default:
		t.Fatal("expected reconnect request after hello ack")
	}
}

func TestHandleOnlineEnvelopeAppliesReplayAfterSequenceGap(t *testing.T) {
	s := newOnlineTurnTestScene()
	s.g.online.lastServerSequence = 2
	s.g.online.awaitingCatchUp = true

	s.handleOnlineGameEnvelope(onlineCommandEnvelope(t, 3, onlineTurnCommand(1, 1)))

	if got, want := s.activePlayerIndex, 1; got != want {
		t.Fatalf("active player = %d, want replayed turn %d", got, want)
	}
	if got, want := s.g.online.lastServerSequence, int64(3); got != want {
		t.Fatalf("last server sequence = %d, want %d", got, want)
	}
	if s.g.online.awaitingCatchUp {
		t.Fatal("client is still awaiting catch-up after replay")
	}
}

func TestHandleOnlineEnvelopeRecoversFromReorderedServerEvents(t *testing.T) {
	client := &onlineClient{send: make(chan protocol.Envelope, 1)}
	s := newOnlineTurnTestScene()
	s.g.online.client = client
	s.g.online.lastServerSequence = 1

	s.handleOnlineGameEnvelope(onlineCommandEnvelope(t, 3, onlineTurnCommand(0, 2)))
	if got, want := s.g.online.lastServerSequence, int64(1); got != want {
		t.Fatalf("last server sequence after gap = %d, want %d", got, want)
	}
	<-client.send

	s.handleOnlineGameEnvelope(onlineCommandEnvelope(t, 2, onlineTurnCommand(1, 1)))
	if got, want := s.activePlayerIndex, 1; got != want {
		t.Fatalf("active player after replayed seq 2 = %d, want %d", got, want)
	}

	s.handleOnlineGameEnvelope(onlineCommandEnvelope(t, 3, onlineTurnCommand(0, 2)))
	if got, want := s.activePlayerIndex, 0; got != want {
		t.Fatalf("active player after replayed seq 3 = %d, want %d", got, want)
	}
	if got, want := s.g.online.lastServerSequence, int64(3); got != want {
		t.Fatalf("last server sequence = %d, want %d", got, want)
	}
}

func TestHandleOnlineEnvelopeIgnoresDuplicateServerSequence(t *testing.T) {
	s := newOnlineTurnTestScene()
	s.g.online.lastServerSequence = 3

	s.handleOnlineGameEnvelope(onlineCommandEnvelope(t, 3, onlineTurnCommand(1, 1)))

	if got, want := s.activePlayerIndex, 0; got != want {
		t.Fatalf("active player = %d, want unchanged %d", got, want)
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

func TestOnlineCanControlMultipleLobbyOwnedPlayers(t *testing.T) {
	s := newOnlineTurnTestScene()
	s.g.online.controlledPlayerIDs = map[string]bool{
		"p1:slot:0": true,
		"p1:slot:2": true,
	}
	s.g.online.state.Players = []gamecore.Player{
		{ID: "p1:slot:0"},
		{ID: "p2:slot:1"},
		{ID: "p1:slot:2"},
	}

	if !s.onlineCanControlPlayer(0) || !s.onlineCanControlPlayer(2) {
		t.Fatal("local client should control both owned lobby players")
	}
	if s.onlineCanControlPlayer(1) {
		t.Fatal("local client should not control remote lobby player")
	}
}

func TestSendOnlineGameCommandUsesControlledSlotPlayerID(t *testing.T) {
	s := newOnlineTurnTestScene()
	client := &onlineClient{send: make(chan protocol.Envelope, 1)}
	s.g.online.client = client
	s.g.online.playerID = "p1"
	s.g.online.state.Players = []gamecore.Player{
		{ID: "p1:slot:0"},
		{ID: "p2:slot:1"},
	}

	s.sendOnlineGameCommand(protocol.OnlineGameCommand{Kind: "fire", PlayerIndex: 0})

	select {
	case env := <-client.send:
		cmd, err := protocol.Decode[protocol.OnlineGameCommand](env)
		if err != nil {
			t.Fatalf("decode online command: %v", err)
		}
		if got, want := cmd.PlayerID, "p1:slot:0"; got != want {
			t.Fatalf("command player id = %q, want %q", got, want)
		}
	default:
		t.Fatal("no online command sent")
	}
}

func TestSyncOnlineTurnUsesPreviousControlledSlotPlayerID(t *testing.T) {
	s := newOnlineTurnTestScene()
	client := &onlineClient{send: make(chan protocol.Envelope, 1)}
	s.g.online.client = client
	s.g.online.controlledPlayerIDs = map[string]bool{"p1:slot:0": true}
	s.g.online.state.Players = []gamecore.Player{
		{ID: "p1:slot:0"},
		{ID: "p2:slot:1"},
	}
	s.activePlayerIndex = 1

	s.syncOnlineTurn(0)

	select {
	case env := <-client.send:
		cmd, err := protocol.Decode[protocol.OnlineGameCommand](env)
		if err != nil {
			t.Fatalf("decode online command: %v", err)
		}
		if got, want := cmd.PlayerID, "p1:slot:0"; got != want {
			t.Fatalf("turn command player id = %q, want previous player %q", got, want)
		}
		if got, want := cmd.PlayerIndex, 1; got != want {
			t.Fatalf("turn command player index = %d, want next player %d", got, want)
		}
	default:
		t.Fatal("no turn command sent")
	}
}

func TestOnlinePlayerConfigsUseLobbySlots(t *testing.T) {
	state := gamecore.MatchState{Players: []gamecore.Player{
		{ID: "host:slot:0", DisplayName: "Host"},
		{ID: "computer:slot:1", DisplayName: "CPU"},
	}}
	slots := []protocol.LobbySlot{
		{PlayerID: "host:slot:0", Kind: "human", Name: "Host", Color: protocol.RGBA{R: 10, G: 20, B: 30, A: 255}},
		{PlayerID: "computer:slot:1", Kind: "computer", Name: "CPU", ComputerID: int(computerplayers.MisterXID), Color: protocol.RGBA{R: 40, G: 50, B: 60, A: 255}},
	}

	players := onlinePlayerConfigs(state, "host", slots, false)

	if got, want := players[1].Kind, PlayerComputer; got != want {
		t.Fatalf("second player kind = %v, want %v", got, want)
	}
	if got, want := players[1].ComputerID, computerplayers.MisterXID; got != want {
		t.Fatalf("second computer id = %v, want %v", got, want)
	}
	if got, want := players[0].Color, (color.RGBA{R: 10, G: 20, B: 30, A: 255}); got != want {
		t.Fatalf("first color = %+v, want %+v", got, want)
	}
}

func TestOnlineGameSceneUsesAuthoritativeWorldSeeds(t *testing.T) {
	state := gamecore.MatchState{
		MatchID:            "match-1",
		Seed:               100,
		TerrainSeed:        200,
		TankSeed:           300,
		PalmSeed:           400,
		Wind:               -42,
		CurrentPlayerIndex: 1,
		TotalRounds:        3,
		Status:             gamecore.MatchInGame,
		Players: []gamecore.Player{
			{ID: "p1:slot:0", DisplayName: "Player 1"},
			{ID: "p2:slot:1", DisplayName: "Player 2"},
			{ID: "p1:slot:2", DisplayName: "CPU"},
		},
	}
	lobbySlots := []protocol.LobbySlot{
		{Index: 0, PlayerID: "p1:slot:0", Kind: "human", Name: "Player 1", Color: protocol.RGBA{R: 10, G: 20, B: 30, A: 255}},
		{Index: 1, PlayerID: "p2:slot:1", Kind: "human", Name: "Player 2", Color: protocol.RGBA{R: 40, G: 50, B: 60, A: 255}},
		{Index: 2, PlayerID: "p1:slot:2", Kind: "computer", Name: "CPU", Color: protocol.RGBA{R: 70, G: 80, B: 90, A: 255}},
	}

	left := newTestOnlineGameScene(t, state, lobbySlots)
	right := newTestOnlineGameScene(t, state, lobbySlots)

	if left.wind != right.wind || left.windDirection != right.windDirection {
		t.Fatalf("wind left=%d/%d right=%d/%d", left.wind, left.windDirection, right.wind, right.windDirection)
	}
	for _, x := range []float64{0, 240, 640, 960, left.worldWidth} {
		if got, want := left.ground.SurfaceY(x), right.ground.SurfaceY(x); got != want {
			t.Fatalf("surface y at %.1f = %f, want %f", x, got, want)
		}
	}
	if got, want := tankPositions(left), tankPositions(right); !vecSlicesEqual(got, want) {
		t.Fatalf("tank positions = %+v, want %+v", got, want)
	}
	if got, want := palmPositions(left), palmPositions(right); !vecSlicesEqual(got, want) {
		t.Fatalf("palm positions = %+v, want %+v", got, want)
	}
}

func TestOnlineDeterministicRNGKeepsCrumblerTerrainInSync(t *testing.T) {
	state := gamecore.MatchState{
		MatchID:     "match-1",
		Seed:        100,
		TerrainSeed: 200,
		TankSeed:    300,
		PalmSeed:    400,
		Wind:        -42,
		Status:      gamecore.MatchInGame,
		Players: []gamecore.Player{
			{ID: "p1:slot:0", DisplayName: "Player 1"},
			{ID: "p2:slot:1", DisplayName: "Player 2"},
		},
	}
	slots := []protocol.LobbySlot{
		{Index: 0, PlayerID: "p1:slot:0", Kind: "human", Name: "Player 1", Color: protocol.RGBA{A: 255}},
		{Index: 1, PlayerID: "p2:slot:1", Kind: "human", Name: "Player 2", Color: protocol.RGBA{A: 255}},
	}
	left := newTestOnlineGameScene(t, state, slots)
	right := newTestOnlineGameScene(t, state, slots)
	cmd := protocol.OnlineGameCommand{
		MatchID:         "match-1",
		PlayerID:        "p1:slot:0",
		PlayerIndex:     0,
		Kind:            "fire",
		TurnSequence:    7,
		CommandSequence: 11,
		WeaponSlot:      3,
		ShotStrength:    64,
	}
	pos := engine.V(520, 300)

	left.withDeterministicOnlineRNG(cmd, func() {
		left.startSmallCrumblerImpact(pos, smallCrumblerConfig())
	})
	right.withDeterministicOnlineRNG(cmd, func() {
		right.startSmallCrumblerImpact(pos, smallCrumblerConfig())
	})
	for i := 0; i < 240; i++ {
		left.updateSmallCrumblerImpacts()
		right.updateSmallCrumblerImpacts()
	}

	if got, want := groundSamples(left), groundSamples(right); got != want {
		t.Fatalf("ground samples diverged after deterministic crumbler impact:\n got  %+v\n want %+v", got, want)
	}
}

func newTestOnlineGameScene(t *testing.T, state gamecore.MatchState, lobbySlots []protocol.LobbySlot) *GameScene {
	t.Helper()
	scene, err := NewOnlineGameSceneWithControl(
		&GameLoop{options: defaultGameOptions(), muted: true},
		&onlineClient{id: onlineIdentity{PlayerID: "p1"}},
		state,
		[]string{"p1:slot:0", "p1:slot:2"},
		lobbySlots,
		false,
	)
	if err != nil {
		t.Fatalf("NewOnlineGameSceneWithControl() error = %v", err)
	}
	gameScene, ok := scene.(*GameScene)
	if !ok {
		t.Fatalf("scene type = %T, want *GameScene", scene)
	}
	return gameScene
}

func groundSamples(s *GameScene) [8]float64 {
	var out [8]float64
	if s.worldWidth <= 0 {
		return out
	}
	for i := range out {
		x := s.worldWidth * float64(i) / float64(len(out)-1)
		out[i] = s.ground.SurfaceY(x)
	}
	return out
}

func tankPositions(s *GameScene) []engine.Vec {
	out := make([]engine.Vec, 0, len(s.tanks))
	for _, tank := range s.tanks {
		if tank == nil || tank.body == nil || tank.body.Pos == nil {
			continue
		}
		out = append(out, *tank.body.Pos)
	}
	return out
}

func palmPositions(s *GameScene) []engine.Vec {
	out := make([]engine.Vec, 0, len(s.palms))
	for _, palm := range s.palms {
		if palm == nil || palm.sprite == nil || palm.sprite.Pos == nil {
			continue
		}
		out = append(out, *palm.sprite.Pos)
	}
	return out
}

func vecSlicesEqual(a, b []engine.Vec) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
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

func onlineCommandEnvelope(t *testing.T, sequence int64, cmd protocol.OnlineGameCommand) protocol.Envelope {
	t.Helper()
	env, err := protocol.Wrap(protocol.TypeOnlineGameCommand, cmd)
	if err != nil {
		t.Fatalf("wrap online command: %v", err)
	}
	env.Sequence = sequence
	env.MatchID = cmd.MatchID
	return env
}
