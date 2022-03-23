package core

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type Game interface {
	Scene
	OnMuted(muted bool)
	Close() error
}

var _ Game = (*GameSceneLoop)(nil)

type GameSceneLoop struct {
	scene Scene
}

func (g *GameSceneLoop) SetScene(scene Scene) error {
	DebugLog("new scene: %T", scene)
	g.scene = scene
	return nil
}

func (g *GameSceneLoop) Update() error {
	if g.scene != nil {
		if err := g.scene.Update(); err != nil {
			return err
		}
	}
	return nil
}

func (g *GameSceneLoop) Draw(screen *ebiten.Image) {
	if g.scene != nil {
		g.scene.Draw(screen)
	}
}

func (g *GameSceneLoop) OnMuted(muted bool) {}
func (g *GameSceneLoop) Close() error       { return nil }
