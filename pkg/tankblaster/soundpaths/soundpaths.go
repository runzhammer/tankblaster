package soundpaths

const (
	None = ""
)

const (
	SoundAbstellen    = "resources/sounds/abstellen.wav"
	SoundAnlassen     = "resources/sounds/anlassen.wav"
	SoundAtom         = "resources/sounds/atom.wav"
	SoundBaumschrei   = "resources/sounds/baumschrei.wav"
	SoundBeep         = "resources/sounds/beep.wav"
	SoundBlitz        = "resources/sounds/blitz.wav"
	SoundBlotsch      = "resources/sounds/blotsch.wav"
	SoundBlubber      = "resources/sounds/blubber.wav"
	SoundBroeselklirr = "resources/sounds/broeselklirr.wav"
	SoundBroesler     = "resources/sounds/broesler.wav"
	SoundBupp         = "resources/sounds/bupp.wav"
	SoundBurning      = "resources/sounds/burning.wav"
	SoundChoose       = "resources/sounds/choose.wav"
	SoundClusterblast = "resources/sounds/clusterblast.wav"
	SoundClusterexplo = "resources/sounds/clusterexplo.wav"
	SoundColorblop    = "resources/sounds/colorblop.wav"
	SoundDampf        = "resources/sounds/dampf.wav"
	SoundDown         = "resources/sounds/down.wav"
	SoundDrehen       = "resources/sounds/drehen.wav"
	SoundDrehenDown   = "resources/sounds/drehen_down.wav"
	SoundDrehenLeft   = "resources/sounds/drehen_left.wav"
	SoundDrehenRight  = "resources/sounds/drehen_right.wav"
	SoundDrehenUp     = "resources/sounds/drehen_up.wav"
	SoundExplosion1   = "resources/sounds/explosion1.wav"
	SoundExplosion2   = "resources/sounds/explosion2.wav"
	SoundExplosion3   = "resources/sounds/explosion3.wav"
	SoundFire         = "resources/sounds/fire.wav"
	SoundFlyin        = "resources/sounds/flyin.wav"
	SoundFlyout       = "resources/sounds/flyout.wav"
	SoundGrowing      = "resources/sounds/growing.wav"
	SoundHahaha       = "resources/sounds/hahaha.wav"
	SoundIncinerator1 = "resources/sounds/incinerator1.wav"
	SoundIncinerator2 = "resources/sounds/incinerator2.wav"
	SoundJet          = "resources/sounds/jet.wav"
	SoundKasse        = "resources/sounds/kasse.wav"
	SoundKette        = "resources/sounds/kette.wav"
	SoundKlick        = "resources/sounds/klick.wav"
	SoundKlirr        = "resources/sounds/klirr.wav"
	SoundKlonk        = "resources/sounds/klonk.wav"
	SoundLaser        = "resources/sounds/laser.wav"
	SoundMole         = "resources/sounds/mole.wav"
	SoundMoney        = "resources/sounds/money.wav"
	SoundMoskitos     = "resources/sounds/moskitos.wav"
	SoundMotor        = "resources/sounds/motor.wav"
	SoundNukeall      = "resources/sounds/nukeall.wav"
	SoundPlasma       = "resources/sounds/plasma.wav"
	SoundPock         = "resources/sounds/pock.wav"
	SoundSchnaeppchen = "resources/sounds/schnaeppchen.wav"
	SoundScream       = "resources/sounds/scream.wav"
	SoundScream2      = "resources/sounds/scream2.wav"
	SoundShockwave    = "resources/sounds/shockwave.wav"
	SoundUp           = "resources/sounds/up.wav"
	SoundWasserstoff  = "resources/sounds/wasserstoff.wav"
)

const (
	Default = SoundKlonk
)

const (
	WeaponFireDefault   = SoundFire
	WeaponImpactDefault = SoundIncinerator1
)

const (
	PlayerSelectionStart   = SoundChoose
	RoundStart             = None
	RoundEnd               = None
	WeaponSelect           = None
	CannonRotateLeft       = SoundDrehenLeft
	CannonRotateRight      = SoundDrehenRight
	TankHit                = None
	TankDestroyed          = None
	PalmHit                = SoundPock
	PalmIgnite             = SoundBurning
	PalmCrumble            = SoundKlirr
	PalmEyes               = SoundPock
	PalmRevenge            = SoundBaumschrei
	CloudSearch            = None
	CloudLightning         = SoundBlitz
	RevengeTankBroken      = SoundHahaha
	WaterFill              = None
	WaterBlubber           = SoundBlubber
	WaterBlotch            = SoundBlotsch
	DudImpact              = SoundSchnaeppchen
	MoleImpact             = SoundMole
	CrumblerImpact         = SoundBroesler
	AirStrikeBeacon        = SoundBeep
	AirStrikeBomb          = SoundJet
	Shockwave              = SoundShockwave
	Mosquitos              = SoundMoskitos
	LaserSmoke             = SoundLaser
	XMV12Ignition          = SoundAnlassen
	XMV12EngineLoop        = SoundMotor
	XMV12TrackLoop         = SoundKette
	XMV12MotorOff          = SoundAbstellen
	ShopListSelect         = SoundBupp
	ShopBuy                = SoundKasse
	ShopNotEnoughMoney     = SoundMoney
	ShopNextPlayer         = None
	ButtonPress            = SoundKlick
	ProjectileReentry      = SoundFlyout
	ProjectileReentryExit  = SoundFlyout
	ProjectileReentryEnter = SoundFlyin
)

const (
	ZeroPowerDust               = SoundDampf
	ZeroPowerExplosion          = SoundExplosion1
	ZeroPowerMushroom           = SoundExplosion3
	ZeroPowerSmoke              = None
	ZeroPowerGrenadeImpact      = SoundExplosion1
	ZeroPowerLargeGrenadeImpact = SoundExplosion2
	ZeroPowerAtomImpact         = SoundAtom
	ZeroPowerScatterProjectiles = SoundClusterblast
)
