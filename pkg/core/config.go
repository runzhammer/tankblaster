package core

import (
	"image/color"
	"log"

	"github.com/runzhammer/gamedemo/pkg/engine"
	r "github.com/runzhammer/gamedemo/resources"

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
		ImpactAnimationSeconds   float64 `yaml:"impact_animation_seconds"`
		ImpactPauseSeconds       float64 `yaml:"impact_pause_seconds"`
		TankHitPauseSeconds      float64 `yaml:"tank_hit_pause_seconds"`
	} `yaml:"gameplay"`
	Debug struct {
		Enabled    bool   `yaml:"enabled"`
		StartScene string `yaml:"start_scene"`
		Game       struct {
			Rounds  int                   `yaml:"rounds"`
			Players []DebugPlayerSettings `yaml:"players"`
		} `yaml:"game"`
	} `yaml:"debug"`
}

type DebugPlayerSettings struct {
	Kind  string        `yaml:"kind"`
	Name  string        `yaml:"name"`
	Color ColorSettings `yaml:"color"`
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
		ImpactAnimationSeconds   float64
		ImpactPauseSeconds       float64
		TankHitPauseSeconds      float64
	}
	Debug struct {
		Enabled    bool
		StartScene string
		Game       struct {
			Rounds  int
			Players []DebugPlayerSettings
		}
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
	s.Gameplay.ImpactAnimationSeconds = fs.Gameplay.ImpactAnimationSeconds
	if s.Gameplay.ImpactAnimationSeconds < 0 {
		s.Gameplay.ImpactAnimationSeconds = 0
	}
	s.Gameplay.ImpactPauseSeconds = fs.Gameplay.ImpactPauseSeconds
	if s.Gameplay.ImpactPauseSeconds < 0 {
		s.Gameplay.ImpactPauseSeconds = 0
	}
	s.Gameplay.TankHitPauseSeconds = fs.Gameplay.TankHitPauseSeconds
	if s.Gameplay.TankHitPauseSeconds < 0 {
		s.Gameplay.TankHitPauseSeconds = 0
	}
	s.Debug.Enabled = fs.Debug.Enabled
	s.Debug.StartScene = fs.Debug.StartScene
	if s.Debug.StartScene == "" {
		s.Debug.StartScene = "game"
	}
	s.Debug.Game.Rounds = fs.Debug.Game.Rounds
	if s.Debug.Game.Rounds <= 0 {
		s.Debug.Game.Rounds = 10
	}
	s.Debug.Game.Players = fs.Debug.Game.Players
	GameSettings = &s
}

func Config() *Settings {
	return GameSettings
}

func (c ColorSettings) RGBA(fallback color.RGBA) color.RGBA {
	if c.A == 0 {
		c.A = fallback.A
	}
	return color.RGBA{R: c.R, G: c.G, B: c.B, A: c.A}
}
