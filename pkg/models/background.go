package models

import (
	_ "embed"

	"github.com/runzhammer/gamedemo/pkg/core"
	"github.com/runzhammer/gamedemo/pkg/engine"
	r "github.com/runzhammer/gamedemo/resources"
)

type Background struct {
	Name     string
	Sprite   *engine.Sprite
	Position engine.Vec
	Size     engine.Vec
}

func NewBackground() Background {

	m := Background{}
	m.Position = engine.Vec{X: 0, Y: 0}
	m.Size = engine.Vec{X: core.Config().Screen.Width, Y: core.Config().Screen.Height}

	m.Sprite = engine.NewSprite(r.BackgroundSprite, r.BackgroundSpec)
	m.Sprite.Tag = "background"
	m.Sprite.Pos = &m.Position
	m.Sprite.Size = &m.Size

	return m
}

func (m *Background) GetSprites() []*engine.Sprite {
	return []*engine.Sprite{m.Sprite}
}
