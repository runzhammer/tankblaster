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

//go:embed player_computer_doedel.png
var PlayerComputerDoedel []byte

//go:embed player_computer_frederik.png
var PlayerComputerFrederik []byte

//go:embed player_computer_mister_x.png
var PlayerComputerMisterX []byte

//go:embed player_computer_dr_nuke.png
var PlayerComputerDrNuke []byte

//go:embed player_computer_harald.png
var PlayerComputerHarald []byte

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

//go:embed palm.png
var Palm []byte

//go:embed cloud_lightning.png
var CloudLightning []byte

//go:embed cloud_1.png
var Cloud1 []byte

//go:embed cloud_2.png
var Cloud2 []byte

//go:embed cloud_3.png
var Cloud3 []byte

//go:embed cloud_4.png
var Cloud4 []byte

//go:embed cloud_5.png
var Cloud5 []byte
