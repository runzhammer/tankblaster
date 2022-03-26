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
	b.Sprite = engine.NewSprite(r.GroundSprite, r.GroundSpec)

	b.Sprite.Tag = "background"
	b.Sprite.Pos = engine.V(0, 0)
	b.Sprite.Size = engine.V(core.Config().Screen.Width, core.Config().Screen.Height)

	return b
}
