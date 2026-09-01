package gamecore

import (
	"errors"
	"math"
	"math/rand"
)

type MatchStatus string

const (
	MatchWaiting  MatchStatus = "waiting"
	MatchInGame   MatchStatus = "in_game"
	MatchFinished MatchStatus = "finished"
)

type Player struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	Rating      int    `json:"rating"`
}

type TankState struct {
	PlayerID string  `json:"player_id"`
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	Health   int     `json:"health"`
	Alive    bool    `json:"alive"`
}

type TerrainCrater struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Radius float64 `json:"radius"`
}

type MatchState struct {
	MatchID            string          `json:"match_id"`
	Seed               int64           `json:"seed"`
	Players            []Player        `json:"players"`
	Tanks              []TankState     `json:"tanks"`
	Terrain            []TerrainCrater `json:"terrain"`
	CurrentPlayerIndex int             `json:"current_player_index"`
	Round              int             `json:"round"`
	TotalRounds        int             `json:"total_rounds"`
	Wind               int             `json:"wind"`
	Status             MatchStatus     `json:"status"`
	WinnerID           string          `json:"winner_id,omitempty"`
}

type FireCommand struct {
	PlayerID string
	Weapon   string
	Angle    float64
	Power    float64
}

type ShotResult struct {
	PlayerID        string          `json:"player_id"`
	Weapon          string          `json:"weapon"`
	Angle           float64         `json:"angle"`
	Power           float64         `json:"power"`
	ImpactX         float64         `json:"impact_x"`
	ImpactY         float64         `json:"impact_y"`
	Damage          map[string]int  `json:"damage"`
	TerrainDelta    []TerrainCrater `json:"terrain_delta"`
	NextPlayerIndex int             `json:"next_player_index"`
	MatchFinished   bool            `json:"match_finished"`
	WinnerID        string          `json:"winner_id,omitempty"`
}

type Engine struct {
	rng  *rand.Rand
	seed int64
}

func NewEngine(seed int64) *Engine {
	return &Engine{
		rng:  rand.New(rand.NewSource(seed)),
		seed: seed,
	}
}

func (e *Engine) NewMatch(matchID string, players []Player) MatchState {
	tanks := make([]TankState, 0, len(players))
	spacing := 800.0 / math.Max(1, float64(len(players)-1))
	for i, player := range players {
		x := 112.0 + float64(i)*spacing
		tanks = append(tanks, TankState{
			PlayerID: player.ID,
			X:        x,
			Y:        560,
			Health:   100,
			Alive:    true,
		})
	}
	wind := e.rng.Intn(101)
	if e.rng.Intn(2) == 0 {
		wind = -wind
	}
	return MatchState{
		MatchID:            matchID,
		Seed:               e.seed,
		Players:            append([]Player(nil), players...),
		Tanks:              tanks,
		CurrentPlayerIndex: 0,
		Round:              1,
		TotalRounds:        1,
		Wind:               wind,
		Status:             MatchInGame,
	}
}

func ApplyFire(state *MatchState, cmd FireCommand) (ShotResult, error) {
	if state == nil || state.Status != MatchInGame {
		return ShotResult{}, errors.New("match is not active")
	}
	if state.CurrentPlayerIndex < 0 || state.CurrentPlayerIndex >= len(state.Players) {
		return ShotResult{}, errors.New("invalid current player")
	}
	if state.Players[state.CurrentPlayerIndex].ID != cmd.PlayerID {
		return ShotResult{}, errors.New("not this player's turn")
	}
	if cmd.Power < 0 || cmd.Power > 100 || cmd.Angle < 0 || cmd.Angle > 180 {
		return ShotResult{}, errors.New("invalid shot parameters")
	}

	shooter := tankForPlayer(state, cmd.PlayerID)
	if shooter == nil || !shooter.Alive {
		return ShotResult{}, errors.New("shooter is not alive")
	}

	rad := cmd.Angle * math.Pi / 180
	impactX := shooter.X + math.Cos(rad)*cmd.Power*7 + float64(state.Wind)*0.8
	impactY := shooter.Y - math.Sin(rad)*cmd.Power*2.8 + cmd.Power*1.7
	radius := weaponRadius(cmd.Weapon)
	damage := map[string]int{}
	for i := range state.Tanks {
		tank := &state.Tanks[i]
		if !tank.Alive || tank.PlayerID == cmd.PlayerID {
			continue
		}
		d := math.Hypot(tank.X-impactX, tank.Y-impactY)
		if d > radius {
			continue
		}
		amount := int(math.Round(float64(weaponDamage(cmd.Weapon)) * (1 - d/radius)))
		if amount < 1 {
			amount = 1
		}
		tank.Health -= amount
		if tank.Health <= 0 {
			tank.Health = 0
			tank.Alive = false
		}
		damage[tank.PlayerID] = amount
	}
	crater := TerrainCrater{X: impactX, Y: impactY, Radius: radius}
	state.Terrain = append(state.Terrain, crater)
	state.CurrentPlayerIndex = nextLivingPlayer(*state, state.CurrentPlayerIndex)
	winner := winnerID(*state)
	if winner != "" {
		state.Status = MatchFinished
		state.WinnerID = winner
	}
	return ShotResult{
		PlayerID:        cmd.PlayerID,
		Weapon:          cmd.Weapon,
		Angle:           cmd.Angle,
		Power:           cmd.Power,
		ImpactX:         impactX,
		ImpactY:         impactY,
		Damage:          damage,
		TerrainDelta:    []TerrainCrater{crater},
		NextPlayerIndex: state.CurrentPlayerIndex,
		MatchFinished:   state.Status == MatchFinished,
		WinnerID:        state.WinnerID,
	}, nil
}

func tankForPlayer(state *MatchState, playerID string) *TankState {
	for i := range state.Tanks {
		if state.Tanks[i].PlayerID == playerID {
			return &state.Tanks[i]
		}
	}
	return nil
}

func nextLivingPlayer(state MatchState, current int) int {
	if len(state.Tanks) == 0 {
		return 0
	}
	for offset := 1; offset <= len(state.Tanks); offset++ {
		idx := (current + offset) % len(state.Tanks)
		if state.Tanks[idx].Alive {
			return idx
		}
	}
	return current
}

func winnerID(state MatchState) string {
	winner := ""
	count := 0
	for _, tank := range state.Tanks {
		if tank.Alive {
			winner = tank.PlayerID
			count++
		}
	}
	if count == 1 {
		return winner
	}
	return ""
}

func weaponRadius(weapon string) float64 {
	switch weapon {
	case "large_grenade":
		return 96
	case "atom_bomb":
		return 140
	default:
		return 64
	}
}

func weaponDamage(weapon string) int {
	switch weapon {
	case "large_grenade":
		return 55
	case "atom_bomb":
		return 75
	default:
		return 40
	}
}
