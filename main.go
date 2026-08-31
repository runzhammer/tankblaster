package main

import (
	"bytes"
	"image"
	"image/png"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/runzhammer/gamedemo/pkg/core"
	"github.com/runzhammer/gamedemo/pkg/tankblaster"
	r "github.com/runzhammer/gamedemo/resources"
)

func main() {

	ebiten.SetWindowSize(int((*core.Config()).Screen.Width), int((*core.Config()).Screen.Height))
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowTitle("Tank Blaster 3.0")
	ebiten.SetWindowIcon(windowIcons())

	var err error
	var tankblasterGame core.Game

	if tankblasterGame, err = tankblaster.NewGame(); err != nil {
		log.Fatal(err)
	}

	if err = ebiten.RunGame(tankblasterGame); err != nil {
		log.Fatal(err)
	}
}

func windowIcons() []image.Image {
	icon, err := png.Decode(bytes.NewReader(r.AppIconPNG))
	if err != nil {
		log.Printf("window icon: %v", err)
		return nil
	}
	return []image.Image{icon}
}
