package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	m "github.com/runzhammer/gamedemo/models"
)

type Game struct {
	Tick  uint64
	Tanks []*m.Tank
}

func (g *Game) Update() error {
	g.Tick++
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// ebitenutil.DebugPrint(screen, "Hello, World!")
	// screen.DrawImage(t.Sprite.Image, nil)
	for _, t := range g.Tanks {
		screen.DrawImage(t.Place(m.Position{}))
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 1024, 768
}
