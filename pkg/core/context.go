package core

import (
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/runzhammer/gamedemo/pkg/tempura"
)

type Context interface {
	Loader() tempura.Loader
	AudioContext() *audio.Context
	Muted() bool
}
