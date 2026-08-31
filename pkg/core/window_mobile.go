//go:build android || ios
// +build android ios

package core

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

var mobileGameViewport image.Rectangle

func (g *GameSceneLoop) updatePlatformWindow() {}

func (g *GameSceneLoop) platformLayout(outsideWidth, outsideHeight int) (int, int) {
	cfgW := int(Config().Screen.Width)
	cfgH := int(Config().Screen.Height)
	mobileGameViewport = image.Rect(0, 0, cfgW, cfgH)
	return cfgW, cfgH
}

func (g *GameSceneLoop) drawPlatformScene(screen *ebiten.Image) {
	if g.scene == nil {
		return
	}
	g.scene.Draw(screen)
}

func GameViewport() image.Rectangle {
	if mobileGameViewport.Empty() {
		return image.Rect(0, 0, int(Config().Screen.Width), int(Config().Screen.Height))
	}
	return mobileGameViewport
}
