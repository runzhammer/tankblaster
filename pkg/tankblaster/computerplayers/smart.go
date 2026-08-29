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
	angleStep            float64
	strengthStep         int
	angleNoise           float64
	strengthNoise        float64
	selfTargetChance     float64
	randomTargetChance   float64
	trainingAmmoChance   float64
	laserAimNoise        float64
	delayMin             int
	delayJitter          int
	obstaclePenalty      float64
	learningRate         float64
	targetBiasLimit      float64
	hitReward            float64
	preferHighDamage     bool
	preferFragileTarget  bool
	preferredWeaponSlots []int
}

const (
	weaponSlotTraining = iota
	weaponSlotGrenade
	weaponSlotLargeGrenade
	weaponSlotAtomBomb
	weaponSlotHBomb
	weaponSlotPlasmaMelter
	weaponSlotWonderPalm
	weaponSlotFireball
	weaponSlotWater
	weaponSlotMoles
	weaponSlotMFSTriple
	weaponSlotSmallCrumblers
	weaponSlotLargeCrumblers
	weaponSlotSurpriseEgg
	weaponSlotMosquitos
	weaponSlotShockwave
	weaponSlotAirStrike
	weaponSlotSplitterBomb
	weaponSlotLaser
)

var (
	frederikProfile = smartProfile{
		angleStep:          3,
		strengthStep:       4,
		angleNoise:         7,
		strengthNoise:      4,
		selfTargetChance:   0.03,
		randomTargetChance: 0.10,
		trainingAmmoChance: 0.06,
		laserAimNoise:      2.4,
		delayMin:           35,
		delayJitter:        60,
		obstaclePenalty:    140,
		learningRate:       0.32,
		targetBiasLimit:    220,
		hitReward:          600,
		preferredWeaponSlots: []int{
			weaponSlotGrenade,
			weaponSlotLargeGrenade,
			weaponSlotMoles,
			weaponSlotFireball,
			weaponSlotSmallCrumblers,
			weaponSlotWater,
			weaponSlotSplitterBomb,
			weaponSlotMFSTriple,
			weaponSlotAtomBomb,
			weaponSlotWonderPalm,
			weaponSlotLargeCrumblers,
			weaponSlotMosquitos,
			weaponSlotShockwave,
			weaponSlotAirStrike,
			weaponSlotHBomb,
			weaponSlotPlasmaMelter,
			weaponSlotLaser,
			weaponSlotSurpriseEgg,
			weaponSlotTraining,
		},
	}
	drNukeProfile = smartProfile{
		angleStep:           1.5,
		strengthStep:        1,
		angleNoise:          1.7,
		strengthNoise:       1,
		selfTargetChance:    0.004,
		randomTargetChance:  0.03,
		trainingAmmoChance:  0.01,
		laserAimNoise:       0.8,
		delayMin:            26,
		delayJitter:         38,
		obstaclePenalty:     360,
		learningRate:        0.5,
		targetBiasLimit:     260,
		hitReward:           900,
		preferHighDamage:    true,
		preferFragileTarget: true,
		preferredWeaponSlots: []int{
			weaponSlotLaser,
			weaponSlotPlasmaMelter,
			weaponSlotHBomb,
			weaponSlotAirStrike,
			weaponSlotShockwave,
			weaponSlotAtomBomb,
			weaponSlotMFSTriple,
			weaponSlotSplitterBomb,
			weaponSlotLargeGrenade,
			weaponSlotLargeCrumblers,
			weaponSlotMosquitos,
			weaponSlotFireball,
			weaponSlotMoles,
			weaponSlotWater,
			weaponSlotSmallCrumblers,
			weaponSlotWonderPalm,
			weaponSlotGrenade,
			weaponSlotSurpriseEgg,
			weaponSlotTraining,
		},
	}
	haraldProfile = smartProfile{
		angleStep:           0.75,
		strengthStep:        1,
		angleNoise:          0.35,
		strengthNoise:       0.25,
		selfTargetChance:    0,
		randomTargetChance:  0,
		trainingAmmoChance:  0,
		laserAimNoise:       0.15,
		delayMin:            18,
		delayJitter:         24,
		obstaclePenalty:     900,
		learningRate:        0.72,
		targetBiasLimit:     320,
		hitReward:           1400,
		preferHighDamage:    true,
		preferFragileTarget: true,
		preferredWeaponSlots: []int{
			weaponSlotLaser,
			weaponSlotPlasmaMelter,
			weaponSlotAirStrike,
			weaponSlotShockwave,
			weaponSlotHBomb,
			weaponSlotAtomBomb,
			weaponSlotMFSTriple,
			weaponSlotSplitterBomb,
			weaponSlotMosquitos,
			weaponSlotLargeGrenade,
			weaponSlotLargeCrumblers,
			weaponSlotFireball,
			weaponSlotWater,
			weaponSlotMoles,
			weaponSlotSmallCrumblers,
			weaponSlotGrenade,
			weaponSlotWonderPalm,
			weaponSlotSurpriseEgg,
			weaponSlotTraining,
		},
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
	laserAngle, laserUseful := directLaserAngle(state, target)
	weaponSlot := smartWeaponSlot(state.AvailableWeaponSlots, rng, profile, laserUseful)
	if weaponSlot == weaponSlotLaser {
		angle = clampFloat(laserAngle+rng.NormFloat64()*profile.laserAimNoise, 0, 180)
		strength = maxStrength
	}

	return Decision{
		TargetIndex:  target.Index,
		WeaponSlot:   weaponSlot,
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
	const (
		gravity              = 0.16
		projectileWindFactor = 0.00195
	)
	windAcceleration := float64(state.WindDirection*state.Wind) * projectileWindFactor
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
		if pointHitsTarget(x, y, target) {
			return math.Max(0, float64(frame)*0.08-profile.hitReward)
		}
		if !blocked && profile.obstaclePenalty > 0 && pointHitsObstacle(x, y, state.Obstacles) {
			blocked = true
		}
		if groundY, ok := groundSurfaceY(x, state.Ground); ok && y >= groundY {
			return math.Hypot(x-target.X, y-target.Y) + 180
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

func pointHitsTarget(x, y float64, target TankState) bool {
	halfWidth := math.Max(16, target.Width*0.42)
	halfHeight := math.Max(10, target.Height*0.42)
	return x >= target.X-halfWidth && x <= target.X+halfWidth &&
		y >= target.Y-halfHeight && y <= target.Y+halfHeight
}

func groundSurfaceY(x float64, samples []GroundSample) (float64, bool) {
	if len(samples) == 0 || x < samples[0].X || x > samples[len(samples)-1].X {
		return 0, false
	}
	for i := 1; i < len(samples); i++ {
		right := samples[i]
		if x > right.X {
			continue
		}
		left := samples[i-1]
		width := right.X - left.X
		if width <= 0 {
			return right.Y, true
		}
		t := (x - left.X) / width
		return left.Y + (right.Y-left.Y)*t, true
	}
	return samples[len(samples)-1].Y, true
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

func smartWeaponSlot(slots []int, rng *rand.Rand, profile smartProfile, laserUseful bool) int {
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
	for _, preferred := range profile.preferredWeaponSlots {
		if preferred == weaponSlotLaser && !laserUseful {
			continue
		}
		for _, slot := range slots {
			if slot == preferred {
				return slot
			}
		}
	}
	if profile.preferHighDamage {
		best := -1
		for _, slot := range slots {
			if slot == weaponSlotLaser && !laserUseful {
				continue
			}
			if slot > best {
				best = slot
			}
		}
		if best >= 0 {
			return best
		}
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

func directLaserAngle(state State, target TankState) (float64, bool) {
	active := activeTank(state)
	dx := target.X - active.X
	dy := target.Y - active.Y
	distance := math.Hypot(dx, dy)
	if distance < 1 {
		return 90, false
	}
	angle := math.Atan2(dy, dx)*180/math.Pi + 180
	for angle < 0 {
		angle += 360
	}
	for angle >= 360 {
		angle -= 360
	}
	if angle > 180 {
		return 0, false
	}
	return angle, laserHasLineOfSight(state, active, target, dx/distance, dy/distance, distance)
}

func laserHasLineOfSight(state State, active, target TankState, dirX, dirY, distance float64) bool {
	step := 4.0
	startSkip := math.Max(active.Width, active.Height) * 0.65
	targetReach := math.Max(target.Width, target.Height) * 0.5
	for traveled := startSkip; traveled <= distance+targetReach; traveled += step {
		x := active.X + dirX*traveled
		y := active.Y + dirY*traveled
		if pointHitsTarget(x, y, target) {
			return true
		}
		for _, tank := range state.Tanks {
			if !tank.Alive || tank.Index == active.Index || tank.Index == target.Index {
				continue
			}
			if pointHitsTarget(x, y, tank) {
				return false
			}
		}
		if pointHitsObstacle(x, y, state.Obstacles) {
			return false
		}
		if groundY, ok := groundSurfaceY(x, state.Ground); ok && y >= groundY {
			return false
		}
	}
	return false
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
