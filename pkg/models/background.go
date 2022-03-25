package models

import (
	_ "embed"

	"github.com/runzhammer/gamedemo/pkg/engine"
	r "github.com/runzhammer/gamedemo/resources"
)

type Background struct {
	Name   string
	Sprite engine.Sprite
}

func NewBackground() Background {

	b := Background{}
	s := engine.NewSprite(r.BackgroundSprite, r.BackgroundSpec)

	b.Sprite = engine.Sprite{
		Tag:  "background",
		Pos:  engine.V(0, 0),
		Size: engine.V(ScreenWidth, ScreenHeight),
	}

	return b
}
