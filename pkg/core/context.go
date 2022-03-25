package core

import (
	"github.com/hajimehoshi/ebiten/v2/audio"
)

type Context interface {
	AudioContext() *audio.Context
	Muted() bool
}
