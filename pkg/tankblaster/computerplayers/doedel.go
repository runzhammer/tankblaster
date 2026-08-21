package computerplayers

import (
	"math"
	"math/rand"
)

type Doedel struct{}

func (Doedel) Decide(state State, rng *rand.Rand) Decision {
	if rng == nil {
		rng = rand.New(rand.NewSource(1))
	}
	maxStrength := clampInt(state.MaxStrength, 0, 100)
	minStrength := 0
	if maxStrength > 0 {
		minStrength = 1
	}

	target := chooseDoedelTarget(state, rng)
	angle := doedelAngle(state, target, rng)
	strength := doedelStrength(state, target, minStrength, maxStrength, rng)

	return Decision{
		TargetIndex:  target.Index,
		WeaponSlot:   doedelWeaponSlot(state.AvailableWeaponSlots, rng),
		Strength:     strength,
		AngleDegrees: angle,
		DelayFrames:  45 + rng.Intn(85),
	}
}

func chooseDoedelTarget(state State, rng *rand.Rand) TankState {
	active := activeTank(state)
	candidates := make([]TankState, 0, len(state.Tanks))
	for _, tank := range state.Tanks {
		if tank.Alive && tank.Index != state.ActiveIndex {
			candidates = append(candidates, tank)
		}
	}
	if len(candidates) == 0 || rng.Float64() < 0.18 {
		return active
	}
	return candidates[rng.Intn(len(candidates))]
}

func activeTank(state State) TankState {
	for _, tank := range state.Tanks {
		if tank.Index == state.ActiveIndex {
			return tank
		}
	}
	return TankState{Index: state.ActiveIndex, Alive: true}
}

func doedelAngle(state State, target TankState, rng *rand.Rand) float64 {
	active := activeTank(state)
	dx := target.X - active.X
	if math.Abs(dx) < 1 {
		dx = rng.Float64()*2 - 1
	}

	base := 35.0
	if dx > 0 {
		base = 145.0
	}
	distance := math.Abs(dx)
	if distance > 450 {
		base += math.Copysign(12, dx)
	}
	angle := base + state.Memory.AngleBias + rng.NormFloat64()*24 + float64(state.WindDirection*state.Wind)*0.035

	switch roll := rng.Float64(); {
	case roll < 0.10:
		angle = 90 + rng.NormFloat64()*42
	case roll < 0.18:
		angle = 180 - angle + rng.NormFloat64()*18
	case roll < 0.24:
		if rng.Intn(2) == 0 {
			angle = rng.Float64() * 18
		} else {
			angle = 162 + rng.Float64()*18
		}
	}

	return clampFloat(angle, 0, 180)
}

func doedelStrength(state State, target TankState, minStrength, maxStrength int, rng *rand.Rand) int {
	if maxStrength <= 0 {
		return 0
	}
	active := activeTank(state)
	distance := math.Abs(target.X - active.X)
	guess := int(math.Round(distance/9.5)) + rng.Intn(31) - 15
	guess += int(math.Round(state.Memory.StrengthBias))
	guess += int(math.Round(float64(state.Wind*state.WindDirection) / 18))

	switch roll := rng.Float64(); {
	case roll < 0.16:
		guess = minStrength + rng.Intn(maxIntForDoedel(1, maxStrength-minStrength+1))
	case roll < 0.25:
		guess = minStrength
	case roll < 0.34:
		guess = maxStrength
	}

	return clampInt(guess, minStrength, maxStrength)
}

func doedelWeaponSlot(slots []int, rng *rand.Rand) int {
	if len(slots) == 0 {
		return 0
	}
	has := func(slot int) bool {
		for _, available := range slots {
			if available == slot {
				return true
			}
		}
		return false
	}
	switch roll := rng.Float64(); {
	case roll < 0.16 && has(0):
		return 0
	case roll < 0.72 && has(1):
		return 1
	case roll < 0.86 && has(2):
		return 2
	case roll < 0.94 && has(3):
		return 3
	default:
		return slots[rng.Intn(len(slots))]
	}
}

func maxIntForDoedel(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func learnDoedel(memory *Memory, lesson Lesson) {
	if lesson.Hit {
		memory.StrengthBias *= 0.96
		memory.AngleBias *= 0.96
		return
	}

	errorX := lesson.TargetX - lesson.ImpactX
	memory.StrengthBias = clampFloat(memory.StrengthBias+errorX*0.006, -14, 14)
	memory.AngleBias = clampFloat(memory.AngleBias+errorX*0.0018, -10, 10)
}
