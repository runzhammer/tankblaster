package main

import (
	"bytes"
	"image"
	"image/png"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/runzhammer/tankblaster/pkg/core"
	"github.com/runzhammer/tankblaster/pkg/tankblaster"
	r "github.com/runzhammer/tankblaster/resources"
)

func main() {
	applyRuntimeConfig()

	ebiten.SetWindowSize(int((*core.Config()).Screen.Width), int((*core.Config()).Screen.Height))
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowTitle("Tank Blaster")
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

func applyRuntimeConfig() {
	cfg := core.Config()
	if value := os.Getenv("TANKBLASTER_DEBUG_ENABLED"); value != "" {
		switch strings.ToLower(strings.TrimSpace(value)) {
		case "1", "true", "yes", "on":
			cfg.Debug.Enabled = true
		case "0", "false", "no", "off":
			cfg.Debug.Enabled = false
		}
	}
	if value := strings.TrimSpace(os.Getenv("TANKBLASTER_DEBUG_START_SCENE")); value != "" {
		cfg.Debug.StartScene = value
	}
	if value := strings.TrimSpace(os.Getenv("TANKBLASTER_DEBUG_ROUNDS")); value != "" {
		if rounds, err := strconv.Atoi(value); err == nil {
			cfg.Debug.Game.Rounds = rounds
		}
	}
	if value := strings.TrimSpace(os.Getenv("TANKBLASTER_ONLINE_SERVER_URL")); value != "" {
		cfg.Online.ServerURL = value
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
