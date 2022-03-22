package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	g "github.com/runzhammer/gamedemo/game"
	m "github.com/runzhammer/gamedemo/models"
)

func main() {

	game := g.Game{Tick: 1}

	b := m.NewBackground(&game.Tick, m.Position{X: 0, Y: 0})
	gr := m.NewGround(&game.Tick, m.Position{X: 0, Y: float64(g.ScreenHeight - 110)})

	t1 := m.NewTank(&game.Tick, m.Position{X: 128, Y: float64(g.ScreenHeight) - 160})
	t1.Name = "Player 1"
	t1.Options.ColorM.RotateHue(140)

	t2 := m.NewTank(&game.Tick, m.Position{X: float64(g.ScreenWidth) - 64 - 128, Y: float64(g.ScreenHeight) - 160})
	t2.Name = "Player 2"
	t2.Options.ColorM.RotateHue(80)

	game.Tanks = append(game.Tanks, &t1, &t2)
	game.Background = b
	game.Ground = gr

	ebiten.SetWindowSize(g.ScreenWidth, g.ScreenHeight)
	ebiten.SetWindowTitle("Tanks!")

	if err := ebiten.RunGame(&game); err != nil {
		log.Fatal(err)
	}
}
