package resources

import _ "embed"

//go:embed config.yaml
var GameConfig []byte

//go:embed tank.png
var TankSprite []byte

//go:embed tank.yaml
var TankSpec []byte

//go:embed cannon.png
var CannonSprite []byte

//go:embed cannon.yaml
var CannonSpec []byte

//go:embed background.png
var BackgroundSprite []byte

//go:embed background.yaml
var BackgroundSpec []byte

//go:embed ground.png
var GroundSprite []byte

//go:embed ground.yaml
var GroundSpec []byte

//go:embed player_selection_base.png
var PlayerSelectionBase []byte

//go:embed player_human.png
var PlayerHuman []byte

//go:embed player_computer.png
var PlayerComputer []byte
