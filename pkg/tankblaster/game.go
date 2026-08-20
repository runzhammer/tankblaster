package tankblaster

import (
	"errors"
	"image/color"
	"strconv"
	"strings"

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

type PlayerKind uint8

const (
	PlayerNone PlayerKind = iota
	PlayerHuman
	PlayerComputer
)

type PlayerConfig struct {
	Kind  PlayerKind
	Name  string
	Color color.RGBA
}

type GameLoop struct {
	core.GameSceneLoop
	context core.Context

	redScore  int
	blueScore int

	rounds  int
	players []PlayerConfig

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
	game := &GameLoop{
		rounds: 10,
	}

	if core.Config().Debug.Enabled {
		game.applyDebugConfig()
		if err := game.SetNewScene(debugStartSceneFactory(core.Config().Debug.StartScene)); err != nil {
			return nil, err
		}
		return game, nil
	}

	if err := game.SetNewScene(NewPlayerSelectionScene); err != nil {
		return nil, err
	}

	// bgm.SetVolume(bgmVolume)
	// bgm.Play()

	return game, nil
}

func (g *GameLoop) applyDebugConfig() {
	cfg := core.Config().Debug.Game
	g.rounds = cfg.Rounds
	g.players = debugPlayers(cfg.Players)
}

func debugPlayers(configured []core.DebugPlayerSettings) []PlayerConfig {
	players := make([]PlayerConfig, 0, len(configured))
	for i, player := range configured {
		kind := debugPlayerKind(player.Kind)
		if kind == PlayerNone {
			continue
		}
		name := player.Name
		if name == "" {
			name = "Spieler " + strconv.Itoa(len(players)+1)
		}
		players = append(players, PlayerConfig{
			Kind:  kind,
			Name:  name,
			Color: player.Color.RGBA(defaultTankColor(i)),
		})
	}
	if len(players) >= 2 {
		return players
	}
	return []PlayerConfig{
		{Kind: PlayerHuman, Name: "Spieler 1", Color: defaultTankColor(0)},
		{Kind: PlayerHuman, Name: "Spieler 2", Color: defaultTankColor(1)},
	}
}

func debugPlayerKind(value string) PlayerKind {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "human", "player", "spieler", "mensch":
		return PlayerHuman
	case "computer", "cpu", "ki":
		return PlayerComputer
	default:
		return PlayerNone
	}
}

func debugStartSceneFactory(scene string) func(*GameLoop) (core.Scene, error) {
	switch strings.ToLower(strings.TrimSpace(scene)) {
	case "title", "title_scene":
		return NewTitleScene
	case "player_selection", "selection", "spieler_auswahl":
		return NewPlayerSelectionScene
	case "game", "battle", "spielmodus", "":
		return NewGameScene
	default:
		return NewGameScene
	}
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
