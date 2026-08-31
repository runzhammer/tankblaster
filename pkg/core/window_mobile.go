//go:build android || ios
// +build android ios

package core

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

var mobileGameViewport image.Rectangle

func (g *GameSceneLoop) updatePlatformWindow() {}

func (g *GameSceneLoop) platformLayout(outsideWidth, outsideHeight int) (int, int) {
	cfgW := int(Config().Screen.Width)
	cfgH := int(Config().Screen.Height)
	if outsideWidth <= 0 || outsideHeight <= 0 {
		mobileGameViewport = image.Rect(0, 0, cfgW, cfgH)
		return cfgW, cfgH
	}

	outsideAspect := float64(outsideWidth) / float64(outsideHeight)
	gameAspect := float64(cfgW) / float64(cfgH)
	screenW, screenH := cfgW, cfgH
	if outsideAspect > gameAspect {
		screenW = int(float64(cfgH)*outsideAspect + 0.5)
		if screenW < cfgW {
			screenW = cfgW
		}
	} else if outsideAspect < gameAspect {
		screenH = int(float64(cfgW)/outsideAspect + 0.5)
		if screenH < cfgH {
			screenH = cfgH
		}
	}
	x := (screenW - cfgW) / 2
	y := (screenH - cfgH) / 2
	mobileGameViewport = image.Rect(x, y, x+cfgW, y+cfgH)
	return screenW, screenH
}

func (g *GameSceneLoop) drawPlatformScene(screen *ebiten.Image) {
	if g.scene == nil {
		return
	}
	cfgW := int(Config().Screen.Width)
	cfgH := int(Config().Screen.Height)
	if g.sceneCanvas == nil || g.sceneCanvas.Bounds().Dx() != cfgW || g.sceneCanvas.Bounds().Dy() != cfgH {
		g.sceneCanvas = ebiten.NewImage(cfgW, cfgH)
	}
	g.sceneCanvas.Clear()
	g.scene.Draw(g.sceneCanvas)

	screen.Fill(color.Black)
	viewport := GameViewport()
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(float64(viewport.Dx())/float64(cfgW), float64(viewport.Dy())/float64(cfgH))
	op.GeoM.Translate(float64(viewport.Min.X), float64(viewport.Min.Y))
	op.Filter = ebiten.FilterNearest
	screen.DrawImage(g.sceneCanvas, op)

	if drawer, ok := g.scene.(MobileOverlayDrawer); ok {
		drawer.DrawMobileOverlay(screen, viewport)
	}
}

func GameViewport() image.Rectangle {
	if mobileGameViewport.Empty() {
		return image.Rect(0, 0, int(Config().Screen.Width), int(Config().Screen.Height))
	}
	return mobileGameViewport
}
