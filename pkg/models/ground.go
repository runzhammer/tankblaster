package models

import (
	_ "embed"

	"github.com/runzhammer/gamedemo/pkg/core"
	"github.com/runzhammer/gamedemo/pkg/engine"
	r "github.com/runzhammer/gamedemo/resources"
)

type Ground struct {
	Name     string
	Sprite   *engine.Sprite
	Position engine.Vec
	Size     engine.Vec
}

func NewGround() Ground {

	m := Ground{Name: "ground"}
	m.Position = engine.Vec{X: 0, Y: core.Config().Screen.Height - 110}
	m.Size = engine.Vec{X: core.Config().Screen.Width, Y: 110}

	m.Sprite = engine.NewSprite(r.GroundSprite, r.GroundSpec)
	m.Sprite.Tag = m.Name
	m.Sprite.Pos = &m.Position
	m.Sprite.Size = &m.Size

	return m
}

func (m *Ground) GetSprites() []*engine.Sprite {
	return []*engine.Sprite{m.Sprite}
}
