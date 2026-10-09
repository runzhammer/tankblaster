package tankblaster

import (
	"image"
	"math"
	"math/rand"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/runzhammer/tankblaster/pkg/engine"
	"github.com/runzhammer/tankblaster/pkg/models"
	weaponspkg "github.com/runzhammer/tankblaster/pkg/tankblaster/weapons"
)

func newScoringTestScene(powers ...int) *GameScene {
	s := &GameScene{
		scores:      make([]int, len(powers)),
		roundScores: make([]int, len(powers)),
		credits:     make([]int, len(powers)),
		inventories: makeShopInventories(len(powers)),
		rng:         rand.New(rand.NewSource(1)),
	}
	for index, power := range powers {
		s.tanks = append(s.tanks, &battleTank{
			playerIndex: index,
			power:       power,
		})
	}
	return s
}

func repeatedWaterSurface(count int, y float64) []float64 {
	surface := make([]float64, count)
	for i := range surface {
		surface[i] = y
	}
	return surface
}

func waterSurfaceFromGround(ground models.Ground, left, right int) []float64 {
	surface := make([]float64, right-left+1)
	for x := left; x <= right; x++ {
		surface[x-left] = ground.SurfaceY(float64(x))
	}
	return surface
}

func TestScoringNormalHitAwardsVictimCredits(t *testing.T) {
	s := newScoringTestScene(100, 100)
	attacker := s.tanks[0]
	victim := s.tanks[1]
	s.credits[1] = 1000

	s.damageTank(victim, 30, attacker, damageCauseDirect)

	if got, want := s.credits[1], 1210; got != want {
		t.Fatalf("victim credits = %d, want %d", got, want)
	}
	if got := s.credits[0]; got != 0 {
		t.Fatalf("attacker credits = %d, want unchanged 0", got)
	}
	if got := s.scores[0]; got != 0 {
		t.Fatalf("attacker points = %d, want unchanged 0", got)
	}
}

func TestScoringShieldedHitUsesNominalDamageForCredits(t *testing.T) {
	s := newScoringTestScene(100, 100)
	attacker := s.tanks[0]
	victim := s.tanks[1]
	s.inventories[1].energyShield = 100

	s.damageTank(victim, 30, attacker, damageCauseDirect)

	if got, want := s.credits[1], 210; got != want {
		t.Fatalf("victim credits = %d, want %d", got, want)
	}
	if got := victim.power; got != 100 {
		t.Fatalf("victim power = %d, want shield to absorb all energy damage", got)
	}
}

func TestScoringOriginalShieldSequence(t *testing.T) {
	s := newScoringTestScene(100)
	player := s.tanks[0]
	s.inventories[0].energyShield = 100

	s.damageTank(player, 100, nil, damageCauseDirect)
	if got, want := player.power, 100; got != want {
		t.Fatalf("power after first hit = %d, want %d", got, want)
	}
	if got, want := s.inventories[0].energyShield, 58; got != want {
		t.Fatalf("shield after first hit = %d, want %d", got, want)
	}

	s.damageTank(player, 100, nil, damageCauseDirect)
	if got, want := player.power, 100; got != want {
		t.Fatalf("power after second hit = %d, want %d", got, want)
	}
	if got, want := s.inventories[0].energyShield, 16; got != want {
		t.Fatalf("shield after second hit = %d, want %d", got, want)
	}

	s.damageTank(player, 100, nil, damageCauseDirect)
	if got, want := player.power, 74; got != want {
		t.Fatalf("power after third hit = %d, want %d", got, want)
	}
	if got, want := s.inventories[0].energyShield, 0; got != want {
		t.Fatalf("shield after third hit = %d, want %d", got, want)
	}

	s.damageTank(player, 100, nil, damageCauseDirect)
	if got, want := player.power, 0; got != want {
		t.Fatalf("power after fourth hit = %d, want clamped %d", got, want)
	}
	if got, want := s.inventories[0].energyShield, 0; got != want {
		t.Fatalf("shield after fourth hit = %d, want %d", got, want)
	}
}

func TestScoringKillAwardsAttackerPointsAndCredits(t *testing.T) {
	s := newScoringTestScene(100, 10)
	attacker := s.tanks[0]
	victim := s.tanks[1]
	s.scores[0] = 4
	s.credits[0] = 500

	s.damageTank(victim, 10, attacker, damageCauseDirect)

	if got, want := s.scores[0], 6; got != want {
		t.Fatalf("attacker points = %d, want %d", got, want)
	}
	if got, want := s.credits[0], 3500; got != want {
		t.Fatalf("attacker credits = %d, want %d", got, want)
	}
}

func TestWaterDrownsTankAndAwardsKillToSource(t *testing.T) {
	s := newScoringTestScene(100, 100)
	s.tanks[1].body = &engine.Sprite{
		Pos:  &engine.Vec{X: 48, Y: 50},
		Size: &engine.Vec{X: 20, Y: 10},
	}
	s.waterBlubberAnimation = spriteAnimation{
		frames:     []*ebiten.Image{nil},
		totalTicks: 3,
	}
	fill := &waterFill{
		leftX:      0,
		rightX:     120,
		topY:       40,
		surfaceY:   repeatedWaterSurface(121, 100),
		age:        1,
		duration:   1,
		hitPlayers: make(map[int]bool),
		source:     s.tanks[0],
	}

	s.damageTanksTouchingWater(fill)

	if got := s.tanks[1].power; got != 0 {
		t.Fatalf("victim power = %d, want 0", got)
	}
	if got, want := s.scores[0], 2; got != want {
		t.Fatalf("attacker points = %d, want %d", got, want)
	}
	if !fill.hitPlayers[1] {
		t.Fatal("victim should be marked as handled by water fill")
	}
	if got := len(s.waterBlubbers); got != 1 {
		t.Fatalf("water blubber count = %d, want 1", got)
	}
}

func TestExpandedWaterAwardsKillToExpandingSource(t *testing.T) {
	s := newScoringTestScene(100, 100, 100)
	s.ground = models.NewGroundWithSize(200, 100)
	s.tanks[2].body = &engine.Sprite{
		Pos:  &engine.Vec{X: 60, Y: 15},
		Size: &engine.Vec{X: 20, Y: 10},
	}
	fill := waterFill{
		leftX:      80,
		rightX:     90,
		topY:       -1,
		surfaceY:   waterSurfaceFromGround(s.ground, 80, 90),
		age:        1,
		duration:   1,
		hitPlayers: make(map[int]bool),
		source:     s.tanks[0],
	}
	s.waterFills = []waterFill{fill}

	s.refillWaterBelowAreaFrom(image.Rect(79, 0, 91, 100), s.tanks[1])
	s.damageTanksTouchingWater(&s.waterFills[0])

	if got := s.tanks[2].power; got != 0 {
		t.Fatalf("victim power = %d, want 0", got)
	}
	if got := s.scores[0]; got != 0 {
		t.Fatalf("original water source points = %d, want 0", got)
	}
	if got, want := s.scores[1], 2; got != want {
		t.Fatalf("expanding source points = %d, want %d", got, want)
	}
}

func TestScoringSuicidePenaltyWithEnoughCredits(t *testing.T) {
	s := newScoringTestScene(10)
	player := s.tanks[0]
	s.scores[0] = 5
	s.credits[0] = 5000

	s.damageTank(player, 10, player, damageCauseDirect)

	if got, want := s.scores[0], 2; got != want {
		t.Fatalf("points = %d, want %d", got, want)
	}
	if got, want := s.credits[0], 4000; got != want {
		t.Fatalf("credits = %d, want %d", got, want)
	}
}

func TestScoringSuicidePenaltyClampsCreditsAtZero(t *testing.T) {
	s := newScoringTestScene(10)
	player := s.tanks[0]
	s.scores[0] = 5
	s.credits[0] = 400

	s.damageTank(player, 10, player, damageCauseDirect)

	if got, want := s.scores[0], 2; got != want {
		t.Fatalf("points = %d, want %d", got, want)
	}
	if got := s.credits[0]; got != 0 {
		t.Fatalf("credits = %d, want clamped 0", got)
	}
}

func TestScoringSuicideAllowsNegativePoints(t *testing.T) {
	s := newScoringTestScene(10)
	player := s.tanks[0]
	s.scores[0] = 1

	s.damageTank(player, 10, player, damageCauseDirect)

	if got, want := s.scores[0], -2; got != want {
		t.Fatalf("points = %d, want %d", got, want)
	}
}

func TestWonderPalmImpactPlantsPalmWithoutDamage(t *testing.T) {
	s := newScoringTestScene(100)
	shooter := s.tanks[0]
	s.worldWidth = 200
	s.ground = models.NewGroundWithSize(200, 100)
	s.layers = engine.NewLayers(numLayers)
	s.palmImage = ebiten.NewImage(10, 20)
	s.palmPixels = image.NewRGBA(image.Rect(0, 0, 10, 20))

	handled := s.onGroundImpact(&projectile{
		pos:                engine.V(80, s.ground.SurfaceY(80)),
		shooter:            shooter,
		effectiveWeapon:    weaponspkg.WonderPalm(),
		hasEffectiveWeapon: true,
	})

	if !handled {
		t.Fatal("wonder palm impact was not handled")
	}
	if got, want := shooter.power, 100; got != want {
		t.Fatalf("shooter power = %d, want unchanged %d", got, want)
	}
	if got, want := len(s.palms), 1; got != want {
		t.Fatalf("palms = %d, want %d", got, want)
	}
	if got := len(s.impacts); got != 0 {
		t.Fatalf("impact animations = %d, want no crater impact", got)
	}
}

func TestScoringRoundWinnerAwardsPointsAndRemainingEnergyCredits(t *testing.T) {
	s := newScoringTestScene(60)
	winner := s.tanks[0]
	s.scores[0] = 4
	s.credits[0] = 1000

	s.rewardRoundWinner(winner)

	if got, want := s.scores[0], 5; got != want {
		t.Fatalf("points = %d, want %d", got, want)
	}
	if got, want := s.credits[0], 1900; got != want {
		t.Fatalf("credits = %d, want %d", got, want)
	}
}

func TestScoringNoRoundWinnerWhenNoSurvivor(t *testing.T) {
	s := newScoringTestScene(0, 0)

	if !s.endRoundIfOnlyOneTankRemains() {
		t.Fatal("round should end when no tanks remain")
	}

	for index := range s.tanks {
		if got := s.scores[index]; got != 0 {
			t.Fatalf("player %d points = %d, want 0", index, got)
		}
		if got := s.credits[index]; got != 0 {
			t.Fatalf("player %d credits = %d, want 0", index, got)
		}
	}
}

func TestScoringKillOfLastOpponentAwardsKillAndRoundWinOnce(t *testing.T) {
	s := newScoringTestScene(60, 40)
	attacker := s.tanks[0]
	victim := s.tanks[1]

	s.damageTank(victim, 40, attacker, damageCauseDirect)
	if !s.endRoundIfOnlyOneTankRemains() {
		t.Fatal("round should end after last opponent is destroyed")
	}

	if got, want := s.scores[0], 3; got != want {
		t.Fatalf("attacker points = %d, want %d", got, want)
	}
	if got, want := s.credits[0], 3900; got != want {
		t.Fatalf("attacker credits = %d, want %d", got, want)
	}
	if got, want := s.credits[1], 280; got != want {
		t.Fatalf("victim final-hit credits = %d, want %d", got, want)
	}

	s.endRoundIfOnlyOneTankRemains()
	if got, want := s.scores[0], 3; got != want {
		t.Fatalf("attacker points after repeated round-end check = %d, want %d", got, want)
	}
	if got, want := s.credits[0], 3900; got != want {
		t.Fatalf("attacker credits after repeated round-end check = %d, want %d", got, want)
	}
}

func TestMFSFiresExactlyThreeProjectilesAtOriginalAngles(t *testing.T) {
	s := newScoringTestScene(100)
	cannon := &engine.Sprite{
		Pos:       &engine.Vec{},
		Size:      &engine.Vec{X: 20, Y: 3},
		RotAnchor: &engine.Vec{},
		Rot:       1.25,
	}
	s.activePlayerIndex = 0
	s.tanks[0].cannon = cannon
	s.tanks[0].selectedWeapon = 10
	s.tanks[0].shotStrength = 20
	s.inventories[0].classA = make([]int, 10)
	s.inventories[0].classA[9] = 1

	s.fireActiveWeapon()

	if got, want := len(s.projectiles), 3; got != want {
		t.Fatalf("projectile count = %d, want %d", got, want)
	}
	offset := 5 * math.Pi / 180
	wantAngles := []float64{cannon.Rot - offset, cannon.Rot, cannon.Rot + offset}
	for i, want := range wantAngles {
		if got := s.projectiles[i].launchRot; math.Abs(got-want) > 0.000001 {
			t.Fatalf("projectile %d angle = %f, want %f", i, got, want)
		}
	}
}

func TestSplitterAndAirStrikeOriginalCounts(t *testing.T) {
	if got, want := splitterBombFragmentCount, 9; got != want {
		t.Fatalf("splitter fragment count = %d, want %d", got, want)
	}
	if got, want := airStrikeBombCount, 8; got != want {
		t.Fatalf("airstrike bomb count = %d, want %d", got, want)
	}
}

func TestAirStrikeUsesOriginalBombOffsets(t *testing.T) {
	got := originalAirStrikeOffsets(1)
	want := []float64{180, 120, 60, 0, 60, -120, -180, -240}
	if len(got) != len(want) {
		t.Fatalf("offset count = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("offset %d = %v, want %v", i, got[i], want[i])
		}
	}
}

func TestSurpriseEggRejectsOriginalExcludedWeaponIDs(t *testing.T) {
	s := newScoringTestScene(100)
	for i := 0; i < 1000; i++ {
		weapon := s.randomSurpriseEggWeaponIndex()
		if weapon == 0 || weapon == 13 || weapon == 18 {
			t.Fatalf("surprise egg selected excluded weapon id %d", weapon)
		}
		if weapon < 1 || weapon > 17 {
			t.Fatalf("surprise egg selected id %d, want 1..12 or 14..17", weapon)
		}
	}
}

func TestDestroyedUnlandedTankDoesNotBlockAllTanksLanded(t *testing.T) {
	s := newScoringTestScene(0, 100)
	s.tanks[0].landed = false
	s.tanks[1].landed = true

	if !s.allTanksLanded() {
		t.Fatal("destroyed unlanded tank should not block battle turn progression")
	}
}

func TestEnsureActivePlayerCanActSkipsDestroyedActiveTank(t *testing.T) {
	s := newScoringTestScene(0, 100, 100)
	s.activePlayerIndex = 0
	for _, tank := range s.tanks {
		tank.landed = true
	}

	if !s.ensureActivePlayerCanAct() {
		t.Fatal("expected another active player to be selected")
	}
	if got, want := s.activePlayerIndex, 1; got != want {
		t.Fatalf("active player = %d, want %d", got, want)
	}
}
