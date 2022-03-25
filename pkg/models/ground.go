package models

import (
	_ "embed"

	"github.com/runzhammer/gamedemo/pkg/engine"
	"github.com/runzhammer/gamedemo/pkg/tankblaster"
	r "github.com/runzhammer/gamedemo/resources"
)

type Ground struct {
	Name   string
	Sprite engine.Sprite
}

func NewGround() Background {

	gr := Ground{}
	gr = engine.NewSprite(r.GroundSprite, r.GroundSpec)

	gr.Sprite = engine.Sprite{
		Tag:  "ground",
		Pos:  engine.V(0, ScreenHeight-110),
		Size: engine.V(ScreenWidth, tankblaster.ScreenHeight),
	}

	return gr
}
