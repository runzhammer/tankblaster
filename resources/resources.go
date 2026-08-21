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

//go:embed zero_power_smoke.gif
var ZeroPowerSmokeGIF []byte

//go:embed fonts/DejaVuSansMono.ttf
var DejaVuSansMono []byte

//go:embed store_background.png
var StoreBackground []byte

//go:embed store_main_left.png
var StoreMainLeft []byte

//go:embed store_main_right.png
var StoreMainRight []byte

//go:embed store_roll.png
var StoreRoll []byte

//go:embed store_icons.png
var StoreIcons []byte

//go:embed taining-ammo.png
var TrainingAmmoSelected []byte

//go:embed taining-ammo-deslected.png
var TrainingAmmoDeselected []byte
