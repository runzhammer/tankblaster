package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	m "github.com/runzhammer/gamedemo/models"
)

type Game struct {
}

func (g *Game) Update() error {
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// ebitenutil.DebugPrint(screen, "Hello, World!")
	t := m.NewTank()
	screen.DrawImage(t.Sprite.Image, nil)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 1024, 768
}
