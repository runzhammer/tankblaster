package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	m "github.com/runzhammer/gamedemo/models"
)

var ScreenWidth, ScreenHeight int = 1024, 768

type Game struct {
	Tick       uint64
	Background m.Background
	Ground     m.Ground
	Tanks      []*m.Tank
}

func (g *Game) Update() error {
	g.Tick++
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {

	// background
	screen.DrawImage(g.Background.Place(m.Position{}))

	// ground
	screen.DrawImage(g.Ground.Place(m.Position{}))

	// ebitenutil.DebugPrint(screen, "Hello, World!")
	// screen.DrawImage(t.Sprite.Image, nil)
	for _, t := range g.Tanks {
		screen.DrawImage(t.Place(m.Position{}))
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return ScreenWidth, ScreenHeight
}
