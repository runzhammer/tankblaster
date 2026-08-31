package tankblaster

import (
	"sync"
	"sync/atomic"
)

var playerNameInputActive atomic.Bool

type playerNameInputCommand struct {
	text      string
	replace   bool
	backspace bool
	finish    bool
}

var playerNameInputCommands struct {
	mu       sync.Mutex
	commands []playerNameInputCommand
}

func setPlayerNameInputActive(active bool) {
	playerNameInputActive.Store(active)
}

func PlayerNameInputActive() bool {
	return playerNameInputActive.Load()
}

func CommitPlayerNameText(value string) {
	if value == "" {
		return
	}
	queuePlayerNameInputCommand(playerNameInputCommand{text: value})
}

func SetPlayerNameText(value string) {
	queuePlayerNameInputCommand(playerNameInputCommand{text: value, replace: true})
}

func DeletePlayerNameText() {
	queuePlayerNameInputCommand(playerNameInputCommand{backspace: true})
}

func FinishPlayerNameInput() {
	queuePlayerNameInputCommand(playerNameInputCommand{finish: true})
}

func drainPlayerNameInputCommands() []playerNameInputCommand {
	playerNameInputCommands.mu.Lock()
	defer playerNameInputCommands.mu.Unlock()
	if len(playerNameInputCommands.commands) == 0 {
		return nil
	}
	commands := append([]playerNameInputCommand(nil), playerNameInputCommands.commands...)
	playerNameInputCommands.commands = playerNameInputCommands.commands[:0]
	return commands
}

func queuePlayerNameInputCommand(command playerNameInputCommand) {
	playerNameInputCommands.mu.Lock()
	defer playerNameInputCommands.mu.Unlock()
	playerNameInputCommands.commands = append(playerNameInputCommands.commands, command)
}
