package models

import (
	_ "embed"

	"github.com/runzhammer/gamedemo/pkg/core"
	"github.com/runzhammer/gamedemo/pkg/engine"
	r "github.com/runzhammer/gamedemo/resources"
)

type Ground struct {
	Name     string
	Sprites  *engine.Sprites
	Position *engine.Vec
	Size     *engine.Vec
}

func NewGround() Ground {

	m := Ground{Name: "ground"}
	m.Sprites = engine.NewSprites()
	m.Position = &engine.Vec{X: 0, Y: core.Config().Screen.Height - 110}
	m.Size = &engine.Vec{X: core.Config().Screen.Width, Y: 110}

	groundSprite := engine.NewSprite(r.GroundSprite, r.GroundSpec)
	groundSprite.Tag = m.Name
	groundSprite.Pos = m.Position
	groundSprite.Size = m.Size

	m.Sprites.Add(groundSprite)

	return m
}
