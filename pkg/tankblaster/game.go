package tankblaster

import (
	"errors"

	"github.com/runzhammer/gamedemo/pkg/core"
	"github.com/runzhammer/gamedemo/pkg/engine"
)

var _ core.Game = (*GameLoop)(nil)

const (
	Title     = "Tank Blaster 3.0"
	bgmVolume = 0.5

	ScreenWidth  = 1024
	ScreenHeight = 768

	AudioSampleRate = 44100
)

var (
	ScreenBounds = engine.R(0, 0, ScreenWidth, ScreenHeight)

	RegularTermination = errors.New("goodbye")
	GameTermination    = errors.New("game over")
)

type GameLoop struct {
	core.GameSceneLoop
	context core.Context

	redScore  int
	blueScore int

	// bgm *audio.Player
}

func NewGame() (core.Game, error) {
	// bgm, err := context.Loader().AudioLoop(context.AudioContext(), "mp3", "music/octane.mp3")
	// if err != nil {
	// 	return nil, err
	// }

	// game := &Game{
	// 	context: context,
	// 	bgm: bgm,
	// }
	game := &GameLoop{}

	if err := game.SetNewScene(NewGameScene); err != nil {
		return nil, err
	}

	// bgm.SetVolume(bgmVolume)
	// bgm.Play()

	return game, nil
}

func (g *GameLoop) SetNewScene(factory func(*GameLoop) (scene core.Scene, err error)) error {
	scene, err := factory(g)
	if err != nil {
		return err
	}
	return g.SetScene(scene)
}

func (g *GameLoop) OnMuted(muted bool) {
	// if muted {
	// 	g.bgm.SetVolume(0)
	// } else {
	// 	g.bgm.SetVolume(bgmVolume)
	// }
}

func (g *GameLoop) Close() error {
	// return g.bgm.Close()
	return nil
}
