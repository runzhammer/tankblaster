package models

import "github.com/runzhammer/gamedemo/pkg/engine"

type Model interface {
	*engine.Vec
	float64
	*engine.Sprite
	setPosition()
}
