package tankblaster

import (
	"errors"
	"image/color"
	"log"
	"strconv"
	"strings"

	"github.com/runzhammer/tankblaster/pkg/buildinfo"
	"github.com/runzhammer/tankblaster/pkg/core"
	"github.com/runzhammer/tankblaster/pkg/engine"
	"github.com/runzhammer/tankblaster/pkg/gamecore"
	"github.com/runzhammer/tankblaster/pkg/tankblaster/computerplayers"
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
	Kind       PlayerKind
	ComputerID computerplayers.ID
	Name       string
	Color      color.RGBA
}

type GameLoop struct {
	core.GameSceneLoop
	context core.Context
	sounds  *soundPlayer
	muted   bool

	redScore  int
	blueScore int

	rounds  int
	players []PlayerConfig
	options gameOptions
	online  *onlineGameRuntime

	// bgm *audio.Player
}

type onlineGameRuntime struct {
	client                   *onlineClient
	playerID                 string
	controlledPlayerIDs      map[string]bool
	state                    gamecore.MatchState
	autoPlay                 bool
	applyingRemoteCmd        bool
	turnSequence             int
	commandSequence          int
	receivedCommandSequences map[string]int
	lastServerSequence       int64
	awaitingCatchUp          bool
	lastAim                  onlineAimState
	matchResultSent          bool
}

type onlineAimState struct {
	valid          bool
	playerIndex    int
	turnSequence   int
	weaponSlot     int
	shotStrength   int
	cannonRotation float64
	cameraX        float64
}

type gameOptions struct {
	projectileReentry int
	palmCount         int
	cloudAggression   int
	quickRoundStart   bool
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
		rounds:  10,
		options: defaultGameOptions(),
		sounds:  newSoundPlayer(),
	}
	log.Printf("tankblaster version: %s", buildinfo.String())
	if err := game.loadUserConfig(); err != nil {
		log.Printf("load %s: %v", userConfigFileName, err)
	}

	if core.Config().Debug.Enabled {
		game.applyDebugConfig()
		factory := debugStartSceneFactory(core.Config().Debug.StartScene)
		if core.Config().Debug.StartShop {
			factory = NewDebugShopScene
		}
		if err := game.SetNewScene(factory); err != nil {
			return nil, err
		}
		return game, nil
	}

	if err := game.SetNewScene(NewIntroScene); err != nil {
		return nil, err
	}

	// bgm.SetVolume(bgmVolume)
	// bgm.Play()

	return game, nil
}

func defaultGameOptions() gameOptions {
	cfg := core.Config().Gameplay
	return gameOptions{
		projectileReentry: cfg.ProjectileReentry,
		palmCount:         cfg.PalmCount,
		cloudAggression:   cfg.CloudAggression,
		quickRoundStart:   cfg.QuickRoundStart,
	}
}

func (g *GameLoop) applyDebugConfig() {
	debug := core.Config().Debug
	cfg := debug.Game
	g.rounds = cfg.Rounds
	g.players = debugPlayersForMode(debug.Mode, cfg.Players)
}

func debugPlayersForMode(mode string, configured []core.DebugPlayerSettings) []PlayerConfig {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "two_humans", "2_humans", "human_vs_human", "menschen":
		return []PlayerConfig{
			{Kind: PlayerHuman, Name: "Spieler 1", Color: defaultTankColor(0)},
			{Kind: PlayerHuman, Name: "Spieler 2", Color: defaultTankColor(1)},
		}
	default:
		return debugPlayers(configured)
	}
}

func debugPlayers(configured []core.DebugPlayerSettings) []PlayerConfig {
	players := make([]PlayerConfig, 0, len(configured))
	for i, player := range configured {
		kind := debugPlayerKind(player.Kind)
		if kind == PlayerNone {
			continue
		}
		computerID := player.ComputerID
		if computerID == 0 {
			computerID = player.ID
		}
		name := player.Name
		if name == "" {
			name = "Spieler " + strconv.Itoa(len(players)+1)
			if kind == PlayerComputer {
				name = computerplayers.Name(computerplayers.ID(computerID))
			}
		}
		players = append(players, PlayerConfig{
			Kind:       kind,
			ComputerID: computerplayers.ID(computerID),
			Name:       name,
			Color:      player.Color.RGBA(defaultTankColor(i)),
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
	case "intro", "intro_scene", "vorspann":
		return NewIntroScene
	case "player_selection", "selection", "spieler_auswahl":
		return NewPlayerSelectionScene
	case "game", "battle", "spielmodus", "":
		return NewGameScene
	case "shop", "store", "laden":
		return NewDebugShopScene
	case "score", "scores", "hall_of_fame", "hall-of-fame", "bestenliste":
		return NewDebugHallOfFameScene
	case "online", "onlinespiel", "online_game":
		return NewEmbeddedOnlinePlayerSelectionScene
	default:
		return NewGameScene
	}
}

func (g *GameLoop) SetNewScene(factory func(*GameLoop) (scene core.Scene, err error)) error {
	setPlayerNameInputActive(false)
	scene, err := factory(g)
	if err != nil {
		return err
	}
	if g.sounds != nil {
		g.sounds.StopAllLoops()
	}
	return g.SetScene(scene)
}

func (g *GameLoop) OnMuted(muted bool) {
	g.muted = muted
	if muted && g.sounds != nil {
		g.sounds.StopAllLoops()
	}
	// if muted {
	// 	g.bgm.SetVolume(0)
	// } else {
	// 	g.bgm.SetVolume(bgmVolume)
	// }
}

func (g *GameLoop) Close() error {
	if g.sounds != nil {
		g.sounds.StopAllLoops()
	}
	// return g.bgm.Close()
	return nil
}
