package models

import (
	_ "embed"

	"github.com/runzhammer/gamedemo/pkg/core"
	"github.com/runzhammer/gamedemo/pkg/engine"
	r "github.com/runzhammer/gamedemo/resources"
)

type Background struct {
	Name   string
	Sprite engine.Sprite
}

func NewBackground(tick *float64) Background {

	b := Background{}
	b.Sprite = engine.NewSprite(r.BackgroundSprite, r.BackgroundSpec)

	b.Sprite = engine.Sprite{
		Tag:  "background",
		Pos:  engine.V(0, 0),
		Size: engine.V(core.Config().Screen.Width, core.Config().Screen.Height),
	}

	return b
}
