package models

import (
	_ "embed"

	"github.com/runzhammer/gamedemo/pkg/core"
	"github.com/runzhammer/gamedemo/pkg/engine"
	r "github.com/runzhammer/gamedemo/resources"
)

type Background struct {
	Name     string
	Sprites  *engine.Sprites
	Position *engine.Vec
	Size     *engine.Vec
}

func NewBackground() Background {
	return NewBackgroundWithWidth(core.Config().Screen.Width)
}

func NewBackgroundWithWidth(width float64) Background {
	return NewBackgroundWithSize(width, core.Config().Screen.Height)
}

func NewBackgroundWithSize(width, height float64) Background {

	m := Background{}
	m.Sprites = engine.NewSprites()
	m.Position = &engine.Vec{X: 0, Y: 0}
	m.Size = &engine.Vec{X: width, Y: height}

	backgroundSprite := engine.NewSprite(r.BackgroundSprite, r.BackgroundSpec)
	backgroundSprite.Tag = "background"
	backgroundSprite.Pos = m.Position
	backgroundSprite.Size = m.Size

	m.Sprites.Add(backgroundSprite)

	return m
}
