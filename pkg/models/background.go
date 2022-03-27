package models

import (
	_ "embed"

	"github.com/runzhammer/gamedemo/pkg/core"
	"github.com/runzhammer/gamedemo/pkg/engine"
	r "github.com/runzhammer/gamedemo/resources"
)

type Background struct {
	Name   string
	Sprite *engine.Sprite
}

func NewBackground() Background {

	b := Background{}
	b.Sprite = engine.NewSprite(r.BackgroundSprite, r.BackgroundSpec)

	b.Sprite.Tag = "background"
	b.Sprite.Pos = engine.Vec{X: 0, Y: 0}
	b.Sprite.Size = engine.Vec{X: core.Config().Screen.Width, Y: core.Config().Screen.Height}

	return b
}
