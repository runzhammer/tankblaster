package core

import (
	"log"

	"github.com/runzhammer/gamedemo/pkg/engine"
	r "github.com/runzhammer/gamedemo/resources"

	"gopkg.in/yaml.v2"
)

type Settings struct {
	Screen struct {
		Width  float64 `yaml:"width"`
		Height float64 `yaml:"height"`
		Bounds engine.Rect
	} `yaml:"screen"`
}

var GameSettings Settings

func init() {

	var err error
	GameSettings = Settings{}

	// load config
	err = yaml.Unmarshal(r.GameConfig, Config)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	GameSettings.Screen.Bounds = engine.R(0, 0, GameSettings.Screen.Width, GameSettings.Screen.Height)
}

func Config() Settings {
	return GameSettings
}
