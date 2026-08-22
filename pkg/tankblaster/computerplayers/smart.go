package computerplayers

import (
	"math"
	"math/rand"
)

type Frederik struct{}
type MisterX struct{}
type DrNuke struct{}
type Harald struct{}

type smartProfile struct {
	angleStep           float64
	strengthStep        int
	angleNoise          float64
	strengthNoise       float64
	selfTargetChance    float64
	randomTargetChance  float64
	trainingAmmoChance  float64
	delayMin            int
	delayJitter         int
	obstaclePenalty     float64
	learningRate        float64
	targetBiasLimit     float64
	preferHighDamage    bool
	preferFragileTarget bool
}

var (
	frederikProfile = smartProfile{
		angleStep:          3,
		strengthStep:       4,
		angleNoise:         12,
		strengthNoise:      7,
		selfTargetChance:   0.06,
		randomTargetChance: 0.18,
		trainingAmmoChance: 0.12,
		delayMin:           35,
		delayJitter:        60,
		obstaclePenalty:    140,
		learningRate:       0.32,
		targetBiasLimit:    220,
	}
	drNukeProfile = smartProfile{
		angleStep:           2,
		strengthStep:        2,
		angleNoise:          3.5,
		strengthNoise:       2,
		selfTargetChance:    0.015,
		randomTargetChance:  0.08,
		trainingAmmoChance:  0.04,
		delayMin:            26,
		delayJitter:         38,
		obstaclePenalty:     360,
		learningRate:        0.5,
		targetBiasLimit:     260,
		preferHighDamage:    true,
		preferFragileTarget: true,
	}
	haraldProfile = smartProfile{
		angleStep:           1,
		strengthStep:        1,
		angleNoise:          1.4,
		strengthNoise:       1,
		selfTargetChance:    0,
		randomTargetChance:  0.015,
		trainingAmmoChance:  0,
		delayMin:            18,
		delayJitter:         24,
		obstaclePenalty:     900,
		learningRate:        0.72,
		targetBiasLimit:     320,
		preferHighDamage:    true,
		preferFragileTarget: true,
	}
)

func (Frederik) Decide(state State, rng *rand.Rand) Decision {
	return smartDecision(state, rngOrDefault(rng), frederikProfile)
}

func (DrNuke) Decide(state State, rng *rand.Rand) Decision {
	return smartDecision(state, rngOrDefault(rng), drNukeProfile)
}

func (Harald) Decide(state State, rng *rand.Rand) Decision {
	return smartDecision(state, rngOrDefault(rng), haraldProfile)
}

func (MisterX) Decide(state State, rng *rand.Rand) Decision {
	return ForID(RandomMisterXID(rngOrDefault(rng))).Decide(state, rngOrDefault(rng))
}

func smartDecision(state State, rng *rand.Rand, profile smartProfile) Decision {
	maxStrength := clampInt(state.MaxStrength, 0, 100)
	minStrength := 0
	if maxStrength > 0 {
		minStrength = 1
	}

	target := chooseSmartTarget(state, rng, profile)
	aimTarget := applyMemoryAimBias(target, state.Memory)
	shot := bestShot(state, aimTarget, minStrength, maxStrength, profile)
	angle := clampFloat(shot.angle+state.Memory.AngleBias+rng.NormFloat64()*profile.angleNoise, 0, 180)
	strength := clampInt(int(math.Round(float64(shot.strength)+state.Memory.StrengthBias+rng.NormFloat64()*profile.strengthNoise)), minStrength, maxStrength)

	return Decision{
		TargetIndex:  target.Index,
		WeaponSlot:   smartWeaponSlot(state.AvailableWeaponSlots, rng, profile),
		Strength:     strength,
		AngleDegrees: angle,
		DelayFrames:  profile.delayMin + rng.Intn(maxIntForDoedel(1, profile.delayJitter)),
	}
}

func chooseSmartTarget(state State, rng *rand.Rand, profile smartProfile) TankState {
	active := activeTank(state)
	enemies := aliveEnemies(state)
	if len(enemies) == 0 || rng.Float64() < profile.selfTargetChance {
		return active
	}
	if rng.Float64() < profile.randomTargetChance {
		return enemies[rng.Intn(len(enemies))]
	}

	best := enemies[0]
	bestScore := math.Inf(1)
	for _, tank := range enemies {
		distance := math.Abs(tank.X - active.X)
		score := distance
		if profile.preferFragileTarget {
			score += float64(tank.Power) * 5
		}
		if score < bestScore {
			bestScore = score
			best = tank
		}
	}
	return best
}

func aliveEnemies(state State) []TankState {
	enemies := make([]TankState, 0, len(state.Tanks))
	for _, tank := range state.Tanks {
		if tank.Alive && tank.Index != state.ActiveIndex {
			enemies = append(enemies, tank)
		}
	}
	return enemies
}

type candidateShot struct {
	angle    float64
	strength int
	error    float64
}

func bestShot(state State, target TankState, minStrength, maxStrength int, profile smartProfile) candidateShot {
	active := activeTank(state)
	if maxStrength <= 0 {
		return candidateShot{angle: 90, strength: 0}
	}
	leftSide := target.X < active.X
	minAngle := 95.0
	maxAngle := 178.0
	if leftSide {
		minAngle = 2
		maxAngle = 85
	}
	best := candidateShot{angle: (minAngle + maxAngle) / 2, strength: minStrength, error: math.Inf(1)}
	strengthStep := maxIntForDoedel(1, profile.strengthStep)
	for strength := minStrength; strength <= maxStrength; strength += strengthStep {
		for angle := minAngle; angle <= maxAngle; angle += profile.angleStep {
			error := simulatedShotError(active, target, strength, angle, state, profile)
			if error < best.error {
				best = candidateShot{angle: angle, strength: strength, error: error}
			}
		}
	}
	return best
}

func simulatedShotError(active, target TankState, strength int, displayAngle float64, state State, profile smartProfile) float64 {
	const gravity = 0.16
	windAcceleration := float64(state.WindDirection*state.Wind) * 0.00065
	speed := 1.4 + float64(strength)*0.32
	rot := (displayAngle - 180) * math.Pi / 180
	x := active.X
	y := active.Y
	vx := math.Cos(rot) * speed
	vy := math.Sin(rot) * speed
	best := math.Inf(1)
	blocked := false
	for frame := 0; frame < 300; frame++ {
		vx += windAcceleration
		vy += gravity
		x += vx
		y += vy
		if !blocked && profile.obstaclePenalty > 0 && pointHitsObstacle(x, y, state.Obstacles) {
			blocked = true
		}
		dx := x - target.X
		dy := y - target.Y
		err := math.Hypot(dx, dy)
		if err < best {
			best = err
		}
		if y > target.Y+220 && frame > 20 {
			break
		}
	}
	if blocked {
		best += profile.obstaclePenalty
	}
	return best
}

func pointHitsObstacle(x, y float64, obstacles []ObstacleState) bool {
	for _, obstacle := range obstacles {
		if x >= obstacle.X && x <= obstacle.X+obstacle.Width &&
			y >= obstacle.Y && y <= obstacle.Y+obstacle.Height {
			return true
		}
	}
	return false
}

func smartWeaponSlot(slots []int, rng *rand.Rand, profile smartProfile) int {
	if len(slots) == 0 {
		return 0
	}
	if rng.Float64() < profile.trainingAmmoChance {
		for _, slot := range slots {
			if slot == 0 {
				return 0
			}
		}
	}
	if profile.preferHighDamage {
		best := slots[0]
		for _, slot := range slots[1:] {
			if slot > best {
				best = slot
			}
		}
		return best
	}
	for _, preferred := range []int{1, 2, 3, 0} {
		for _, slot := range slots {
			if slot == preferred {
				return slot
			}
		}
	}
	return slots[rng.Intn(len(slots))]
}

func learnSmart(memory *Memory, lesson Lesson, id ID) {
	if lesson.Hit {
		memory.StrengthBias *= 0.86
		memory.AngleBias *= 0.86
		memory.TargetXBias *= 0.55
		return
	}
	profile := frederikProfile
	if id == DrNukeID {
		profile = drNukeProfile
	}
	if id == HaraldID {
		profile = haraldProfile
	}
	errorX := lesson.TargetX - lesson.ImpactX
	if memory.HasTarget && memory.LastTarget != lesson.TargetIndex {
		memory.TargetXBias *= 0.25
	}
	memory.LastTarget = lesson.TargetIndex
	memory.HasTarget = true
	direction := math.Copysign(1, lesson.TargetX-lesson.ActiveX)
	rangeError := errorX * direction
	memory.TargetXBias = clampFloat(memory.TargetXBias+errorX*profile.learningRate, -profile.targetBiasLimit, profile.targetBiasLimit)
	memory.StrengthBias = clampFloat(memory.StrengthBias+rangeError*0.012*profile.learningRate, -10, 10)
	memory.AngleBias = clampFloat(memory.AngleBias+rangeError*0.0022*profile.learningRate, -7, 7)
}

func rngOrDefault(rng *rand.Rand) *rand.Rand {
	if rng != nil {
		return rng
	}
	return rand.New(rand.NewSource(1))
}
