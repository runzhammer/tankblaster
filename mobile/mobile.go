package mobile

import (
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/mobile"
	"github.com/runzhammer/gamedemo/pkg/tankblaster"
)

func init() {
	mobile.SetGame(&lazyGame{})
}

// Dummy forces gomobile to compile this package.
func Dummy() {}

type lazyGame struct {
	game ebiten.Game
	err  error
}

func (g *lazyGame) Update() error {
	if err := g.ensureGame(); err != nil {
		return err
	}
	return g.game.Update()
}

func (g *lazyGame) Draw(screen *ebiten.Image) {
	if g.game == nil {
		screen.Fill(color.Black)
		return
	}
	g.game.Draw(screen)
}

func (g *lazyGame) Layout(outsideWidth, outsideHeight int) (int, int) {
	if g.game == nil {
		return tankblaster.ScreenWidth, tankblaster.ScreenHeight
	}
	return g.game.Layout(outsideWidth, outsideHeight)
}

func (g *lazyGame) ensureGame() error {
	if g.game != nil || g.err != nil {
		return g.err
	}
	log.Printf("tankblaster mobile: creating game")
	g.game, g.err = tankblaster.NewGame()
	if g.err != nil {
		log.Printf("tankblaster mobile: creating game failed: %v", g.err)
		return g.err
	}
	log.Printf("tankblaster mobile: game created")
	return nil
}
