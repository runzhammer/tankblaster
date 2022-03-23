package tanks

import "github.com/runzhammer/gamedemo/pkg/games"

func init() {
	games.RegisterGameFactory("tanks", NewGame)
}
