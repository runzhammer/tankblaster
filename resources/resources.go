package resources

import "embed"

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

//go:embed shop_entry.png
var ShopEntry []byte

//go:embed shop_class_a.png
var ShopClassA []byte

//go:embed shop_class_b.png
var ShopClassB []byte

//go:embed taining-ammo.png
var TrainingAmmoSelected []byte

//go:embed taining-ammo-deslected.png
var TrainingAmmoDeselected []byte

//go:embed class-a-*.png class-b.png
var ShopListScreens embed.FS
