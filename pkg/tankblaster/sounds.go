package tankblaster

import "github.com/runzhammer/tankblaster/pkg/tankblaster/soundpaths"

type soundEvent string

const (
	soundEventPlayerSelectionStart   soundEvent = "player_selection_start"
	soundEventRoundStart             soundEvent = "round_start"
	soundEventRoundEnd               soundEvent = "round_end"
	soundEventWeaponSelect           soundEvent = "weapon_select"
	soundEventCannonPowerUp          soundEvent = "cannon_power_up"
	soundEventCannonPowerDown        soundEvent = "cannon_power_down"
	soundEventCannonRotateLeft       soundEvent = "cannon_rotate_left"
	soundEventCannonRotateRight      soundEvent = "cannon_rotate_right"
	soundEventTankHit                soundEvent = "tank_hit"
	soundEventTankDestroyed          soundEvent = "tank_destroyed"
	soundEventPalmHit                soundEvent = "palm_hit"
	soundEventPalmIgnite             soundEvent = "palm_ignite"
	soundEventPalmCrumble            soundEvent = "palm_crumble"
	soundEventPalmEyes               soundEvent = "palm_eyes"
	soundEventPalmRevenge            soundEvent = "palm_revenge"
	soundEventCloudSearch            soundEvent = "cloud_search"
	soundEventCloudLightning         soundEvent = "cloud_lightning"
	soundEventRevengeTankBroken      soundEvent = "revenge_tank_broken"
	soundEventWaterFill              soundEvent = "water_fill"
	soundEventWaterBlubber           soundEvent = "water_blubber"
	soundEventWaterBlotch            soundEvent = "water_blotch"
	soundEventDudImpact              soundEvent = "dud_impact"
	soundEventMoleImpact             soundEvent = "mole_impact"
	soundEventCrumblerImpact         soundEvent = "crumbler_impact"
	soundEventSplitterBombSplit      soundEvent = "splitter_bomb_split"
	soundEventAirStrikeBeacon        soundEvent = "air_strike_beacon"
	soundEventAirStrikeBomb          soundEvent = "air_strike_bomb"
	soundEventAirStrikeJet           soundEvent = "air_strike_jet"
	soundEventShockwave              soundEvent = "shockwave"
	soundEventMosquitos              soundEvent = "mosquitos"
	soundEventMosquitoScream         soundEvent = "mosquito_scream"
	soundEventLaser                  soundEvent = "laser"
	soundEventLaserSmoke             soundEvent = "laser_smoke"
	soundEventXMV12Ignition          soundEvent = "xm_v12_ignition"
	soundEventXMV12EngineLoop        soundEvent = "xm_v12_engine_loop"
	soundEventXMV12TrackLoop         soundEvent = "xm_v12_track_loop"
	soundEventXMV12MotorOff          soundEvent = "xm_v12_motor_off"
	soundEventShopListSelect         soundEvent = "shop_list_select"
	soundEventShopBuy                soundEvent = "shop_buy"
	soundEventShopNotEnoughMoney     soundEvent = "shop_not_enough_money"
	soundEventShopNextPlayer         soundEvent = "shop_next_player"
	soundEventButtonPress            soundEvent = "button_press"
	soundEventProjectileReentry      soundEvent = "projectile_reentry"
	soundEventProjectileReentryExit  soundEvent = "projectile_reentry_exit"
	soundEventProjectileReentryEnter soundEvent = "projectile_reentry_enter"
)

type zeroPowerSound string

const (
	zeroPowerSoundDust               zeroPowerSound = "dust"
	zeroPowerSoundExplosion          zeroPowerSound = "explosion"
	zeroPowerSoundMushroom           zeroPowerSound = "mushroom"
	zeroPowerSoundSmoke              zeroPowerSound = "smoke"
	zeroPowerSoundGrenadeImpact      zeroPowerSound = "grenade_impact"
	zeroPowerSoundLargeGrenadeImpact zeroPowerSound = "large_grenade_impact"
	zeroPowerSoundAtomImpact         zeroPowerSound = "atom_impact"
	zeroPowerSoundScatterProjectiles zeroPowerSound = "scatter_projectiles"
	zeroPowerSoundScatterImpact      zeroPowerSound = "scatter_impact"
)

type audioConfig struct {
	Events    map[soundEvent]string
	ZeroPower map[zeroPowerSound]string
	Options   map[string]soundOptions
}

type soundOptions struct {
	Loop               bool
	DurationSeconds    float64
	PlaybackSpeed      float64
	RepeatEverySeconds float64
	MaxRepeats         int
	Volume             float64
	StartOffsetSeconds float64
	TrimSilenceForLoop bool
	KeepSilenceForLoop bool
	SilenceThreshold   int16
	AlternatePaths     []string
}

var tankBlasterSounds = audioConfig{
	Events: map[soundEvent]string{
		soundEventPlayerSelectionStart:   soundpaths.PlayerSelectionStart,
		soundEventRoundStart:             soundpaths.RoundStart,
		soundEventRoundEnd:               soundpaths.RoundEnd,
		soundEventWeaponSelect:           soundpaths.WeaponSelect,
		soundEventCannonPowerUp:          soundpaths.CannonPowerUp,
		soundEventCannonPowerDown:        soundpaths.CannonPowerDown,
		soundEventCannonRotateLeft:       soundpaths.CannonRotateLeft,
		soundEventCannonRotateRight:      soundpaths.CannonRotateRight,
		soundEventTankHit:                soundpaths.TankHit,
		soundEventTankDestroyed:          soundpaths.TankDestroyed,
		soundEventPalmHit:                soundpaths.PalmHit,
		soundEventPalmIgnite:             soundpaths.PalmIgnite,
		soundEventPalmCrumble:            soundpaths.PalmCrumble,
		soundEventPalmEyes:               soundpaths.PalmEyes,
		soundEventPalmRevenge:            soundpaths.PalmRevenge,
		soundEventCloudSearch:            soundpaths.CloudSearch,
		soundEventCloudLightning:         soundpaths.CloudLightning,
		soundEventRevengeTankBroken:      soundpaths.RevengeTankBroken,
		soundEventWaterFill:              soundpaths.WaterFill,
		soundEventWaterBlubber:           soundpaths.WaterBlubber,
		soundEventWaterBlotch:            soundpaths.WaterBlotch,
		soundEventDudImpact:              soundpaths.DudImpact,
		soundEventMoleImpact:             soundpaths.MoleImpact,
		soundEventCrumblerImpact:         soundpaths.CrumblerImpact,
		soundEventSplitterBombSplit:      soundpaths.SplitterBombSplit,
		soundEventAirStrikeBeacon:        soundpaths.AirStrikeBeacon,
		soundEventAirStrikeBomb:          soundpaths.AirStrikeBomb,
		soundEventAirStrikeJet:           soundpaths.AirStrikeJet,
		soundEventShockwave:              soundpaths.Shockwave,
		soundEventMosquitos:              soundpaths.Mosquitos,
		soundEventMosquitoScream:         soundpaths.MosquitoScream,
		soundEventLaser:                  soundpaths.Laser,
		soundEventLaserSmoke:             soundpaths.LaserSmoke,
		soundEventXMV12Ignition:          soundpaths.XMV12Ignition,
		soundEventXMV12EngineLoop:        soundpaths.XMV12EngineLoop,
		soundEventXMV12TrackLoop:         soundpaths.XMV12TrackLoop,
		soundEventXMV12MotorOff:          soundpaths.XMV12MotorOff,
		soundEventShopListSelect:         soundpaths.ShopListSelect,
		soundEventShopBuy:                soundpaths.ShopBuy,
		soundEventShopNotEnoughMoney:     soundpaths.ShopNotEnoughMoney,
		soundEventShopNextPlayer:         soundpaths.ShopNextPlayer,
		soundEventButtonPress:            soundpaths.ButtonPress,
		soundEventProjectileReentry:      soundpaths.ProjectileReentry,
		soundEventProjectileReentryExit:  soundpaths.ProjectileReentryExit,
		soundEventProjectileReentryEnter: soundpaths.ProjectileReentryEnter,
	},
	ZeroPower: map[zeroPowerSound]string{
		zeroPowerSoundDust:               soundpaths.ZeroPowerDust,
		zeroPowerSoundExplosion:          soundpaths.ZeroPowerExplosion,
		zeroPowerSoundMushroom:           soundpaths.ZeroPowerMushroom,
		zeroPowerSoundSmoke:              soundpaths.ZeroPowerSmoke,
		zeroPowerSoundGrenadeImpact:      soundpaths.ZeroPowerGrenadeImpact,
		zeroPowerSoundLargeGrenadeImpact: soundpaths.ZeroPowerLargeGrenadeImpact,
		zeroPowerSoundAtomImpact:         soundpaths.ZeroPowerAtomImpact,
		zeroPowerSoundScatterProjectiles: soundpaths.ZeroPowerScatterProjectiles,
		zeroPowerSoundScatterImpact:      soundpaths.ZeroPowerScatterImpact,
	},
	Options: map[string]soundOptions{
		soundpaths.SoundBeep: {KeepSilenceForLoop: true},
	},
}
