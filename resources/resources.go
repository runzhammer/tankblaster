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

//go:embed zero_power_dust_explosion.png
var ZeroPowerDustExplosionPNG []byte

//go:embed zero_power_explosion.png
var ZeroPowerExplosionPNG []byte

//go:embed zero_power_mushroom_explosion.png
var ZeroPowerMushroomExplosionPNG []byte

//go:embed zero_power_player_smoke.png
var ZeroPowerPlayerSmokePNG []byte

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

//go:embed weaponbar_active.png
var WeaponbarActive []byte

//go:embed weaponbar_onstock.png
var WeaponbarOnStock []byte

//go:embed weaponbar_outofstock.png
var WeaponbarOutOfStock []byte
