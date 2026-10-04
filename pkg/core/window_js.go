//go:build js

package core

import (
	"image"
	"image/color"
	"strings"
	"syscall/js"

	"github.com/hajimehoshi/ebiten/v2"
)

var webGameViewport image.Rectangle

func (g *GameSceneLoop) updatePlatformWindow() {}

func (g *GameSceneLoop) platformLayout(outsideWidth, outsideHeight int) (int, int) {
	if webMobileLayout() {
		return g.mobilePlatformLayout(outsideWidth, outsideHeight)
	}
	return int((*Config()).Screen.Width), int((*Config()).Screen.Height)
}

func (g *GameSceneLoop) drawPlatformScene(screen *ebiten.Image) {
	if !webMobileLayout() {
		if g.scene != nil {
			g.scene.Draw(screen)
		}
		return
	}
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

func (g *GameSceneLoop) mobilePlatformLayout(outsideWidth, outsideHeight int) (int, int) {
	cfgW := int(Config().Screen.Width)
	cfgH := int(Config().Screen.Height)
	if outsideWidth <= 0 || outsideHeight <= 0 {
		webGameViewport = image.Rect(0, 0, cfgW, cfgH)
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
	webGameViewport = image.Rect(x, y, x+cfgW, y+cfgH)
	return screenW, screenH
}

func GameViewport() image.Rectangle {
	if !webMobileLayout() || webGameViewport.Empty() {
		return image.Rect(0, 0, int(Config().Screen.Width), int(Config().Screen.Height))
	}
	return webGameViewport
}

func webMobileLayout() bool {
	search := js.Global().Get("location").Get("search").String()
	if strings.Contains(search, "desktop=1") {
		return false
	}
	if strings.Contains(search, "mobile=1") {
		return true
	}

	nav := js.Global().Get("navigator")
	maxTouchPoints := nav.Get("maxTouchPoints").Int()
	ua := strings.ToLower(nav.Get("userAgent").String())
	mobileUA := strings.Contains(ua, "android") ||
		strings.Contains(ua, "iphone") ||
		strings.Contains(ua, "ipad") ||
		strings.Contains(ua, "ipod") ||
		strings.Contains(ua, "mobile")
	coarsePointer := false
	if matchMedia := js.Global().Get("matchMedia"); matchMedia.Type() == js.TypeFunction {
		coarsePointer = matchMedia.Invoke("(pointer: coarse)").Get("matches").Bool()
	}
	return maxTouchPoints > 0 && (mobileUA || coarsePointer)
}
