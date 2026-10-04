//go:build (darwin || freebsd || linux || windows) && !android && !ios
// +build darwin freebsd linux windows
// +build !android
// +build !ios

package core

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

func (g *GameSceneLoop) updatePlatformWindow() {
	g.enforceWindowAspectRatio()
}

func (g *GameSceneLoop) platformLayout(outsideWidth, outsideHeight int) (int, int) {
	return int((*Config()).Screen.Width), int((*Config()).Screen.Height)
}

func (g *GameSceneLoop) drawPlatformScene(screen *ebiten.Image) {
	if g.scene != nil {
		g.scene.Draw(screen)
	}
}

func GameViewport() image.Rectangle {
	return image.Rect(0, 0, int(Config().Screen.Width), int(Config().Screen.Height))
}

func (g *GameSceneLoop) enforceWindowAspectRatio() {
	if ebiten.IsFullscreen() {
		return
	}
	width, height := ebiten.WindowSize()
	if width <= 0 || height <= 0 {
		return
	}
	if g.windowWidth == 0 || g.windowHeight == 0 {
		g.windowWidth = width
		g.windowHeight = height
		return
	}
	if g.resizingWindow {
		g.windowWidth = width
		g.windowHeight = height
		g.resizingWindow = false
		return
	}
	if width == g.windowWidth && height == g.windowHeight {
		return
	}

	targetWidth := width
	targetHeight := height
	widthDelta := absInt(width - g.windowWidth)
	heightDelta := absInt(height - g.windowHeight)
	if widthDelta >= heightDelta {
		targetHeight = maxInt(1, int(float64(width)*Config().Screen.Height/Config().Screen.Width+0.5))
	} else {
		targetWidth = maxInt(1, int(float64(height)*Config().Screen.Width/Config().Screen.Height+0.5))
	}
	if targetWidth != width || targetHeight != height {
		g.resizingWindow = true
		ebiten.SetWindowSize(targetWidth, targetHeight)
	}
	g.windowWidth = targetWidth
	g.windowHeight = targetHeight
}

func absInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

func maxInt(left, right int) int {
	if left > right {
		return left
	}
	return right
}
