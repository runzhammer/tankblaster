package mobile

import (
	"github.com/hajimehoshi/ebiten/v2/mobile"
	"github.com/runzhammer/gamedemo/pkg/tankblaster"
)

func init() {
	game, err := tankblaster.NewGame()
	if err != nil {
		panic(err)
	}
	mobile.SetGame(game)
}
