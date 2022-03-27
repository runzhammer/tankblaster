package models

import (
	_ "embed"

	"github.com/runzhammer/gamedemo/pkg/core"
	"github.com/runzhammer/gamedemo/pkg/engine"
	r "github.com/runzhammer/gamedemo/resources"
)

type Ground struct {
	Name   string
	Sprite *engine.Sprite
}

func NewGround() Ground {

	gr := Ground{Name: "ground"}
	gr.Sprite = engine.NewSprite(r.GroundSprite, r.GroundSpec)

	gr.Sprite.Tag = "ground"
	gr.Sprite.Pos = engine.Vec{X: 0, Y: core.Config().Screen.Height - 110}
	gr.Sprite.Size = engine.V(core.Config().Screen.Width, 110)

	return gr
}
