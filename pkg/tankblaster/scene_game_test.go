package tankblaster

import (
	"math/rand"
	"testing"

	"github.com/runzhammer/tankblaster/pkg/engine"
	"github.com/runzhammer/tankblaster/pkg/tankblaster/computerplayers"
)

func TestLightningCloudAggressionIncrease(t *testing.T) {
	tests := []struct {
		aggression int
		want       float64
	}{
		{-10, 0},
		{0, 0},
		{35, 12.25},
		{50, 25},
		{75, 56.25},
		{100, 100},
		{120, 100},
	}

	for _, tt := range tests {
		if got := lightningCloudAggressionIncrease(tt.aggression); got != tt.want {
			t.Fatalf("lightningCloudAggressionIncrease(%d) = %v, want %v", tt.aggression, got, tt.want)
		}
	}
}

func TestShuffledIndexesReturnsPermutation(t *testing.T) {
	const count = 6
	indexes := shuffledIndexes(rand.New(rand.NewSource(1)), count)
	seen := make([]bool, count)

	for _, index := range indexes {
		if index < 0 || index >= count {
			t.Fatalf("index %d outside range 0..%d", index, count-1)
		}
		if seen[index] {
			t.Fatalf("duplicate index %d in %v", index, indexes)
		}
		seen[index] = true
	}
}

func TestQuickStartChoosesRandomStartingPlayer(t *testing.T) {
	seenNonFirstPlayer := false
	for seed := int64(1); seed <= 20; seed++ {
		s := newScoringTestScene(100, 100, 100)
		s.g = &GameLoop{}
		s.rng = rand.New(rand.NewSource(seed))
		s.activePlayerIndex = -1
		for _, tank := range s.tanks {
			tank.landed = true
		}

		if !s.ensureActivePlayerCanAct() {
			t.Fatal("expected active player")
		}
		if s.activePlayerIndex != 0 {
			seenNonFirstPlayer = true
			break
		}
	}

	if !seenNonFirstPlayer {
		t.Fatal("quick start always selected first player")
	}
}

func TestFollowingRoundStartsWithPreviousLoser(t *testing.T) {
	s := newScoringTestScene(100, 100, 100)
	s.g = &GameLoop{}
	s.rng = rand.New(rand.NewSource(1))
	s.roundNumber = 2
	s.nextRoundStartingPlayerIndex = 2
	s.activePlayerIndex = -1
	for _, tank := range s.tanks {
		tank.landed = true
	}

	s.chooseStartingPlayerAfterLanding()

	if got, want := s.activePlayerIndex, 2; got != want {
		t.Fatalf("active player = %d, want %d", got, want)
	}
}

func TestFirstRoundIgnoresPreviousLoserStartSlot(t *testing.T) {
	seenDifferentStarter := false
	for seed := int64(1); seed <= 20; seed++ {
		s := newScoringTestScene(100, 100, 100)
		s.g = &GameLoop{}
		s.rng = rand.New(rand.NewSource(seed))
		s.roundNumber = 1
		s.nextRoundStartingPlayerIndex = 2
		s.activePlayerIndex = -1
		for _, tank := range s.tanks {
			tank.landed = true
		}

		s.chooseStartingPlayerAfterLanding()
		if s.activePlayerIndex != 2 {
			seenDifferentStarter = true
			break
		}
	}

	if !seenDifferentStarter {
		t.Fatal("first round always used previous loser start slot")
	}
}

func TestTurnOrderFollowsTankPositionsLeftToRight(t *testing.T) {
	tanks := []*battleTank{
		testTankAtX(300),
		testTankAtX(100),
		testTankAtX(200),
	}

	order := turnOrderLeftToRight(tanks)

	want := []int{1, 2, 0}
	for i := range want {
		if order[i] != want[i] {
			t.Fatalf("turn order = %v, want %v", order, want)
		}
	}
}

func TestNextActivePlayerFollowsTurnOrderLeftToRight(t *testing.T) {
	s := newScoringTestScene(100, 100, 100)
	s.tanks[0].body = testTankBodyAtX(300)
	s.tanks[1].body = testTankBodyAtX(100)
	s.tanks[2].body = testTankBodyAtX(200)
	for _, tank := range s.tanks {
		tank.landed = true
	}
	s.setTurnOrderLeftToRight()

	s.activePlayerIndex = 1
	if got, want := s.nextActivePlayerIndex(), 2; got != want {
		t.Fatalf("next after left tank = %d, want %d", got, want)
	}

	s.activePlayerIndex = 0
	if got, want := s.nextActivePlayerIndex(), 1; got != want {
		t.Fatalf("next after right tank = %d, want %d", got, want)
	}
}

func TestComputerIntelligenceIncreasesWhenOnlyComputersCanAct(t *testing.T) {
	s := newScoringTestScene(100, 0, 100)
	s.effectiveComputerIDs = []computerplayers.ID{
		computerplayers.DoedelID,
		computerplayers.DoedelID,
		computerplayers.HaraldID,
	}
	s.tanks[0].player.Kind = PlayerComputer
	s.tanks[1].player.Kind = PlayerHuman
	s.tanks[2].player.Kind = PlayerComputer
	for _, tank := range s.tanks {
		tank.landed = true
	}

	s.increaseComputerIntelligenceIfOnlyComputersRemain(s.tanks[0])
	if got, want := s.effectiveComputerIDs[0], computerplayers.FrederikID; got != want {
		t.Fatalf("computer 0 id = %v, want %v", got, want)
	}

	s.increaseComputerIntelligenceIfOnlyComputersRemain(s.tanks[2])
	if got, want := s.effectiveComputerIDs[2], computerplayers.HaraldID; got != want {
		t.Fatalf("computer 2 id = %v, want %v", got, want)
	}
}

func TestComputerIntelligenceDoesNotIncreaseWhileHumanCanAct(t *testing.T) {
	s := newScoringTestScene(100, 100)
	s.effectiveComputerIDs = []computerplayers.ID{
		computerplayers.DoedelID,
		computerplayers.DoedelID,
	}
	s.tanks[0].player.Kind = PlayerComputer
	s.tanks[1].player.Kind = PlayerHuman
	for _, tank := range s.tanks {
		tank.landed = true
	}

	s.increaseComputerIntelligenceIfOnlyComputersRemain(s.tanks[0])
	if got, want := s.effectiveComputerIDs[0], computerplayers.DoedelID; got != want {
		t.Fatalf("computer 0 id = %v, want %v", got, want)
	}
}

func TestRoundLoserUsesLowestRoundScoreThenPower(t *testing.T) {
	s := newScoringTestScene(0, 25, 0)
	s.roundScores = []int{2, -1, -1}

	if got, want := s.roundLoserIndex(), 2; got != want {
		t.Fatalf("round loser = %d, want %d", got, want)
	}
}

func TestRoundEndStoresNextRoundStarter(t *testing.T) {
	s := newScoringTestScene(40, 0, 0)
	s.roundScores = []int{1, -2, 0}
	s.nextRoundStartingPlayerIndex = -1

	if !s.endRoundIfOnlyOneTankRemains() {
		t.Fatal("round should end")
	}
	if got, want := s.nextRoundStartingPlayerIndex, 1; got != want {
		t.Fatalf("next round starter = %d, want %d", got, want)
	}
}

func TestFinishPalmRevengeResetsPalmAggressionWhenTargetDestroyed(t *testing.T) {
	palm := &battlePalm{
		aggression:        palmRevengeTrigger,
		initialAggression: 12,
	}
	cloud := &battleCloud{
		aggression:        palmRevengeTrigger,
		initialAggression: 7,
	}
	s := &GameScene{}

	s.finishPalmRevenge(&palmRevengeEvent{
		palm:   palm,
		cloud:  cloud,
		target: &battleTank{power: 0},
	})

	if got, want := palm.aggression, 12.0; got != want {
		t.Fatalf("palm aggression = %v, want %v", got, want)
	}
	if got, want := cloud.aggression, palmRevengeTrigger; got != want {
		t.Fatalf("cloud aggression = %v, want %v", got, want)
	}
}

func TestFinishCloudRevengeResetsCloudAggressionWhenTargetDestroyed(t *testing.T) {
	cloud := &battleCloud{
		aggression:        palmRevengeTrigger,
		initialAggression: 9,
	}
	s := &GameScene{}

	s.finishPalmRevenge(&palmRevengeEvent{
		cloud:  cloud,
		target: &battleTank{power: 0},
	})

	if got, want := cloud.aggression, 9.0; got != want {
		t.Fatalf("cloud aggression = %v, want %v", got, want)
	}
}

func TestFinishRevengeKeepsAggressionWhenTargetSurvives(t *testing.T) {
	palm := &battlePalm{
		aggression:        88,
		initialAggression: 12,
	}
	cloud := &battleCloud{
		aggression:        77,
		initialAggression: 9,
	}
	s := &GameScene{}

	s.finishPalmRevenge(&palmRevengeEvent{
		palm:   palm,
		cloud:  cloud,
		target: &battleTank{power: 1},
	})

	if got, want := palm.aggression, 88.0; got != want {
		t.Fatalf("palm aggression = %v, want %v", got, want)
	}
	if got, want := cloud.aggression, 77.0; got != want {
		t.Fatalf("cloud aggression = %v, want %v", got, want)
	}
}

func testTankAtX(x float64) *battleTank {
	return &battleTank{
		power:  100,
		landed: true,
		body:   testTankBodyAtX(x),
	}
}

func testTankBodyAtX(x float64) *engine.Sprite {
	return &engine.Sprite{
		Pos:  &engine.Vec{X: x, Y: 0},
		Size: &engine.Vec{X: 20, Y: 10},
	}
}
