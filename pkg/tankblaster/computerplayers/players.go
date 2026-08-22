package computerplayers

import (
	"math"
	"math/rand"
	"strings"
)

const (
	DoedelName   = "D. Dödel"
	FrederikName = "Frederik"
	MisterXName  = "Mister X"
	DrNukeName   = "Dr. Nuke"
	HaraldName   = "Harald"
)

type ID int

const (
	DoedelID ID = iota
	FrederikID
	MisterXID
	DrNukeID
	HaraldID
)

var orderedIDs = []ID{
	DoedelID,
	FrederikID,
	MisterXID,
	DrNukeID,
	HaraldID,
}

type TankState struct {
	Index int
	X     float64
	Y     float64
	Power int
	Alive bool
}

type ObstacleState struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
}

type State struct {
	ActiveIndex          int
	Tanks                []TankState
	Obstacles            []ObstacleState
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
	TargetXBias  float64
	LastTarget   int
	HasTarget    bool
}

type Lesson struct {
	ActiveX     float64
	TargetIndex int
	TargetX     float64
	ImpactX     float64
	Hit         bool
}

type Player interface {
	Decide(State, *rand.Rand) Decision
}

func Name(id ID) string {
	switch id {
	case DoedelID:
		return DoedelName
	case FrederikID:
		return FrederikName
	case MisterXID:
		return MisterXName
	case DrNukeID:
		return DrNukeName
	case HaraldID:
		return HaraldName
	default:
		return DoedelName
	}
}

func ForID(id ID) Player {
	switch id {
	case DoedelID:
		return Doedel{}
	case FrederikID:
		return Frederik{}
	case MisterXID:
		return MisterX{}
	case DrNukeID:
		return DrNuke{}
	case HaraldID:
		return Harald{}
	default:
		return Doedel{}
	}
}

func IDForName(name string) ID {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case strings.ToLower(DoedelName), "doedel", "dödel", "d. doedel":
		return DoedelID
	case strings.ToLower(FrederikName):
		return FrederikID
	case strings.ToLower(MisterXName), "misterx":
		return MisterXID
	case strings.ToLower(DrNukeName), "dr nuke":
		return DrNukeID
	case strings.ToLower(HaraldName):
		return HaraldID
	default:
		return DoedelID
	}
}

func IDs() []ID {
	ids := make([]ID, len(orderedIDs))
	copy(ids, orderedIDs)
	return ids
}

func NextID(id ID) ID {
	for i, candidate := range orderedIDs {
		if candidate == id {
			return orderedIDs[(i+1)%len(orderedIDs)]
		}
	}
	return DoedelID
}

func RandomMisterXID(rng *rand.Rand) ID {
	choices := []ID{DoedelID, FrederikID, DrNukeID, HaraldID}
	if rng == nil {
		rng = rand.New(rand.NewSource(1))
	}
	return choices[rng.Intn(len(choices))]
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
	case FrederikID, MisterXID, DrNukeID, HaraldID:
		learnSmart(memory, lesson, id)
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
