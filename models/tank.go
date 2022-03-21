package models

import _ "embed"

//go:embed resources/tank.png
var sprite []byte

//go:embed resources/tank.yaml
var animations []byte

type Tank struct {
	Name   string
	Sprite Sprite
}

func NewTank() Tank {
	var t Tank
	t.Sprite.Init(sprite, animations)
	return t
}
