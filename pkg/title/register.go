package title

import "github.com/runzhammer/gamedemo/pkg/games"

func init() {
	games.RegisterGameFactory("title", NewGame)
}
