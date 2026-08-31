package core

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

type Game interface {
	Scene
	OnMuted(muted bool)
	Close() error
	Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int)
}

var _ Game = (*GameSceneLoop)(nil)

type GameSceneLoop struct {
	scene          Scene
	windowWidth    int
	windowHeight   int
	resizingWindow bool
	sceneCanvas    *ebiten.Image
}

type MobileOverlayDrawer interface {
	DrawMobileOverlay(screen *ebiten.Image, viewport image.Rectangle)
}

func (g *GameSceneLoop) SetScene(scene Scene) error {
	g.scene = scene
	return nil
}

func (g *GameSceneLoop) Update() error {
	g.updatePlatformWindow()
	if g.scene != nil {
		if err := g.scene.Update(); err != nil {
			return err
		}
	}
	return nil
}

func (g *GameSceneLoop) Draw(screen *ebiten.Image) {
	g.drawPlatformScene(screen)
}

func (g *GameSceneLoop) OnMuted(muted bool) {}
func (g *GameSceneLoop) Close() error       { return nil }

func (g *GameSceneLoop) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return g.platformLayout(outsideWidth, outsideHeight)
}
