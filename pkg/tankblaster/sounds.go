package tankblaster

import "github.com/runzhammer/gamedemo/pkg/tankblaster/soundpaths"

const defaultSoundPath = soundpaths.Default

type soundEvent string

const (
	soundEventRoundStart       soundEvent = "round_start"
	soundEventRoundEnd         soundEvent = "round_end"
	soundEventWeaponSelect     soundEvent = "weapon_select"
	soundEventTankHit          soundEvent = "tank_hit"
	soundEventTankDestroyed    soundEvent = "tank_destroyed"
	soundEventPalmHit          soundEvent = "palm_hit"
	soundEventPalmIgnite       soundEvent = "palm_ignite"
	soundEventPalmCrumble      soundEvent = "palm_crumble"
	soundEventPalmEyes         soundEvent = "palm_eyes"
	soundEventPalmRevenge      soundEvent = "palm_revenge"
	soundEventCloudSearch      soundEvent = "cloud_search"
	soundEventCloudLightning   soundEvent = "cloud_lightning"
	soundEventWaterFill        soundEvent = "water_fill"
	soundEventWaterBlubber     soundEvent = "water_blubber"
	soundEventWaterBlotch      soundEvent = "water_blotch"
	soundEventDudImpact        soundEvent = "dud_impact"
	soundEventMoleImpact       soundEvent = "mole_impact"
	soundEventCrumblerImpact   soundEvent = "crumbler_impact"
	soundEventAirStrikeBeacon  soundEvent = "air_strike_beacon"
	soundEventAirStrikeBomb    soundEvent = "air_strike_bomb"
	soundEventShockwave        soundEvent = "shockwave"
	soundEventMosquitos        soundEvent = "mosquitos"
	soundEventLaserSmoke       soundEvent = "laser_smoke"
	soundEventXMV12Ignition    soundEvent = "xm_v12_ignition"
	soundEventXMV12MotorOff    soundEvent = "xm_v12_motor_off"
	soundEventShopBuy          soundEvent = "shop_buy"
	soundEventShopNextPlayer   soundEvent = "shop_next_player"
	soundEventButtonPress      soundEvent = "button_press"
	soundEventProjectileReentry soundEvent = "projectile_reentry"
)

type zeroPowerSound string

const (
	zeroPowerSoundDust              zeroPowerSound = "dust"
	zeroPowerSoundExplosion         zeroPowerSound = "explosion"
	zeroPowerSoundMushroom          zeroPowerSound = "mushroom"
	zeroPowerSoundSmoke             zeroPowerSound = "smoke"
	zeroPowerSoundGrenadeImpact     zeroPowerSound = "grenade_impact"
	zeroPowerSoundLargeGrenadeImpact zeroPowerSound = "large_grenade_impact"
	zeroPowerSoundAtomImpact        zeroPowerSound = "atom_impact"
	zeroPowerSoundScatterProjectiles zeroPowerSound = "scatter_projectiles"
)

type audioConfig struct {
	Events    map[soundEvent]string
	ZeroPower map[zeroPowerSound]string
}

var tankBlasterSounds = audioConfig{
	Events: map[soundEvent]string{
		soundEventRoundStart:        defaultSoundPath,
		soundEventRoundEnd:          defaultSoundPath,
		soundEventWeaponSelect:      defaultSoundPath,
		soundEventTankHit:           defaultSoundPath,
		soundEventTankDestroyed:     defaultSoundPath,
		soundEventPalmHit:           defaultSoundPath,
		soundEventPalmIgnite:        defaultSoundPath,
		soundEventPalmCrumble:       defaultSoundPath,
		soundEventPalmEyes:          defaultSoundPath,
		soundEventPalmRevenge:       defaultSoundPath,
		soundEventCloudSearch:       defaultSoundPath,
		soundEventCloudLightning:    defaultSoundPath,
		soundEventWaterFill:         defaultSoundPath,
		soundEventWaterBlubber:      defaultSoundPath,
		soundEventWaterBlotch:       defaultSoundPath,
		soundEventDudImpact:         defaultSoundPath,
		soundEventMoleImpact:        defaultSoundPath,
		soundEventCrumblerImpact:    defaultSoundPath,
		soundEventAirStrikeBeacon:   defaultSoundPath,
		soundEventAirStrikeBomb:     defaultSoundPath,
		soundEventShockwave:         defaultSoundPath,
		soundEventMosquitos:         defaultSoundPath,
		soundEventLaserSmoke:        defaultSoundPath,
		soundEventXMV12Ignition:     defaultSoundPath,
		soundEventXMV12MotorOff:     defaultSoundPath,
		soundEventShopBuy:           defaultSoundPath,
		soundEventShopNextPlayer:    defaultSoundPath,
		soundEventButtonPress:       defaultSoundPath,
		soundEventProjectileReentry: defaultSoundPath,
	},
	ZeroPower: map[zeroPowerSound]string{
		zeroPowerSoundDust:               defaultSoundPath,
		zeroPowerSoundExplosion:          defaultSoundPath,
		zeroPowerSoundMushroom:           defaultSoundPath,
		zeroPowerSoundSmoke:              defaultSoundPath,
		zeroPowerSoundGrenadeImpact:      defaultSoundPath,
		zeroPowerSoundLargeGrenadeImpact: defaultSoundPath,
		zeroPowerSoundAtomImpact:         defaultSoundPath,
		zeroPowerSoundScatterProjectiles: defaultSoundPath,
	},
}
