package tankblaster

import (
	"image/color"
	"math"
	"math/rand"
	"testing"

	"github.com/runzhammer/tankblaster/pkg/engine"
	"github.com/runzhammer/tankblaster/pkg/gamecore"
	"github.com/runzhammer/tankblaster/pkg/models"
	"github.com/runzhammer/tankblaster/pkg/protocol"
	"github.com/runzhammer/tankblaster/pkg/tankblaster/computerplayers"
	weaponspkg "github.com/runzhammer/tankblaster/pkg/tankblaster/weapons"
)

func newOnlineTurnTestScene() *GameScene {
	return &GameScene{
		g: &GameLoop{
			muted: true,
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

func TestSyncOnlineShopBuyIgnoresLocalSlotEcho(t *testing.T) {
	s := newOnlineTurnTestScene()
	client := &onlineClient{send: make(chan protocol.Envelope, 1)}
	s.g.online.client = client
	s.g.online.playerID = "p1"
	s.g.online.controlledPlayerIDs = map[string]bool{"p1:slot:0": true}
	s.g.online.state.Players = []gamecore.Player{
		{ID: "p1:slot:0"},
		{ID: "p2:slot:1"},
	}
	s.phase = phaseShop
	s.players = []PlayerConfig{{Name: "Player 1"}, {Name: "Player 2"}}
	s.credits = []int{5000, 5000}
	s.inventories = makeShopInventories(2)
	s.shopPlayerOrder = []int{0, 1}
	s.shopPlayerCursor = 0
	s.shopMode = shopModeClassA
	s.shopSelectedIndex = 1
	initialStock := s.inventories[0].classA[1]

	s.syncOnlineShopBuy()

	select {
	case env := <-client.send:
		cmd, err := protocol.Decode[protocol.OnlineGameCommand](env)
		if err != nil {
			t.Fatalf("decode online command: %v", err)
		}
		s.applyOnlineGameCommand(cmd)
	default:
		t.Fatal("no shop buy command sent")
	}
	if got, want := s.credits[0], 3000; got != want {
		t.Fatalf("credits after local echo = %d, want one purchase %d", got, want)
	}
	if got, want := s.inventories[0].classA[1], initialStock+10; got != want {
		t.Fatalf("inventory after local echo = %d, want one purchase stock %d", got, want)
	}
}

func TestApplyOnlineShopBuyUsesEconomySnapshotWithoutDoubleBuy(t *testing.T) {
	s := newOnlineTurnTestScene()
	s.g.online.playerID = "p1"
	s.g.online.controlledPlayerIDs = map[string]bool{"p1:slot:0": true}
	s.g.online.state.Players = []gamecore.Player{
		{ID: "p1:slot:0"},
		{ID: "p2:slot:1"},
	}
	s.phase = phaseShop
	s.players = []PlayerConfig{{Name: "Player 1"}, {Name: "Player 2"}}
	s.credits = []int{9999, 9999}
	s.inventories = makeShopInventories(2)
	s.shopPlayerOrder = []int{1}
	s.shopPlayerCursor = 0
	s.shopMode = shopModeClassA
	s.shopSelectedIndex = 1

	remoteInventory := protocol.OnlineInventoryState{
		ClassA: []int{50, 10},
		ClassB: []int{0, 0},
	}
	s.applyOnlineGameCommand(protocol.OnlineGameCommand{
		MatchID:     "match-1",
		PlayerID:    "p2:slot:1",
		PlayerIndex: 1,
		Kind:        "shop_buy",
		Economy: &protocol.OnlineEconomy{
			Scores:       []int{3, 4},
			RoundScores:  []int{30, 40},
			Credits:      []int{1111, 3000},
			Inventories:  []protocol.OnlineInventoryState{{ClassA: []int{50}, ClassB: []int{0}}, remoteInventory},
			ShopOrder:    []int{1},
			ShopCursor:   0,
			ShopMode:     int(shopModeClassA),
			ShopSelected: 1,
		},
	})

	if got, want := s.credits[1], 3000; got != want {
		t.Fatalf("remote credits = %d, want snapshot value %d", got, want)
	}
	if got, want := s.inventories[1].classA[1], 10; got != want {
		t.Fatalf("remote inventory = %d, want snapshot value %d", got, want)
	}
	if got, want := s.scores[1], 4; got != want {
		t.Fatalf("remote scores = %d, want %d", got, want)
	}
}

func TestSendOnlineGameCommandAttachesSharedState(t *testing.T) {
	s := newOnlineTurnTestScene()
	client := &onlineClient{send: make(chan protocol.Envelope, 1)}
	s.g.online.client = client
	s.g.online.state.Players = []gamecore.Player{{ID: "p1"}, {ID: "p2"}}
	s.clouds = []*battleCloud{{
		sprite:     &engine.Sprite{Pos: &engine.Vec{X: 10, Y: 20}},
		speed:      1.5,
		kind:       cloudKindNormal,
		aggression: 42,
	}}
	s.palms = []*battlePalm{{
		state:                palmStateAlive,
		aggression:           77,
		initialAggression:    12,
		aggressionMultiplier: 1.2,
	}}
	s.scores = []int{8, 9}
	s.roundScores = []int{18, 19}
	s.credits = []int{2800, 3100}
	s.inventories = makeShopInventories(2)
	s.inventories[1].classA[1] = 10

	s.sendOnlineGameCommand(protocol.OnlineGameCommand{Kind: "aim", PlayerIndex: 0})

	select {
	case env := <-client.send:
		cmd, err := protocol.Decode[protocol.OnlineGameCommand](env)
		if err != nil {
			t.Fatalf("decode online command: %v", err)
		}
		if len(cmd.Clouds) != 1 || cmd.Clouds[0].Aggression != 42 {
			t.Fatalf("cloud snapshot = %+v, want aggression 42", cmd.Clouds)
		}
		if len(cmd.Palms) != 1 || cmd.Palms[0].Aggression != 77 {
			t.Fatalf("palm snapshot = %+v, want aggression 77", cmd.Palms)
		}
		if cmd.Economy == nil || len(cmd.Economy.Credits) != 2 || cmd.Economy.Credits[1] != 3100 {
			t.Fatalf("economy snapshot = %+v, want credits for both players", cmd.Economy)
		}
	default:
		t.Fatal("no online command sent")
	}
}

func TestApplyOnlineSharedStateSyncsPalmAndCloudAggression(t *testing.T) {
	s := newOnlineTurnTestScene()
	s.g.online.state.Players = []gamecore.Player{{ID: "p1"}, {ID: "p2"}}
	s.clouds = []*battleCloud{{
		sprite:     &engine.Sprite{Pos: &engine.Vec{X: 10, Y: 20}},
		speed:      1,
		kind:       cloudKindLightning,
		aggression: 3,
	}}
	s.palms = []*battlePalm{{
		state:                palmStateAlive,
		aggression:           4,
		initialAggression:    2,
		aggressionMultiplier: 1,
	}}

	s.applyOnlineGameCommand(protocol.OnlineGameCommand{
		MatchID:     "match-1",
		PlayerID:    "p2",
		PlayerIndex: 1,
		Kind:        "aim",
		Clouds: []protocol.OnlineCloudState{{
			Kind:       int(cloudKindLightning),
			X:          30,
			Y:          40,
			Speed:      2,
			Aggression: 88,
		}},
		Palms: []protocol.OnlinePalmState{{
			State:                int(palmStateAlive),
			Aggression:           99,
			InitialAggression:    7,
			AggressionMultiplier: 1.3,
			Screaming:            true,
			ScreamAge:            5,
		}},
	})

	if got, want := s.clouds[0].aggression, 88.0; got != want {
		t.Fatalf("cloud aggression = %v, want %v", got, want)
	}
	if got, want := s.clouds[0].sprite.Pos.X, 30.0; got != want {
		t.Fatalf("cloud x = %v, want %v", got, want)
	}
	if got, want := s.palms[0].aggression, 99.0; got != want {
		t.Fatalf("palm aggression = %v, want %v", got, want)
	}
	if !s.palms[0].screaming || s.palms[0].screamAge != 5 {
		t.Fatalf("palm scream state = screaming %v age %d, want true age 5", s.palms[0].screaming, s.palms[0].screamAge)
	}
}

func TestOnlineZeroPowerAnimationChoiceIgnoresLocalRNGState(t *testing.T) {
	left := newOnlineTurnTestScene()
	right := newOnlineTurnTestScene()
	left.rng = rand.New(rand.NewSource(1))
	right.rng = rand.New(rand.NewSource(999))
	left.g.online.state.Seed = 4242
	right.g.online.state.Seed = 4242
	left.g.online.turnSequence = 7
	right.g.online.turnSequence = 7
	left.roundNumber = 3
	right.roundNumber = 3
	left.lastDamageSource = left.tanks[0]
	right.lastDamageSource = right.tanks[0]
	left.tanks[1].body = &engine.Sprite{Pos: &engine.Vec{X: 120, Y: 200}, Size: &engine.Vec{X: 20, Y: 12}}
	right.tanks[1].body = &engine.Sprite{Pos: &engine.Vec{X: 120, Y: 200}, Size: &engine.Vec{X: 20, Y: 12}}
	left.tanks[1].cannon = &engine.Sprite{}
	right.tanks[1].cannon = &engine.Sprite{}

	left.startZeroPowerAnimation(left.tanks[1])
	right.startZeroPowerAnimation(right.tanks[1])

	if len(left.zeroPowerEffects) != 1 || len(right.zeroPowerEffects) != 1 {
		t.Fatalf("zero power effect counts = %d/%d, want 1/1", len(left.zeroPowerEffects), len(right.zeroPowerEffects))
	}
	leftEffect := left.zeroPowerEffects[0]
	rightEffect := right.zeroPowerEffects[0]
	if leftEffect.kind != rightEffect.kind || leftEffect.randomSeed != rightEffect.randomSeed {
		t.Fatalf("zero power choices differ: left kind %d seed %d, right kind %d seed %d", leftEffect.kind, leftEffect.randomSeed, rightEffect.kind, rightEffect.randomSeed)
	}
}

func TestZeroPowerScatterProjectilesUseEffectSeed(t *testing.T) {
	left := newOnlineTurnTestScene()
	right := newOnlineTurnTestScene()
	left.rng = rand.New(rand.NewSource(1))
	right.rng = rand.New(rand.NewSource(999))
	left.tanks[0].body = &engine.Sprite{Pos: &engine.Vec{X: 120, Y: 200}, Size: &engine.Vec{X: 20, Y: 12}}
	right.tanks[0].body = &engine.Sprite{Pos: &engine.Vec{X: 120, Y: 200}, Size: &engine.Vec{X: 20, Y: 12}}
	left.tanks[0].shotStrength = 60
	right.tanks[0].shotStrength = 60

	left.fireZeroPowerScatterProjectiles(left.tanks[0], 12345)
	right.fireZeroPowerScatterProjectiles(right.tanks[0], 12345)

	if len(left.projectiles) != len(right.projectiles) {
		t.Fatalf("projectile counts = %d/%d, want equal", len(left.projectiles), len(right.projectiles))
	}
	for i := range left.projectiles {
		l := left.projectiles[i]
		r := right.projectiles[i]
		if l.scatterImpactColor != r.scatterImpactColor || l.randomSeed != r.randomSeed || l.velocity != r.velocity {
			t.Fatalf("projectile %d differs: left color %+v seed %d velocity %+v, right color %+v seed %d velocity %+v", i, l.scatterImpactColor, l.randomSeed, l.velocity, r.scatterImpactColor, r.randomSeed, r.velocity)
		}
	}
}

func TestSplitterFragmentsUseProjectileSeed(t *testing.T) {
	left := newOnlineTurnTestScene()
	right := newOnlineTurnTestScene()
	left.rng = rand.New(rand.NewSource(1))
	right.rng = rand.New(rand.NewSource(999))
	left.ground = models.NewGroundWithSize(400, 200)
	right.ground = models.NewGroundWithSize(400, 200)
	parent := func() *projectile {
		return &projectile{
			pos:                engine.V(180, 40),
			prev:               engine.V(180, 40),
			velocity:           engine.V(0, 0.2),
			effectiveWeapon:    weaponspkg.SplitterBomb(),
			hasEffectiveWeapon: true,
			splitterArmed:      true,
			randomSeed:         98765,
		}
	}

	leftFragments := left.splitProjectile(parent(), 0.16, 0.01)
	rightFragments := right.splitProjectile(parent(), 0.16, 0.01)

	if len(leftFragments) != len(rightFragments) {
		t.Fatalf("fragment counts = %d/%d, want equal", len(leftFragments), len(rightFragments))
	}
	for i := range leftFragments {
		l := leftFragments[i]
		r := rightFragments[i]
		if l.pos != r.pos || l.velocity != r.velocity || l.randomSeed != r.randomSeed || l.launchRot != r.launchRot {
			t.Fatalf("fragment %d differs: left pos %+v velocity %+v seed %d rot %f, right pos %+v velocity %+v seed %d rot %f", i, l.pos, l.velocity, l.randomSeed, l.launchRot, r.pos, r.velocity, r.randomSeed, r.launchRot)
		}
	}
}

func TestMFSTripleProjectilesUseOnlineCommandSeed(t *testing.T) {
	left := newOnlineTurnTestScene()
	right := newOnlineTurnTestScene()
	left.rng = rand.New(rand.NewSource(1))
	right.rng = rand.New(rand.NewSource(999))
	left.g.online.state.Seed = 4242
	right.g.online.state.Seed = 4242
	for _, s := range []*GameScene{left, right} {
		s.players = []PlayerConfig{{Name: "Player 1"}, {Name: "Player 2"}}
		s.inventories = makeShopInventories(2)
		s.inventories[0].classA[9] = 1
		s.activePlayerIndex = 0
		s.tanks[0].selectedWeapon = 10
		s.tanks[0].shotStrength = 50
		s.tanks[0].cannon = &engine.Sprite{
			Pos:  &engine.Vec{X: 100, Y: 80},
			Size: &engine.Vec{X: 20, Y: 4},
			Rot:  -math.Pi / 4,
		}
	}
	cmd := protocol.OnlineGameCommand{
		MatchID:         "match-1",
		PlayerID:        "p1",
		PlayerIndex:     0,
		Kind:            "fire",
		CommandSequence: 5,
		TurnSequence:    0,
		WeaponSlot:      10,
		ShotStrength:    50,
		CannonRotation:  -math.Pi / 4,
	}

	left.withDeterministicOnlineRNG(cmd, func() { left.fireActiveWeapon() })
	right.withDeterministicOnlineRNG(cmd, func() { right.fireActiveWeapon() })

	if len(left.projectiles) != 3 || len(right.projectiles) != 3 {
		t.Fatalf("projectile counts = %d/%d, want 3/3", len(left.projectiles), len(right.projectiles))
	}
	for i := range left.projectiles {
		l := left.projectiles[i]
		r := right.projectiles[i]
		if l.pos != r.pos || l.velocity != r.velocity || l.randomSeed != r.randomSeed || l.launchRot != r.launchRot {
			t.Fatalf("mfs projectile %d differs: left pos %+v velocity %+v seed %d rot %f, right pos %+v velocity %+v seed %d rot %f", i, l.pos, l.velocity, l.randomSeed, l.launchRot, r.pos, r.velocity, r.randomSeed, r.launchRot)
		}
	}
}

func TestAirStrikeBombPlanUsesProjectileSeed(t *testing.T) {
	left := newOnlineTurnTestScene()
	right := newOnlineTurnTestScene()
	left.rng = rand.New(rand.NewSource(1))
	right.rng = rand.New(rand.NewSource(999))
	left.ground = models.NewGroundWithSize(800, 240)
	right.ground = models.NewGroundWithSize(800, 240)
	left.worldWidth = 800
	right.worldWidth = 800

	left.startAirStrikeImpact(engine.V(400, 120), 54321)
	right.startAirStrikeImpact(engine.V(400, 120), 54321)

	if len(left.airStrikeImpacts) != 1 || len(right.airStrikeImpacts) != 1 {
		t.Fatalf("air strike impact counts = %d/%d, want 1/1", len(left.airStrikeImpacts), len(right.airStrikeImpacts))
	}
	leftBombs := left.airStrikeImpacts[0].bombs
	rightBombs := right.airStrikeImpacts[0].bombs
	if len(leftBombs) != len(rightBombs) {
		t.Fatalf("bomb counts = %d/%d, want equal", len(leftBombs), len(rightBombs))
	}
	for i := range leftBombs {
		if leftBombs[i] != rightBombs[i] {
			t.Fatalf("bomb %d differs: left %+v, right %+v", i, leftBombs[i], rightBombs[i])
		}
	}
}

func TestApplyOnlineFireIgnoresDuplicateCommandSequence(t *testing.T) {
	s := newOnlineTurnTestScene()
	s.rng = rand.New(rand.NewSource(1))
	s.activePlayerIndex = 0
	s.tanks[0].shotStrength = 50
	s.tanks[0].cannon = &engine.Sprite{
		Pos:  &engine.Vec{X: 100, Y: 80},
		Size: &engine.Vec{X: 20, Y: 4},
		Rot:  -math.Pi / 4,
	}
	cmd := protocol.OnlineGameCommand{
		MatchID:         "match-1",
		PlayerID:        "p1",
		PlayerIndex:     0,
		Kind:            "fire",
		CommandSequence: 5,
		TurnSequence:    0,
		ShotStrength:    50,
		CannonRotation:  -math.Pi / 4,
	}

	s.applyOnlineFireCommand(cmd)
	if len(s.projectiles) != 1 {
		t.Fatalf("projectiles after first fire = %d, want 1", len(s.projectiles))
	}
	s.projectiles = nil
	s.applyOnlineFireCommand(cmd)
	if len(s.projectiles) != 0 {
		t.Fatalf("duplicate fire command spawned %d projectiles, want 0", len(s.projectiles))
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
