package resources

import (
	"embed"
	"fmt"
	"path/filepath"
	"strings"
)

//go:embed config.yaml
var GameConfig []byte

//go:embed sounds/*.wav
var Sounds embed.FS

//go:embed images/tank.png
var TankSprite []byte

//go:embed tank.yaml
var TankSpec []byte

//go:embed images/tank_small.png
var TankSmallSprite []byte

//go:embed tank_small.yaml
var TankSmallSpec []byte

//go:embed images/xm-v12-tank.png
var XMV12TankSprite []byte

//go:embed images/tankanzeige.png
var FuelGaugePNG []byte

//go:embed images/slopemeter.png
var SlopeMeterPNG []byte

//go:embed images/button-ignition.png
var ButtonIgnitionPNG []byte

//go:embed images/cannon.png
var CannonSprite []byte

//go:embed images/tankblaster.ico
var AppIconICO []byte

//go:embed images/tankblaster_icon.png
var AppIconPNG []byte

//go:embed cannon.yaml
var CannonSpec []byte

//go:embed images/background.png
var BackgroundSprite []byte

//go:embed background.yaml
var BackgroundSpec []byte

//go:embed images/ground.png
var GroundSprite []byte

//go:embed ground.yaml
var GroundSpec []byte

//go:embed images/player_selection_base.png
var PlayerSelectionBase []byte

//go:embed images/player_human.png
var PlayerHuman []byte

//go:embed images/player_computer_doedel.png
var PlayerComputerDoedel []byte

//go:embed images/player_computer_frederik.png
var PlayerComputerFrederik []byte

//go:embed images/player_computer_mister_x.png
var PlayerComputerMisterX []byte

//go:embed images/player_computer_dr_nuke.png
var PlayerComputerDrNuke []byte

//go:embed images/player_computer_harald.png
var PlayerComputerHarald []byte

//go:embed images/klecks.png
var PaintSplotchPNG []byte

//go:embed images/zero_power_dust_explosion.png
var ZeroPowerDustExplosionPNG []byte

//go:embed images/zero_power_explosion.png
var ZeroPowerExplosionPNG []byte

//go:embed images/zero_power_mushroom_explosion.png
var ZeroPowerMushroomExplosionPNG []byte

//go:embed images/zero_power_player_smoke.png
var ZeroPowerPlayerSmokePNG []byte

//go:embed fonts/DejaVuSansMono.ttf
var DejaVuSansMono []byte

//go:embed images/store_background.png
var StoreBackground []byte

//go:embed images/store_main_left.png
var StoreMainLeft []byte

//go:embed images/store_main_right.png
var StoreMainRight []byte

//go:embed images/store_roll.png
var StoreRoll []byte

//go:embed images/store_icons.png
var StoreIcons []byte

//go:embed images/weaponbar_active.png
var WeaponbarActive []byte

//go:embed images/weaponbar_onstock.png
var WeaponbarOnStock []byte

//go:embed images/weaponbar_outofstock.png
var WeaponbarOutOfStock []byte

//go:embed images/symbol_reentry.png
var SymbolReentry []byte

//go:embed images/earth_reentry.png
var EarthReentry []byte

//go:embed images/palm.png
var Palm []byte

//go:embed images/palm_leaves.png
var PalmLeavesPNG []byte

//go:embed images/palm_eyes_open.png
var PalmEyesOpenPNG []byte

//go:embed images/palm_eyes_close.png
var PalmEyesClosePNG []byte

//go:embed images/palm_scream.png
var PalmScreamPNG []byte

//go:embed images/palm_grin.png
var PalmGrinPNG []byte

//go:embed images/moskitos.png
var MoskitosPNG []byte

//go:embed images/fragezeichen.png
var FragezeichenPNG []byte

//go:embed images/fragezeichen-dialog.png
var FragezeichenDialogPNG []byte

//go:embed images/blinkboje.png
var BlinkBojePNG []byte

//go:embed images/bullet_bomb.png
var BulletBombPNG []byte

//go:embed images/lasersmoke.png
var LaserSmokePNG []byte

//go:embed images/fireball_impact.png
var FireballImpactPNG []byte

//go:embed images/palm_fire.png
var PalmFirePNG []byte

//go:embed images/palm_skeleton.png
var PalmSkeletonPNG []byte

//go:embed images/palm_smoke.png
var PalmSmokePNG []byte

//go:embed images/palm_crumble.png
var PalmCrumblePNG []byte

//go:embed images/water_texture.png
var WaterTexturePNG []byte

//go:embed images/water_blubber.png
var WaterBlubberPNG []byte

//go:embed images/water_blotch.png
var WaterBlotchPNG []byte

//go:embed images/cloud_lightning.png
var CloudLightning []byte

//go:embed images/cloud_aerger.png
var CloudAngryPNG []byte

//go:embed images/cloud_searching.png
var CloudSearchingPNG []byte

//go:embed images/cloud_aerger2grinse.png
var CloudAngryToGrinPNG []byte

//go:embed images/cloud_grinse.png
var CloudGrinPNG []byte

//go:embed images/lightning.png
var LightningPNG []byte

//go:embed images/cloud_1.png
var Cloud1 []byte

//go:embed images/cloud_2.png
var Cloud2 []byte

//go:embed images/cloud_3.png
var Cloud3 []byte

//go:embed images/cloud_4.png
var Cloud4 []byte

//go:embed images/cloud_5.png
var Cloud5 []byte

func SoundBytes(path string) ([]byte, error) {
	name := strings.TrimSpace(path)
	if name == "" {
		return nil, nil
	}
	name = filepath.ToSlash(name)
	name = strings.TrimPrefix(name, "resources/")
	if !strings.HasPrefix(name, "sounds/") {
		return nil, fmt.Errorf("sound path must be under resources/sounds: %s", path)
	}
	return Sounds.ReadFile(name)
}
