package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	g "github.com/runzhammer/gamedemo/game"
	m "github.com/runzhammer/gamedemo/models"
)

func main() {
	game := g.Game{Tick: 1}
	t1 := m.NewTank(&game.Tick, m.Position{X: 0, Y: 0})
	t1.Name = "Player 1"
	t2 := m.NewTank(&game.Tick, m.Position{X: 0, Y: 64})
	t2.Name = "Player 2"
	game.Tanks = append(game.Tanks, &t1, &t2)

	ebiten.SetWindowSize(1024, 768)
	ebiten.SetWindowTitle("Tanks!")
	if err := ebiten.RunGame(&game); err != nil {
		log.Fatal(err)
	}
}
