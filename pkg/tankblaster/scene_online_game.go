package tankblaster

import (
	"image/color"

	"github.com/runzhammer/tankblaster/pkg/core"
	"github.com/runzhammer/tankblaster/pkg/gamecore"
	"github.com/runzhammer/tankblaster/pkg/tankblaster/computerplayers"
)

func NewOnlineGameScene(game *GameLoop, client *onlineClient, state gamecore.MatchState, autoPlay bool) (core.Scene, error) {
	game.online = &onlineGameRuntime{
		client:                   client,
		playerID:                 client.id.PlayerID,
		state:                    state,
		autoPlay:                 autoPlay,
		receivedCommandSequences: make(map[string]int),
	}
	game.rounds = normalizedOnlineRounds(state.TotalRounds)
	game.players = onlinePlayerConfigs(state, client.id.PlayerID, autoPlay)
	return NewGameScene(game)
}

func onlinePlayerConfigs(state gamecore.MatchState, localPlayerID string, autoPlay bool) []PlayerConfig {
	players := make([]PlayerConfig, 0, len(state.Players))
	for i, player := range state.Players {
		name := player.DisplayName
		if name == "" {
			name = "Player"
		}
		kind := PlayerHuman
		if autoPlay && player.ID == localPlayerID {
			kind = PlayerComputer
		}
		players = append(players, PlayerConfig{
			Kind:       kind,
			ComputerID: computerIDForOnlineAutoPlayer(i),
			Name:       name,
			Color:      onlinePlayerColor(i),
		})
	}
	return players
}

func computerIDForOnlineAutoPlayer(index int) computerplayers.ID {
	ids := computerplayers.IDs()
	if len(ids) == 0 {
		return computerplayers.DoedelID
	}
	return ids[index%len(ids)]
}

func onlinePlayerColor(index int) color.RGBA {
	return defaultTankColor(index)
}
