package core

import (
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
	GameSettings = &s
}

func Config() *Settings {
	return GameSettings
}
