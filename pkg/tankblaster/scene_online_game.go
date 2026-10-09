package tankblaster

import (
	"image/color"

	"github.com/runzhammer/tankblaster/pkg/core"
	"github.com/runzhammer/tankblaster/pkg/gamecore"
	"github.com/runzhammer/tankblaster/pkg/protocol"
	"github.com/runzhammer/tankblaster/pkg/tankblaster/computerplayers"
)

func NewOnlineGameScene(game *GameLoop, client *onlineClient, state gamecore.MatchState, autoPlay bool) (core.Scene, error) {
	return NewOnlineGameSceneWithControl(game, client, state, nil, nil, autoPlay)
}

func NewOnlineGameSceneWithControl(game *GameLoop, client *onlineClient, state gamecore.MatchState, controlledPlayerIDs []string, lobbySlots []protocol.LobbySlot, autoPlay bool) (core.Scene, error) {
	game.online = &onlineGameRuntime{
		client:                   client,
		playerID:                 client.id.PlayerID,
		controlledPlayerIDs:      controlledPlayerIDSet(controlledPlayerIDs),
		state:                    state,
		autoPlay:                 autoPlay,
		receivedCommandSequences: make(map[string]int),
	}
	game.rounds = normalizedOnlineRounds(state.TotalRounds)
	game.players = onlinePlayerConfigs(state, client.id.PlayerID, lobbySlots, autoPlay)
	return NewGameScene(game)
}

func controlledPlayerIDSet(ids []string) map[string]bool {
	if len(ids) == 0 {
		return nil
	}
	set := make(map[string]bool, len(ids))
	for _, id := range ids {
		if id != "" {
			set[id] = true
		}
	}
	return set
}

func onlinePlayerConfigs(state gamecore.MatchState, localPlayerID string, lobbySlots []protocol.LobbySlot, autoPlay bool) []PlayerConfig {
	players := make([]PlayerConfig, 0, len(state.Players))
	for i, player := range state.Players {
		if slot, ok := lobbySlotForPlayerID(lobbySlots, player.ID); ok {
			kind := PlayerHuman
			if slot.Kind == "computer" {
				kind = PlayerComputer
			}
			players = append(players, PlayerConfig{
				Kind:       kind,
				ComputerID: computerplayers.ID(slot.ComputerID),
				Name:       slot.Name,
				Color:      color.RGBA{R: slot.Color.R, G: slot.Color.G, B: slot.Color.B, A: slot.Color.A},
			})
			continue
		}
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

func lobbySlotForPlayerID(slots []protocol.LobbySlot, playerID string) (protocol.LobbySlot, bool) {
	for _, slot := range slots {
		if slot.PlayerID == playerID {
			return slot, true
		}
	}
	return protocol.LobbySlot{}, false
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
