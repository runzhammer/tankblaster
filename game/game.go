package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/runzhammer/gamedemo/models"
	"github.com/runzhammer/gamedemo/pkg/core"
)

var ScreenWidth, ScreenHeight int = 1024, 768

const (
	Title     = "Tank Blaster 3.0"
	bgmVolume = 0.5
)

type Game struct {
	
	core.GameSceneLoop
	context core.Context

	Background models.Background
	Ground     models.Ground
	Tanks      []*models.Tank
}

func (g *Game) Update() error {
	g.Tick++
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {

	// background
	screen.DrawImage(g.Background.Place(models.Position{}))

	// ground
	screen.DrawImage(g.Ground.Place(models.Position{}))

	// ebitenutil.DebugPrint(screen, "Hello, World!")
	// screen.DrawImage(t.Sprite.Image, nil)
	for _, t := range g.Tanks {
		screen.DrawImage(t.Place(models.Position{}))
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return ScreenWidth, ScreenHeight
}
