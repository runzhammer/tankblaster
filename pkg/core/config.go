package core

import (
	"image/color"
	"log"

	"github.com/runzhammer/tankblaster/pkg/engine"
	r "github.com/runzhammer/tankblaster/resources"

	"gopkg.in/yaml.v2"
)

type FileSettings struct {
	Window struct {
		Width  int `yaml:"width"`
		Height int `yaml:"height"`
	} `yaml:"window"`
	Tanks struct {
		Count int `yaml:"count"`
	} `yaml:"tanks"`
	Gameplay struct {
		SpawnLandingPauseSeconds float64 `yaml:"spawn_landing_pause_seconds"`
		ImpactPauseSeconds       float64 `yaml:"impact_pause_seconds"`
		TankHitPauseSeconds      float64 `yaml:"tank_hit_pause_seconds"`
		PalmHitPauseSeconds      float64 `yaml:"palm_hit_pause_seconds"`
		ProjectileReentry        *int    `yaml:"projectile_reentry"`
		PalmCount                *int    `yaml:"palm_count"`
		CloudAggression          *int    `yaml:"cloud_aggression"`
		QuickRoundStart          *bool   `yaml:"quick_round_start"`
		Palms                    struct {
			MinCount int `yaml:"min_count"`
			MaxCount int `yaml:"max_count"`
		} `yaml:"palms"`
		Clouds struct {
			MinCount int     `yaml:"min_count"`
			MaxCount int     `yaml:"max_count"`
			MinSpeed float64 `yaml:"min_speed"`
			MaxSpeed float64 `yaml:"max_speed"`
		} `yaml:"clouds"`
		Scoring FileScoringSettings `yaml:"scoring"`
	} `yaml:"gameplay"`
	Debug struct {
		Enabled             bool     `yaml:"enabled"`
		Mode                string   `yaml:"mode"`
		StartScene          string   `yaml:"start_scene"`
		StartShop           bool     `yaml:"start_shop"`
		ZeroPowerAnimations []string `yaml:"zero_power_animations"`
		Game                struct {
			Rounds  int                   `yaml:"rounds"`
			Players []DebugPlayerSettings `yaml:"players"`
		} `yaml:"game"`
	} `yaml:"debug"`
	Online struct {
		Enabled   *bool  `yaml:"enabled"`
		ServerURL string `yaml:"server_url"`
	} `yaml:"online"`
}

type FileScoringSettings struct {
	DamageReceivedCreditMultiplier *int `yaml:"damage_received_credit_multiplier"`
	Kill                           struct {
		Points  *int `yaml:"points"`
		Credits *int `yaml:"credits"`
	} `yaml:"kill"`
	Suicide struct {
		PointsPenalty  *int `yaml:"points_penalty"`
		CreditsPenalty *int `yaml:"credits_penalty"`
	} `yaml:"suicide"`
	RoundWin struct {
		Points                   *int `yaml:"points"`
		CreditPerRemainingEnergy *int `yaml:"credit_per_remaining_energy"`
	} `yaml:"round_win"`
}

type ScoringSettings struct {
	DamageReceivedCreditMultiplier int
	Kill                           struct {
		Points  int
		Credits int
	}
	Suicide struct {
		PointsPenalty  int
		CreditsPenalty int
	}
	RoundWin struct {
		Points                   int
		CreditPerRemainingEnergy int
	}
}

type DebugPlayerSettings struct {
	Kind       string        `yaml:"kind"`
	ID         int           `yaml:"id"`
	ComputerID int           `yaml:"computer_id"`
	Name       string        `yaml:"name"`
	Color      ColorSettings `yaml:"color"`
}

type ColorSettings struct {
	R uint8 `yaml:"r"`
	G uint8 `yaml:"g"`
	B uint8 `yaml:"b"`
	A uint8 `yaml:"a"`
}

type Settings struct {
	Screen struct {
		Width  float64
		Height float64
		Bounds engine.Rect
	}
	Tanks struct {
		Count int
	}
	Gameplay struct {
		SpawnLandingPauseSeconds float64
		ImpactPauseSeconds       float64
		TankHitPauseSeconds      float64
		PalmHitPauseSeconds      float64
		ProjectileReentry        int
		PalmCount                int
		CloudAggression          int
		QuickRoundStart          bool
		Palms                    struct {
			MinCount int
			MaxCount int
		}
		Clouds struct {
			MinCount int
			MaxCount int
			MinSpeed float64
			MaxSpeed float64
		}
		Scoring ScoringSettings
	}
	Debug struct {
		Enabled             bool
		Mode                string
		StartScene          string
		StartShop           bool
		ZeroPowerAnimations []string
		Game                struct {
			Rounds  int
			Players []DebugPlayerSettings
		}
	}
	Online struct {
		Enabled   bool
		ServerURL string
	}
}

var GameSettings *Settings

func init() {

	var err error
	fs := FileSettings{}

	// load config
	err = yaml.Unmarshal(r.GameConfig, &fs)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	s := Settings{}

	s.Screen.Width = float64(fs.Window.Width)
	s.Screen.Height = float64(fs.Window.Height)
	s.Screen.Bounds = engine.R(0, 0, s.Screen.Width, s.Screen.Height)
	s.Tanks.Count = fs.Tanks.Count
	if s.Tanks.Count <= 0 {
		s.Tanks.Count = 1
	}
	s.Gameplay.SpawnLandingPauseSeconds = fs.Gameplay.SpawnLandingPauseSeconds
	if s.Gameplay.SpawnLandingPauseSeconds < 0 {
		s.Gameplay.SpawnLandingPauseSeconds = 0
	}
	s.Gameplay.ImpactPauseSeconds = fs.Gameplay.ImpactPauseSeconds
	if s.Gameplay.ImpactPauseSeconds < 0 {
		s.Gameplay.ImpactPauseSeconds = 0
	}
	s.Gameplay.TankHitPauseSeconds = fs.Gameplay.TankHitPauseSeconds
	if s.Gameplay.TankHitPauseSeconds < 0 {
		s.Gameplay.TankHitPauseSeconds = 0
	}
	s.Gameplay.PalmHitPauseSeconds = fs.Gameplay.PalmHitPauseSeconds
	if s.Gameplay.PalmHitPauseSeconds <= 0 {
		s.Gameplay.PalmHitPauseSeconds = 2
	}
	s.Gameplay.ProjectileReentry = 1
	if fs.Gameplay.ProjectileReentry != nil {
		s.Gameplay.ProjectileReentry = *fs.Gameplay.ProjectileReentry
	}
	if s.Gameplay.ProjectileReentry < 0 || s.Gameplay.ProjectileReentry > 2 {
		s.Gameplay.ProjectileReentry = 1
	}
	s.Gameplay.PalmCount = -1
	if fs.Gameplay.PalmCount != nil {
		s.Gameplay.PalmCount = *fs.Gameplay.PalmCount
	}
	if s.Gameplay.PalmCount < -1 || s.Gameplay.PalmCount > 2 {
		s.Gameplay.PalmCount = -1
	}
	s.Gameplay.CloudAggression = 35
	if fs.Gameplay.CloudAggression != nil {
		s.Gameplay.CloudAggression = *fs.Gameplay.CloudAggression
	}
	if s.Gameplay.CloudAggression < 0 {
		s.Gameplay.CloudAggression = 0
	}
	if s.Gameplay.CloudAggression > 100 {
		s.Gameplay.CloudAggression = 100
	}
	s.Gameplay.QuickRoundStart = true
	if fs.Gameplay.QuickRoundStart != nil {
		s.Gameplay.QuickRoundStart = *fs.Gameplay.QuickRoundStart
	}
	s.Gameplay.Palms.MinCount = fs.Gameplay.Palms.MinCount
	s.Gameplay.Palms.MaxCount = fs.Gameplay.Palms.MaxCount
	s.Gameplay.Clouds.MinCount = fs.Gameplay.Clouds.MinCount
	if s.Gameplay.Clouds.MinCount <= 0 {
		s.Gameplay.Clouds.MinCount = 3
	}
	s.Gameplay.Clouds.MaxCount = fs.Gameplay.Clouds.MaxCount
	if s.Gameplay.Clouds.MaxCount <= 0 {
		s.Gameplay.Clouds.MaxCount = 6
	}
	if s.Gameplay.Clouds.MaxCount < s.Gameplay.Clouds.MinCount {
		s.Gameplay.Clouds.MaxCount = s.Gameplay.Clouds.MinCount
	}
	s.Gameplay.Clouds.MinSpeed = fs.Gameplay.Clouds.MinSpeed
	if s.Gameplay.Clouds.MinSpeed <= 0 {
		s.Gameplay.Clouds.MinSpeed = 0.035
	}
	s.Gameplay.Clouds.MaxSpeed = fs.Gameplay.Clouds.MaxSpeed
	if s.Gameplay.Clouds.MaxSpeed <= 0 {
		s.Gameplay.Clouds.MaxSpeed = 0.13
	}
	if s.Gameplay.Clouds.MaxSpeed < s.Gameplay.Clouds.MinSpeed {
		s.Gameplay.Clouds.MaxSpeed = s.Gameplay.Clouds.MinSpeed
	}
	s.Gameplay.Scoring = scoringSettingsWithDefaults(fs.Gameplay.Scoring)
	s.Debug.Enabled = fs.Debug.Enabled
	s.Debug.Mode = fs.Debug.Mode
	s.Debug.StartScene = fs.Debug.StartScene
	s.Debug.StartShop = fs.Debug.StartShop
	s.Debug.ZeroPowerAnimations = fs.Debug.ZeroPowerAnimations
	if s.Debug.StartScene == "" {
		s.Debug.StartScene = "game"
	}
	s.Debug.Game.Rounds = fs.Debug.Game.Rounds
	if s.Debug.Game.Rounds <= 0 {
		s.Debug.Game.Rounds = 10
	}
	s.Debug.Game.Players = fs.Debug.Game.Players
	s.Online.Enabled = true
	if fs.Online.Enabled != nil {
		s.Online.Enabled = *fs.Online.Enabled
	}
	s.Online.ServerURL = fs.Online.ServerURL
	if s.Online.ServerURL == "" {
		s.Online.ServerURL = "ws://127.0.0.1:8765/game"
	}
	GameSettings = &s
}

func Config() *Settings {
	return GameSettings
}

func scoringSettingsWithDefaults(fs FileScoringSettings) ScoringSettings {
	// Defaults mirror the Points/Credits economy of Tank Blaster II 1.3.0.0.
	var s ScoringSettings
	s.DamageReceivedCreditMultiplier = 7
	s.Kill.Points = 2
	s.Kill.Credits = 3000
	s.Suicide.PointsPenalty = 3
	s.Suicide.CreditsPenalty = 1000
	s.RoundWin.Points = 1
	s.RoundWin.CreditPerRemainingEnergy = 15

	if fs.DamageReceivedCreditMultiplier != nil {
		s.DamageReceivedCreditMultiplier = *fs.DamageReceivedCreditMultiplier
	}
	if fs.Kill.Points != nil {
		s.Kill.Points = *fs.Kill.Points
	}
	if fs.Kill.Credits != nil {
		s.Kill.Credits = *fs.Kill.Credits
	}
	if fs.Suicide.PointsPenalty != nil {
		s.Suicide.PointsPenalty = *fs.Suicide.PointsPenalty
	}
	if fs.Suicide.CreditsPenalty != nil {
		s.Suicide.CreditsPenalty = *fs.Suicide.CreditsPenalty
	}
	if fs.RoundWin.Points != nil {
		s.RoundWin.Points = *fs.RoundWin.Points
	}
	if fs.RoundWin.CreditPerRemainingEnergy != nil {
		s.RoundWin.CreditPerRemainingEnergy = *fs.RoundWin.CreditPerRemainingEnergy
	}
	return s
}

func (c ColorSettings) RGBA(fallback color.RGBA) color.RGBA {
	if c.A == 0 {
		c.A = fallback.A
	}
	return color.RGBA{R: c.R, G: c.G, B: c.B, A: c.A}
}
