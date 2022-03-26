package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/runzhammer/gamedemo/pkg/core"
	"github.com/runzhammer/gamedemo/pkg/tankblaster"
)

func main() {

	ebiten.SetWindowSize(int(core.Config().Screen.Width), int(core.Config().Screen.Height))
	ebiten.SetWindowTitle("Tank Blaster 3.0")

	var err error
	var tankblasterGame core.Game

	if tankblasterGame, err = tankblaster.NewGame(); err != nil {
		log.Fatal(err)
	}

	if err = ebiten.RunGame(tankblasterGame); err != nil {
		log.Fatal(err)
	}
}
