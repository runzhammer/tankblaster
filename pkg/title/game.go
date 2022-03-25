package title

import (
	"github.com/runzhammer/gamedemo/pkg/core"
	"github.com/runzhammer/gamedemo/pkg/engine"
)

var _ core.Game = (*Game)(nil)

type Game struct {
	core.GameSceneLoop
	context core.Context
}

func NewGame(context core.Context) (core.Game, error) {
	if core.Debug {
		defer engine.LogStart("Title init").End()
	}
	game := &Game{
		context: context,
	}

	if err := game.SetNewScene(NewTitleScene); err != nil {
		return nil, err
	}

	return game, nil
}

func (g *Game) SetNewScene(factory func(*Game) (scene core.Scene, err error)) error {
	if core.Debug {
		defer engine.LogStart("New title scene").End()
	}
	scene, err := factory(g)
	if err != nil {
		return err
	}
	return g.SetScene(scene)
}
