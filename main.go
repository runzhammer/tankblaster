package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	g "github.com/runzhammer/gamedemo/game"
)

func main() {
	ebiten.SetWindowSize(1024, 768)
	ebiten.SetWindowTitle("Tanks!")
	if err := ebiten.RunGame(&g.Game{}); err != nil {
		log.Fatal(err)
	}
}
