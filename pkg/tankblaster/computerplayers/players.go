package computerplayers

import (
	"math"
	"math/rand"
	"strings"
)

const DoedelName = "D. Dödel"

type ID int

const DoedelID ID = 0

type TankState struct {
	Index int
	X     float64
	Y     float64
	Power int
	Alive bool
}

type State struct {
	ActiveIndex          int
	Tanks                []TankState
	AvailableWeaponSlots []int
	MaxStrength          int
	Wind                 int
	WindDirection        int
	Memory               Memory
}

type Decision struct {
	TargetIndex  int
	WeaponSlot   int
	Strength     int
	AngleDegrees float64
	DelayFrames  int
}

type Memory struct {
	StrengthBias float64
	AngleBias    float64
}

type Lesson struct {
	TargetX float64
	ImpactX float64
	Hit     bool
}

type Player interface {
	Decide(State, *rand.Rand) Decision
}

func Name(id ID) string {
	switch id {
	case DoedelID:
		return DoedelName
	default:
		return DoedelName
	}
}

func ForID(id ID) Player {
	switch id {
	case DoedelID:
		return Doedel{}
	default:
		return Doedel{}
	}
}

func IDForName(name string) ID {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case strings.ToLower(DoedelName), "doedel", "dödel", "d. doedel":
		return DoedelID
	default:
		return DoedelID
	}
}

func Decide(id ID, state State, rng *rand.Rand) Decision {
	return ForID(id).Decide(state, rng)
}

func Learn(id ID, memory *Memory, lesson Lesson) {
	if memory == nil {
		return
	}
	switch id {
	case DoedelID:
		learnDoedel(memory, lesson)
	default:
		learnDoedel(memory, lesson)
	}
}

func clampInt(value, min, max int) int {
	if max < min {
		return min
	}
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func clampFloat(value, min, max float64) float64 {
	return math.Max(min, math.Min(max, value))
}
