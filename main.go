package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	g "github.com/runzhammer/gamedemo/game"
)

func main() {

	ebiten.SetWindowSize(g.ScreenWidth, g.ScreenHeight)
	ebiten.SetWindowTitle("Tank Blaster 3.0")

	if err := ebiten.RunGame(&game); err != nil {
		log.Fatal(err)
	}
}
