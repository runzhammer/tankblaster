//go:build (darwin || freebsd || linux || windows) && !android && !ios
// +build darwin freebsd linux windows
// +build !android
// +build !ios

package tankblaster

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

func (s *GameScene) handleMobileSideControls(tank *battleTank, strengthStep int) {}

func (s *GameScene) handleMobileXMV12SideControls(tank *battleTank) {}

func (s *GameScene) drawMobileSideControls(screen *ebiten.Image, viewport image.Rectangle) {}
