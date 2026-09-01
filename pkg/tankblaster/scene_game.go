package tankblaster

import (
	"bytes"
	"image"
	"image/color"
	"math"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"

	_ "image/jpeg"
	_ "image/png"

	"github.com/runzhammer/gamedemo/pkg/core"
	"github.com/runzhammer/gamedemo/pkg/engine"
	"github.com/runzhammer/gamedemo/pkg/models"
	"github.com/runzhammer/gamedemo/pkg/protocol"
	"github.com/runzhammer/gamedemo/pkg/tankblaster/computerplayers"
	"github.com/runzhammer/gamedemo/pkg/tankblaster/soundpaths"
	weaponspkg "github.com/runzhammer/gamedemo/pkg/tankblaster/weapons"
	r "github.com/runzhammer/gamedemo/resources"
	"golang.org/x/image/colornames"
)

var _ core.Scene = (*GameScene)(nil)

var solidWhiteImage *ebiten.Image

func getSolidWhiteImage() *ebiten.Image {
	if solidWhiteImage == nil {
		img := ebiten.NewImage(1, 1)
		img.Fill(colornames.White)
		solidWhiteImage = img
	}
	return solidWhiteImage
}

type Phase uint8

const (
	phaseBattle Phase = iota
	phaseShop
	phaseBlueVictory
	phaseRedVictory
)

const (
	layerBackground = iota
	layerClouds
	layerGround
	layerWater
	layerPalms
	layerTanks
	layerProjectiles
	numLayers
)

const gameHUDHeight = 132

const weaponbarSlotCount = 20

const reentryAnimationFrames = 170

const (
	projectileRadius           = 4
	impactRadiusMultiplier     = 4
	sandFallFrames             = 12
	zeroPowerFrames            = 216
	zeroPowerDissolveFrames    = 12
	debugShopStartingCredits   = 20000
	roundTransitionSeconds     = 10.0 / 3.0
	computerAdjustSpeedFactor  = 1.6
	additionalVisibleSkyHeight = 500.0
	palmRevengeAggroMin        = 20.0
	palmRevengeAggroMax        = 40.0
	palmRevengeTrigger         = 100.0
	palmRevengeFocusFrames     = 150
	palmRevengeAttackFrames    = 120
	palmRevengeLightningFrames = 240
	palmRevengeRecoverFrames   = 150
)

const (
	plasmaImpactVisualYOffset = 38
	plasmaRingSpacing         = 2
	plasmaRingThickness       = 1
	plasmaBuildProgress       = 2.0 / 6.5
	plasmaGreenProgress       = 2.5 / 6.5
	moleTunnelFrames          = 44
	moleStarFrames            = 28
	moleTunnelDepth           = 76.0
	moleTunnelWidth           = 7.0
	moleStarLineCount         = 36
	moleStarRadius            = 110.25
	moleStarLineThickness     = 2.0
	smallCrumblerCount        = 80
	smallCrumblerWidth        = 2
	smallCrumblerHeight       = 2
	smallCrumblerMaxDepth     = 300.0
	smallCrumblerFreezeFrames = 12
	smallCrumblerMinSpeed     = 1.6
	smallCrumblerMaxSpeed     = 3.2
	smallCrumblerStartWidth   = 50.0
	smallCrumblerHiddenStart  = 68.0
	smallCrumblerFunnelSpread = 280.0
	smallCrumblerJitter       = 4.8
	smallCrumblerWobble       = 3.2
	largeCrumblerCount        = 120
	largeCrumblerWidth        = 2
	largeCrumblerHeight       = 3
	largeCrumblerMaxDepth     = 450.0
	largeCrumblerFreezeFrames = 12
	largeCrumblerMinSpeed     = 1.6
	largeCrumblerMaxSpeed     = 3.2
	largeCrumblerStartWidth   = 100.0
	largeCrumblerHiddenStart  = 68.0
	largeCrumblerFunnelSpread = 320.0
	largeCrumblerJitter       = 4.8
	largeCrumblerWobble       = 3.2
	mosquitoPreviewFrames     = 10
	mosquitoTouchFrames       = 8
	mosquitoRiseFrames        = 20
	mosquitoHoverHeight       = 30.0
	mosquitoSeekSpeed         = 1.35
	mosquitoDiveSpeed         = 6.0
	mosquitoAttachedFrames    = 60
	mosquitoColorDelayFrames  = 30
	shockwavePulseFrames      = 30
	shockwavePulseCount       = 4
	shockwaveLineWidth        = 6.0
	shockwaveInnerLineWidth   = 2.0
	shockwaveCameraOrbit      = 34.0
	shockwaveCameraAngular    = 0.82
	airStrikeWaitFrames       = 540
	airStrikeBombCount        = 10
	airStrikeBombSpacing      = 80.0
	airStrikeBombDelayFrames  = 20
	airStrikeBombDelayWindow  = airStrikeBombDelayFrames * (airStrikeBombCount - 1)
	airStrikeBombFallFrames   = 46
	airStrikeBombAngle        = 25 * math.Pi / 180
	airStrikeImpactScale      = 1.5
	splitterBombFragmentCount = 9
	splitterBombSpreadWidth   = 300.0
	projectileWindFactor      = 0.00195
	laserHoldFrames           = 60
	laserDrillSpeed           = 4.0
	laserLineThickness        = 2.0
	laserSmokeSpacing         = 12.0
	laserMaxSmokeCount        = 3
	humanCannonRepeatStart    = 16
	humanCannonRepeatFrames   = 6
	humanCannonStepDegrees    = 1.0
	xmV12EngineOffDelayFrames = 30
	xmV12OutOfBoundsDelay     = 1.0
	xmV12EngineLoopKey        = "xm_v12_engine"
	xmV12TrackLoopKey         = "xm_v12_track"
	moskitosLoopKey           = "moskitos"
	fireballBurningLoopKey    = "fireball_burning"
	moleBroeslerLoopKey       = "mole_broesler"
	crumblerBroeslerLoopKey   = "crumbler_broesler"
	laserLoopKey              = "laser"
	airStrikeBeaconLoopKey    = "air_strike_beacon"
	airStrikeJetLoopKey       = "air_strike_jet"
	abortRoundFlashFrames     = 300
	abortRoundFlashWhiteFrame = 180
)

type gameConfirmDialog uint8

const (
	gameConfirmNone gameConfirmDialog = iota
	gameConfirmAbortRound
	gameConfirmQuit
)

type damageCause uint8

const (
	damageCauseDirect damageCause = iota
	damageCauseFall
)

type battleTank struct {
	playerIndex    int
	player         PlayerConfig
	body           *engine.Sprite
	cannon         *engine.Sprite
	landed         bool
	falling        bool
	fallStartY     float64
	fallTargetY    float64
	fallDamage     bool
	power          int
	score          int
	tint           color.RGBA
	zeroPowerShown bool
	zeroPowerGone  bool
	terrainLocked  bool
	selectedWeapon int
	shotStrength   int
	computerPlan   *computerTurnPlan
}

type battlePalm struct {
	sprite               *engine.Sprite
	pixels               *image.RGBA
	state                palmState
	age                  int
	eyeAge               int
	eyesOn               bool
	screamAge            int
	screaming            bool
	grinAge              int
	grinning             bool
	grinHideAt           int
	aggression           float64
	aggressionMultiplier float64
}

type palmState uint8

const (
	palmStateAlive palmState = iota
	palmStateBurning
	palmStateSkeletonSmoking
	palmStateSkeleton
	palmStateCrumbling
)

type battleCloud struct {
	sprite        *engine.Sprite
	speed         float64
	kind          cloudKind
	image         *ebiten.Image
	revengeActive bool
	aggression    float64
}

type cloudSearchEffect struct {
	cloud     *battleCloud
	age       int
	delay     int
	duration  int
	animation spriteAnimation
}

type palmRevengePhase uint8

const (
	palmRevengeNone palmRevengePhase = iota
	palmRevengeCloudFocus
	palmRevengeCloudAttack
	palmRevengeLightning
	palmRevengeRecover
)

type palmRevengeEvent struct {
	phase              palmRevengePhase
	age                int
	palm               *battlePalm
	cloud              *battleCloud
	target             *battleTank
	cloudStart         engine.Vec
	cloudAttack        engine.Vec
	cloudOriginal      engine.Vec
	targetCameraX      float64
	tankBlackened      bool
	damageDone         bool
	smokeDone          bool
	smokeAge           int
	lightningSoundDone [2]bool
}

type cloudAsset struct {
	image *ebiten.Image
	size  engine.Vec
	kind  cloudKind
}

type cloudKind uint8

const (
	cloudKindNormal cloudKind = iota
	cloudKindLightning
)

type computerTurnPlan struct {
	phase          computerTurnPhase
	delay          int
	adjustProgress float64
	decision       computerplayers.Decision
	targetStrength int
	targetAngle    float64
}

type computerTurnPhase uint8

const (
	computerTurnWaitCamera computerTurnPhase = iota
	computerTurnAdjustStrength
	computerTurnAdjustAngle
	computerTurnSelectWeapon
	computerTurnFireDelay
)

type computerShotRecord struct {
	active      bool
	playerIndex int
	computerID  computerplayers.ID
	targetIndex int
	activeX     float64
	targetX     float64
}

type projectile struct {
	pos                engine.Vec
	prev               engine.Vec
	velocity           engine.Vec
	weaponIndex        int
	effectiveWeapon    weaponspkg.Weapon
	hasEffectiveWeapon bool
	classBDud          bool
	zeroPowerScatter   bool
	scatterImpactColor color.RGBA
	trail              []engine.Vec
	shooter            *battleTank
	launchRot          float64
	angeredClouds      map[*battleCloud]bool
	searchingCloud     *battleCloud
	mosquitoPreview    bool
	splitterArmed      bool
}

type projectileReentryAnimation struct {
	projectile *projectile
	shooter    *battleTank
	age        int
	duration   int
	direction  int
	cannonRot  float64
}

type impactAnimation struct {
	pos            engine.Vec
	radius         float64
	age            int
	duration       int
	cycles         int
	color          color.RGBA
	damage         int
	radialDamage   weaponspkg.RadialDamageProfile
	attacker       *battleTank
	damageApplied  map[int]bool
	outward        bool
	style          weaponspkg.ImpactAnimationStyle
	terrainApplied bool
}

type animatedImpact struct {
	pos       engine.Vec
	age       int
	duration  int
	animation spriteAnimation
	damage    int
	applied   bool
	loopKey   string
}

type waterFill struct {
	leftX      int
	rightX     int
	topY       float64
	surfaceY   []float64
	age        int
	duration   int
	image      *ebiten.Image
	cacheAge   int
	cacheFrame int
	hitPlayers map[int]bool
}

type waterBlubberEffect struct {
	start     engine.Vec
	end       engine.Vec
	age       int
	duration  int
	animation spriteAnimation
}

func (e waterBlubberEffect) position() engine.Vec {
	progress := 1.0
	if e.duration > 0 && e.age < e.duration {
		progress = easeOut(float64(e.age) / float64(e.duration))
	}
	return engine.V(
		e.start.X+(e.end.X-e.start.X)*progress,
		e.start.Y+(e.end.Y-e.start.Y)*progress,
	)
}

type waterSurfaceImpact struct {
	pos       engine.Vec
	age       int
	duration  int
	animation spriteAnimation
}

type moleImpact struct {
	start      engine.Vec
	pos        engine.Vec
	age        int
	phase      moleImpactPhase
	tunnelPath []engine.Vec
	starLines  []moleStarLine
	editArea   image.Rectangle
}

type moleImpactPhase uint8

const (
	molePhaseTunnel moleImpactPhase = iota
	molePhaseStar
)

type moleStarLine struct {
	angle  float64
	length float64
}

type smallCrumblerImpact struct {
	start     engine.Vec
	crumbs    []smallCrumb
	cfg       crumblerConfig
	age       int
	frozen    bool
	freezeAge int
	editArea  image.Rectangle
}

type crumblerConfig struct {
	count        int
	width        int
	height       int
	maxDepth     float64
	freezeFrames int
	minSpeed     float64
	maxSpeed     float64
	startWidth   float64
	hiddenStart  float64
	funnelSpread float64
	jitter       float64
	wobble       float64
}

type smallCrumb struct {
	pos      engine.Vec
	fan      float64
	speed    float64
	drift    float64
	wobble   float64
	stopped  bool
	lastArea image.Rectangle
}

type moskitoPhase uint8

const (
	moskitoPhaseTouch moskitoPhase = iota
	moskitoPhaseRise
	moskitoPhaseQuestion
	moskitoPhaseSeek
	moskitoPhaseDive
	moskitoPhaseAttached
)

type moskitoEffect struct {
	pos      engine.Vec
	ground   engine.Vec
	hover    engine.Vec
	phase    moskitoPhase
	age      int
	target   *battleTank
	shooter  *battleTank
	colored  bool
	finished bool
}

type shockwaveImpact struct {
	pos      engine.Vec
	radius   float64
	age      int
	duration int
}

type airStrikeImpact struct {
	pos             engine.Vec
	age             int
	duration        int
	bojeHidden      bool
	beepsPlayed     int
	lastBeaconCycle int
	jetStarted      bool
	bombs           []airStrikeBomb
}

type airStrikeBomb struct {
	start    engine.Vec
	target   engine.Vec
	pos      engine.Vec
	delay    int
	impacted bool
}

type laserEffect struct {
	start       engine.Vec
	end         engine.Vec
	tip         engine.Vec
	dir         engine.Vec
	maxDistance float64
	age         int
	duration    int
	traveling   bool
	drilling    bool
	finished    bool
	settled     bool
	editArea    image.Rectangle
	smokes      []laserSmoke
	smokeFrom   engine.Vec
}

type laserSmoke struct {
	pos engine.Vec
	age int
}

type sandFallAnimation struct {
	pixels   []models.SandFallPixel
	age      int
	duration int
}

type palmLeafFall struct {
	image    *ebiten.Image
	pos      engine.Vec
	velocity engine.Vec
	age      int
	landed   bool
	maxAge   int
}

type zeroPowerAnimation struct {
	tank      *battleTank
	animation spriteAnimation
	kind      zeroPowerEffectKind
	weapon    weaponspkg.Weapon
	delay     int
	age       int
	duration  int
	triggered bool
}

type zeroPowerEffectKind uint8

const (
	zeroPowerEffectSprite zeroPowerEffectKind = iota
	zeroPowerEffectGrenadeImpact
	zeroPowerEffectLargeGrenadeImpact
	zeroPowerEffectAtomImpact
	zeroPowerEffectScatterProjectiles
)

type zeroPowerChoice struct {
	name      string
	spriteIdx int
	kind      zeroPowerEffectKind
}

type spriteAnimation struct {
	frames     []*ebiten.Image
	pixels     []*image.RGBA
	delays     []int
	totalTicks int
	width      int
	height     int
	scaleX     float64
	scaleY     float64
	anchor     zeroPowerAnimationAnchor
}

type zeroPowerAnimationAnchor uint8

const (
	zeroPowerAnchorTankCenter zeroPowerAnimationAnchor = iota
	zeroPowerAnchorTankBottom
)

type GameScene struct {
	g    *GameLoop
	time float64

	phase Phase

	// shot engine.Drawable

	layers                  engine.Layers
	ground                  models.Ground
	worldWidth              float64
	cameraX                 float64
	cameraY                 float64
	cameraGoal              float64
	cameraGoalY             float64
	rng                     *rand.Rand
	palmImage               *ebiten.Image
	palmPixels              *image.RGBA
	palmSkeletonImage       *ebiten.Image
	palmSkeletonPixels      *image.RGBA
	palmFireAnimation       spriteAnimation
	palmSmokeAnimation      spriteAnimation
	palmCrumbleAnimation    spriteAnimation
	palmEyesAnimation       spriteAnimation
	palmEyesCloseAnimation  spriteAnimation
	palmScreamImage         *ebiten.Image
	palmGrinImage           *ebiten.Image
	palmLeafImages          []*ebiten.Image
	cloudAngryImage         *ebiten.Image
	cloudSearchingAnimation spriteAnimation
	cloudGrinImage          *ebiten.Image
	cloudGrinAnimation      spriteAnimation
	lightningImage          *ebiten.Image
	questionIconImage       *ebiten.Image
	reentrySymbol           *ebiten.Image
	reentryEarth            *ebiten.Image
	fuelGaugeImage          *ebiten.Image
	slopeMeterImage         *ebiten.Image
	ignitionFrames          []*ebiten.Image
	humanPortrait           *ebiten.Image
	computerPortraits       map[computerplayers.ID]*ebiten.Image
	zeroPowerSmoke          spriteAnimation
	fireballAnimation       spriteAnimation
	waterAnimation          spriteAnimation
	waterBlubberAnimation   spriteAnimation
	waterBlotchAnimation    spriteAnimation
	dudImpactAnimation      spriteAnimation
	moskitosAnimation       spriteAnimation
	questionAnimation       spriteAnimation
	blinkBojeAnimation      spriteAnimation
	bulletBombImage         *ebiten.Image
	laserSmokeAnimation     spriteAnimation
	clouds                  []*battleCloud
	cloudAssets             []cloudAsset

	tanks                []*battleTank
	palms                []*battlePalm
	players              []PlayerConfig
	scores               []int
	roundScores          []int
	credits              []int
	inventories          []shopInventory
	computerMemories     []computerplayers.Memory
	effectiveComputerIDs []computerplayers.ID
	roundNumber          int
	spawnIndex           int
	spawnPauseFrames     int
	activePlayerIndex    int
	wind                 int
	windDirection        int
	projectileReentry    bool
	reentryAnimation     *projectileReentryAnimation
	projectile           *projectile
	projectiles          []*projectile
	impacts              []impactAnimation
	animatedImpacts      []animatedImpact
	waterFills           []waterFill
	waterBlubbers        []*waterBlubberEffect
	waterBlotches        []*waterSurfaceImpact
	moleImpacts          []*moleImpact
	smallCrumblerImpacts []*smallCrumblerImpact
	moskitoEffects       []*moskitoEffect
	shockwaveImpacts     []*shockwaveImpact
	airStrikeImpacts     []*airStrikeImpact
	laserEffects         []*laserEffect
	sandFalls            []sandFallAnimation
	palmLeafFalls        []palmLeafFall
	cloudSearchEffects   []*cloudSearchEffect
	palmRevenge          *palmRevengeEvent
	palmRevengeRemoval   *battleTank
	zeroPowerEffects     []zeroPowerAnimation
	zeroPowerAnimations  []spriteAnimation
	zeroPowerNames       []string
	shop                 shopAssets
	turnAdvanceDelay     int
	roundTransitionDelay int
	roundSeriesComplete  bool
	showScoreTable       bool
	showPlayerNames      bool
	gameHelpOpen         bool
	playerInfoOpen       bool
	playerInfoIndex      int
	pressedDialogButton  string
	confirmDialog        gameConfirmDialog
	abortRoundFlashAge   int
	abortRoundFlashing   bool
	gamePaused           bool
	lastDamageSource     *battleTank
	shopPlayerOrder      []int
	shopPlayerCursor     int
	shopHoverClass       int
	shopMode             shopMode
	shopSelectedIndex    int
	shopClassAStock      []int
	shopClassBStock      []int
	shopClassBItems      []int
	shopComputerPlan     *shopComputerPlan
	shopTouchScroll      shopTouchScrollState
	scrollBarDragging    bool
	zeroPowerStartDelay  int
	lastComputerShot     computerShotRecord
	palmCameraFocus      *battlePalm
	waterCameraFocus     *waterBlubberEffect
	waterBlotchFocus     *waterSurfaceImpact
	crumblerCameraFocus  *engine.Vec
	moskitoCameraFocus   *engine.Vec
	xmV12DriveMode       bool
	xmV12DriveDirection  int
	xmV12EngineOffDelay  int
	xmV12IdleOffset      engine.Vec
}

func NewGameScene(game *GameLoop) (core.Scene, error) {
	// loader := game.context.Loader()
	rngSeed := time.Now().UnixNano()
	if game.online != nil && game.online.state.Seed != 0 {
		rngSeed = game.online.state.Seed
	}

	s := &GameScene{
		g:                 game,
		phase:             phaseBattle,
		rng:               rand.New(rand.NewSource(rngSeed)),
		activePlayerIndex: -1,
		roundNumber:       1,
		showPlayerNames:   false,
		playerInfoIndex:   -1,
		reentrySymbol:     mustImageFromPNG(r.SymbolReentry),
		reentryEarth:      mustImageFromPNG(r.EarthReentry),
		fuelGaugeImage:    mustImageFromPNG(r.FuelGaugePNG),
		slopeMeterImage:   mustImageFromPNG(r.SlopeMeterPNG),
		humanPortrait:     mustImageFromPNG(r.PlayerHuman),
		computerPortraits: map[computerplayers.ID]*ebiten.Image{
			computerplayers.DoedelID:   mustImageFromPNG(r.PlayerComputerDoedel),
			computerplayers.FrederikID: mustImageFromPNG(r.PlayerComputerFrederik),
			computerplayers.MisterXID:  mustImageFromPNG(r.PlayerComputerMisterX),
			computerplayers.DrNukeID:   mustImageFromPNG(r.PlayerComputerDrNuke),
			computerplayers.HaraldID:   mustImageFromPNG(r.PlayerComputerHarald),
		},
	}
	s.ignitionFrames = splitImageFrames(mustImageFromPNG(r.ButtonIgnitionPNG), 31)
	zeroPowerAnimations, zeroPowerNames, err := loadZeroPowerAnimations()
	if err != nil {
		return nil, err
	}
	s.zeroPowerAnimations = zeroPowerAnimations
	s.zeroPowerNames = zeroPowerNames
	s.shop = shopAssets{
		human:               mustImageFromPNG(r.PlayerHuman),
		computer:            mustImageFromPNG(r.PlayerComputerDoedel),
		storeBg:             mustImageFromPNG(r.StoreBackground),
		storeMainLeft:       mustImageFromPNG(r.StoreMainLeft),
		storeMainRight:      mustImageFromPNG(r.StoreMainRight),
		storeRoll:           mustImageFromPNG(r.StoreRoll),
		storeIcons:          mustImageFromPNG(r.StoreIcons),
		weaponbarActive:     mustImageFromPNG(r.WeaponbarActive),
		weaponbarOnStock:    mustImageFromPNG(r.WeaponbarOnStock),
		weaponbarOutOfStock: mustImageFromPNG(r.WeaponbarOutOfStock),
	}
	palmImage, palmPixels, err := loadPalmAsset()
	if err != nil {
		return nil, err
	}
	s.palmImage = palmImage
	s.palmPixels = palmPixels
	palmLeafImages, err := loadPalmLeafImages(r.PalmLeavesPNG)
	if err != nil {
		return nil, err
	}
	s.palmLeafImages = palmLeafImages
	palmSkeletonImage, palmSkeletonPixels, err := loadImageWithPixels(r.PalmSkeletonPNG)
	if err != nil {
		return nil, err
	}
	s.palmSkeletonImage = palmSkeletonImage
	s.palmSkeletonPixels = palmSkeletonPixels
	fireballAnimation, err := loadSpriteAnimation(zeroPowerAnimationSheet{data: r.FireballImpactPNG, frameWidth: 36, delay: 8})
	if err != nil {
		return nil, err
	}
	s.fireballAnimation = repeatAnimation(fireballAnimation, 5, fireballAnimation.totalTicks*5)
	palmFireAnimation, err := loadSpriteAnimation(zeroPowerAnimationSheet{data: r.PalmFirePNG, frameWidth: 86, delay: 8})
	if err != nil {
		return nil, err
	}
	s.palmFireAnimation = repeatAnimation(palmFireAnimation, 3, secondsToFrames(2))
	palmSmokeAnimation, err := loadSpriteAnimation(zeroPowerAnimationSheet{data: r.PalmSmokePNG, frameWidth: 73, delay: 8})
	if err != nil {
		return nil, err
	}
	s.palmSmokeAnimation = fitAnimationDuration(palmSmokeAnimation, secondsToFrames(1.5))
	palmCrumbleAnimation, err := loadSpriteAnimation(zeroPowerAnimationSheet{data: r.PalmCrumblePNG, frameWidth: 121, frameHeight: 152, delay: 8})
	if err != nil {
		return nil, err
	}
	s.palmCrumbleAnimation = fitAnimationDuration(palmCrumbleAnimation, secondsToFrames(1.0))
	palmEyesAnimation, err := loadSpriteAnimation(zeroPowerAnimationSheet{data: r.PalmEyesOpenPNG, frameWidth: 30, delay: 3})
	if err != nil {
		return nil, err
	}
	s.palmEyesAnimation = fitAnimationDuration(palmEyesAnimation, secondsToFrames(0.5))
	palmEyesCloseAnimation, err := loadSpriteAnimation(zeroPowerAnimationSheet{data: r.PalmEyesClosePNG, frameWidth: 30, delay: 3})
	if err != nil {
		return nil, err
	}
	s.palmEyesCloseAnimation = fitAnimationDuration(palmEyesCloseAnimation, secondsToFrames(1))
	s.palmScreamImage = mustImageFromPNG(r.PalmScreamPNG)
	s.palmGrinImage = mustImageFromPNG(r.PalmGrinPNG)
	s.cloudAngryImage = mustImageFromPNG(r.CloudAngryPNG)
	cloudSearchingAnimation, err := loadSpriteAnimation(zeroPowerAnimationSheet{data: r.CloudSearchingPNG, frameWidth: 54, delay: 120})
	if err != nil {
		return nil, err
	}
	s.cloudSearchingAnimation = cloudSearchingAnimation
	s.cloudGrinImage = mustImageFromPNG(r.CloudGrinPNG)
	s.lightningImage = mustImageFromPNG(r.LightningPNG)
	s.questionIconImage = mustImageFromPNG(r.FragezeichenDialogPNG)
	cloudGrinAnimation, err := loadSpriteAnimation(zeroPowerAnimationSheet{data: r.CloudAngryToGrinPNG, frameWidth: 54, delay: 6})
	if err != nil {
		return nil, err
	}
	s.cloudGrinAnimation = cloudGrinAnimation
	zeroPowerSmoke, err := loadSpriteAnimation(zeroPowerAnimationSheet{data: r.ZeroPowerPlayerSmokePNG, frameWidth: 15, delay: 6, scaleX: 1.0, scaleY: 1.0})
	if err != nil {
		return nil, err
	}
	s.zeroPowerSmoke = zeroPowerSmoke
	waterAnimation, err := loadSpriteAnimation(zeroPowerAnimationSheet{data: r.WaterTexturePNG, frameWidth: 64, delay: 8})
	if err != nil {
		return nil, err
	}
	s.waterAnimation = waterAnimation
	waterBlubberAnimation, err := loadSpriteAnimation(zeroPowerAnimationSheet{data: r.WaterBlubberPNG, frameWidth: 19, delay: 8})
	if err != nil {
		return nil, err
	}
	s.waterBlubberAnimation = waterBlubberAnimation
	waterBlotchAnimation, err := loadSpriteAnimation(zeroPowerAnimationSheet{data: r.WaterBlotchPNG, frameWidth: 22, delay: 8, scaleX: 1.5, scaleY: 1.5})
	if err != nil {
		return nil, err
	}
	s.waterBlotchAnimation = waterBlotchAnimation
	dudImpactAnimation, err := loadSpriteAnimation(zeroPowerAnimationSheet{data: r.ZeroPowerDustExplosionPNG, frameWidth: 20, delay: 6, scaleX: 1, scaleY: 1})
	if err != nil {
		return nil, err
	}
	s.dudImpactAnimation = dudImpactAnimation
	moskitosAnimation, err := loadSpriteAnimation(zeroPowerAnimationSheet{data: r.MoskitosPNG, frameWidth: 23, delay: 8})
	if err != nil {
		return nil, err
	}
	s.moskitosAnimation = moskitosAnimation
	questionAnimation, err := loadSpriteAnimation(zeroPowerAnimationSheet{data: r.FragezeichenPNG, frameWidth: 32, delay: 8})
	if err != nil {
		return nil, err
	}
	s.questionAnimation = fitAnimationDuration(questionAnimation, secondsToFrames(2))
	blinkBojeAnimation, err := loadSpriteAnimation(zeroPowerAnimationSheet{data: r.BlinkBojePNG, frameWidth: 6, delay: 8})
	if err != nil {
		return nil, err
	}
	s.blinkBojeAnimation = fitAnimationDuration(blinkBojeAnimation, secondsToFrames(0.8))
	s.bulletBombImage = mustImageFromPNG(r.BulletBombPNG)
	laserSmokeAnimation, err := loadSpriteAnimation(zeroPowerAnimationSheet{data: r.LaserSmokePNG, frameWidth: 10, delay: 12})
	if err != nil {
		return nil, err
	}
	s.laserSmokeAnimation = laserSmokeAnimation
	s.cloudAssets = loadCloudAssets()
	s.players = s.playersForRound()
	s.scores = make([]int, len(s.players))
	s.roundScores = make([]int, len(s.players))
	s.credits = make([]int, len(s.players))
	s.ensureDebugHumanCredits()
	s.inventories = makeShopInventories(len(s.players))
	s.computerMemories = make([]computerplayers.Memory, len(s.players))
	s.startRound()

	return s, nil
}

func NewDebugShopScene(game *GameLoop) (core.Scene, error) {
	scene, err := NewGameScene(game)
	if err != nil {
		return nil, err
	}
	s, ok := scene.(*GameScene)
	if !ok {
		return scene, nil
	}
	s.ensureDebugHumanCredits()
	s.beginShop()
	return s, nil
}

func (s *GameScene) ensureDebugHumanCredits() {
	if !core.Config().Debug.Enabled {
		return
	}
	for i, player := range s.players {
		if player.Kind != PlayerHuman || i < 0 || i >= len(s.credits) {
			continue
		}
		if s.credits[i] < debugShopStartingCredits {
			s.credits[i] = debugShopStartingCredits
		}
	}
}

func (s *GameScene) startRound() {
	s.playEventSound(soundEventRoundStart)
	s.phase = phaseBattle
	s.layers = engine.NewLayers(numLayers)
	s.spawnIndex = 0
	s.spawnPauseFrames = 0
	s.activePlayerIndex = -1
	s.projectile = nil
	s.projectiles = nil
	s.laserEffects = nil
	s.reentryAnimation = nil
	s.impacts = nil
	s.sandFalls = nil
	s.palmLeafFalls = nil
	s.cloudSearchEffects = nil
	s.palmRevenge = nil
	s.palmRevengeRemoval = nil
	s.zeroPowerEffects = nil
	s.moskitoCameraFocus = nil
	s.turnAdvanceDelay = 0
	s.roundTransitionDelay = 0
	s.roundSeriesComplete = false
	s.lastDamageSource = nil
	s.lastComputerShot = computerShotRecord{}
	s.confirmDialog = gameConfirmNone
	s.abortRoundFlashAge = 0
	s.abortRoundFlashing = false
	s.shopPlayerOrder = nil
	s.shopPlayerCursor = 0
	s.shopHoverClass = 0
	s.shopMode = shopModeEntry
	s.shopSelectedIndex = 0
	s.roundScores = make([]int, len(s.players))
	s.assignEffectiveComputerIDs()
	s.wind = s.rng.Intn(101)
	if s.rng.Intn(2) == 0 {
		s.windDirection = -1
	} else {
		s.windDirection = 1
	}
	s.projectileReentry = s.projectileReentryEnabledForRound()
	s.cameraY = 0
	s.cameraGoalY = 0
	s.xmV12DriveMode = false
	s.xmV12DriveDirection = 0
	s.xmV12EngineOffDelay = 0
	s.xmV12IdleOffset = engine.Vec{}

	s.worldWidth = worldWidthForPlayers(len(s.players))

	battlefieldHeight := s.battlefieldHeight()
	skyExtra := s.skyExtraHeight()
	b := models.NewBackgroundWithSize(s.worldWidth, battlefieldHeight+skyExtra)
	if b.Position != nil {
		b.Position.Y = -skyExtra
	}
	gr := models.NewRandomGroundWithSize(s.worldWidth, battlefieldHeight, s.rng.Int63())
	s.ground = gr
	s.tanks = nil
	s.palms = nil
	s.clouds = nil
	s.animatedImpacts = nil
	s.waterFills = nil
	s.waterBlubbers = nil
	s.waterBlotches = nil
	s.moleImpacts = nil
	s.smallCrumblerImpacts = nil
	s.moskitoEffects = nil
	s.shockwaveImpacts = nil
	s.airStrikeImpacts = nil
	s.stopBattleEffectLoops()

	for tankIndex, player := range s.players {
		tank := models.NewTank(player.Name, player.Color)
		if s.playerHasXMV12(tankIndex) {
			tank = models.NewXMV12Tank(player.Name, player.Color)
		}
		battleTank := &battleTank{
			playerIndex:    tankIndex,
			player:         player,
			power:          100,
			score:          s.scoreForPlayer(tankIndex),
			tint:           player.Color,
			selectedWeapon: 1,
			shotStrength:   20,
		}
		tankBody := tank.Body()
		if tankBody != nil {
			tankBody.Pos = randomTankDropPosition(s.rng, tankIndex, len(s.players), tankBody.Size, s.worldWidth, battlefieldHeight)
			tankBody.Velocity = engine.Vec{Y: 2 + s.rng.Float64()*4}
			battleTank.body = tankBody
		}

		cannon := tank.Cannon()
		if cannon != nil {
			cannon.Steps = engine.MakeBehaviors(
				s.behaviorRotateActiveCannon,
			)
			cannon.PostSteps = engine.MakeBehaviors(
				s.behaviorAttachCannonToTank(tankBody),
			)
			battleTank.cannon = cannon
		}

		s.tanks = append(s.tanks, battleTank)
		s.layers[layerTanks] = engine.AddSprites(s.layers[layerTanks], tank.Sprites)
	}
	s.createClouds(battlefieldHeight)
	s.createPalms()
	if s.g.options.quickRoundStart {
		s.placeTanksOnGroundForQuickStart()
	}
	s.cameraGoal = s.cameraTargetForTank(0)
	s.cameraGoalY = 0
	s.layers[layerGround] = engine.AddSprites(s.layers[layerGround], gr.Sprites)
	s.layers[layerPalms] = engine.AddSprites(s.layers[layerPalms], s.palmSprites())
	s.layers[layerClouds] = engine.AddSprites(s.layers[layerClouds], s.cloudSprites())
	s.layers[layerBackground] = engine.AddSprites(s.layers[layerBackground], b.Sprites)
}

func (s *GameScene) projectileReentryEnabledForRound() bool {
	switch s.g.options.projectileReentry {
	case 1:
		return true
	case 2:
		return s.rng.Intn(2) == 0
	default:
		return false
	}
}

func loadCloudAssets() []cloudAsset {
	clouds := []struct {
		data []byte
		kind cloudKind
	}{
		{data: r.CloudLightning, kind: cloudKindLightning},
		{data: r.Cloud1, kind: cloudKindNormal},
		{data: r.Cloud2, kind: cloudKindNormal},
		{data: r.Cloud3, kind: cloudKindNormal},
		{data: r.Cloud4, kind: cloudKindNormal},
		{data: r.Cloud5, kind: cloudKindNormal},
	}
	assets := make([]cloudAsset, 0, len(clouds))
	for _, cloud := range clouds {
		img := mustImageFromPNG(cloud.data)
		bounds := img.Bounds()
		assets = append(assets, cloudAsset{
			image: img,
			size:  engine.V(float64(bounds.Dx()), float64(bounds.Dy())),
			kind:  cloud.kind,
		})
	}
	return assets
}

func (s *GameScene) createClouds(battlefieldHeight float64) {
	if len(s.cloudAssets) == 0 {
		return
	}
	count := s.randomCloudCount()
	cfg := core.Config().Gameplay.Clouds
	speedRange := cfg.MaxSpeed - cfg.MinSpeed
	windBoost := 0.45 + float64(s.wind)/100*0.7
	minY := -s.skyExtraHeight() + 20.0
	maxY := math.Max(minY+1, battlefieldHeight*0.28)
	laneWidth := s.worldWidth / float64(count)
	lightningLane := s.rng.Intn(count)

	for i := 0; i < count; i++ {
		asset := s.randomNormalCloudAsset()
		if i == lightningLane {
			if lightning := s.cloudAssetByKind(cloudKindLightning); lightning.image != nil {
				asset = lightning
			}
		}
		scale := 1.0
		size := engine.V(asset.size.X*scale, asset.size.Y*scale)
		laneCenter := laneWidth*float64(i) + laneWidth/2
		x := laneCenter - size.X/2 + (s.rng.Float64()-0.5)*laneWidth*0.5
		x = math.Max(-size.X, math.Min(s.worldWidth, x))
		y := minY + s.rng.Float64()*(maxY-minY)
		speed := (cfg.MinSpeed + s.rng.Float64()*speedRange) * windBoost
		cloud := &battleCloud{
			kind:  asset.kind,
			speed: speed,
			image: asset.image,
			sprite: &engine.Sprite{
				Tag:      cloudSpriteTag(asset.kind),
				Pos:      &engine.Vec{X: x, Y: y},
				Size:     &engine.Vec{X: size.X, Y: size.Y},
				Drawable: engine.NewImageDrawable(asset.image),
			},
		}
		cloud.sprite.Steps = engine.MakeBehaviors(s.behaviorDriftCloud(cloud))
		s.clouds = append(s.clouds, cloud)
	}
}

func (s *GameScene) cloudAssetByKind(kind cloudKind) cloudAsset {
	for _, asset := range s.cloudAssets {
		if asset.kind == kind {
			return asset
		}
	}
	return cloudAsset{}
}

func (s *GameScene) randomNormalCloudAsset() cloudAsset {
	normal := make([]cloudAsset, 0, len(s.cloudAssets))
	for _, asset := range s.cloudAssets {
		if asset.kind == cloudKindNormal {
			normal = append(normal, asset)
		}
	}
	if len(normal) == 0 {
		return s.cloudAssetByKind(cloudKindLightning)
	}
	return normal[s.rng.Intn(len(normal))]
}

func cloudSpriteTag(kind cloudKind) string {
	if kind == cloudKindLightning {
		return "lightning-cloud"
	}
	return "cloud"
}

func (s *GameScene) randomCloudCount() int {
	cfg := core.Config().Gameplay.Clouds
	perScreen := cfg.MinCount + s.rng.Intn(cfg.MaxCount-cfg.MinCount+1)
	screenWidth := core.Config().Screen.Width
	if screenWidth <= 0 {
		return perScreen
	}
	screenCount := math.Max(1, s.worldWidth/screenWidth)
	return maxInt(1, int(math.Round(float64(perScreen)*math.Sqrt(screenCount))))
}

func (s *GameScene) behaviorDriftCloud(cloud *battleCloud) engine.Behavior {
	return func(source *engine.Sprite) {
		if source == nil || source.Pos == nil || source.Size == nil {
			return
		}
		if cloud != nil && cloud.revengeActive {
			return
		}
		direction := s.windDirection
		if direction == 0 {
			direction = 1
		}
		source.Pos.X += float64(direction) * cloud.speed

		screenWidth := core.Config().Screen.Width
		leftEdge := s.cameraX - source.Size.X
		rightEdge := s.cameraX + screenWidth + source.Size.X
		if direction > 0 && source.Pos.X > rightEdge {
			source.Pos.X = leftEdge
		}
		if direction > 0 && source.Pos.X+source.Size.X < s.cameraX-screenWidth {
			source.Pos.X = s.cameraX + screenWidth
		}
		if direction < 0 && source.Pos.X+source.Size.X < s.cameraX {
			source.Pos.X = s.cameraX + screenWidth
		}
		if direction < 0 && source.Pos.X > s.cameraX+screenWidth*2 {
			source.Pos.X = leftEdge
		}
	}
}

func (s *GameScene) cloudSprites() *engine.Sprites {
	sprites := engine.NewSprites()
	for _, cloud := range s.clouds {
		if cloud != nil && cloud.kind != cloudKindLightning && cloud.sprite != nil {
			sprites.Add(cloud.sprite)
		}
	}
	for _, cloud := range s.clouds {
		if cloud != nil && cloud.kind == cloudKindLightning && cloud.sprite != nil {
			sprites.Add(cloud.sprite)
		}
	}
	return sprites
}

func (s *GameScene) cloudImageForKind(kind cloudKind) *ebiten.Image {
	for _, asset := range s.cloudAssets {
		if asset.kind == kind {
			return asset.image
		}
	}
	return nil
}

func (s *GameScene) createPalms() {
	if s.palmImage == nil || s.palmPixels == nil {
		return
	}
	count := s.randomPalmCount()
	if count <= 0 {
		return
	}
	palmBounds := s.palmPixels.Bounds()
	palmSize := engine.V(float64(palmBounds.Dx()), float64(palmBounds.Dy()))
	minCenterX := palmSize.X / 2
	maxCenterX := s.worldWidth - palmSize.X/2
	if maxCenterX <= minCenterX {
		return
	}

	minDistance := s.minimumPalmDistance()
	occupied := s.tankCenterXs()
	attempts := 0
	for len(s.palms) < count && attempts < count*120 {
		attempts++
		centerX := minCenterX + s.rng.Float64()*(maxCenterX-minCenterX)
		if tooCloseToAny(centerX, occupied, minDistance) {
			continue
		}
		s.addPalm(centerX, palmSize)
		occupied = append(occupied, centerX)
	}
	for len(s.palms) < count && attempts < count*180 {
		attempts++
		centerX := minCenterX + s.rng.Float64()*(maxCenterX-minCenterX)
		if tooCloseToAny(centerX, occupied, minDistance*0.55) {
			continue
		}
		s.addPalm(centerX, palmSize)
		occupied = append(occupied, centerX)
	}
}

func (s *GameScene) addPalm(centerX float64, palmSize engine.Vec) *battlePalm {
	surfaceY := s.ground.SurfaceY(centerX)
	palm := &battlePalm{
		pixels:               s.palmPixels,
		aggressionMultiplier: 1.0 + s.rng.Float64()*0.35,
		sprite: &engine.Sprite{
			Tag:      "palm",
			Pos:      &engine.Vec{X: centerX - palmSize.X/2, Y: surfaceY - palmSize.Y},
			Size:     &engine.Vec{X: palmSize.X, Y: palmSize.Y},
			Drawable: engine.NewImageDrawable(s.palmImage),
		},
	}
	s.palms = append(s.palms, palm)
	return palm
}

func (s *GameScene) updatePalmGroundSupport(palm *battlePalm) {
	if palm == nil || palm.sprite == nil || palm.sprite.Pos == nil || palm.sprite.Size == nil {
		return
	}
	if palm.state == palmStateCrumbling {
		return
	}

	centerX := palm.sprite.Pos.X + palm.sprite.Size.X/2
	targetBottom := s.ground.SurfaceY(centerX)
	currentBottom := palm.sprite.Pos.Y + palm.sprite.Size.Y
	if targetBottom <= currentBottom+0.5 {
		return
	}

	const palmSlideSpeed = 4.0
	palm.sprite.Pos.Y += math.Min(palmSlideSpeed, targetBottom-currentBottom)
}

func (s *GameScene) plantPalmAtImpact(pos engine.Vec) {
	if s.palmImage == nil || s.palmPixels == nil {
		return
	}
	s.playSound(soundpaths.SoundGrowing)
	palmBounds := s.palmPixels.Bounds()
	palmSize := engine.V(float64(palmBounds.Dx()), float64(palmBounds.Dy()))
	centerX := math.Max(palmSize.X/2, math.Min(s.worldWidth-palmSize.X/2, pos.X))
	palm := s.addPalm(centerX, palmSize)
	if palm != nil && palm.sprite != nil && s.layers[layerPalms] != nil {
		s.layers[layerPalms].Add(palm.sprite)
	}
}

func (s *GameScene) ignitePalm(palm *battlePalm) {
	if palm == nil || palm.state != palmStateAlive {
		return
	}
	s.playEventSound(soundEventPalmIgnite)
	palm.state = palmStateBurning
	palm.age = 0
}

func (s *GameScene) crumblePalm(palm *battlePalm) {
	if palm == nil || palm.state == palmStateCrumbling {
		return
	}
	s.playEventSound(soundEventPalmCrumble)
	if palm.sprite != nil && s.layers[layerPalms] != nil {
		s.layers[layerPalms].Remove(palm.sprite)
	}
	palm.state = palmStateCrumbling
	palm.age = 0
}

func (s *GameScene) setPalmSkeleton(palm *battlePalm, smoking bool) {
	if palm == nil {
		return
	}
	if s.palmSkeletonImage != nil && s.palmSkeletonPixels != nil {
		palm.pixels = s.palmSkeletonPixels
		if palm.sprite != nil {
			palm.sprite.Drawable = engine.NewImageDrawable(s.palmSkeletonImage)
		}
	}
	if smoking {
		palm.state = palmStateSkeletonSmoking
	} else {
		palm.state = palmStateSkeleton
	}
	palm.age = 0
}

func (s *GameScene) spawnPalmLeafFall(palm *battlePalm) {
	if palm == nil || palm.sprite == nil || len(s.palmLeafImages) == 0 {
		return
	}
	s.playEventSound(soundEventPalmHit)
	bounds := palm.sprite.Bounds()
	leafCount := 18
	for i := 0; i < leafCount; i++ {
		crownWidth := bounds.W() * 0.86
		x := bounds.Center().X - crownWidth/2 + s.rng.Float64()*crownWidth
		y := bounds.Min.Y + bounds.H()*(0.27+s.rng.Float64()*0.08)
		leaf := palmLeafFall{
			image:    s.palmLeafImages[s.rng.Intn(len(s.palmLeafImages))],
			pos:      engine.V(x, y),
			velocity: engine.V((s.rng.Float64()-0.5)*0.45, 0.75+s.rng.Float64()*1.65),
			maxAge:   240,
		}
		s.palmLeafFalls = append(s.palmLeafFalls, leaf)
	}
}

func (s *GameScene) startPalmEyes(palm *battlePalm) {
	if palm == nil || palm.state != palmStateAlive || len(s.palmEyesAnimation.frames) == 0 {
		return
	}
	s.playEventSound(soundEventPalmEyes)
	palm.eyesOn = true
	palm.eyeAge = 0
}

func (s *GameScene) addPalmAggression(palm *battlePalm, attacker *battleTank) bool {
	if palm == nil || palm.state != palmStateAlive || attacker == nil || s.palmRevenge != nil {
		return false
	}
	increase := (palmRevengeAggroMin + s.rng.Float64()*(palmRevengeAggroMax-palmRevengeAggroMin)) * palm.aggressionMultiplier
	palm.aggression += increase
	if palm.aggression < palmRevengeTrigger {
		return false
	}
	palm.aggression = palmRevengeTrigger
	s.startPalmRevenge(palm, attacker)
	return true
}

func (s *GameScene) startPalmRevenge(palm *battlePalm, target *battleTank) {
	cloud := s.lightningCloudForRevenge()
	if palm == nil {
		return
	}
	s.startLightningCloudRevenge(cloud, palm, target)
}

func (s *GameScene) addLightningCloudAggressionForProjectile(p *projectile) bool {
	if p == nil || p.shooter == nil || s.palmRevenge != nil {
		return false
	}
	increase := float64(maxInt(0, s.g.options.cloudAggression))
	if increase <= 0 {
		return false
	}
	if p.angeredClouds == nil {
		p.angeredClouds = make(map[*battleCloud]bool)
	}
	radius := projectileRadiusForWeapon(s.weaponForProjectile(p))
	for _, cloud := range s.clouds {
		if cloud == nil || cloud.kind != cloudKindLightning || cloud.sprite == nil || cloud.revengeActive {
			continue
		}
		if p.angeredClouds[cloud] || !projectilePassesLightningCloud(p.pos, radius, cloud) {
			continue
		}
		p.angeredClouds[cloud] = true
		cloud.aggression += increase
		if cloud.aggression < palmRevengeTrigger {
			if p.searchingCloud == nil && cloud.aggression+increase >= palmRevengeTrigger {
				p.searchingCloud = cloud
			}
			continue
		}
		cloud.aggression = palmRevengeTrigger
		if !s.tankCanAct(p.shooter) {
			p.searchingCloud = cloud
			return true
		}
		s.startLightningCloudRevenge(cloud, nil, p.shooter)
		return true
	}
	return false
}

func (s *GameScene) scheduleCloudSearchForProjectile(p *projectile) {
	if p == nil || p.searchingCloud == nil || p.searchingCloud.sprite == nil || s.palmRevenge != nil || len(s.cloudSearchingAnimation.frames) == 0 {
		return
	}
	if s.cloudSearchAlreadyScheduled(p.searchingCloud) {
		return
	}
	delay := maxInt(0, s.turnAdvanceDelay)
	effect := &cloudSearchEffect{
		cloud:     p.searchingCloud,
		delay:     delay,
		duration:  s.cloudSearchingAnimation.totalTicks,
		animation: s.cloudSearchingAnimation,
	}
	s.cloudSearchEffects = append(s.cloudSearchEffects, effect)
	s.playEventSound(soundEventCloudSearch)
	s.delayTurnAdvance(delay + effect.duration + secondsToFrames(0.25))
}

func (s *GameScene) cloudSearchAlreadyScheduled(cloud *battleCloud) bool {
	for _, effect := range s.cloudSearchEffects {
		if effect != nil && effect.cloud == cloud && effect.age < effect.duration {
			return true
		}
	}
	return false
}

func projectilePassesLightningCloud(pos engine.Vec, projectileRadius float64, cloud *battleCloud) bool {
	if cloud == nil || cloud.sprite == nil {
		return false
	}
	bounds := cloud.sprite.Bounds()
	margin := math.Max(24, projectileRadius+18)
	zone := engine.R(
		bounds.Min.X-margin,
		bounds.Min.Y-margin,
		bounds.Max.X+margin,
		bounds.Max.Y+margin,
	)
	projectileBounds := engine.R(
		pos.X-projectileRadius,
		pos.Y-projectileRadius,
		pos.X+projectileRadius,
		pos.Y+projectileRadius,
	)
	return engine.Collision(zone, projectileBounds)
}

func (s *GameScene) startLightningCloudRevenge(cloud *battleCloud, palm *battlePalm, target *battleTank) {
	if cloud == nil || cloud.sprite == nil || cloud.sprite.Pos == nil || cloud.sprite.Size == nil || target == nil || target.body == nil {
		return
	}
	s.playEventSound(soundEventPalmRevenge)
	if palm != nil {
		palm.eyesOn = false
		palm.screaming = true
		palm.screamAge = 0
	}
	cloud.revengeActive = true
	if s.layers[layerClouds] != nil {
		s.layers[layerClouds].Remove(cloud.sprite)
	}
	start := *cloud.sprite.Pos
	targetBounds := target.body.Bounds()
	attack := engine.V(
		targetBounds.Center().X-cloud.sprite.Size.X/2,
		math.Max(0, targetBounds.Min.Y-130-cloud.sprite.Size.Y),
	)
	s.palmCameraFocus = nil
	s.palmRevenge = &palmRevengeEvent{
		phase:         palmRevengeCloudFocus,
		palm:          palm,
		cloud:         cloud,
		target:        target,
		cloudStart:    start,
		cloudOriginal: start,
		cloudAttack:   attack,
		targetCameraX: s.cameraTargetForWorldX(targetBounds.Center().X),
	}
	s.delayTurnAdvance(palmRevengeFocusFrames + palmRevengeAttackFrames + palmRevengeLightningFrames + palmRevengeRecoverFrames + secondsToFrames(0.5))
}

func (s *GameScene) lightningCloudForRevenge() *battleCloud {
	for _, cloud := range s.clouds {
		if cloud != nil && cloud.kind == cloudKindLightning && cloud.sprite != nil {
			return cloud
		}
	}
	for _, asset := range s.cloudAssets {
		if asset.kind != cloudKindLightning {
			continue
		}
		size := engine.V(asset.size.X, asset.size.Y)
		cloud := &battleCloud{
			kind:  asset.kind,
			speed: 0.35,
			image: asset.image,
			sprite: &engine.Sprite{
				Tag:      cloudSpriteTag(asset.kind),
				Pos:      &engine.Vec{X: s.cameraX + core.Config().Screen.Width*0.5, Y: 30},
				Size:     &engine.Vec{X: size.X, Y: size.Y},
				Drawable: engine.NewImageDrawable(asset.image),
			},
		}
		cloud.sprite.Steps = engine.MakeBehaviors(s.behaviorDriftCloud(cloud))
		s.clouds = append(s.clouds, cloud)
		if s.layers[layerClouds] != nil {
			s.layers[layerClouds].Add(cloud.sprite)
		}
		return cloud
	}
	return nil
}

func (s *GameScene) updatePalmEyes(palm *battlePalm) {
	if palm == nil || !palm.eyesOn {
		return
	}
	palm.eyeAge++
	if palm.eyeAge > s.palmEyesTotalTicks() {
		palm.eyesOn = false
		palm.eyeAge = 0
	}
}

func (s *GameScene) updatePalmGrin(palm *battlePalm) {
	if palm == nil || !palm.grinning {
		return
	}
	palm.grinAge++
	if palm.grinHideAt > 0 && palm.grinAge >= palm.grinHideAt {
		palm.grinning = false
		palm.grinAge = 0
		palm.grinHideAt = 0
	}
}

func (s *GameScene) palmEyesTotalTicks() int {
	return s.palmEyesAnimation.totalTicks + secondsToFrames(1) + s.palmEyesCloseAnimation.totalTicks
}

func (s *GameScene) palmEyesFrame(age int) *ebiten.Image {
	openFrames := s.palmEyesAnimation.frames
	if len(openFrames) == 0 {
		return nil
	}
	openTicks := s.palmEyesAnimation.totalTicks
	if age < openTicks {
		return s.palmEyesAnimation.frameAt(age)
	}
	age -= openTicks
	if age < secondsToFrames(1) {
		return openFrames[len(openFrames)-1]
	}
	age -= secondsToFrames(1)
	if len(s.palmEyesCloseAnimation.frames) > 0 && age < s.palmEyesCloseAnimation.totalTicks {
		return s.palmEyesCloseAnimation.frameAt(age)
	}
	return openFrames[0]
}

func (s *GameScene) palmEffectDelayFrames(palm *battlePalm) int {
	if palm == nil {
		return s.palmHitPauseFrames()
	}
	pause := secondsToFrames(0.5)
	if palm.eyesOn {
		return maxInt(1, s.palmEyesTotalTicks()-palm.eyeAge) + pause
	}
	switch palm.state {
	case palmStateBurning:
		return maxInt(1, s.palmFireAnimation.totalTicks-palm.age) + s.palmSmokeAnimation.totalTicks + pause
	case palmStateSkeletonSmoking:
		return maxInt(1, s.palmSmokeAnimation.totalTicks-palm.age) + pause
	case palmStateCrumbling:
		return maxInt(1, s.palmCrumbleAnimation.totalTicks-palm.age) + pause
	default:
		return s.palmHitPauseFrames()
	}
}

func (s *GameScene) randomPalmCount() int {
	if s.g != nil && s.g.options.palmCount >= 0 {
		return s.g.options.palmCount
	}
	cfg := core.Config().Gameplay.Palms
	minCount := cfg.MinCount
	if minCount <= 0 {
		minCount = 1
	}
	maxCount := cfg.MaxCount
	if maxCount <= 0 {
		maxCount = len(s.players)
	}
	if maxCount < minCount {
		maxCount = minCount
	}
	return minCount + s.rng.Intn(maxCount-minCount+1)
}

func (s *GameScene) placeTanksOnGroundForQuickStart() {
	for _, tank := range s.tanks {
		if tank == nil || tank.body == nil {
			continue
		}
		tank.body.Velocity = engine.Vec{}
		tank.body.Steps = nil
		s.alignTankBodyToSurface(tank.body)
		tank.landed = true
		tank.falling = false
		tank.fallDamage = false
		tank.fallTargetY = 0
		if tank.cannon != nil {
			s.clampCannonRotationToTank(tank.cannon, tank.body)
		}
	}
	s.spawnIndex = len(s.tanks)
	s.spawnPauseFrames = 0
}

func (s *GameScene) replaceTankModel(playerIndex int) {
	if playerIndex < 0 || playerIndex >= len(s.tanks) {
		return
	}
	battleTank := s.tanks[playerIndex]
	if battleTank == nil || !s.playerHasXMV12(playerIndex) {
		return
	}

	var oldPos engine.Vec
	var oldRot float64
	var oldVelocity engine.Vec
	if battleTank.body != nil {
		oldPos = *battleTank.body.Pos
		oldRot = battleTank.body.Rot
		oldVelocity = battleTank.body.Velocity
	} else {
		oldPos = engine.V(200, 600)
	}

	tank := models.NewXMV12Tank(battleTank.player.Name, battleTank.tint)
	body := tank.Body()
	cannon := tank.Cannon()
	if body != nil {
		body.Pos = &engine.Vec{X: oldPos.X, Y: oldPos.Y}
		body.Rot = oldRot
		body.Velocity = oldVelocity
	}
	if cannon != nil {
		cannon.Steps = engine.MakeBehaviors(
			s.behaviorRotateActiveCannon,
		)
		cannon.PostSteps = engine.MakeBehaviors(
			s.behaviorAttachCannonToTank(body),
		)
	}

	s.removeTankSprites(battleTank)
	battleTank.body = body
	battleTank.cannon = cannon
	if battleTank.landed && body != nil {
		s.alignTankBodyToSurface(body)
	}
	if s.layers[layerTanks] != nil {
		s.layers[layerTanks] = engine.AddSprites(s.layers[layerTanks], tank.Sprites)
	}
}

func (s *GameScene) minimumPalmDistance() float64 {
	playerCount := maxInt(1, len(s.players))
	laneWidth := s.worldWidth / float64(playerCount)
	return math.Max(110, laneWidth*0.42)
}

func (s *GameScene) tankCenterXs() []float64 {
	centers := make([]float64, 0, len(s.tanks))
	for _, tank := range s.tanks {
		if tank == nil || tank.body == nil {
			continue
		}
		centers = append(centers, tank.body.Pos.X+tank.body.Size.X/2)
	}
	return centers
}

func tooCloseToAny(x float64, occupied []float64, minDistance float64) bool {
	for _, other := range occupied {
		if math.Abs(x-other) < minDistance {
			return true
		}
	}
	return false
}

func (s *GameScene) palmSprites() *engine.Sprites {
	sprites := engine.NewSprites()
	for _, palm := range s.palms {
		if palm != nil && palm.sprite != nil {
			sprites.Add(palm.sprite)
		}
	}
	return sprites
}

func (s *GameScene) playersForRound() []PlayerConfig {
	if len(s.g.players) > 0 {
		return s.g.players
	}

	count := core.Config().Tanks.Count
	if count <= 0 {
		count = 1
	}
	players := make([]PlayerConfig, count)
	for i := range players {
		players[i] = PlayerConfig{
			Kind:  PlayerHuman,
			Name:  "Player " + strconv.Itoa(i+1),
			Color: defaultTankColor(i),
		}
	}
	return players
}

func defaultTankColor(index int) color.RGBA {
	colors := []color.RGBA{
		{R: 11, G: 31, B: 255, A: 255},
		{R: 230, G: 34, B: 45, A: 255},
	}
	return colors[index%len(colors)]
}

func worldWidthForPlayers(playerCount int) float64 {
	screen := core.Config().Screen
	if playerCount < 1 {
		playerCount = 1
	}
	return screen.Width + float64(playerCount)*screen.Width*0.75
}

func randomTankDropPosition(rng *rand.Rand, index, count int, size *engine.Vec, worldWidth, battlefieldHeight float64) *engine.Vec {
	screen := core.Config().Screen
	laneWidth := worldWidth / float64(count)
	laneCenter := laneWidth*float64(index) + laneWidth/2
	jitter := (rng.Float64() - 0.5) * laneWidth * 0.28
	centerX := laneCenter + jitter
	centerX = math.Max(screen.Width/2, math.Min(worldWidth-screen.Width/2, centerX))
	x := centerX - size.X/2
	x = math.Max(0, math.Min(worldWidth-size.X, x))

	return &engine.Vec{
		X: x,
		Y: -size.Y - rng.Float64()*battlefieldHeight*0.55,
	}
}

func (g *GameScene) Movement(source *engine.Sprite) {
	if MoveLeft() {
		// move left
		source.Velocity = source.Velocity.Rotated(source.Rot)
		engine.Movement(source)
	} else {
		source.Velocity = source.Velocity.Rotated(source.Rot)
		engine.Movement(source)
	}
}

func (s *GameScene) Update() error {
	s.time += 1
	s.consumeOnlineGameCommands()
	if s.phase != phaseShop {
		if err := s.handleGameDialogInput(); err != nil {
			return err
		}
		if s.abortRoundFlashing {
			s.updateAbortRoundFlash()
			return nil
		}
		if s.gameHelpOpen || s.playerInfoOpen || s.gamePaused {
			return nil
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		s.showScoreTable = !s.showScoreTable
	}
	if s.allTanksLanded() {
		s.handleCameraScrollControls()
	}

	switch s.phase {
	// countdownTime := s.time * 2
	// if countdownTime >= 3 {
	// 	s.phase = phaseBattle
	// 	break
	// }
	// seconds := 3 - int(countdownTime)

	// countdownColorIndex := 3 - seconds
	// if countdownColorIndex < 0 {
	// 	countdownColorIndex = 0
	// }
	// text := engine.NewText(s.messageFace, countdownColors[countdownColorIndex], strconv.Itoa(seconds))
	// s.message = &text
	case phaseBattle:
		s.updateSpawnSequence()
		s.updateImpacts()
		s.updateAnimatedImpacts()
		s.updateWaterFills()
		s.updateWaterBlubbers()
		s.updateWaterBlotches()
		s.updateMoskitoEffects()
		s.updateShockwaveImpacts()
		s.updateAirStrikeImpacts()
		s.updateLaserEffects()
		s.updateMoleImpacts()
		s.updateSmallCrumblerImpacts()
		s.updateSandFalls()
		s.updatePalms()
		s.updatePalmLeafFalls()
		s.updatePalmRevenge()
		s.updateZeroPowerEffects()
		s.syncBattleEffectLoops()
		if s.roundTransitionDelay > 0 || s.roundSeriesComplete {
			return s.updateRoundTransition()
		}
		if s.allTanksLanded() {
			if s.projectilesActive() {
				s.updateProjectile()
			} else if s.turnAdvanceDelay > 0 {
				s.updateTurnAdvanceDelay()
			} else if s.turnAdvanceBlocked() {
				s.updateBattleCamera()
			} else if s.xmV12DriveMode {
				s.updateXMV12DriveMode()
			} else {
				if s.ensureActivePlayerCanAct() {
					s.clampActiveShotStrength()
					s.handleBattleInput()
					s.handleComputerTurn()
					s.updateProjectile()
				}
				if s.turnAdvanceDelay == 0 {
					s.updateBattleCamera()
				}
			}
		}
		s.layers.Update()
		s.advanceSpawnSequence()
		s.chooseStartingPlayerAfterLanding()
	case phaseShop:
		s.handleShopInput()
	case phaseBlueVictory:
		fallthrough
	case phaseRedVictory:
		// if s.victoryTime <= 0 {
		// 	return s.g.SetNewScene(NewTitleScene)
		// }
	}

	return nil
}

func (s *GameScene) updateSpawnSequence() {
	if s.allTanksLanded() || s.spawnIndex >= len(s.tanks) {
		return
	}

	s.cameraGoal = s.cameraTargetForTank(s.spawnIndex)
	s.cameraGoalY = 0
	s.cameraX = approach(s.cameraX, s.cameraGoal, 0.08, 0.35)
	s.cameraY = approach(s.cameraY, s.cameraGoalY, 0.08, 0.35)

	tank := s.tanks[s.spawnIndex]
	if tank == nil || tank.body == nil || tank.falling || tank.landed {
		return
	}
	if math.Abs(s.cameraX-s.cameraGoal) > 1 {
		return
	}

	tank.falling = true
	tank.fallDamage = false
	tank.body.Steps = engine.MakeBehaviors(
		s.behaviorFallOntoGround(s.ground, tank, true),
	)
}

func (s *GameScene) advanceSpawnSequence() {
	if s.spawnIndex >= len(s.tanks) {
		return
	}
	tank := s.tanks[s.spawnIndex]
	if tank == nil || !tank.landed {
		return
	}
	if s.spawnPauseFrames > 0 {
		s.spawnPauseFrames--
		return
	}

	s.spawnIndex++
	if s.spawnIndex < len(s.tanks) {
		s.cameraGoal = s.cameraTargetForTank(s.spawnIndex)
		s.cameraGoalY = 0
	}
}

func (s *GameScene) cameraTargetForTank(index int) float64 {
	if index < 0 || index >= len(s.tanks) || s.tanks[index] == nil || s.tanks[index].body == nil {
		return s.cameraX
	}

	body := s.tanks[index].body
	centerX := body.Pos.X + body.Size.X/2
	return s.cameraTargetForWorldX(centerX)
}

func (s *GameScene) cameraTargetForWorldX(centerX float64) float64 {
	screenWidth := core.Config().Screen.Width
	return math.Max(0, math.Min(s.worldWidth-screenWidth, centerX-screenWidth/2))
}

func (s *GameScene) cameraTargetForWorldY(centerY float64) float64 {
	target := centerY - s.battlefieldHeight()*0.35
	return math.Max(-s.skyExtraHeight(), math.Min(0, target))
}

func approach(current, target, smoothing, minStep float64) float64 {
	delta := target - current
	if math.Abs(delta) <= minStep {
		return target
	}
	return current + delta*smoothing
}

type zeroPowerAnimationSheet struct {
	name        string
	data        []byte
	frameWidth  int
	frameHeight int
	delay       int
	anchor      zeroPowerAnimationAnchor
	scaleX      float64
	scaleY      float64
}

func loadSpriteAnimation(spec zeroPowerAnimationSheet) (spriteAnimation, error) {
	decoded, _, err := image.Decode(bytes.NewReader(spec.data))
	if err != nil {
		return spriteAnimation{}, err
	}

	bounds := decoded.Bounds()
	frameHeight := spec.frameHeight
	if frameHeight <= 0 {
		frameHeight = bounds.Dy()
	}
	cols := bounds.Dx() / spec.frameWidth
	rows := bounds.Dy() / frameHeight
	scaleX := spec.scaleX
	if scaleX <= 0 {
		scaleX = 1
	}
	scaleY := spec.scaleY
	if scaleY <= 0 {
		scaleY = 1
	}
	animation := spriteAnimation{
		frames: make([]*ebiten.Image, 0, cols*rows),
		pixels: make([]*image.RGBA, 0, cols*rows),
		delays: make([]int, 0, cols*rows),
		width:  spec.frameWidth,
		height: frameHeight,
		scaleX: scaleX,
		scaleY: scaleY,
		anchor: spec.anchor,
	}
	ticks := maxInt(1, int(math.Round(float64(spec.delay)*60/100)))
	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			src := image.Rect(
				bounds.Min.X+col*spec.frameWidth,
				bounds.Min.Y+row*frameHeight,
				bounds.Min.X+(col+1)*spec.frameWidth,
				bounds.Min.Y+(row+1)*frameHeight,
			)
			frame := image.NewRGBA(image.Rect(0, 0, spec.frameWidth, frameHeight))
			for y := 0; y < frameHeight; y++ {
				for x := 0; x < spec.frameWidth; x++ {
					frame.Set(x, y, decoded.At(src.Min.X+x, src.Min.Y+y))
				}
			}
			animation.frames = append(animation.frames, ebiten.NewImageFromImage(frame))
			animation.pixels = append(animation.pixels, frame)
			animation.delays = append(animation.delays, ticks)
			animation.totalTicks += ticks
		}
	}
	if animation.totalTicks <= 0 {
		animation.totalTicks = zeroPowerFrames
	}
	return animation, nil
}

func loadZeroPowerAnimations() ([]spriteAnimation, []string, error) {
	sources := []zeroPowerAnimationSheet{
		{name: "dust", data: r.ZeroPowerDustExplosionPNG, frameWidth: 20, delay: 6, scaleX: 1, scaleY: 1},
		{name: "explosion", data: r.ZeroPowerExplosionPNG, frameWidth: 65, frameHeight: 59, delay: 6},
		{name: "mushroom", data: r.ZeroPowerMushroomExplosionPNG, frameWidth: 51, frameHeight: 57, delay: 6, anchor: zeroPowerAnchorTankBottom},
		{name: "smoke", data: r.ZeroPowerPlayerSmokePNG, frameWidth: 15, delay: 6, scaleX: 1.0, scaleY: 1.0},
	}
	animations := make([]spriteAnimation, 0, len(sources))
	names := make([]string, 0, len(sources))
	for _, source := range sources {
		animation, err := loadSpriteAnimation(source)
		if err != nil {
			return nil, nil, err
		}
		if len(animation.frames) > 0 {
			animations = append(animations, animation)
			names = append(names, source.name)
		}
	}
	return animations, names, nil
}

func loadPalmAsset() (*ebiten.Image, *image.RGBA, error) {
	return loadImageWithPixels(r.Palm)
}

func loadPalmLeafImages(data []byte) ([]*ebiten.Image, error) {
	decoded, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	bounds := decoded.Bounds()
	if bounds.Dx() < 3 || bounds.Dy() == 0 {
		return nil, nil
	}
	sheet := ebiten.NewImageFromImage(decoded)
	leaves := make([]*ebiten.Image, 0, bounds.Dx()/3)
	for x := bounds.Min.X; x+3 <= bounds.Max.X; x += 3 {
		rect := image.Rect(x, bounds.Min.Y, x+3, bounds.Max.Y)
		if leaf, ok := sheet.SubImage(rect).(*ebiten.Image); ok {
			leaves = append(leaves, leaf)
		}
	}
	return leaves, nil
}

func loadImageWithPixels(data []byte) (*ebiten.Image, *image.RGBA, error) {
	decoded, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, nil, err
	}
	bounds := decoded.Bounds()
	pixels := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			pixels.Set(x, y, decoded.At(bounds.Min.X+x, bounds.Min.Y+y))
		}
	}
	return ebiten.NewImageFromImage(pixels), pixels, nil
}

func fitAnimationDuration(animation spriteAnimation, totalTicks int) spriteAnimation {
	if len(animation.frames) == 0 || totalTicks <= 0 {
		return animation
	}
	ticks := maxInt(1, int(math.Round(float64(totalTicks)/float64(len(animation.frames)))))
	animation.delays = make([]int, len(animation.frames))
	animation.totalTicks = 0
	for i := range animation.delays {
		animation.delays[i] = ticks
		animation.totalTicks += ticks
	}
	return animation
}

func repeatAnimation(animation spriteAnimation, repeats, totalTicks int) spriteAnimation {
	if repeats <= 1 || len(animation.frames) == 0 {
		return fitAnimationDuration(animation, totalTicks)
	}
	frames := animation.frames
	pixels := animation.pixels
	animation.frames = make([]*ebiten.Image, 0, len(frames)*repeats)
	animation.pixels = make([]*image.RGBA, 0, len(pixels)*repeats)
	for i := 0; i < repeats; i++ {
		animation.frames = append(animation.frames, frames...)
		animation.pixels = append(animation.pixels, pixels...)
	}
	return fitAnimationDuration(animation, totalTicks)
}

func (a spriteAnimation) frameAt(tick int) *ebiten.Image {
	if len(a.frames) == 0 {
		return nil
	}
	if a.totalTicks > 0 && tick >= a.totalTicks {
		tick = a.totalTicks - 1
	}
	for i, delay := range a.delays {
		if tick < delay {
			return a.frames[i]
		}
		tick -= delay
	}
	return a.frames[len(a.frames)-1]
}

func (s *GameScene) Draw(screen *ebiten.Image) {
	if s.phase == phaseShop {
		s.drawShop(screen)
		return
	}

	camera := ebiten.GeoM{}
	camera.Translate(-s.cameraX, -s.cameraY)

	for layerIndex, layer := range s.layers {
		if layerIndex == layerWater {
			s.drawWaterFills(screen, &camera)
		}
		layer.Draw(&camera, screen)
	}
	s.drawImpacts(screen, &camera)
	s.drawAnimatedImpacts(screen, &camera)
	s.drawCloudSearchEffects(screen, &camera)
	s.drawWaterBlubbers(screen, &camera)
	s.drawWaterBlotches(screen, &camera)
	s.drawMoskitoEffects(screen, &camera)
	s.drawShockwaveImpacts(screen, &camera)
	s.drawAirStrikeImpacts(screen, &camera)
	s.drawMoleImpacts(screen, &camera)
	s.drawSmallCrumblerImpacts(screen, &camera)
	s.drawPalmEffects(screen, &camera)
	s.drawPalmLeafFalls(screen, &camera)
	s.drawPalmRevenge(screen, &camera)
	s.drawSandFalls(screen, &camera)
	s.drawZeroPowerEffects(screen, &camera)
	s.drawProjectile(screen, &camera)
	s.drawLaserEffects(screen, &camera)
	s.drawGameHUD(screen)
	s.drawPalmRevengeScreenFlash(screen)
	s.drawAbortRoundFlash(screen)
	s.drawDebugScrollBar(screen)
	s.drawPlayerNames(screen, &camera)
	s.drawScoreTable(screen)
	s.drawRoundTransitionBanner(screen)
	s.drawProjectileReentryAnimation(screen)
	s.drawGameDialogs(screen)

	switch s.phase {
	case phaseBattle:
	case phaseBlueVictory:
		fallthrough
	case phaseRedVictory:
		// fallthrough
		// case phaseCountdown:
		// if s.message == nil {
		// 	return
		// }
		// s.message.Draw(image, core.Config().Screen.Width/2, core.Config().Screen.Height/2+s.message.H/2, engine.AlignCenter)
	}
}

func (s *GameScene) handleCameraScrollControls() {
	if s.g.online != nil && !s.onlineCanControlActivePlayer() {
		return
	}
	if !s.scrollBarAvailable() {
		return
	}
	previousCameraX := s.cameraX
	if primaryPointerJustPressed() {
		x, y := primaryPointerPosition()
		s.scrollBarDragging = image.Pt(x, y).In(debugScrollBarRect())
	}
	if !primaryPointerPressed() {
		s.scrollBarDragging = false
	}
	if s.scrollBarDragging {
		x, _ := primaryPointerPosition()
		s.setCameraFromScrollBarX(x)
	}

	if !s.scrollOMatActive() {
		return
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		s.cameraX = math.Max(0, s.cameraX-scrollOMatKeyboardStep)
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		maxCameraX := math.Max(0, s.worldWidth-core.Config().Screen.Width)
		s.cameraX = math.Min(maxCameraX, s.cameraX+scrollOMatKeyboardStep)
	}
	if s.cameraX != previousCameraX {
		s.syncOnlineAim(s.activeTank())
	}
}

func (s *GameScene) handleGameDialogInput() error {
	if s.abortRoundFlashing {
		return nil
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyF1) {
		s.gameHelpOpen = true
		s.playerInfoOpen = false
		s.confirmDialog = gameConfirmNone
		s.pressedDialogButton = ""
		return nil
	}
	if s.confirmDialog != gameConfirmNone {
		return s.handleConfirmDialogInput()
	}
	if s.gameHelpOpen || s.playerInfoOpen {
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeyKPEnter) {
			s.gameHelpOpen = false
			s.playerInfoOpen = false
			s.pressedDialogButton = ""
			return nil
		}
		if primaryPointerJustPressed() {
			x, y := primaryPointerPosition()
			if s.gameHelpOpen && image.Pt(x, y).In(gameHelpOKRect()) {
				s.pressedDialogButton = "game_help_ok"
				return nil
			}
			if s.playerInfoOpen && image.Pt(x, y).In(playerInfoOKRect()) {
				s.pressedDialogButton = "player_info_ok"
				return nil
			}
		}
		if primaryPointerJustReleased() {
			x, y := primaryPointerPosition()
			p := image.Pt(x, y)
			button := s.pressedDialogButton
			s.pressedDialogButton = ""
			switch {
			case button == "game_help_ok" && s.gameHelpOpen && p.In(gameHelpOKRect()):
				s.gameHelpOpen = false
				return nil
			case button == "player_info_ok" && s.playerInfoOpen && p.In(playerInfoOKRect()):
				s.playerInfoOpen = false
				return nil
			}
		}
		if index, ok := pressedPlayerInfoIndex(); ok {
			s.openPlayerInfo(index)
		}
		return nil
	}
	if index, ok := pressedPlayerInfoIndex(); ok {
		s.openPlayerInfo(index)
		return nil
	}
	if controlPressed() {
		switch {
		case inpututil.IsKeyJustPressed(ebiten.KeyQ):
			s.openConfirmDialog(gameConfirmQuit)
			return nil
		case inpututil.IsKeyJustPressed(ebiten.KeyE):
			s.openConfirmDialog(gameConfirmAbortRound)
			return nil
		case inpututil.IsKeyJustPressed(ebiten.KeyP):
			s.gamePaused = !s.gamePaused
			return nil
		case inpututil.IsKeyJustPressed(ebiten.KeyN):
			s.showPlayerNames = !s.showPlayerNames
			return nil
		}
	}
	return nil
}

func (s *GameScene) openConfirmDialog(kind gameConfirmDialog) {
	s.confirmDialog = kind
	s.gameHelpOpen = false
	s.playerInfoOpen = false
	s.pressedDialogButton = ""
}

func (s *GameScene) handleConfirmDialogInput() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyN) {
		s.closeConfirmDialog()
		return nil
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeyKPEnter) || inpututil.IsKeyJustPressed(ebiten.KeyY) || inpututil.IsKeyJustPressed(ebiten.KeyJ) {
		return s.confirmDialogYes()
	}
	if primaryPointerJustPressed() {
		x, y := primaryPointerPosition()
		switch {
		case image.Pt(x, y).In(confirmDialogYesRect()):
			s.pressedDialogButton = "confirm_yes"
		case image.Pt(x, y).In(confirmDialogNoRect()):
			s.pressedDialogButton = "confirm_no"
		}
		return nil
	}
	if primaryPointerJustReleased() {
		x, y := primaryPointerPosition()
		p := image.Pt(x, y)
		button := s.pressedDialogButton
		s.pressedDialogButton = ""
		switch {
		case button == "confirm_yes" && p.In(confirmDialogYesRect()):
			return s.confirmDialogYes()
		case button == "confirm_no" && p.In(confirmDialogNoRect()):
			s.closeConfirmDialog()
		}
	}
	return nil
}

func (s *GameScene) confirmDialogYes() error {
	switch s.confirmDialog {
	case gameConfirmQuit:
		s.stopBattleEffectLoops()
		return RegularTermination
	case gameConfirmAbortRound:
		s.closeConfirmDialog()
		s.startAbortRoundFlash()
	}
	return nil
}

func (s *GameScene) closeConfirmDialog() {
	s.confirmDialog = gameConfirmNone
	s.pressedDialogButton = ""
}

func (s *GameScene) startAbortRoundFlash() {
	s.stopBattleEffectLoops()
	s.playSound(soundpaths.SoundNukeall)
	s.abortRoundFlashing = true
	s.abortRoundFlashAge = 0
	s.projectile = nil
	s.projectiles = nil
	s.turnAdvanceDelay = 0
}

func (s *GameScene) updateAbortRoundFlash() {
	if !s.abortRoundFlashing {
		return
	}
	s.abortRoundFlashAge++
	if s.abortRoundFlashAge < abortRoundFlashFrames {
		return
	}
	s.abortRoundFlashing = false
	s.abortRoundFlashAge = 0
	s.abortCurrentRound()
}

func (s *GameScene) abortCurrentRound() {
	if s.roundNumber >= maxInt(1, s.g.rounds) {
		s.roundSeriesComplete = true
		s.roundTransitionDelay = 0
		return
	}
	s.roundNumber++
	s.beginShop()
}

func (s *GameScene) openPlayerInfo(index int) {
	if index < 0 || index >= len(s.tanks) {
		return
	}
	s.playerInfoIndex = index
	s.playerInfoOpen = true
	s.gameHelpOpen = false
}

func pressedPlayerInfoIndex() (int, bool) {
	keys := []ebiten.Key{
		ebiten.Key1, ebiten.Key2, ebiten.Key3, ebiten.Key4, ebiten.Key5,
		ebiten.Key6, ebiten.Key7, ebiten.Key8, ebiten.Key9, ebiten.Key0,
	}
	keypad := []ebiten.Key{
		ebiten.KeyKP1, ebiten.KeyKP2, ebiten.KeyKP3, ebiten.KeyKP4, ebiten.KeyKP5,
		ebiten.KeyKP6, ebiten.KeyKP7, ebiten.KeyKP8, ebiten.KeyKP9, ebiten.KeyKP0,
	}
	for index, key := range keys {
		if inpututil.IsKeyJustPressed(key) || inpututil.IsKeyJustPressed(keypad[index]) {
			return index, true
		}
	}
	return 0, false
}

func controlPressed() bool {
	return ebiten.IsKeyPressed(ebiten.KeyControl) || ebiten.IsKeyPressed(ebiten.KeyControlLeft) || ebiten.IsKeyPressed(ebiten.KeyControlRight)
}

func shiftPressed() bool {
	return ebiten.IsKeyPressed(ebiten.KeyShift) || ebiten.IsKeyPressed(ebiten.KeyShiftLeft) || ebiten.IsKeyPressed(ebiten.KeyShiftRight)
}

func (s *GameScene) setCameraFromScrollBarX(x int) {
	bar := debugScrollBarRect()
	if bar.Dx() <= 0 {
		return
	}
	t := float64(x-bar.Min.X) / float64(bar.Dx())
	t = math.Max(0, math.Min(1, t))
	s.cameraX = math.Max(0, math.Min(s.worldWidth-core.Config().Screen.Width, t*(s.worldWidth-core.Config().Screen.Width)))
}

func (s *GameScene) drawDebugScrollBar(screen *ebiten.Image) {
	if !s.scrollBarAvailable() {
		return
	}

	bar := debugScrollBarRect()
	drawFrame(screen, bar, color.RGBA{R: 45, G: 45, B: 45, A: 230}, colornames.Black)

	visiblePart := core.Config().Screen.Width / s.worldWidth
	handleW := math.Max(36, float64(bar.Dx())*visiblePart)
	maxCameraX := s.worldWidth - core.Config().Screen.Width
	t := 0.0
	if maxCameraX > 0 {
		t = s.cameraX / maxCameraX
	}
	handleX := float64(bar.Min.X) + (float64(bar.Dx())-handleW)*t
	handle := image.Rect(int(handleX), bar.Min.Y+3, int(handleX+handleW), bar.Max.Y-3)
	drawFrame(screen, handle, color.RGBA{R: 230, G: 230, B: 230, A: 240}, colornames.White)
}

func (s *GameScene) scrollBarAvailable() bool {
	if s.worldWidth <= core.Config().Screen.Width {
		return false
	}
	return s.scrollOMatActive() || s.scrollBarDragging
}

func (s *GameScene) handleBattleInput() {
	if !s.activePlayerCanAdjustShot() {
		return
	}
	tank := s.activeTank()
	if tank == nil {
		return
	}
	if tank.player.Kind == PlayerComputer {
		return
	}
	if !s.onlineCanControlPlayer(tank.playerIndex) {
		return
	}

	strengthStep := 1
	if shiftPressed() {
		strengthStep = 10
	}
	if shouldAdjustStrength(ebiten.KeyArrowUp) {
		s.adjustShotStrength(tank, strengthStep)
	}
	if shouldAdjustStrength(ebiten.KeyArrowDown) {
		s.adjustShotStrength(tank, -strengthStep)
	}
	s.handleBattleHUDButtons(tank, strengthStep)
	s.handleMobileSideControls(tank, strengthStep)
	if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		if shiftPressed() {
			s.selectPreviousWeapon(tank)
		} else {
			s.selectNextWeapon(tank)
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEqual) || inpututil.IsKeyJustPressed(ebiten.KeyKPAdd) {
		s.adjustShotStrength(tank, 10)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyMinus) || inpututil.IsKeyJustPressed(ebiten.KeyKPSubtract) {
		s.adjustShotStrength(tank, -10)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyI) {
		s.invertCannonAngle(tank)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyS) {
		s.toggleScrollOMat(tank)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyM) && s.playerHasXMV12(tank.playerIndex) && s.dieselForPlayer(tank.playerIndex) > 0 {
		s.requestXMV12Start(tank)
		return
	}

	if primaryPointerJustPressed() {
		x, y := primaryPointerPosition()
		for i := 0; i < s.weaponSlotCount(); i++ {
			if image.Pt(x, y).In(s.weaponSlotRect(i)) && s.canSelectWeaponSlot(tank, i) {
				s.setSelectedWeapon(tank, i)
				break
			}
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeyKPEnter) {
		s.requestFireActiveWeapon()
	}
}

func (s *GameScene) handleBattleHUDButtons(tank *battleTank, strengthStep int) {
	if tank == nil || !s.hudMouseAction() {
		return
	}
	x, y := primaryPointerPosition()
	cursor := image.Pt(x, y)
	switch {
	case cursor.In(s.hudStrengthMinusRect()):
		s.adjustShotStrength(tank, -strengthStep)
	case cursor.In(s.hudStrengthPlusRect()):
		s.adjustShotStrength(tank, strengthStep)
	case cursor.In(s.hudAngleMinusRect()):
		s.adjustTankCannon(tank, -s.humanCannonStep())
	case cursor.In(s.hudAnglePlusRect()):
		s.adjustTankCannon(tank, s.humanCannonStep())
	case cursor.In(s.hudFireButtonRect()):
		s.requestFireActiveWeapon()
	case cursor.In(s.hudIgnitionRect()) && s.playerHasXMV12(tank.playerIndex) && s.dieselForPlayer(tank.playerIndex) > 0:
		s.requestXMV12Start(tank)
	}
}

func (s *GameScene) adjustShotStrength(tank *battleTank, delta int) {
	if tank == nil || delta == 0 {
		return
	}
	previous := tank.shotStrength
	tank.shotStrength = maxInt(s.minShotStrength(), minInt(s.maxShotStrength(), tank.shotStrength+delta))
	if tank.shotStrength == previous {
		return
	}
	defer s.syncOnlineAim(tank)
	if delta > 0 {
		s.playEventSound(soundEventCannonPowerUp)
		return
	}
	s.playEventSound(soundEventCannonPowerDown)
}

func (s *GameScene) hudMouseAction() bool {
	if primaryPointerJustPressed() {
		return true
	}
	return primaryPointerPressed() && int(s.time)%humanCannonRepeatFrames == 0
}

func (s *GameScene) handleComputerTurn() {
	if s.projectilesActive() || s.activePlayerIndex < 0 || s.activePlayerIndex >= len(s.tanks) {
		return
	}
	tank := s.activeTank()
	if tank == nil || tank.player.Kind != PlayerComputer || !s.tankCanAct(tank) {
		return
	}
	if tank.computerPlan == nil {
		decisionRNG := s.rng
		if s.g.online != nil {
			decisionRNG = rand.New(rand.NewSource(s.onlineComputerDecisionSeed(tank)))
		}
		decision := computerplayers.Decide(s.effectiveComputerID(tank), s.computerPlayerState(tank), decisionRNG)
		tank.computerPlan = &computerTurnPlan{
			phase:          computerTurnWaitCamera,
			decision:       decision,
			targetStrength: s.clampedComputerStrength(decision.Strength),
			targetAngle:    math.Max(0, math.Min(180, decision.AngleDegrees)),
		}
	}
	s.updateComputerTurnPlan(tank)
}

func (s *GameScene) updateComputerTurnPlan(tank *battleTank) {
	plan := tank.computerPlan
	if plan == nil {
		return
	}

	switch plan.phase {
	case computerTurnWaitCamera:
		s.cameraGoal = s.cameraTargetForTank(s.activePlayerIndex)
		s.cameraGoalY = 0
		if math.Abs(s.cameraX-s.cameraGoal) > 2 || math.Abs(s.cameraY-s.cameraGoalY) > 2 {
			return
		}
		plan.phase = computerTurnAdjustStrength
	case computerTurnAdjustStrength:
		if tank.shotStrength == plan.targetStrength {
			plan.adjustProgress = 0
			plan.phase = computerTurnAdjustAngle
			return
		}
		plan.adjustProgress += computerAdjustSpeedFactor
		adjustInterval := float64(4)
		if plan.adjustProgress < adjustInterval {
			return
		}
		plan.adjustProgress -= adjustInterval
		if tank.shotStrength < plan.targetStrength {
			s.adjustShotStrength(tank, 1)
			return
		}
		if tank.shotStrength > plan.targetStrength {
			s.adjustShotStrength(tank, -1)
			return
		}
	case computerTurnAdjustAngle:
		if s.adjustComputerCannonAngle(tank, plan.targetAngle) {
			return
		}
		plan.phase = computerTurnSelectWeapon
	case computerTurnSelectWeapon:
		if s.canSelectWeaponSlot(tank, plan.decision.WeaponSlot) {
			tank.selectedWeapon = plan.decision.WeaponSlot
		} else {
			tank.selectedWeapon = 0
		}
		plan.delay = maxInt(1, plan.decision.DelayFrames/4)
		plan.phase = computerTurnFireDelay
	case computerTurnFireDelay:
		plan.delay--
		if plan.delay > 0 {
			return
		}
		s.requestFireActiveWeapon()
	}
}

func (s *GameScene) updateXMV12DriveMode() {
	tank := s.activeTank()
	if tank == nil || !s.playerHasXMV12(tank.playerIndex) || tank.player.Kind == PlayerComputer {
		s.clearXMV12IdleVibration(tank)
		s.stopXMV12LoopSounds()
		s.xmV12DriveMode = false
		return
	}
	s.updateXMV12LoopSounds(tank)
	if s.onlineCanControlPlayer(tank.playerIndex) {
		s.handleXMV12HUDInput(tank)
		s.handleMobileXMV12SideControls(tank)
	}
	if s.xmV12EngineOffDelay > 0 {
		s.clearXMV12IdleVibration(tank)
		s.stopXMV12LoopSounds()
		s.xmV12EngineOffDelay--
		if s.xmV12EngineOffDelay == 0 {
			s.xmV12DriveMode = false
			s.xmV12DriveDirection = 0
			s.turnAdvanceDelay = 1
		}
		s.updateBattleCamera()
		return
	}
	if s.xmV12DriveDirection == 0 {
		s.applyXMV12IdleVibration(tank)
		s.updateBattleCamera()
		return
	}
	if s.dieselForPlayer(tank.playerIndex) <= 0 {
		s.clearXMV12IdleVibration(tank)
		s.endXMV12Turn()
		return
	}
	if tank.body == nil || tank.body.Pos == nil {
		s.clearXMV12IdleVibration(tank)
		s.endXMV12Turn()
		return
	}
	s.clearXMV12IdleVibration(tank)
	s.updateXMV12Facing(tank, s.xmV12DriveDirection)
	tank.body.Pos.X += float64(s.xmV12DriveDirection) * xmV12DriveSpeed
	s.ensureInventory(tank.playerIndex)
	s.inventories[tank.playerIndex].diesel = math.Max(0, s.inventories[tank.playerIndex].diesel-xmV12DieselPerFrame)
	if tank.body.Pos.X+tank.body.Size.X < 0 || tank.body.Pos.X > s.worldWidth {
		s.destroyTankOutOfBounds(tank)
		return
	}
	s.alignTankBodyToSurface(tank.body)
	if tank.cannon != nil {
		if s.xmV12DriveDirection < 0 {
			tank.cannon.Rot = tank.body.Rot - math.Pi
		} else {
			tank.cannon.Rot = tank.body.Rot
		}
		s.clampCannonRotationToTank(tank.cannon, tank.body)
	}
	s.cameraGoal = s.cameraTargetForTank(s.activePlayerIndex)
	s.cameraGoalY = 0
	s.updateBattleCamera()
	if s.dieselForPlayer(tank.playerIndex) <= 0 {
		s.endXMV12Turn()
	}
}

func (s *GameScene) applyXMV12IdleVibration(tank *battleTank) {
	if tank == nil || tank.body == nil || tank.body.Pos == nil || !s.playerHasXMV12(tank.playerIndex) {
		return
	}
	s.clearXMV12IdleVibration(tank)
	centerX := tank.body.Pos.X + tank.body.Size.X/2
	s.alignTankBodyToSurface(tank.body)
	tank.body.Pos.X = centerX - tank.body.Size.X/2
	if int(s.time)%4 >= 2 {
		return
	}
	normal := engine.V(-math.Sin(tank.body.Rot), -math.Cos(tank.body.Rot))
	s.xmV12IdleOffset = normal.Scaled(2)
	tank.body.Pos = tank.body.Pos.Add(s.xmV12IdleOffset)
	if tank.cannon != nil {
		s.clampCannonRotationToTank(tank.cannon, tank.body)
	}
}

func (s *GameScene) clearXMV12IdleVibration(tank *battleTank) {
	if s.xmV12IdleOffset.X == 0 && s.xmV12IdleOffset.Y == 0 {
		return
	}
	if tank != nil && tank.body != nil && tank.body.Pos != nil {
		pos := tank.body.Pos.Sub(s.xmV12IdleOffset)
		tank.body.Pos = &engine.Vec{X: pos.X, Y: pos.Y}
	}
	s.xmV12IdleOffset = engine.Vec{}
}

func (s *GameScene) updateXMV12LoopSounds(tank *battleTank) {
	if tank == nil || !s.xmV12DriveMode || s.xmV12EngineOffDelay > 0 {
		s.stopXMV12LoopSounds()
		return
	}
	s.playEventSoundLoop(xmV12EngineLoopKey, soundEventXMV12EngineLoop)
	if s.xmV12DriveDirection != 0 {
		s.playEventSoundLoop(xmV12TrackLoopKey, soundEventXMV12TrackLoop)
		return
	}
	s.stopSoundLoop(xmV12TrackLoopKey)
}

func (s *GameScene) stopXMV12LoopSounds() {
	s.stopSoundLoop(xmV12EngineLoopKey)
	s.stopSoundLoop(xmV12TrackLoopKey)
}

func (s *GameScene) syncBattleEffectLoops() {
	if len(s.laserEffects) == 0 {
		s.stopSoundLoop(laserLoopKey)
	}
	if len(s.moskitoEffects) == 0 {
		s.stopSoundLoop(moskitosLoopKey)
	}
	if len(s.moleImpacts) == 0 {
		s.stopSoundLoop(moleBroeslerLoopKey)
	}
	if len(s.smallCrumblerImpacts) == 0 {
		s.stopSoundLoop(crumblerBroeslerLoopKey)
	}
	if len(s.airStrikeImpacts) == 0 {
		s.stopSoundLoop(airStrikeBeaconLoopKey)
	}
}

func (s *GameScene) stopBattleEffectLoops() {
	s.stopXMV12LoopSounds()
	s.stopSoundLoop(moskitosLoopKey)
	s.stopSoundLoop(fireballBurningLoopKey)
	s.stopSoundLoop(moleBroeslerLoopKey)
	s.stopSoundLoop(crumblerBroeslerLoopKey)
	s.stopSoundLoop(laserLoopKey)
	s.stopSoundLoop(airStrikeBeaconLoopKey)
	s.stopSoundLoop(airStrikeJetLoopKey)
}

func (s *GameScene) updateXMV12Facing(tank *battleTank, facing int) {
	if tank == nil || tank.body == nil || !s.playerHasXMV12(tank.playerIndex) || facing == 0 {
		return
	}
	models.SetTankFacing(tank.body, tank.tint, facing)
}

func (s *GameScene) updateXMV12FacingFromCannon(tank *battleTank) {
	if tank == nil || !s.playerHasXMV12(tank.playerIndex) {
		return
	}
	if s.cannonAngleDegreesForTank(tank) < 90 {
		s.updateXMV12Facing(tank, -1)
		return
	}
	s.updateXMV12Facing(tank, 1)
}

func (s *GameScene) handleXMV12HUDInput(tank *battleTank) {
	if tank == nil || !primaryPointerJustPressed() {
		return
	}
	x, y := primaryPointerPosition()
	cursor := image.Pt(x, y)
	switch {
	case cursor.In(s.xmV12LeftButtonRect()):
		s.requestXMV12Direction(tank, -1)
	case cursor.In(s.xmV12StopButtonRect()):
		s.requestXMV12Direction(tank, 0)
	case cursor.In(s.xmV12RightButtonRect()):
		s.requestXMV12Direction(tank, 1)
	case cursor.In(s.xmV12MotorOffRect()):
		s.requestXMV12MotorOff(tank)
	}
}

func (s *GameScene) endXMV12Turn() {
	s.clearXMV12IdleVibration(s.activeTank())
	s.stopXMV12LoopSounds()
	s.xmV12DriveMode = false
	s.xmV12DriveDirection = 0
	s.xmV12EngineOffDelay = 0
	s.turnAdvanceDelay = 1
}

func (s *GameScene) destroyTankOutOfBounds(tank *battleTank) {
	if tank == nil || tank.power <= 0 {
		return
	}
	s.clearXMV12IdleVibration(tank)
	s.stopXMV12LoopSounds()
	tank.power = 0
	tank.shotStrength = 0
	tank.zeroPowerShown = true
	tank.zeroPowerGone = true
	s.awardZeroPowerScore(tank, tank, damageCauseDirect)
	s.removeTankSprites(tank)
	s.xmV12DriveMode = false
	s.xmV12DriveDirection = 0
	s.xmV12EngineOffDelay = 0
	s.turnAdvanceDelay = secondsToFrames(xmV12OutOfBoundsDelay)
}

func (s *GameScene) computerPlayerState(active *battleTank) computerplayers.State {
	state := computerplayers.State{
		ActiveIndex:          active.playerIndex,
		MaxStrength:          s.maxShotStrength(),
		Wind:                 s.wind,
		WindDirection:        s.windDirection,
		AvailableWeaponSlots: s.availableComputerWeaponSlots(active),
		Memory:               s.computerMemory(active.playerIndex),
		Tanks:                make([]computerplayers.TankState, 0, len(s.tanks)),
	}
	for _, tank := range s.tanks {
		if tank == nil || tank.body == nil {
			continue
		}
		center := tank.body.Bounds().Center()
		state.Tanks = append(state.Tanks, computerplayers.TankState{
			Index:  tank.playerIndex,
			X:      center.X,
			Y:      center.Y,
			Width:  tank.body.Bounds().W(),
			Height: tank.body.Bounds().H(),
			Power:  tank.power,
			Alive:  tank.power > 0 && tank.landed && !tank.falling,
		})
	}
	for _, palm := range s.palms {
		if palm == nil || palm.sprite == nil {
			continue
		}
		bounds := palm.sprite.Bounds()
		state.Obstacles = append(state.Obstacles, computerplayers.ObstacleState{
			X:      bounds.Min.X,
			Y:      bounds.Min.Y,
			Width:  bounds.W(),
			Height: bounds.H(),
		})
	}
	const groundSampleStep = 16
	for x := 0.0; x <= s.worldWidth; x += groundSampleStep {
		state.Ground = append(state.Ground, computerplayers.GroundSample{
			X: x,
			Y: s.ground.SurfaceY(x),
		})
	}
	if len(state.Ground) == 0 || state.Ground[len(state.Ground)-1].X < s.worldWidth {
		state.Ground = append(state.Ground, computerplayers.GroundSample{
			X: s.worldWidth,
			Y: s.ground.SurfaceY(s.worldWidth),
		})
	}
	return state
}

func (s *GameScene) assignEffectiveComputerIDs() {
	s.effectiveComputerIDs = make([]computerplayers.ID, len(s.players))
	for i, player := range s.players {
		id := player.ComputerID
		if player.Kind == PlayerComputer && id == computerplayers.MisterXID {
			id = computerplayers.RandomMisterXID(s.rng)
		}
		s.effectiveComputerIDs[i] = id
	}
}

func (s *GameScene) effectiveComputerID(tank *battleTank) computerplayers.ID {
	if tank == nil || tank.playerIndex < 0 || tank.playerIndex >= len(s.effectiveComputerIDs) {
		if tank != nil {
			return tank.player.ComputerID
		}
		return computerplayers.DoedelID
	}
	return s.effectiveComputerIDs[tank.playerIndex]
}

func (s *GameScene) availableComputerWeaponSlots(tank *battleTank) []int {
	const maxComputerProjectileSlot = 18
	slots := make([]int, 0, maxComputerProjectileSlot+1)
	for slot := 0; slot <= maxComputerProjectileSlot && slot < s.weaponSlotCount(); slot++ {
		if s.canSelectWeaponSlot(tank, slot) {
			slots = append(slots, slot)
		}
	}
	if len(slots) == 0 {
		return []int{0}
	}
	return slots
}

func (s *GameScene) adjustComputerCannonAngle(tank *battleTank, targetAngle float64) bool {
	if tank.cannon != nil && tank.body != nil {
		current := s.cannonAngleDegreesForTank(tank)
		delta := targetAngle - current
		step := math.Max(1, float64(tank.cannon.RotationSpeed)*computerAdjustSpeedFactor)
		if math.Abs(delta) <= step {
			leftLimit := tank.body.Rot - math.Pi
			tank.cannon.Rot = leftLimit + engine.DegToRad(targetAngle)
			s.clampCannonRotationToTank(tank.cannon, tank.body)
			s.syncOnlineAim(tank)
			return false
		}
		if delta < 0 {
			step = -step
		}
		tank.cannon.Rot += engine.DegToRad(step)
		s.clampCannonRotationToTank(tank.cannon, tank.body)
		s.syncOnlineAim(tank)
		return true
	}
	return false
}

func (s *GameScene) clampedComputerStrength(strength int) int {
	strength = minInt(strength, s.maxShotStrength())
	strength = maxInt(s.minShotStrength(), strength)
	return strength
}

func (s *GameScene) cannonAngleDegreesForTank(tank *battleTank) float64 {
	if tank == nil || tank.cannon == nil || tank.body == nil {
		return 0
	}
	leftLimit := tank.body.Rot - math.Pi
	displayAngle := engine.RadToDeg(tank.cannon.Rot - leftLimit)
	return math.Max(0, math.Min(180, displayAngle))
}

func (s *GameScene) cannonDisplayAngleDegreesForTank(tank *battleTank) float64 {
	rawAngle := s.cannonAngleDegreesForTank(tank)
	displayAngle := 90 - math.Abs(rawAngle-90)
	return math.Max(0, math.Min(90, displayAngle))
}

func (s *GameScene) computerMemory(playerIndex int) computerplayers.Memory {
	if playerIndex < 0 || playerIndex >= len(s.computerMemories) {
		return computerplayers.Memory{}
	}
	return s.computerMemories[playerIndex]
}

func (s *GameScene) computerShotRecordFor(tank *battleTank) computerShotRecord {
	if tank == nil || tank.computerPlan == nil {
		return computerShotRecord{}
	}
	record := computerShotRecord{
		active:      true,
		playerIndex: tank.playerIndex,
		computerID:  s.effectiveComputerID(tank),
		targetIndex: tank.computerPlan.decision.TargetIndex,
	}
	if tank.body != nil {
		record.activeX = tank.body.Bounds().Center().X
	}
	if target := s.tankByPlayerIndex(record.targetIndex); target != nil && target.body != nil {
		record.targetX = target.body.Bounds().Center().X
	} else if tank.body != nil {
		record.targetX = tank.body.Bounds().Center().X
	}
	return record
}

func (s *GameScene) reportComputerShot(impact engine.Vec, hitPlayerIndex int, directHit bool) {
	record := s.lastComputerShot
	if !record.active || record.playerIndex < 0 || record.playerIndex >= len(s.computerMemories) {
		s.lastComputerShot = computerShotRecord{}
		return
	}
	computerplayers.Learn(record.computerID, &s.computerMemories[record.playerIndex], computerplayers.Lesson{
		ActiveX:     record.activeX,
		TargetIndex: record.targetIndex,
		TargetX:     record.targetX,
		ImpactX:     impact.X,
		Hit:         directHit && hitPlayerIndex == record.targetIndex,
	})
	s.lastComputerShot = computerShotRecord{}
}

func (s *GameScene) tankByPlayerIndex(playerIndex int) *battleTank {
	for _, tank := range s.tanks {
		if tank != nil && tank.playerIndex == playerIndex {
			return tank
		}
	}
	return nil
}

func shouldAdjustStrength(key ebiten.Key) bool {
	if inpututil.IsKeyJustPressed(key) {
		return true
	}
	held := inpututil.KeyPressDuration(key)
	return held > 18 && held%4 == 0
}

func shouldAdjustCannon(key ebiten.Key) bool {
	if inpututil.IsKeyJustPressed(key) {
		return true
	}
	held := inpututil.KeyPressDuration(key)
	return held > humanCannonRepeatStart && held%humanCannonRepeatFrames == 0
}

func (s *GameScene) canSelectWeaponSlot(tank *battleTank, slot int) bool {
	if tank == nil || slot < 0 {
		return false
	}
	if slot == 0 {
		return true
	}
	return s.ammoForWeaponSlot(tank.playerIndex, slot) > 0
}

func (s *GameScene) selectNextWeapon(tank *battleTank) {
	if tank == nil {
		return
	}
	for step := 1; step <= s.weaponSlotCount(); step++ {
		slot := (tank.selectedWeapon + step) % s.weaponSlotCount()
		if s.canSelectWeaponSlot(tank, slot) {
			s.setSelectedWeapon(tank, slot)
			return
		}
	}
}

func (s *GameScene) selectPreviousWeapon(tank *battleTank) {
	if tank == nil {
		return
	}
	count := s.weaponSlotCount()
	for step := 1; step <= count; step++ {
		slot := (tank.selectedWeapon - step + count) % count
		if s.canSelectWeaponSlot(tank, slot) {
			s.setSelectedWeapon(tank, slot)
			return
		}
	}
}

func (s *GameScene) setSelectedWeapon(tank *battleTank, slot int) {
	if tank == nil || tank.selectedWeapon == slot {
		return
	}
	tank.selectedWeapon = slot
	s.playEventSound(soundEventWeaponSelect)
	s.syncOnlineAim(tank)
}

func (s *GameScene) invertCannonAngle(tank *battleTank) {
	if tank == nil || tank.cannon == nil || tank.body == nil {
		return
	}
	current := s.cannonAngleDegreesForTank(tank)
	leftLimit := tank.body.Rot - math.Pi
	tank.cannon.Rot = leftLimit + engine.DegToRad(180-current)
	s.clampCannonRotationToTank(tank.cannon, tank.body)
	s.syncOnlineAim(tank)
}

func (s *GameScene) toggleScrollOMat(tank *battleTank) {
	if tank == nil {
		return
	}
	scrollSlot := s.scrollOMatItemIndex() + 1
	if scrollSlot <= 0 || scrollSlot >= s.weaponSlotCount() || !s.canSelectWeaponSlot(tank, scrollSlot) {
		return
	}
	if tank.selectedWeapon == scrollSlot {
		tank.selectedWeapon = 0
		s.cameraGoal = s.cameraTargetForTank(s.activePlayerIndex)
		s.cameraGoalY = 0
		s.syncOnlineAim(tank)
		return
	}
	tank.selectedWeapon = scrollSlot
	s.syncOnlineAim(tank)
}

func (s *GameScene) consumeSelectedWeaponAmmo(tank *battleTank) (bool, bool) {
	if tank == nil {
		return false, false
	}
	slot := tank.selectedWeapon
	if slot == 0 {
		return true, false
	}
	itemIndex := s.itemIndexForWeaponSlot(slot)
	if itemIndex < 0 || s.shopItemCountForPlayer(tank.playerIndex, itemIndex) <= 0 {
		return false, false
	}
	if s.isScrollOMatItem(itemIndex) {
		return true, false
	}
	s.ensureInventory(tank.playerIndex)
	if itemIndex < len(s.inventories[tank.playerIndex].classA) && s.inventories[tank.playerIndex].classA[itemIndex] > 0 {
		s.inventories[tank.playerIndex].classA[itemIndex]--
		return true, false
	}
	if itemIndex < len(s.inventories[tank.playerIndex].classB) && s.inventories[tank.playerIndex].classB[itemIndex] > 0 {
		s.inventories[tank.playerIndex].classB[itemIndex]--
		return true, s.rng.Intn(2) == 0
	}
	return false, false
}

func (s *GameScene) ammoForWeaponSlot(playerIndex, slot int) int {
	if slot == 0 {
		return 1
	}
	return s.shopItemCountForPlayer(playerIndex, s.itemIndexForWeaponSlot(slot))
}

func (s *GameScene) itemIndexForWeaponSlot(slot int) int {
	if slot <= 0 {
		return -1
	}
	return slot - 1
}

func (s *GameScene) fireActiveWeapon() {
	tank := s.activeTank()
	if tank == nil || tank.cannon == nil {
		return
	}
	if s.scrollOMatActive() {
		return
	}
	consumed, classBDud := s.consumeSelectedWeaponAmmo(tank)
	if !consumed {
		return
	}

	muzzle := s.cannonMuzzle(tank.cannon)
	speed := 1.4 + float64(tank.shotStrength)*0.32
	weapon := s.weaponForSlot(tank.selectedWeapon)
	hasEffectiveWeapon := false
	if weapon.SurpriseEgg {
		weapon = s.randomSurpriseEggWeapon()
		hasEffectiveWeapon = true
	}
	s.lastDamageSource = tank
	if tank.player.Kind == PlayerComputer {
		s.lastComputerShot = s.computerShotRecordFor(tank)
	}
	tank.computerPlan = nil
	s.playWeaponFireSound(weapon)
	if weapon.Laser {
		s.fireLaserWeapon(tank, *muzzle, tank.cannon.Rot, weapon, classBDud)
		return
	}
	angles := []float64{tank.cannon.Rot}
	if weapon.TripleShot {
		if !hasEffectiveWeapon {
			weapon = s.mfsEffectiveWeapon(tank.playerIndex, weapon)
			hasEffectiveWeapon = true
		}
		offset := 5 * math.Pi / 180
		angles = []float64{tank.cannon.Rot - offset, tank.cannon.Rot, tank.cannon.Rot + offset}
	}
	projectiles := make([]*projectile, 0, len(angles))
	for _, angle := range angles {
		velocity := engine.V(speed, 0).Rotated(angle)
		projectiles = append(projectiles, &projectile{
			pos:                *muzzle,
			prev:               *muzzle,
			velocity:           velocity,
			weaponIndex:        tank.selectedWeapon,
			effectiveWeapon:    weapon,
			hasEffectiveWeapon: hasEffectiveWeapon,
			classBDud:          classBDud,
			shooter:            tank,
			launchRot:          angle,
			trail:              []engine.Vec{*muzzle},
			splitterArmed:      weapon.SplitterBomb && velocity.Y < -0.05,
		})
	}
	s.setProjectiles(projectiles)
}

func (s *GameScene) requestFireActiveWeapon() {
	if s.g.online == nil {
		s.fireActiveWeapon()
		return
	}
	tank := s.activeTank()
	if tank == nil || tank.cannon == nil || !s.onlineCanControlPlayer(tank.playerIndex) {
		return
	}
	s.sendOnlineGameCommand(protocol.OnlineGameCommand{
		Kind:           "fire",
		PlayerIndex:    tank.playerIndex,
		TurnSequence:   s.g.online.turnSequence,
		WeaponSlot:     tank.selectedWeapon,
		ShotStrength:   tank.shotStrength,
		CannonRotation: tank.cannon.Rot,
		CameraX:        s.cameraX,
	})
}

func (s *GameScene) mfsEffectiveWeapon(playerIndex int, fallback weaponspkg.Weapon) weaponspkg.Weapon {
	state, boosted := s.consumeMFSBoosterState(playerIndex)
	if !boosted {
		return fallback
	}
	switch state {
	case 0:
		return weaponspkg.LargeGrenade()
	case 1:
		return weaponspkg.AtomBomb()
	case 2:
		return weaponspkg.HBomb()
	default:
		return weaponspkg.Grenade()
	}
}

func (s *GameScene) cannonMuzzle(cannon *engine.Sprite) *engine.Vec {
	if cannon == nil {
		return &engine.Vec{}
	}
	bounds := cannon.Bounds()
	if cannon.RotAnchor == nil {
		return bounds.Center().Add(engine.V(bounds.W()/2+7, 0).Rotated(cannon.Rot))
	}
	scaleX := 1.0
	scaleY := 1.0
	if cannon.Drawable != nil {
		source := cannon.Drawable.Bounds()
		if source.W() != 0 {
			scaleX = cannon.Size.X / source.W()
		}
		if source.H() != 0 {
			scaleY = cannon.Size.Y / source.H()
		}
	}
	anchorWorld := engine.V(
		cannon.Pos.X+cannon.RotAnchor.X*scaleX,
		cannon.Pos.Y+cannon.RotAnchor.Y*scaleY,
	)
	length := math.Max(1, cannon.Size.X-cannon.RotAnchor.X*scaleX)
	return anchorWorld.Add(engine.V(length+2, 0).Rotated(cannon.Rot))
}

func (s *GameScene) weaponForSlot(slot int) weaponspkg.Weapon {
	return s.weaponForProjectile(&projectile{weaponIndex: slot})
}

func (s *GameScene) updateProjectile() {
	if !s.projectilesActive() {
		return
	}
	if s.updateProjectileReentryAnimation() {
		return
	}

	const gravity = 0.16
	windAcceleration := float64(s.windDirection*s.wind) * projectileWindFactor

	active := s.projectiles[:0]
	spawned := make([]*projectile, 0)
	for _, p := range s.projectiles {
		if p == nil {
			continue
		}
		alive, children := s.updateSingleProjectile(p, gravity, windAcceleration)
		if len(children) > 0 {
			spawned = append(spawned, children...)
		}
		if alive {
			active = append(active, p)
		}
	}
	active = append(active, spawned...)
	s.setProjectiles(active)
	if !s.projectilesActive() && s.turnAdvanceDelay <= 0 && !s.turnAdvanceBlocked() {
		s.finishProjectiles()
	}
}

func (s *GameScene) updateSingleProjectile(p *projectile, gravity, windAcceleration float64) (bool, []*projectile) {
	p.prev = p.pos
	previousVelocityY := p.velocity.Y
	p.velocity.X += windAcceleration
	p.velocity.Y += gravity
	if s.shouldSplitProjectile(p, previousVelocityY) {
		s.playEventSound(soundEventSplitterBombSplit)
		return false, s.splitProjectile(p, gravity, windAcceleration)
	}
	p.pos = *p.pos.Add(p.velocity)
	p.trail = append(p.trail, p.pos)
	if !p.zeroPowerScatter && len(p.trail) > 260 {
		p.trail = p.trail[len(p.trail)-260:]
	}
	if alive, paused := s.handleProjectileWorldEdge(p); !alive {
		s.reportComputerShot(p.pos, -1, false)
		return false, nil
	} else if paused {
		return true, nil
	}

	if p.pos.X >= 0 && p.pos.X <= s.worldWidth {
		s.cameraGoal = s.cameraTargetForWorldX(p.pos.X)
		s.cameraGoalY = s.cameraTargetForWorldY(p.pos.Y)
		s.cameraX = approach(s.cameraX, s.cameraGoal, 0.12, 0.4)
		s.cameraY = approach(s.cameraY, s.cameraGoalY, 0.12, 0.4)
	}
	if s.addLightningCloudAggressionForProjectile(p) {
		s.reportComputerShot(p.pos, -1, false)
		s.scheduleCloudSearchForProjectile(p)
		return false, nil
	}

	battlefieldHeight := s.battlefieldHeight()
	if p.pos.Y > battlefieldHeight+80 {
		s.reportComputerShot(p.pos, -1, false)
		return false, nil
	}

	weapon := s.weaponForProjectile(p)
	if !p.classBDud && weapon.Mosquitos && s.projectileGroundImpactWithinFrames(p, gravity, windAcceleration, mosquitoPreviewFrames) {
		p.mosquitoPreview = true
	}
	if hit, ok := s.projectileHitsWaterSurface(p, projectileRadiusForWeapon(weapon)); ok {
		if p.classBDud {
			s.reportComputerShot(hit, -1, false)
			s.startDudImpact(hit)
			s.scheduleCloudSearchForProjectile(p)
			return false, nil
		}
		s.reportComputerShot(hit, -1, false)
		s.startWaterSurfaceImpact(hit)
		s.scheduleCloudSearchForProjectile(p)
		return false, nil
	}
	if palm := s.projectileHitsPalm(p, projectileRadiusForWeapon(weapon)); palm != nil {
		if p.classBDud {
			s.reportComputerShot(p.pos, -1, false)
			s.handleDudPalmHit(palm)
			s.startDudImpact(p.pos)
			s.scheduleCloudSearchForProjectile(p)
			return false, nil
		}
		s.reportComputerShot(p.pos, -1, false)
		s.spawnPalmLeafFall(palm)
		if weapon.ImpactAnimationStyle == weaponspkg.ImpactAnimationFireball && palm.state == palmStateAlive {
			s.ignitePalm(palm)
		} else if palm.state == palmStateSkeleton || palm.state == palmStateSkeletonSmoking {
			s.crumblePalm(palm)
		} else {
			if !s.addPalmAggression(palm, s.lastDamageSource) {
				s.startPalmEyes(palm)
			}
		}
		if s.palmRevenge == nil {
			s.focusPalmCamera(palm)
		}
		s.delayTurnAdvance(s.palmEffectDelayFrames(palm))
		s.scheduleCloudSearchForProjectile(p)
		return false, nil
	}

	if (p.pos.X < 0 || p.pos.X > s.worldWidth) && p.pos.Y >= s.ground.SurfaceY(p.pos.X) {
		s.reportComputerShot(p.pos, -1, false)
		return false, nil
	}

	if p.pos.Y >= s.ground.SurfaceY(p.pos.X) {
		if p.zeroPowerScatter {
			s.onZeroPowerScatterGroundImpact(p)
			s.scheduleCloudSearchForProjectile(p)
			return false, nil
		}
		if s.onGroundImpact(p) {
			s.scheduleCloudSearchForProjectile(p)
			return false, nil
		}
		s.reportComputerShot(p.pos, -1, false)
		return false, nil
	}

	hitRadius := projectileRadiusForWeapon(weapon)
	hitBounds := engine.R(p.pos.X-hitRadius, p.pos.Y-hitRadius, p.pos.X+hitRadius, p.pos.Y+hitRadius)
	for _, tank := range s.tanks {
		if tank == nil || tank.body == nil {
			continue
		}
		if engine.Collision(hitBounds, tank.body.Bounds().ScaledAtCenter(0.78)) {
			if p.classBDud {
				s.reportComputerShot(p.pos, tank.playerIndex, false)
				s.startDudImpact(p.pos)
				s.scheduleCloudSearchForProjectile(p)
				return false, nil
			}
			if p.zeroPowerScatter {
				s.onZeroPowerScatterGroundImpact(p)
				s.scheduleCloudSearchForProjectile(p)
				return false, nil
			}
			if damage := weapon.Damage; damage > 0 && !weapon.PlantsPalm {
				s.damageTank(tank, damage, s.lastDamageSource, damageCauseDirect)
				s.darkenTank(tank, 0.10)
			} else if weaponHasImpactEffect(weapon) {
				s.onGroundImpact(p)
			}
			s.reportComputerShot(p.pos, tank.playerIndex, true)
			s.delayTurnAdvance(s.tankHitPauseFrames())
			s.scheduleCloudSearchForProjectile(p)
			return false, nil
		}
	}
	return true, nil
}

func weaponHasImpactEffect(weapon weaponspkg.Weapon) bool {
	return weapon.DamagesTerrain ||
		weapon.PlantsPalm ||
		weapon.FillsWater ||
		weapon.Moles ||
		weapon.SmallCrumblers ||
		weapon.LargeCrumblers ||
		weapon.Mosquitos ||
		weapon.Shockwave ||
		weapon.AirStrike ||
		weapon.ImpactAnimationStyle == weaponspkg.ImpactAnimationFireball
}

func (s *GameScene) shouldSplitProjectile(p *projectile, previousVelocityY float64) bool {
	if p == nil || p.classBDud || !p.splitterArmed {
		return false
	}
	weapon := s.weaponForProjectile(p)
	return weapon.SplitterBomb && previousVelocityY < 0 && p.velocity.Y >= 0
}

func (s *GameScene) splitProjectile(p *projectile, gravity, windAcceleration float64) []*projectile {
	if p == nil || splitterBombFragmentCount <= 0 {
		return nil
	}
	fragmentWeapon := weaponspkg.SplitterBombFragment()
	fragments := make([]*projectile, 0, splitterBombFragmentCount)
	fallFrames := s.estimateSplitterFragmentFallFrames(p.pos, 0.45, gravity)
	center := float64(splitterBombFragmentCount-1) / 2
	for i := 0; i < splitterBombFragmentCount; i++ {
		offset := 0.0
		if splitterBombFragmentCount > 1 {
			offset = (float64(i) - center) / center * (splitterBombSpreadWidth / 2)
		}
		offset += (s.rng.Float64()*2 - 1) * 8
		frames := math.Max(1, float64(fallFrames))
		windDrift := windAcceleration * frames * (frames + 1) / 2
		vx := (offset - windDrift) / frames
		vy := 0.35 + s.rng.Float64()*0.25
		fragments = append(fragments, &projectile{
			pos:                p.pos,
			prev:               p.pos,
			velocity:           engine.V(vx, vy),
			weaponIndex:        p.weaponIndex,
			effectiveWeapon:    fragmentWeapon,
			hasEffectiveWeapon: true,
			shooter:            p.shooter,
			launchRot:          math.Atan2(vy, vx),
			trail:              []engine.Vec{p.pos},
		})
	}
	return fragments
}

func (s *GameScene) estimateSplitterFragmentFallFrames(pos engine.Vec, initialYVelocity, gravity float64) int {
	groundY := s.ground.SurfaceY(pos.X)
	distance := math.Max(40, groundY-pos.Y)
	if gravity <= 0 {
		return 60
	}
	discriminant := initialYVelocity*initialYVelocity + 2*gravity*distance
	frames := (-initialYVelocity + math.Sqrt(math.Max(0, discriminant))) / gravity
	return max(30, minInt(120, int(math.Round(frames))))
}

type laserHitType uint8

const (
	laserHitNone laserHitType = iota
	laserHitTerrain
	laserHitPalm
	laserHitTank
)

type laserHit struct {
	kind laserHitType
	pos  engine.Vec
	dist float64
	palm *battlePalm
	tank *battleTank
}

func (s *GameScene) fireLaserWeapon(shooter *battleTank, muzzle engine.Vec, angle float64, weapon weaponspkg.Weapon, classBDud bool) {
	s.playEventSoundLoop(laserLoopKey, soundEventLaser)
	dir := engine.V(1, 0).Rotated(angle)
	dir = normalizedVec(dir)
	maxDistance := s.laserMaxDistanceToWorld(muzzle, dir)
	if maxDistance <= 0 {
		s.delayTurnAdvance(laserHoldFrames)
		return
	}

	hit := s.traceLaser(muzzle, dir, maxDistance)
	end := muzzle.Add(dir.Scaled(maxDistance))
	effect := &laserEffect{
		start:       muzzle,
		end:         *end,
		tip:         *end,
		dir:         dir,
		maxDistance: maxDistance,
		duration:    laserHoldFrames,
		finished:    true,
	}

	delay := laserHoldFrames
	if classBDud {
		effect.end = hit.pos
		effect.tip = hit.pos
		if hit.kind == laserHitNone {
			effect.end = *end
			effect.tip = *end
		}
		if hit.kind == laserHitPalm {
			s.handleDudPalmHit(hit.palm)
			s.startDudImpact(effect.tip)
		} else {
			s.startDudImpact(effect.tip)
		}
		s.reportComputerShot(effect.tip, -1, false)
		s.laserEffects = append(s.laserEffects, effect)
		s.delayTurnAdvance(maxInt(delay, s.impactPauseFrames()))
		return
	}
	switch hit.kind {
	case laserHitTank:
		effect.end = hit.pos
		effect.tip = hit.pos
		if hit.tank != nil {
			s.damageTank(hit.tank, weapon.Damage, shooter, damageCauseDirect)
			s.darkenTank(hit.tank, 0.10)
			s.reportComputerShot(hit.pos, hit.tank.playerIndex, true)
			delay = maxInt(delay, s.tankHitPauseFrames())
		} else {
			s.reportComputerShot(hit.pos, -1, false)
		}
	case laserHitPalm:
		effect.end = hit.pos
		effect.tip = hit.pos
		s.handleLaserPalmHit(hit.palm)
		s.reportComputerShot(hit.pos, -1, false)
		delay = maxInt(delay, s.palmEffectDelayFrames(hit.palm))
	case laserHitTerrain:
		effect.end = hit.pos
		effect.tip = hit.pos
		effect.drilling = true
		effect.finished = false
		effect.smokeFrom = hit.pos
		effect.duration = int(math.Ceil(math.Max(0, maxDistance-hit.dist)/laserDrillSpeed)) + laserHoldFrames
		s.playEventSound(soundEventLaserSmoke)
		effect.smokes = append(effect.smokes, laserSmoke{pos: hit.pos})
		s.reportComputerShot(hit.pos, -1, false)
		delay = maxInt(delay, effect.duration)
	default:
		effect.end = muzzle
		effect.tip = muzzle
		effect.traveling = true
		effect.finished = false
		effect.duration = int(math.Ceil(maxDistance/laserDrillSpeed)) + laserHoldFrames
		s.reportComputerShot(*end, -1, false)
		delay = maxInt(delay, effect.duration)
	}

	s.laserEffects = append(s.laserEffects, effect)
	s.delayTurnAdvance(delay)
}

func (s *GameScene) handleLaserPalmHit(palm *battlePalm) {
	if palm == nil {
		return
	}
	s.spawnPalmLeafFall(palm)
	if palm.state == palmStateAlive {
		s.ignitePalm(palm)
	} else if palm.state == palmStateSkeleton || palm.state == palmStateSkeletonSmoking {
		s.crumblePalm(palm)
	}
	if s.palmRevenge == nil {
		s.focusPalmCamera(palm)
	}
}

func (s *GameScene) handleDudPalmHit(palm *battlePalm) {
	if palm == nil {
		return
	}
	s.spawnPalmLeafFall(palm)
	if palm.state == palmStateSkeleton || palm.state == palmStateSkeletonSmoking {
		s.crumblePalm(palm)
	} else if !s.addPalmAggression(palm, s.lastDamageSource) {
		s.startPalmEyes(palm)
	}
	if s.palmRevenge == nil {
		s.focusPalmCamera(palm)
	}
	s.delayTurnAdvance(s.palmEffectDelayFrames(palm))
}

func (s *GameScene) updateLaserEffects() {
	if len(s.laserEffects) == 0 {
		return
	}
	active := s.laserEffects[:0]
	for _, effect := range s.laserEffects {
		if effect == nil {
			continue
		}
		s.updateLaserEffect(effect)
		effect.age++
		if effect.age < effect.duration {
			active = append(active, effect)
		}
	}
	s.laserEffects = active
}

func (s *GameScene) updateLaserEffect(effect *laserEffect) {
	if effect == nil {
		return
	}
	activeSmokes := effect.smokes[:0]
	for _, smoke := range effect.smokes {
		smoke.age++
		if smoke.age < s.laserSmokeAnimation.totalTicks {
			activeSmokes = append(activeSmokes, smoke)
		}
	}
	effect.smokes = activeSmokes
	if (!effect.drilling && !effect.traveling) || effect.finished {
		return
	}
	previous := effect.tip
	next := effect.tip.Add(effect.dir.Scaled(laserDrillSpeed))
	effect.tip = *next
	if effect.maxDistance > 0 && effect.tip.Sub(effect.start).Len() >= effect.maxDistance {
		effect.tip = *effect.start.Add(effect.dir.Scaled(effect.maxDistance))
		effect.finished = true
		effect.duration = minInt(effect.duration, effect.age+laserHoldFrames)
	}
	if hit := s.laserSegmentHitsPalm(previous, effect.tip); hit.palm != nil {
		effect.tip = hit.pos
		effect.end = hit.pos
		effect.finished = true
		effect.drilling = false
		effect.traveling = false
		effect.duration = minInt(effect.duration, effect.age+maxInt(laserHoldFrames, s.palmEffectDelayFrames(hit.palm)))
		s.handleLaserPalmHit(hit.palm)
		s.reportComputerShot(hit.pos, -1, false)
		if s.turnAdvanceDelay > 0 {
			s.turnAdvanceDelay = minInt(s.turnAdvanceDelay, maxInt(laserHoldFrames, s.palmEffectDelayFrames(hit.palm)))
		}
		s.finishLaserDrillingIfNeeded(effect)
		return
	}
	effect.end = effect.tip
	tipInTerrain := false
	if effect.drilling {
		tipInTerrain = s.ground.ColorAt(effect.tip.X, effect.tip.Y).A > 0
		if s.laserSegmentTouchesTerrain(previous, effect.tip) {
			area := s.ground.ClearLine(previous.X, previous.Y, effect.tip.X, effect.tip.Y, laserLineThickness)
			s.accumulateLaserEditArea(effect, area)
			s.refillWaterBelowArea(area)
		}
	}
	if tipInTerrain && len(effect.smokes) < laserMaxSmokeCount && effect.tip.Sub(effect.smokeFrom).Len() >= laserSmokeSpacing {
		s.playEventSound(soundEventLaserSmoke)
		effect.smokes = append(effect.smokes, laserSmoke{pos: effect.tip})
		effect.smokeFrom = effect.tip
	}
	s.finishLaserDrillingIfNeeded(effect)
}

func (s *GameScene) laserSegmentHitsPalm(start, end engine.Vec) laserHit {
	delta := end.Sub(start)
	steps := maxInt(1, int(math.Ceil(delta.Len())))
	for i := 0; i <= steps; i++ {
		t := float64(i) / float64(steps)
		pos := engine.V(start.X+delta.X*t, start.Y+delta.Y*t)
		if palm := s.laserPointHitsPalm(pos); palm != nil {
			return laserHit{kind: laserHitPalm, pos: pos, palm: palm}
		}
	}
	return laserHit{}
}

func (s *GameScene) laserSegmentTouchesTerrain(start, end engine.Vec) bool {
	delta := end.Sub(start)
	steps := maxInt(1, int(math.Ceil(delta.Len())))
	for i := 0; i <= steps; i++ {
		t := float64(i) / float64(steps)
		pos := engine.V(start.X+delta.X*t, start.Y+delta.Y*t)
		if s.ground.ColorAt(pos.X, pos.Y).A > 0 {
			return true
		}
	}
	return false
}

func (s *GameScene) accumulateLaserEditArea(effect *laserEffect, area image.Rectangle) {
	if effect == nil || area.Empty() {
		return
	}
	area.Max.Y = maxInt(area.Max.Y, int(math.Ceil(s.battlefieldHeight())))
	if effect.editArea.Empty() {
		effect.editArea = area
		return
	}
	effect.editArea = effect.editArea.Union(area)
}

func (s *GameScene) finishLaserDrillingIfNeeded(effect *laserEffect) {
	if effect == nil || !effect.finished || effect.settled || effect.editArea.Empty() {
		return
	}
	effect.settled = true
	if falls := s.ground.SettleArea(effect.editArea); len(falls) > 0 {
		s.refillWaterBelowArea(effect.editArea)
		s.sandFalls = append(s.sandFalls, sandFallAnimation{
			pixels:   falls,
			duration: sandFallFrames,
		})
	}
	s.dropUnsupportedTanks()
}

func (s *GameScene) traceLaser(start, dir engine.Vec, maxDistance float64) laserHit {
	step := 1.0
	for dist := 0.0; dist <= maxDistance; dist += step {
		pos := *start.Add(dir.Scaled(dist))
		if dist > 3 {
			if tank := s.laserPointHitsTank(pos); tank != nil {
				return laserHit{kind: laserHitTank, pos: pos, dist: dist, tank: tank}
			}
			if palm := s.laserPointHitsPalm(pos); palm != nil {
				return laserHit{kind: laserHitPalm, pos: pos, dist: dist, palm: palm}
			}
		}
		if s.ground.ColorAt(pos.X, pos.Y).A > 0 {
			return laserHit{kind: laserHitTerrain, pos: pos, dist: dist}
		}
	}
	end := start.Add(dir.Scaled(maxDistance))
	return laserHit{kind: laserHitNone, pos: *end, dist: maxDistance}
}

func (s *GameScene) laserPointHitsTank(pos engine.Vec) *battleTank {
	for _, tank := range s.tanks {
		if tank == nil || tank.body == nil || tank.power <= 0 || tank.zeroPowerGone {
			continue
		}
		if rectContainsPoint(tank.body.Bounds().ScaledAtCenter(0.78), pos) {
			return tank
		}
	}
	return nil
}

func (s *GameScene) laserPointHitsPalm(pos engine.Vec) *battlePalm {
	for _, palm := range s.palms {
		if palm == nil || palm.sprite == nil || palm.pixels == nil || palm.state == palmStateCrumbling {
			continue
		}
		bounds := palm.sprite.Bounds()
		if !rectContainsPoint(bounds, pos) {
			continue
		}
		sourceX := int(math.Floor(pos.X - bounds.Min.X))
		sourceY := int(math.Floor(pos.Y - bounds.Min.Y))
		if sourceX < 0 || sourceY < 0 || sourceX >= palm.pixels.Bounds().Dx() || sourceY >= palm.pixels.Bounds().Dy() {
			continue
		}
		if palm.pixels.RGBAAt(sourceX, sourceY).A > 0 {
			return palm
		}
	}
	return nil
}

func (s *GameScene) laserMaxDistanceToWorld(start, dir engine.Vec) float64 {
	const epsilon = 0.0001
	left := 0.0
	right := s.worldWidth
	top := -s.skyExtraHeight()
	bottom := s.battlefieldHeight()
	maxDistance := math.Inf(1)
	if dir.X > epsilon {
		maxDistance = math.Min(maxDistance, (right-start.X)/dir.X)
	} else if dir.X < -epsilon {
		maxDistance = math.Min(maxDistance, (left-start.X)/dir.X)
	}
	if dir.Y > epsilon {
		maxDistance = math.Min(maxDistance, (bottom-start.Y)/dir.Y)
	} else if dir.Y < -epsilon {
		maxDistance = math.Min(maxDistance, (top-start.Y)/dir.Y)
	}
	if math.IsInf(maxDistance, 1) || maxDistance < 0 {
		return 0
	}
	return maxDistance
}

func normalizedVec(v engine.Vec) engine.Vec {
	length := v.Len()
	if length <= 0 {
		return engine.V(1, 0)
	}
	return v.Scaled(1 / length)
}

func rectContainsPoint(rect engine.Rect, point engine.Vec) bool {
	return point.X >= rect.Min.X && point.X <= rect.Max.X && point.Y >= rect.Min.Y && point.Y <= rect.Max.Y
}

func (s *GameScene) projectileGroundImpactWithinFrames(p *projectile, gravity, windAcceleration float64, frames int) bool {
	if p == nil || frames <= 0 {
		return false
	}
	pos := p.pos
	velocity := p.velocity
	for i := 0; i < frames; i++ {
		velocity.X += windAcceleration
		velocity.Y += gravity
		pos = *pos.Add(velocity)
		if pos.X < 0 || pos.X > s.worldWidth {
			continue
		}
		if pos.Y >= s.ground.SurfaceY(pos.X) {
			return true
		}
	}
	return false
}

func (s *GameScene) handleProjectileWorldEdge(p *projectile) (bool, bool) {
	if p == nil || s.worldWidth <= 0 || (p.pos.X >= 0 && p.pos.X <= s.worldWidth) {
		return true, false
	}
	if !s.projectileReentry {
		return false, false
	}

	direction := 1
	if p.pos.X > s.worldWidth {
		direction = -1
	}
	for p.pos.X < 0 {
		p.pos.X += s.worldWidth
	}
	for p.pos.X > s.worldWidth {
		p.pos.X -= s.worldWidth
	}
	p.prev = p.pos
	p.trail = []engine.Vec{p.pos}
	s.playEventSound(soundEventProjectileReentryExit)
	s.reentryAnimation = &projectileReentryAnimation{
		projectile: p,
		shooter:    p.shooter,
		duration:   reentryAnimationFrames,
		direction:  direction,
		cannonRot:  p.launchRot,
	}
	return true, true
}

func (s *GameScene) updateProjectileReentryAnimation() bool {
	if s.reentryAnimation == nil {
		return false
	}
	s.reentryAnimation.age++
	if s.reentryAnimation.age >= s.reentryAnimation.duration {
		s.playEventSound(soundEventProjectileReentryEnter)
		s.reentryAnimation = nil
	}
	return true
}

func (s *GameScene) projectileHitsPalm(p *projectile, radius float64) *battlePalm {
	if p == nil || len(s.palms) == 0 {
		return nil
	}
	segmentLength := math.Hypot(p.pos.X-p.prev.X, p.pos.Y-p.prev.Y)
	stepLength := math.Max(1, radius)
	steps := maxInt(1, int(math.Ceil(segmentLength/stepLength)))
	for step := 0; step <= steps; step++ {
		t := float64(step) / float64(steps)
		center := engine.V(
			p.prev.X+(p.pos.X-p.prev.X)*t,
			p.prev.Y+(p.pos.Y-p.prev.Y)*t,
		)
		if palm := s.projectileCenterHitsPalm(center, radius); palm != nil {
			return palm
		}
	}
	return nil
}

func (s *GameScene) projectileCenterHitsPalm(center engine.Vec, radius float64) *battlePalm {
	minX := int(math.Floor(center.X - radius))
	maxX := int(math.Ceil(center.X + radius))
	minY := int(math.Floor(center.Y - radius))
	maxY := int(math.Ceil(center.Y + radius))
	r2 := radius * radius

	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			dx := float64(x) - center.X
			dy := float64(y) - center.Y
			if dx*dx+dy*dy > r2 {
				continue
			}
			if palm := s.visiblePalmPixelAt(float64(x), float64(y)); palm != nil {
				return palm
			}
		}
	}
	return nil
}

func (s *GameScene) visiblePalmPixelAt(worldX, worldY float64) *battlePalm {
	for _, palm := range s.palms {
		if palm == nil || palm.sprite == nil || palm.pixels == nil || palm.state == palmStateCrumbling {
			continue
		}
		bounds := palm.sprite.Bounds()
		if worldX < bounds.Min.X || worldX >= bounds.Max.X || worldY < bounds.Min.Y || worldY >= bounds.Max.Y {
			continue
		}
		sourceX := int((worldX - bounds.Min.X) / bounds.W() * float64(palm.pixels.Bounds().Dx()))
		sourceY := int((worldY - bounds.Min.Y) / bounds.H() * float64(palm.pixels.Bounds().Dy()))
		if sourceX < 0 || sourceX >= palm.pixels.Bounds().Dx() || sourceY < 0 || sourceY >= palm.pixels.Bounds().Dy() {
			continue
		}
		if palm.pixels.RGBAAt(sourceX, sourceY).A > 0 {
			return palm
		}
	}
	return nil
}

func (s *GameScene) projectilesActive() bool {
	return len(s.projectiles) > 0
}

func (s *GameScene) turnAdvanceBlocked() bool {
	return len(s.waterBlubbers) > 0 || len(s.smallCrumblerImpacts) > 0 || len(s.moskitoEffects) > 0 || len(s.shockwaveImpacts) > 0 || len(s.airStrikeImpacts) > 0 || len(s.laserEffects) > 0
}

func (s *GameScene) activePlayerCanAdjustShot() bool {
	return s.phase == phaseBattle &&
		s.allTanksLanded() &&
		!s.projectilesActive() &&
		!s.turnAdvanceBlocked() &&
		!s.xmV12DriveMode &&
		s.turnAdvanceDelay <= 0 &&
		s.roundTransitionDelay <= 0 &&
		!s.roundSeriesComplete &&
		s.palmRevenge == nil &&
		s.activePlayerIndex >= 0 &&
		s.activePlayerIndex < len(s.tanks)
}

func (s *GameScene) setProjectiles(projectiles []*projectile) {
	s.projectiles = projectiles
	s.projectile = nil
	if len(projectiles) > 0 {
		s.projectile = projectiles[0]
	}
}

func (s *GameScene) finishProjectiles() {
	s.projectile = nil
	s.projectiles = nil
	if len(s.tanks) == 0 {
		return
	}
	s.advanceActivePlayer()
}

func (s *GameScene) advanceActivePlayer() {
	if len(s.tanks) == 0 {
		return
	}
	previousPlayerIndex := s.activePlayerIndex
	s.clearXMV12IdleVibration(s.activeTank())
	if s.endRoundIfOnlyOneTankRemains() {
		return
	}
	s.lastDamageSource = nil
	s.palmCameraFocus = nil
	s.waterCameraFocus = nil
	s.waterBlotchFocus = nil
	s.crumblerCameraFocus = nil
	s.moskitoCameraFocus = nil
	s.cloudSearchEffects = nil
	s.stopXMV12LoopSounds()
	s.xmV12DriveMode = false
	s.xmV12DriveDirection = 0
	s.xmV12EngineOffDelay = 0
	next := s.nextActivePlayerIndex()
	if next < 0 {
		return
	}
	s.activePlayerIndex = next
	s.resetComputerTurnPlans()
	s.clampActiveShotStrength()
	s.cameraGoal = s.cameraTargetForTank(s.activePlayerIndex)
	s.cameraGoalY = 0
	s.syncOnlineTurn(previousPlayerIndex)
}

func (s *GameScene) nextActivePlayerIndex() int {
	if len(s.tanks) == 0 {
		return -1
	}
	start := s.activePlayerIndex
	for offset := 1; offset <= len(s.tanks); offset++ {
		index := (start + offset) % len(s.tanks)
		if s.tankCanAct(s.tanks[index]) {
			return index
		}
	}
	return -1
}

func (s *GameScene) delayTurnAdvance(frames int) {
	if frames <= 0 {
		if s.turnAdvanceDelay <= 0 {
			s.advanceActivePlayer()
		}
		return
	}
	if s.turnAdvanceDelay < frames {
		s.turnAdvanceDelay = frames
	}
}

func (s *GameScene) onGroundImpact(p *projectile) bool {
	if p == nil {
		return false
	}
	weapon := s.weaponForProjectile(p)
	if p.classBDud {
		s.reportComputerShot(p.pos, -1, false)
		s.startDudImpact(p.pos)
		return true
	}
	if !weaponHasImpactEffect(weapon) {
		s.reportComputerShot(p.pos, -1, false)
		return false
	}
	if weapon.AirStrike {
		s.reportComputerShot(p.pos, -1, false)
		s.startAirStrikeImpact(p.pos)
		return true
	}

	if weapon.Shockwave {
		s.reportComputerShot(p.pos, -1, false)
		s.startShockwaveImpact(p.pos, weapon)
		return true
	}

	if weapon.Mosquitos {
		s.reportComputerShot(p.pos, -1, false)
		s.startMoskitoImpact(p.pos, p.shooter)
		return true
	}

	if weapon.SmallCrumblers {
		s.reportComputerShot(p.pos, -1, false)
		s.startSmallCrumblerImpact(p.pos, smallCrumblerConfig())
		return true
	}

	if weapon.LargeCrumblers {
		s.reportComputerShot(p.pos, -1, false)
		s.startSmallCrumblerImpact(p.pos, largeCrumblerConfig())
		return true
	}

	if weapon.Moles {
		s.reportComputerShot(p.pos, -1, false)
		s.startMoleImpact(p.pos)
		return true
	}

	if weapon.FillsWater {
		s.reportComputerShot(p.pos, -1, false)
		s.startWaterFill(p.pos)
		return true
	}

	if weapon.ImpactAnimationStyle == weaponspkg.ImpactAnimationFireball {
		s.reportComputerShot(p.pos, -1, false)
		s.startFireballImpact(p.pos, weapon)
		return true
	}

	if weapon.PlantsPalm {
		s.reportComputerShot(p.pos, -1, false)
		s.plantPalmAtImpact(p.pos)
		s.damageTank(p.shooter, weapon.Damage, p.shooter, damageCauseDirect)
		s.delayTurnAdvance(s.palmHitPauseFrames())
		return true
	}

	radius := impactRadiusForWeapon(weapon)
	duration := s.impactAnimationFramesForWeapon(weapon)
	s.playWeaponImpactSound(weapon)
	s.zeroPowerStartDelay = (duration * 2) / 3
	defer func() {
		s.zeroPowerStartDelay = 0
	}()
	s.reportComputerShot(p.pos, -1, false)

	impactPos := p.pos
	terrainApplied := true
	damageApplied := map[int]bool(nil)
	if weapon.ImpactAnimationStyle == weaponspkg.ImpactAnimationPlasma {
		impactPos.Y -= plasmaImpactVisualYOffset
		terrainApplied = false
		damageApplied = make(map[int]bool)
	} else {
		s.damageTanksInImpactRadius(p.pos, weapon)
		falls := s.applyCraterAndRefillWater(p.pos.X, p.pos.Y, radius)
		if len(falls) > 0 {
			s.sandFalls = append(s.sandFalls, sandFallAnimation{
				pixels:   falls,
				duration: sandFallFrames,
			})
		}
		s.dropUnsupportedTanks()
	}
	s.impacts = append(s.impacts, impactAnimation{
		pos:            impactPos,
		radius:         radius,
		duration:       duration,
		cycles:         impactCyclesForWeapon(weapon),
		damage:         weapon.Damage,
		radialDamage:   weapon.RadialDamage,
		attacker:       s.lastDamageSource,
		damageApplied:  damageApplied,
		outward:        weapon.ImpactGradientOutward,
		style:          weapon.ImpactAnimationStyle,
		terrainApplied: terrainApplied,
	})
	if minimumDelay := duration + s.impactPauseFrames(); s.turnAdvanceDelay < minimumDelay {
		s.turnAdvanceDelay = minimumDelay
	}
	return true
}

func (s *GameScene) damageTanksInImpactRadius(center engine.Vec, weapon weaponspkg.Weapon) {
	s.damageTanksInRadialProfile(center, weapon.RadialDamage, s.lastDamageSource)
}

func (s *GameScene) damageTanksInRadialProfile(center engine.Vec, profile weaponspkg.RadialDamageProfile, attacker *battleTank) {
	if profile.MaxDamage <= 0 {
		return
	}
	for _, tank := range s.tanks {
		if tank == nil || tank.body == nil || tank.power <= 0 {
			continue
		}
		damage := radialDamageToTank(center, tank, profile)
		if damage <= 0 {
			continue
		}
		s.damageTank(tank, damage, attacker, damageCauseDirect)
	}
}

func radialDamageToTank(center engine.Vec, tank *battleTank, profile weaponspkg.RadialDamageProfile) int {
	if tank == nil || tank.body == nil {
		return 0
	}
	tankCenter := tank.body.Bounds().Center()
	dx := int(tankCenter.X) - int(center.X)
	dy := int(tankCenter.Y) - int(center.Y)
	distance := int(math.Sqrt(float64(dx*dx + dy*dy)))
	return weaponspkg.CalculateRadialDamage(distance, profile)
}

func (s *GameScene) startZeroPowerImpactAtTank(tank *battleTank, weapon weaponspkg.Weapon) {
	center := s.zeroPowerTankImpactCenter(tank)
	previous := s.lastDamageSource
	s.lastDamageSource = tank
	silentWeapon := weapon
	silentWeapon.ImpactSound = ""
	s.startTerrainImpact(center, silentWeapon, true)
	s.lastDamageSource = previous
}

func (s *GameScene) startTerrainImpact(pos engine.Vec, weapon weaponspkg.Weapon, damage bool) {
	s.playWeaponImpactSound(weapon)
	radius := impactRadiusForWeapon(weapon)
	duration := s.impactAnimationFramesForWeapon(weapon)
	impactPos := pos
	terrainApplied := true
	damageApplied := map[int]bool(nil)
	if weapon.ImpactAnimationStyle == weaponspkg.ImpactAnimationPlasma {
		impactPos.Y -= plasmaImpactVisualYOffset
		terrainApplied = false
		if damage {
			damageApplied = make(map[int]bool)
		}
	} else {
		if damage {
			s.damageTanksInImpactRadius(pos, weapon)
		}
		falls := s.applyCraterAndRefillWater(pos.X, pos.Y, radius)
		if len(falls) > 0 {
			s.sandFalls = append(s.sandFalls, sandFallAnimation{
				pixels:   falls,
				duration: sandFallFrames,
			})
		}
		s.dropUnsupportedTanks()
	}
	s.impacts = append(s.impacts, impactAnimation{
		pos:            impactPos,
		radius:         radius,
		duration:       duration,
		cycles:         impactCyclesForWeapon(weapon),
		damage:         weapon.Damage,
		radialDamage:   weapon.RadialDamage,
		attacker:       s.lastDamageSource,
		damageApplied:  damageApplied,
		outward:        weapon.ImpactGradientOutward,
		style:          weapon.ImpactAnimationStyle,
		terrainApplied: terrainApplied,
	})
	if minimumDelay := duration + s.impactPauseFrames(); s.turnAdvanceDelay < minimumDelay {
		s.turnAdvanceDelay = minimumDelay
	}
}

func (s *GameScene) applyCraterAndRefillWater(x, y, radius float64) []models.SandFallPixel {
	falls := s.ground.ApplyCrater(x, y, radius)
	area := image.Rect(
		int(math.Floor(x-radius))-1,
		int(math.Floor(y-radius))-1,
		int(math.Ceil(x+radius))+1,
		int(math.Ceil(y+radius))+1,
	)
	s.refillWaterBelowArea(area)
	return falls
}

func (s *GameScene) applyRingCraterAndRefillWater(x, y, radius, spacing, thickness float64) []models.SandFallPixel {
	falls := s.ground.ApplyRingCrater(x, y, radius, spacing, thickness)
	area := image.Rect(
		int(math.Floor(x-radius))-1,
		int(math.Floor(y-radius))-1,
		int(math.Ceil(x+radius))+1,
		int(math.Ceil(y+radius))+1,
	)
	s.refillWaterBelowArea(area)
	return falls
}

func (s *GameScene) refillWaterBelowArea(area image.Rectangle) {
	if area.Empty() || len(s.waterFills) == 0 {
		return
	}
	for i := range s.waterFills {
		fill := &s.waterFills[i]
		if area.Max.X < fill.leftX-1 || area.Min.X > fill.rightX+1 {
			continue
		}
		if s.rebuildWaterFillSurface(fill) {
			fill.cacheAge = -1
			fill.image = nil
		}
	}
}

func (s *GameScene) rebuildWaterFillSurface(fill *waterFill) bool {
	if fill == nil || len(fill.surfaceY) == 0 {
		return false
	}
	left := fill.leftX
	right := fill.rightX
	for left > 0 && s.ground.SurfaceY(float64(left-1)) > fill.topY {
		left--
	}
	maxX := maxInt(0, int(math.Round(s.worldWidth)))
	for right < maxX && s.ground.SurfaceY(float64(right+1)) > fill.topY {
		right++
	}

	nextSurface := make([]float64, right-left+1)
	for x := left; x <= right; x++ {
		nextSurface[x-left] = s.ground.SurfaceY(float64(x))
	}
	changed := left != fill.leftX || right != fill.rightX || len(nextSurface) != len(fill.surfaceY)
	if !changed {
		for i, y := range nextSurface {
			if y != fill.surfaceY[i] {
				changed = true
				break
			}
		}
	}
	if !changed {
		return false
	}
	fill.leftX = left
	fill.rightX = right
	fill.surfaceY = nextSurface
	return true
}

func (s *GameScene) zeroPowerTankImpactCenter(tank *battleTank) engine.Vec {
	if tank == nil || tank.body == nil {
		return engine.Vec{}
	}
	body := tank.body.Bounds()
	return engine.V(body.Center().X, body.Max.Y)
}

func (s *GameScene) fireZeroPowerScatterProjectiles(tank *battleTank) {
	if tank == nil || tank.body == nil {
		return
	}
	center := s.zeroPowerTankScatterOrigin(tank)
	strength := maxInt(30, tank.shotStrength)
	speed := 1.4 + float64(strength)*0.32
	colors := []color.RGBA{
		{R: 255, G: 230, B: 40, A: 255},
		{R: 0, G: 220, B: 210, A: 255},
		{R: 50, G: 95, B: 255, A: 255},
		{R: 255, G: 45, B: 35, A: 255},
		{R: 145, G: 255, B: 80, A: 255},
	}
	angleOffset := 15 * math.Pi / 180
	angles := []float64{
		-math.Pi/2 - 2*angleOffset,
		-math.Pi/2 - angleOffset,
		-math.Pi / 2,
		-math.Pi/2 + angleOffset,
		-math.Pi/2 + 2*angleOffset,
	}
	projectiles := make([]*projectile, 0, len(angles))
	for _, angle := range angles {
		velocity := engine.V(speed, 0).Rotated(angle)
		weapon := weaponspkg.Grenade()
		weapon.Color = colors[s.rng.Intn(len(colors))]
		projectiles = append(projectiles, &projectile{
			pos:                center,
			prev:               center,
			velocity:           velocity,
			weaponIndex:        1,
			effectiveWeapon:    weapon,
			hasEffectiveWeapon: true,
			zeroPowerScatter:   true,
			scatterImpactColor: weapon.Color,
			shooter:            tank,
			launchRot:          angle,
			trail:              []engine.Vec{center},
		})
	}
	s.setProjectiles(append(s.projectiles, projectiles...))
}

func (s *GameScene) zeroPowerTankScatterOrigin(tank *battleTank) engine.Vec {
	if tank == nil || tank.body == nil {
		return engine.Vec{}
	}
	return tank.body.Bounds().Center()
}

func (s *GameScene) onZeroPowerScatterGroundImpact(p *projectile) {
	if p == nil {
		return
	}
	s.playZeroPowerSound(zeroPowerSoundScatterImpact)
	radius := 30 + s.rng.Intn(71)
	profile := weaponspkg.RadialDamageProfile{
		InnerRadius: (2 * radius) / 3,
		OuterRadius: radius,
		MaxDamage:   100,
	}
	duration := s.impactAnimationFramesForWeapon(weaponspkg.Grenade())
	s.damageTanksInRadialProfile(p.pos, profile, p.shooter)
	falls := s.applyCraterAndRefillWater(p.pos.X, p.pos.Y, float64(radius))
	if len(falls) > 0 {
		s.sandFalls = append(s.sandFalls, sandFallAnimation{
			pixels:   falls,
			duration: sandFallFrames,
		})
	}
	s.dropUnsupportedTanks()
	s.impacts = append(s.impacts, impactAnimation{
		pos:            p.pos,
		radius:         float64(radius),
		duration:       duration,
		cycles:         1,
		color:          p.scatterImpactColor,
		terrainApplied: true,
	})
	if minimumDelay := duration + s.impactPauseFrames(); s.turnAdvanceDelay < minimumDelay {
		s.turnAdvanceDelay = minimumDelay
	}
}

func (s *GameScene) startWaterFill(pos engine.Vec) {
	fill, ok := s.waterFillAt(pos)
	if !ok {
		s.delayTurnAdvance(s.impactPauseFrames())
		return
	}
	s.playEventSound(soundEventWaterFill)
	s.waterFills = append(s.waterFills, fill)
	s.delayTurnAdvance(fill.duration + s.impactPauseFrames())
}

func (s *GameScene) waterFillAt(pos engine.Vec) (waterFill, bool) {
	const (
		boundaryRise = 10.0
		minHeight    = 3.0
	)
	impactX := int(math.Round(pos.X))
	if impactX < 0 || impactX >= int(s.worldWidth) {
		return waterFill{}, false
	}
	impactY := s.ground.SurfaceY(float64(impactX))
	targetY := impactY - boundaryRise

	leftX := -1
	leftY := impactY
	for x := impactX; x >= 0; x-- {
		y := s.ground.SurfaceY(float64(x))
		if y <= targetY {
			leftX = x
			leftY = y
			break
		}
	}
	rightX := -1
	rightY := impactY
	for x := impactX; x <= int(s.worldWidth); x++ {
		y := s.ground.SurfaceY(float64(x))
		if y <= targetY {
			rightX = x
			rightY = y
			break
		}
	}
	if leftX < 0 || rightX < 0 || rightX-leftX < 2 {
		return waterFill{}, false
	}

	topY := math.Max(leftY, rightY)
	if impactY-topY < minHeight {
		return waterFill{}, false
	}

	surface := make([]float64, rightX-leftX+1)
	visibleColumns := 0
	for x := leftX; x <= rightX; x++ {
		y := s.ground.SurfaceY(float64(x))
		surface[x-leftX] = y
		if y > topY+minHeight {
			visibleColumns++
		}
	}
	if visibleColumns == 0 {
		return waterFill{}, false
	}

	return waterFill{
		leftX:      leftX,
		rightX:     rightX,
		topY:       topY,
		surfaceY:   surface,
		duration:   secondsToFrames(1),
		hitPlayers: make(map[int]bool),
	}, true
}

func (s *GameScene) projectileHitsWaterSurface(p *projectile, radius float64) (engine.Vec, bool) {
	if p == nil || len(s.waterFills) == 0 {
		return engine.Vec{}, false
	}
	delta := p.pos.Sub(p.prev)
	steps := maxInt(4, int(math.Ceil(delta.Len()/2)))
	for step := 0; step <= steps; step++ {
		t := float64(step) / float64(steps)
		point := engine.V(p.prev.X+delta.X*t, p.prev.Y+delta.Y*t)
		for i := range s.waterFills {
			top, bottom, ok := s.waterColumnAt(&s.waterFills[i], point.X)
			if !ok {
				continue
			}
			if point.Y+radius < top || point.Y-radius > bottom {
				continue
			}
			return engine.V(point.X, top), true
		}
	}
	return engine.Vec{}, false
}

func (s *GameScene) startMoleImpact(pos engine.Vec) {
	s.damageTanksInImpactRadius(pos, weaponspkg.Moles())
	s.playEventSoundLoop(moleBroeslerLoopKey, soundEventCrumblerImpact)
	lines := make([]moleStarLine, moleStarLineCount)
	for i := range lines {
		angle := -math.Pi + float64(i)*2*math.Pi/float64(len(lines)) + (s.rng.Float64()-0.5)*0.08
		lines[i] = moleStarLine{angle: angle, length: moleStarRadius}
	}
	impact := &moleImpact{
		start:      pos,
		pos:        pos,
		tunnelPath: []engine.Vec{pos},
		starLines:  lines,
	}
	impact.editArea = s.ground.ClearCircle(pos.X, pos.Y, moleTunnelWidth/2)
	s.refillWaterBelowArea(impact.editArea)
	s.moleImpacts = append(s.moleImpacts, impact)
	s.delayTurnAdvance(moleTunnelFrames + moleStarFrames + s.impactPauseFrames())
}

func (s *GameScene) startSmallCrumblerImpact(pos engine.Vec, cfg crumblerConfig) {
	s.playEventSoundLoop(crumblerBroeslerLoopKey, soundEventCrumblerImpact)
	crumbs := make([]smallCrumb, cfg.count)
	for i := range crumbs {
		fan := 0.0
		if len(crumbs) > 1 {
			fan = float64(i)/float64(len(crumbs)-1)*2 - 1
		}
		startX := pos.X + (s.rng.Float64()-0.5)*cfg.startWidth
		startX = math.Max(0, math.Min(s.worldWidth-1, startX))
		crumbs[i] = smallCrumb{
			pos:    engine.V(startX, pos.Y-cfg.hiddenStart),
			fan:    fan,
			speed:  cfg.minSpeed + s.rng.Float64()*(cfg.maxSpeed-cfg.minSpeed),
			drift:  fan*0.34 + (s.rng.Float64()-0.5)*0.7,
			wobble: s.rng.Float64() * math.Pi * 2,
		}
	}

	impact := &smallCrumblerImpact{
		start:  pos,
		crumbs: crumbs,
		cfg:    cfg,
	}
	impact.editArea = s.ground.ClearRects(s.smallCrumblerAreas(impact))
	s.refillWaterBelowArea(impact.editArea)
	s.smallCrumblerImpacts = append(s.smallCrumblerImpacts, impact)
	s.crumblerCameraFocus = &impact.start
}

func (s *GameScene) startWaterSurfaceImpact(pos engine.Vec) {
	if len(s.waterBlotchAnimation.frames) == 0 {
		s.delayTurnAdvance(s.impactPauseFrames())
		return
	}
	s.playEventSound(soundEventWaterBlotch)
	effect := &waterSurfaceImpact{
		pos:       pos,
		duration:  s.waterBlotchAnimation.totalTicks,
		animation: s.waterBlotchAnimation,
	}
	s.waterBlotches = append(s.waterBlotches, effect)
	s.waterBlotchFocus = effect
	s.updateWaterBlotchCamera()
	s.delayTurnAdvance(effect.duration + secondsToFrames(0.5))
}

func (s *GameScene) startAirStrikeImpact(pos engine.Vec) {
	s.playEventSoundLoop(airStrikeBeaconLoopKey, soundEventAirStrikeBeacon)
	bombs := make([]airStrikeBomb, airStrikeBombCount)
	startY := -s.skyExtraHeight() - 80
	lastDelay := airStrikeWaitFrames
	for i := range bombs {
		offset := (float64(i) - float64(airStrikeBombCount-1)/2) * airStrikeBombSpacing
		offset += (s.rng.Float64()*2 - 1) * 6
		targetX := math.Max(0, math.Min(s.worldWidth, pos.X+offset))
		targetY := s.ground.SurfaceY(targetX)
		fallDistance := targetY - startY
		startX := targetX - math.Tan(airStrikeBombAngle)*fallDistance
		delay := airStrikeWaitFrames
		if airStrikeBombDelayWindow > 0 {
			delay += s.rng.Intn(airStrikeBombDelayWindow + 1)
		}
		lastDelay = max(lastDelay, delay)
		bombs[i] = airStrikeBomb{
			start:  engine.V(startX, startY),
			target: engine.V(targetX, targetY),
			pos:    engine.V(startX, startY),
			delay:  delay,
		}
	}
	duration := lastDelay + airStrikeBombFallFrames + s.impactAnimationFramesForWeapon(weaponspkg.AtomBomb()) + s.impactPauseFrames()
	s.airStrikeImpacts = append(s.airStrikeImpacts, &airStrikeImpact{
		pos:      pos,
		duration: duration,
		bombs:    bombs,
	})
	s.delayTurnAdvance(duration)
}

func (s *GameScene) updateAirStrikeImpacts() {
	if len(s.airStrikeImpacts) == 0 {
		return
	}
	active := s.airStrikeImpacts[:0]
	for _, impact := range s.airStrikeImpacts {
		if impact == nil {
			continue
		}
		s.updateAirStrikeImpact(impact)
		impact.age++
		if impact.age < impact.duration {
			active = append(active, impact)
		}
	}
	s.airStrikeImpacts = active
}

func (s *GameScene) updateAirStrikeImpact(impact *airStrikeImpact) {
	s.updateAirStrikeBeaconSound(impact)
	for i := range impact.bombs {
		bomb := &impact.bombs[i]
		if bomb.impacted || impact.age < bomb.delay {
			continue
		}
		progress := math.Min(1, float64(impact.age-bomb.delay)/math.Max(1, float64(airStrikeBombFallFrames)))
		bomb.pos = engine.V(
			bomb.start.X+(bomb.target.X-bomb.start.X)*progress,
			bomb.start.Y+(bomb.target.Y-bomb.start.Y)*progress,
		)
		if tank := s.airStrikeBombHitsTank(bomb.pos); tank != nil {
			bomb.impacted = true
			impact.bojeHidden = true
			s.applyAirStrikeBombImpact(bomb.pos)
			continue
		}
		groundY := s.ground.SurfaceY(bomb.pos.X)
		if progress >= 1 || bomb.pos.Y >= groundY {
			bomb.impacted = true
			impact.bojeHidden = true
			s.applyAirStrikeBombImpact(engine.V(bomb.pos.X, groundY))
		}
	}
}

func (s *GameScene) updateAirStrikeBeaconSound(impact *airStrikeImpact) {
	if impact == nil || s.blinkBojeAnimation.totalTicks <= 0 {
		return
	}
	if impact.bojeHidden {
		s.stopSoundLoop(airStrikeBeaconLoopKey)
		return
	}
	cycle := impact.age / s.blinkBojeAnimation.totalTicks
	if impact.beepsPlayed >= 5 && !impact.jetStarted {
		impact.jetStarted = true
		s.stopSoundLoop(airStrikeBeaconLoopKey)
		s.playEventSound(soundEventAirStrikeJet)
		return
	}
	if cycle <= impact.lastBeaconCycle {
		return
	}
	impact.lastBeaconCycle = cycle
	impact.beepsPlayed++
}

func (s *GameScene) airStrikeBombHitsTank(pos engine.Vec) *battleTank {
	radius := 4.0
	bounds := engine.R(pos.X-radius, pos.Y-radius, pos.X+radius, pos.Y+radius)
	for _, tank := range s.tanks {
		if tank == nil || tank.body == nil || tank.power <= 0 {
			continue
		}
		if engine.Collision(bounds, tank.body.Bounds().ScaledAtCenter(0.78)) {
			return tank
		}
	}
	return nil
}

func (s *GameScene) applyAirStrikeBombImpact(pos engine.Vec) {
	s.playEventSound(soundEventAirStrikeBomb)
	weapon := weaponspkg.AtomBomb()
	radius := impactRadiusForWeapon(weapon) * airStrikeImpactScale
	duration := s.impactAnimationFramesForWeapon(weapon)
	s.zeroPowerStartDelay = (duration * 2) / 3
	s.damageTanksInImpactRadius(pos, weapon)
	s.zeroPowerStartDelay = 0
	if falls := s.applyCraterAndRefillWater(pos.X, pos.Y, radius); len(falls) > 0 {
		s.sandFalls = append(s.sandFalls, sandFallAnimation{
			pixels:   falls,
			duration: sandFallFrames,
		})
	}
	s.dropUnsupportedTanks()
	s.impacts = append(s.impacts, impactAnimation{
		pos:            pos,
		radius:         radius,
		duration:       duration,
		cycles:         impactCyclesForWeapon(weapon),
		outward:        weapon.ImpactGradientOutward,
		style:          weapon.ImpactAnimationStyle,
		terrainApplied: true,
	})
}

func (s *GameScene) startShockwaveImpact(pos engine.Vec, weapon weaponspkg.Weapon) {
	s.playEventSound(soundEventShockwave)
	radius := impactRadiusForWeapon(weapon)
	duration := shockwavePulseFrames * shockwavePulseCount
	effect := &shockwaveImpact{
		pos:      pos,
		radius:   radius,
		duration: duration,
	}
	s.shockwaveImpacts = append(s.shockwaveImpacts, effect)
	s.damageTanksInShockwave(pos, radius)
	s.delayTurnAdvance(duration + s.impactPauseFrames())
}

func (s *GameScene) updateShockwaveImpacts() {
	if len(s.shockwaveImpacts) == 0 {
		return
	}
	active := s.shockwaveImpacts[:0]
	for _, impact := range s.shockwaveImpacts {
		if impact == nil {
			continue
		}
		impact.age++
		if impact.age < impact.duration {
			active = append(active, impact)
		}
	}
	s.shockwaveImpacts = active
}

func (s *GameScene) damageTanksInShockwave(center engine.Vec, visibleRadius float64) {
	if visibleRadius <= 0 {
		return
	}
	s.damageTanksInImpactRadius(center, weaponspkg.Shockwave())
}

func (s *GameScene) startMoskitoImpact(pos engine.Vec, shooter *battleTank) {
	groundY := s.ground.SurfaceY(pos.X)
	ground := engine.V(pos.X, groundY)
	hover := engine.V(pos.X, groundY-mosquitoHoverHeight)
	target := s.nearestLivingTank(hover)
	if target == nil {
		s.delayTurnAdvance(s.impactPauseFrames())
		return
	}
	s.playEventSoundLoop(moskitosLoopKey, soundEventMosquitos)
	s.moskitoEffects = append(s.moskitoEffects, &moskitoEffect{
		pos:     ground,
		ground:  ground,
		hover:   hover,
		target:  target,
		shooter: shooter,
	})
}

func (s *GameScene) updateMoskitoEffects() {
	if len(s.moskitoEffects) == 0 {
		return
	}
	active := s.moskitoEffects[:0]
	for _, effect := range s.moskitoEffects {
		if effect == nil {
			continue
		}
		s.updateMoskitoEffect(effect)
		if effect.finished {
			continue
		}
		active = append(active, effect)
	}
	s.moskitoEffects = active
	if len(s.moskitoEffects) == 0 {
		s.stopSoundLoop(moskitosLoopKey)
	}
}

func (s *GameScene) updateMoskitoEffect(effect *moskitoEffect) {
	s.moskitoCameraFocus = &effect.pos
	switch effect.phase {
	case moskitoPhaseTouch:
		effect.pos = effect.ground
		effect.age++
		if effect.age >= mosquitoTouchFrames {
			effect.phase = moskitoPhaseRise
			effect.age = 0
		}
	case moskitoPhaseRise:
		progress := easeOut(float64(effect.age) / math.Max(1, float64(mosquitoRiseFrames-1)))
		effect.pos = engine.V(
			effect.ground.X+(effect.hover.X-effect.ground.X)*progress,
			effect.ground.Y+(effect.hover.Y-effect.ground.Y)*progress,
		)
		effect.age++
		if effect.age >= mosquitoRiseFrames {
			effect.pos = effect.hover
			effect.phase = moskitoPhaseQuestion
			effect.age = 0
		}
	case moskitoPhaseQuestion:
		effect.pos = effect.hover
		effect.age++
		if effect.age >= s.moskitoQuestionDuration() {
			effect.phase = moskitoPhaseSeek
			effect.age = 0
		}
	case moskitoPhaseSeek:
		if !s.moskitoHasLiveTarget(effect) {
			effect.target = s.nearestLivingTank(effect.pos)
		}
		if !s.moskitoHasLiveTarget(effect) {
			effect.finished = true
			s.delayTurnAdvance(s.impactPauseFrames())
			return
		}
		if s.moveMoskitoAlongTerrain(effect, effect.target, mosquitoSeekSpeed) {
			effect.phase = moskitoPhaseDive
			effect.age = 0
			effect.pos = s.moskitoTargetHover(effect.target)
		}
	case moskitoPhaseDive:
		if !s.moskitoHasLiveTarget(effect) {
			effect.finished = true
			s.delayTurnAdvance(s.impactPauseFrames())
			return
		}
		if s.moveMoskitoToward(effect, effect.target.body.Bounds().Center(), mosquitoDiveSpeed) {
			effect.phase = moskitoPhaseAttached
			effect.age = 0
		}
	case moskitoPhaseAttached:
		if effect.target == nil || effect.target.body == nil {
			effect.finished = true
			s.delayTurnAdvance(s.impactPauseFrames())
			return
		}
		effect.pos = effect.target.body.Bounds().Center()
		if !effect.colored && effect.age >= mosquitoColorDelayFrames {
			s.recolorTankToGround(effect.target)
			effect.colored = true
		}
		effect.age++
		if effect.age >= mosquitoAttachedFrames {
			previousPower := effect.target.power
			s.damageTankAsTerrain(effect.target, 100, effect.shooter, damageCauseDirect)
			if previousPower > 0 && effect.target.power == 0 {
				s.playEventSound(soundEventMosquitoScream)
			}
			effect.finished = true
			s.delayTurnAdvance(s.tankHitPauseFrames())
		}
	}
}

func (s *GameScene) moskitoQuestionDuration() int {
	if s.questionAnimation.totalTicks > 0 {
		return s.questionAnimation.totalTicks
	}
	return secondsToFrames(2)
}

func (s *GameScene) moskitoHasLiveTarget(effect *moskitoEffect) bool {
	return effect != nil && effect.target != nil && effect.target.body != nil && effect.target.power > 0
}

func (s *GameScene) moskitoTargetHover(tank *battleTank) engine.Vec {
	if tank == nil || tank.body == nil {
		return engine.Vec{}
	}
	bounds := tank.body.Bounds()
	return engine.V(bounds.Center().X, bounds.Min.Y-mosquitoHoverHeight)
}

func (s *GameScene) moveMoskitoToward(effect *moskitoEffect, target engine.Vec, speed float64) bool {
	delta := target.Sub(effect.pos)
	distance := delta.Len()
	if distance <= speed || distance <= 0.01 {
		effect.pos = target
		return true
	}
	effect.pos = *effect.pos.Add(delta.Scaled(speed / distance))
	return false
}

func (s *GameScene) moveMoskitoAlongTerrain(effect *moskitoEffect, target *battleTank, speed float64) bool {
	if effect == nil || target == nil || target.body == nil {
		return true
	}
	targetX := target.body.Bounds().Center().X
	dx := targetX - effect.pos.X
	if math.Abs(dx) <= speed {
		return true
	}
	nextX := effect.pos.X + math.Copysign(speed, dx)
	nextX = math.Max(0, math.Min(s.worldWidth, nextX))
	nextY := s.ground.SurfaceY(nextX) - mosquitoHoverHeight
	effect.pos.X = nextX
	effect.pos.Y = nextY
	return false
}

func (s *GameScene) nearestLivingTank(pos engine.Vec) *battleTank {
	var nearest *battleTank
	nearestDistance := math.MaxFloat64
	for _, tank := range s.tanks {
		if tank == nil || tank.body == nil || tank.power <= 0 {
			continue
		}
		distance := tank.body.Bounds().Center().Sub(pos).Len()
		if distance < nearestDistance {
			nearest = tank
			nearestDistance = distance
		}
	}
	return nearest
}

func (s *GameScene) recolorTankToGround(tank *battleTank) {
	if tank == nil || tank.body == nil {
		return
	}
	bounds := tank.body.Bounds()
	x := bounds.Center().X
	startY := math.Ceil(s.ground.SurfaceY(x))
	groundHeight := 0
	if s.ground.Size != nil {
		groundHeight = int(s.ground.Size.Y)
	}
	for y := int(startY); y < groundHeight; y++ {
		c := s.ground.ColorAt(x, float64(y))
		if c.A == 0 {
			continue
		}
		c.A = 255
		tank.tint = c
		models.RecolorTankBody(tank.body, tank.tint)
		models.RecolorCannon(tank.cannon, tank.tint)
		return
	}
}

func (s *GameScene) updateMoleImpacts() {
	if len(s.moleImpacts) == 0 {
		return
	}
	active := s.moleImpacts[:0]
	for _, impact := range s.moleImpacts {
		if impact == nil {
			continue
		}
		switch impact.phase {
		case molePhaseTunnel:
			s.updateMoleTunnel(impact)
		case molePhaseStar:
			s.updateMoleStar(impact)
		}
		impact.age++
		if impact.phase == molePhaseTunnel && impact.age >= moleTunnelFrames {
			impact.phase = molePhaseStar
			impact.age = 0
			s.playEventSoundLoop(moleBroeslerLoopKey, soundEventCrumblerImpact)
		}
		if impact.phase == molePhaseStar && impact.age >= moleStarFrames {
			if falls := s.ground.SettleArea(impact.editArea); len(falls) > 0 {
				s.refillWaterBelowArea(impact.editArea)
				s.sandFalls = append(s.sandFalls, sandFallAnimation{
					pixels:   falls,
					duration: sandFallFrames,
				})
			}
			s.dropUnsupportedTanks()
			s.stopSoundLoop(moleBroeslerLoopKey)
			s.playEventSound(soundEventMoleImpact)
			continue
		}
		active = append(active, impact)
	}
	s.moleImpacts = active
}

func (s *GameScene) updateMoleTunnel(impact *moleImpact) {
	progress := easeInOut(float64(impact.age) / math.Max(1, moleTunnelFrames-1))
	wobble := math.Sin(float64(impact.age)*0.82)*5 + math.Sin(float64(impact.age)*2.17)*2
	next := engine.V(impact.start.X+wobble, impact.start.Y+moleTunnelDepth*progress)
	impact.pos = next
	impact.tunnelPath = append(impact.tunnelPath, next)
	area := s.ground.ClearCircle(next.X, next.Y, moleTunnelWidth/2)
	s.refillWaterBelowArea(area)
	impact.editArea = unionRect(impact.editArea, area)
}

func (s *GameScene) updateMoleStar(impact *moleImpact) {
	progress := easeOut(float64(impact.age) / math.Max(1, moleStarFrames-1))
	for _, line := range impact.starLines {
		end := impact.pos.Add(engine.V(line.length*progress, 0).Rotated(line.angle))
		area := s.ground.ClearLine(impact.pos.X, impact.pos.Y, end.X, end.Y, moleStarLineThickness)
		s.refillWaterBelowArea(area)
		impact.editArea = unionRect(impact.editArea, area)
	}
}

func (s *GameScene) updateSmallCrumblerImpacts() {
	if len(s.smallCrumblerImpacts) == 0 {
		return
	}
	active := s.smallCrumblerImpacts[:0]
	for _, impact := range s.smallCrumblerImpacts {
		if impact == nil {
			continue
		}
		if impact.frozen {
			impact.freezeAge++
			if impact.freezeAge >= impact.cfg.freezeFrames {
				delay := s.impactPauseFrames()
				if falls := s.ground.SettleArea(impact.editArea); len(falls) > 0 {
					s.refillWaterBelowArea(impact.editArea)
					s.sandFalls = append(s.sandFalls, sandFallAnimation{
						pixels:   falls,
						duration: sandFallFrames,
					})
					delay += sandFallFrames
				}
				s.dropUnsupportedTanks()
				s.stopSoundLoop(crumblerBroeslerLoopKey)
				s.playEventSound(soundEventMoleImpact)
				s.delayTurnAdvance(delay)
				continue
			}
			active = append(active, impact)
			continue
		}

		impact.age++
		reachedEnd := false
		for i := range impact.crumbs {
			crumb := &impact.crumbs[i]
			if crumb.stopped {
				continue
			}
			depth := math.Max(0, crumb.pos.Y-impact.start.Y)
			startHalfWidth := impact.cfg.startWidth / 2
			spread := math.Max(2, startHalfWidth+(impact.cfg.funnelSpread-startHalfWidth)*math.Min(1, depth/impact.cfg.maxDepth))
			crumb.pos.Y += crumb.speed
			targetX := impact.start.X + crumb.fan*spread
			randomStep := (s.rng.Float64()*2 - 1) * impact.cfg.jitter
			wobble := math.Sin(float64(impact.age)*1.35+crumb.wobble) * impact.cfg.wobble
			steer := 0.0
			if crumb.pos.Y >= impact.start.Y {
				steer = (targetX - crumb.pos.X) * 0.07
			}
			crumb.pos.X += steer + crumb.drift + wobble + randomStep
			if crumb.pos.Y >= impact.start.Y {
				crumb.pos.X = math.Max(impact.start.X-spread, math.Min(impact.start.X+spread, crumb.pos.X))
			}
			crumb.pos.X = math.Max(0, math.Min(s.worldWidth-1, crumb.pos.X))
			crumb.lastArea = smallCrumblerArea(crumb.pos, impact.cfg)

			if crumb.pos.Y-impact.start.Y >= impact.cfg.maxDepth || crumb.pos.Y >= s.battlefieldHeight()-2 {
				crumb.stopped = true
				reachedEnd = true
			}
		}

		if area := s.ground.ClearRects(s.smallCrumblerAreas(impact)); !area.Empty() {
			s.refillWaterBelowArea(area)
			impact.editArea = unionRect(impact.editArea, area)
		}
		if reachedEnd {
			impact.frozen = true
			impact.freezeAge = 0
			for i := range impact.crumbs {
				impact.crumbs[i].stopped = true
			}
		}
		active = append(active, impact)
	}
	s.smallCrumblerImpacts = active
}

func (s *GameScene) smallCrumblerAreas(impact *smallCrumblerImpact) []image.Rectangle {
	if impact == nil {
		return nil
	}
	areas := make([]image.Rectangle, 0, len(impact.crumbs))
	for i := range impact.crumbs {
		crumb := &impact.crumbs[i]
		if crumb.stopped && !impact.frozen {
			continue
		}
		if crumb.pos.Y < impact.start.Y {
			continue
		}
		area := crumb.lastArea
		if area.Empty() {
			area = smallCrumblerArea(crumb.pos, impact.cfg)
		}
		areas = append(areas, area)
	}
	return areas
}

func smallCrumblerArea(pos engine.Vec, cfg crumblerConfig) image.Rectangle {
	x := int(math.Round(pos.X))
	y := int(math.Round(pos.Y))
	return image.Rect(x, y, x+cfg.width, y+cfg.height)
}

func smallCrumblerConfig() crumblerConfig {
	return crumblerConfig{
		count:        smallCrumblerCount,
		width:        smallCrumblerWidth,
		height:       smallCrumblerHeight,
		maxDepth:     smallCrumblerMaxDepth,
		freezeFrames: smallCrumblerFreezeFrames,
		minSpeed:     smallCrumblerMinSpeed,
		maxSpeed:     smallCrumblerMaxSpeed,
		startWidth:   smallCrumblerStartWidth,
		hiddenStart:  smallCrumblerHiddenStart,
		funnelSpread: smallCrumblerFunnelSpread,
		jitter:       smallCrumblerJitter,
		wobble:       smallCrumblerWobble,
	}
}

func largeCrumblerConfig() crumblerConfig {
	return crumblerConfig{
		count:        largeCrumblerCount,
		width:        largeCrumblerWidth,
		height:       largeCrumblerHeight,
		maxDepth:     largeCrumblerMaxDepth,
		freezeFrames: largeCrumblerFreezeFrames,
		minSpeed:     largeCrumblerMinSpeed,
		maxSpeed:     largeCrumblerMaxSpeed,
		startWidth:   largeCrumblerStartWidth,
		hiddenStart:  largeCrumblerHiddenStart,
		funnelSpread: largeCrumblerFunnelSpread,
		jitter:       largeCrumblerJitter,
		wobble:       largeCrumblerWobble,
	}
}

func unionRect(a, b image.Rectangle) image.Rectangle {
	if a.Empty() {
		return b
	}
	if b.Empty() {
		return a
	}
	return a.Union(b)
}

func (s *GameScene) damageTanksTouchingWater(fill *waterFill) {
	if fill == nil {
		return
	}
	for _, tank := range s.tanks {
		if tank == nil || tank.body == nil || tank.power <= 0 {
			continue
		}
		if !s.tankTouchesWater(tank, fill) {
			continue
		}
		if fill.hitPlayers != nil && fill.hitPlayers[tank.playerIndex] {
			continue
		}
		s.drownTank(tank, fill)
	}
}

func (s *GameScene) tankTouchesWater(tank *battleTank, fill *waterFill) bool {
	if tank == nil || tank.body == nil || fill == nil {
		return false
	}
	bounds := tank.body.Bounds().ScaledAtCenter(0.78)
	if bounds.Max.X < float64(fill.leftX) || bounds.Min.X > float64(fill.rightX) {
		return false
	}
	const samples = 7
	for i := 0; i < samples; i++ {
		t := 0.0
		if samples > 1 {
			t = float64(i) / float64(samples-1)
		}
		x := bounds.Min.X + bounds.W()*t
		top, bottom, ok := s.waterColumnAt(fill, x)
		if !ok {
			continue
		}
		if bounds.Max.Y >= top && bounds.Min.Y <= bottom {
			return true
		}
	}
	return false
}

func (s *GameScene) waterColumnAt(fill *waterFill, x float64) (float64, float64, bool) {
	if fill == nil {
		return 0, 0, false
	}
	column := int(math.Round(x)) - fill.leftX
	if column < 0 || column >= len(fill.surfaceY) {
		return 0, 0, false
	}
	bottom := fill.surfaceY[column]
	if bottom <= fill.topY {
		return 0, 0, false
	}
	top := s.currentWaterTop(fill)
	if bottom-top < 1 {
		return 0, 0, false
	}
	return top, bottom, true
}

func (s *GameScene) currentWaterTop(fill *waterFill) float64 {
	if fill == nil || len(fill.surfaceY) == 0 {
		return 0
	}
	deepestY := fill.topY
	for _, y := range fill.surfaceY {
		if y > deepestY {
			deepestY = y
		}
	}
	progress := 1.0
	if fill.duration > 0 && fill.age < fill.duration {
		progress = easeOut(float64(fill.age) / float64(fill.duration))
	}
	return deepestY - (deepestY-fill.topY)*progress
}

func (s *GameScene) drownTank(tank *battleTank, fill *waterFill) {
	if tank == nil || tank.body == nil || tank.power <= 0 {
		return
	}
	if fill != nil {
		if fill.hitPlayers == nil {
			fill.hitPlayers = make(map[int]bool)
		}
		fill.hitPlayers[tank.playerIndex] = true
	}
}

func (s *GameScene) removeTankSprites(tank *battleTank) {
	if tank == nil || s.layers == nil || layerTanks >= len(s.layers) || s.layers[layerTanks] == nil {
		return
	}
	if tank.body != nil {
		s.layers[layerTanks].Remove(tank.body)
	}
	if tank.cannon != nil {
		s.layers[layerTanks].Remove(tank.cannon)
	}
}

func (s *GameScene) startWaterBlubberForTank(tank *battleTank, fill *waterFill) {
	if tank == nil || tank.body == nil || len(s.waterBlubberAnimation.frames) == 0 {
		return
	}
	s.playEventSound(soundEventWaterBlubber)
	center := tank.body.Bounds().Center()
	start := center
	end := engine.V(center.X, fill.topY)
	distance := math.Max(1, start.Y-end.Y)
	duration := maxInt(s.waterBlubberAnimation.totalTicks*4, int(math.Round(distance*1.2)))
	effect := &waterBlubberEffect{
		start:     start,
		end:       end,
		duration:  duration,
		animation: s.waterBlubberAnimation,
	}
	s.waterBlubbers = append(s.waterBlubbers, effect)
	s.waterCameraFocus = effect
	if minimumDelay := duration + secondsToFrames(0.5); s.turnAdvanceDelay < minimumDelay {
		s.turnAdvanceDelay = minimumDelay
	}
}

func (s *GameScene) startFireballImpact(pos engine.Vec, weapon weaponspkg.Weapon) {
	animation := s.fireballAnimation
	if len(animation.frames) == 0 {
		return
	}
	s.playEventSound(soundEventPalmIgnite)
	duration := maxInt(animation.totalTicks, s.impactAnimationFramesForWeapon(weapon))
	effect := animatedImpact{
		pos:       pos,
		duration:  duration,
		animation: animation,
	}
	s.animatedImpacts = append(s.animatedImpacts, effect)
	s.damageTanksInImpactRadius(pos, weapon)
	if minimumDelay := duration + s.impactPauseFrames(); s.turnAdvanceDelay < minimumDelay {
		s.turnAdvanceDelay = minimumDelay
	}
}

func (s *GameScene) startDudImpact(pos engine.Vec) {
	animation := s.dudImpactAnimation
	if len(animation.frames) == 0 {
		return
	}
	s.playEventSound(soundEventDudImpact)
	duration := maxInt(1, animation.totalTicks)
	s.animatedImpacts = append(s.animatedImpacts, animatedImpact{
		pos:       pos,
		duration:  duration,
		animation: animation,
	})
	if minimumDelay := duration + s.impactPauseFrames(); s.turnAdvanceDelay < minimumDelay {
		s.turnAdvanceDelay = minimumDelay
	}
}

func positiveMod(value, divisor int) int {
	if divisor <= 0 {
		return 0
	}
	result := value % divisor
	if result < 0 {
		result += divisor
	}
	return result
}

func (s *GameScene) animationWorldRect(center engine.Vec, animation spriteAnimation) engine.Rect {
	width := float64(animation.width) * animation.scaleX
	height := float64(animation.height) * animation.scaleY
	return engine.R(center.X-width/2, center.Y-height, center.X+width/2, center.Y)
}

func distancePointToRect(point engine.Vec, rect engine.Rect) float64 {
	closestX := math.Max(rect.Min.X, math.Min(point.X, rect.Max.X))
	closestY := math.Max(rect.Min.Y, math.Min(point.Y, rect.Max.Y))
	return math.Hypot(point.X-closestX, point.Y-closestY)
}

func (s *GameScene) updateImpacts() {
	if len(s.impacts) == 0 {
		return
	}
	active := s.impacts[:0]
	for _, impact := range s.impacts {
		impact.age++
		if impact.style == weaponspkg.ImpactAnimationPlasma {
			progress := float64(impact.age) / math.Max(1, float64(impact.duration))
			s.damageTanksTouchedByPlasmaImpact(&impact, progress)
			if !impact.terrainApplied {
				if progress >= plasmaGreenProgress {
					falls := s.applyRingCraterAndRefillWater(impact.pos.X, impact.pos.Y, impact.radius, plasmaRingSpacing, plasmaRingThickness)
					if len(falls) > 0 {
						s.sandFalls = append(s.sandFalls, sandFallAnimation{
							pixels:   falls,
							duration: sandFallFrames,
						})
					}
					s.dropUnsupportedTanks()
					impact.terrainApplied = true
				}
			}
		}
		if impact.age < impact.duration {
			active = append(active, impact)
		}
	}
	s.impacts = active
}

func (s *GameScene) damageTanksTouchedByPlasmaImpact(impact *impactAnimation, progress float64) {
	if impact == nil || impact.radialDamage.MaxDamage <= 0 || impact.damageApplied == nil {
		return
	}
	radius := plasmaVisibleRadius(impact.radius, progress)
	if radius <= 0 {
		return
	}
	for _, tank := range s.tanks {
		if tank == nil || tank.body == nil || tank.power <= 0 || impact.damageApplied[tank.playerIndex] {
			continue
		}
		distance := distancePointToRect(impact.pos, tank.body.Bounds().ScaledAtCenter(0.78))
		if distance > radius {
			continue
		}
		damage := radialDamageToTank(impact.pos, tank, impact.radialDamage)
		if damage <= 0 {
			continue
		}
		impact.damageApplied[tank.playerIndex] = true
		s.damageTank(tank, damage, impact.attacker, damageCauseDirect)
	}
}

func plasmaVisibleRadius(radius, progress float64) float64 {
	progress = math.Max(0, math.Min(1, progress))
	if progress < plasmaBuildProgress {
		return radius * easeOut(progress/plasmaBuildProgress)
	}
	return radius
}

func (s *GameScene) updateAnimatedImpacts() {
	if len(s.animatedImpacts) == 0 {
		return
	}
	active := s.animatedImpacts[:0]
	for _, impact := range s.animatedImpacts {
		impact.age++
		if impact.age < impact.duration {
			active = append(active, impact)
		} else if impact.loopKey != "" {
			s.stopSoundLoop(impact.loopKey)
		}
	}
	s.animatedImpacts = active
}

func (s *GameScene) updateWaterFills() {
	for i := range s.waterFills {
		if s.waterFills[i].age < s.waterFills[i].duration {
			s.waterFills[i].age++
		}
		s.damageTanksTouchingWater(&s.waterFills[i])
	}
}

func (s *GameScene) updateWaterBlubbers() {
	if len(s.waterBlubbers) == 0 {
		return
	}
	active := s.waterBlubbers[:0]
	for _, effect := range s.waterBlubbers {
		if effect == nil {
			continue
		}
		effect.age++
		if effect.age < effect.duration {
			active = append(active, effect)
		}
	}
	s.waterBlubbers = active
	if len(s.waterBlubbers) == 0 && s.turnAdvanceDelay <= 0 {
		s.delayTurnAdvance(s.impactPauseFrames())
	}
}

func (s *GameScene) updateWaterBlotches() {
	if len(s.waterBlotches) == 0 {
		return
	}
	active := s.waterBlotches[:0]
	for _, effect := range s.waterBlotches {
		if effect == nil {
			continue
		}
		effect.age++
		if effect.age < effect.duration {
			active = append(active, effect)
		}
	}
	s.waterBlotches = active
}

func (s *GameScene) updatePalms() {
	if len(s.palms) == 0 {
		return
	}
	active := s.palms[:0]
	for _, palm := range s.palms {
		if palm == nil {
			continue
		}
		s.updatePalmGroundSupport(palm)
		s.updatePalmEyes(palm)
		s.updatePalmGrin(palm)
		palm.age++
		switch palm.state {
		case palmStateBurning:
			if palm.age >= s.palmFireAnimation.totalTicks {
				s.setPalmSkeleton(palm, true)
			}
		case palmStateSkeletonSmoking:
			if palm.age >= s.palmSmokeAnimation.totalTicks {
				palm.state = palmStateSkeleton
				palm.age = 0
			}
		case palmStateCrumbling:
			if palm.age >= s.palmCrumbleAnimation.totalTicks {
				if palm.sprite != nil && s.layers[layerPalms] != nil {
					s.layers[layerPalms].Remove(palm.sprite)
				}
				continue
			}
		}
		active = append(active, palm)
	}
	s.palms = active
}

func (s *GameScene) updatePalmLeafFalls() {
	if len(s.palmLeafFalls) == 0 {
		return
	}
	active := s.palmLeafFalls[:0]
	for _, leaf := range s.palmLeafFalls {
		leaf.age++
		surfaceY := s.ground.SurfaceY(leaf.pos.X)
		if leaf.landed {
			if surfaceY <= leaf.pos.Y+3.5 {
				active = append(active, leaf)
				continue
			}
			leaf.landed = false
			leaf.age = 0
			leaf.velocity = engine.V((s.rng.Float64()-0.5)*0.2, 0.65+s.rng.Float64()*0.8)
		}

		leaf.pos.X += leaf.velocity.X
		leaf.pos.Y += leaf.velocity.Y
		leaf.velocity.Y = math.Min(leaf.velocity.Y+0.025, 2.8)
		if leaf.pos.Y+3 >= surfaceY {
			leaf.pos.Y = surfaceY - 3
			leaf.velocity = engine.Vec{}
			leaf.landed = true
		}
		if leaf.age < leaf.maxAge || leaf.landed {
			active = append(active, leaf)
		}
	}
	s.palmLeafFalls = active
}

func (s *GameScene) updatePalmRevenge() {
	if s.palmRevenge == nil || s.palmRevenge.phase == palmRevengeNone {
		return
	}
	event := s.palmRevenge
	event.age++
	if event.palm != nil && event.palm.screaming {
		event.palm.screamAge++
	}

	switch event.phase {
	case palmRevengeCloudFocus:
		if event.cloud != nil && event.cloud.sprite != nil {
			s.cameraGoal = s.cameraTargetForWorldX(event.cloud.sprite.Bounds().Center().X)
			s.cameraGoalY = s.cameraTargetForWorldY(event.cloud.sprite.Bounds().Center().Y)
			shake := math.Sin(float64(event.age)*math.Pi*26/float64(palmRevengeFocusFrames)) * 9
			s.cameraX = approach(s.cameraX, s.cameraGoal+shake, 0.18, 0.8)
			s.cameraY = approach(s.cameraY, s.cameraGoalY, 0.18, 0.8)
		}
		if event.age >= palmRevengeFocusFrames {
			event.age = 0
			event.phase = palmRevengeCloudAttack
			if event.palm != nil {
				event.palm.screaming = false
				event.palm.grinning = true
				event.palm.grinAge = 0
				event.palm.grinHideAt = 0
			}
		}
	case palmRevengeCloudAttack:
		progress := easeOut(float64(event.age) / float64(palmRevengeAttackFrames))
		s.moveRevengeCloud(event, progress)
		if event.cloud != nil && event.cloud.sprite != nil {
			s.cameraGoal = s.cameraTargetForWorldX(event.cloud.sprite.Bounds().Center().X)
			s.cameraGoalY = s.cameraTargetForWorldY(event.cloud.sprite.Bounds().Center().Y)
			s.cameraX = approach(s.cameraX, s.cameraGoal, 0.2, 0.8)
			s.cameraY = approach(s.cameraY, s.cameraGoalY, 0.2, 0.8)
		}
		if event.age >= palmRevengeAttackFrames {
			event.age = 0
			event.phase = palmRevengeLightning
		}
	case palmRevengeLightning:
		s.cameraGoal = event.targetCameraX
		s.cameraGoalY = 0
		s.cameraX = approach(s.cameraX, s.cameraGoal, 0.22, 0.8)
		s.cameraY = approach(s.cameraY, s.cameraGoalY, 0.22, 0.8)
		s.playPalmRevengeLightningSound(event)
		if s.palmRevengeLightningVisible(event) && !event.tankBlackened {
			s.blackenPalmRevengeTank(event)
		}
		if event.age >= 48 && event.age < 66 && !event.smokeDone {
			s.startPalmRevengeSmoke(event)
		}
		if event.age >= 140 && !event.damageDone {
			s.damagePalmRevengeTank(event)
		}
		if event.smokeDone {
			event.smokeAge++
		}
		if event.age >= palmRevengeLightningFrames {
			event.age = 0
			event.phase = palmRevengeRecover
		}
	case palmRevengeRecover:
		if event.smokeDone {
			event.smokeAge++
		}
		progress := easeOut(float64(event.age) / float64(palmRevengeRecoverFrames))
		if event.cloud != nil && event.cloud.sprite != nil && event.cloud.sprite.Pos != nil {
			event.cloud.sprite.Pos.X = event.cloudAttack.X + (event.cloudOriginal.X-event.cloudAttack.X)*progress
			event.cloud.sprite.Pos.Y = event.cloudAttack.Y + (event.cloudOriginal.Y-event.cloudAttack.Y)*progress
		}
		s.cameraGoal = event.targetCameraX
		s.cameraGoalY = 0
		s.cameraX = approach(s.cameraX, s.cameraGoal, 0.22, 0.8)
		s.cameraY = approach(s.cameraY, s.cameraGoalY, 0.22, 0.8)
		if event.age >= palmRevengeRecoverFrames {
			s.finishPalmRevenge(event)
		}
	}
}

func (s *GameScene) moveRevengeCloud(event *palmRevengeEvent, progress float64) {
	if event == nil || event.cloud == nil || event.cloud.sprite == nil || event.cloud.sprite.Pos == nil {
		return
	}
	event.cloud.sprite.Pos.X = event.cloudStart.X + (event.cloudAttack.X-event.cloudStart.X)*progress
	event.cloud.sprite.Pos.Y = event.cloudStart.Y + (event.cloudAttack.Y-event.cloudStart.Y)*progress
}

func (s *GameScene) blackenPalmRevengeTank(event *palmRevengeEvent) {
	if event == nil || event.tankBlackened || event.target == nil {
		return
	}
	event.tankBlackened = true
	tank := event.target
	tank.tint = color.RGBA{A: 255}
	models.RecolorTankBody(tank.body, tank.tint)
	models.RecolorCannon(tank.cannon, tank.tint)
	if event.palm != nil && event.palm.grinning {
		event.palm.grinAge = 0
		event.palm.grinHideAt = secondsToFrames(2)
	}
}

func (s *GameScene) damagePalmRevengeTank(event *palmRevengeEvent) {
	if event == nil || event.damageDone || event.target == nil {
		return
	}
	event.damageDone = true
	tank := event.target
	previousPower := tank.power
	nominalDamage := 100
	appliedDamage := s.applyEnergyShieldDamage(tank, nominalDamage)
	tank.power = maxInt(0, tank.power-appliedDamage)
	tank.shotStrength = minInt(tank.shotStrength, maxInt(0, tank.power))
	if previousPower > 0 && tank.power == 0 {
		s.playEventSound(soundEventRevengeTankBroken)
		s.applySuicidePenalty(tank)
		tank.zeroPowerShown = true
	}
}

func (s *GameScene) playPalmRevengeLightningSound(event *palmRevengeEvent) {
	if event == nil || event.phase != palmRevengeLightning {
		return
	}
	if event.age < 18 && !event.lightningSoundDone[0] {
		event.lightningSoundDone[0] = true
		s.playEventSound(soundEventCloudLightning)
		return
	}
	if event.age >= 48 && event.age < 66 && !event.lightningSoundDone[1] {
		event.lightningSoundDone[1] = true
		s.playEventSound(soundEventCloudLightning)
	}
}

func (s *GameScene) startPalmRevengeSmoke(event *palmRevengeEvent) {
	if event == nil || event.smokeDone || event.target == nil || len(s.zeroPowerSmoke.frames) == 0 {
		return
	}
	event.smokeDone = true
	event.smokeAge = 0
}

func (s *GameScene) finishPalmRevenge(event *palmRevengeEvent) {
	if event == nil {
		return
	}
	if event.target != nil && event.target.power <= 0 {
		s.palmRevengeRemoval = event.target
	}
	if event.cloud != nil {
		event.cloud.revengeActive = false
		if event.cloud.sprite != nil && s.layers[layerClouds] != nil {
			s.layers[layerClouds].Add(event.cloud.sprite)
		}
	}
	s.palmRevenge = nil
}

func (s *GameScene) removeZeroPowerEffectsForTank(tank *battleTank) {
	if tank == nil || len(s.zeroPowerEffects) == 0 {
		return
	}
	active := s.zeroPowerEffects[:0]
	for _, effect := range s.zeroPowerEffects {
		if effect.tank != tank {
			active = append(active, effect)
		}
	}
	s.zeroPowerEffects = active
}

func (s *GameScene) updateSandFalls() {
	if len(s.sandFalls) == 0 {
		return
	}
	active := s.sandFalls[:0]
	for _, fall := range s.sandFalls {
		fall.age++
		if fall.age < fall.duration {
			active = append(active, fall)
		}
	}
	s.sandFalls = active
}

func (s *GameScene) updateZeroPowerEffects() {
	if len(s.zeroPowerEffects) == 0 {
		return
	}
	active := s.zeroPowerEffects[:0]
	for _, effect := range s.zeroPowerEffects {
		if effect.delay > 0 {
			effect.delay--
			active = append(active, effect)
			continue
		}
		if !effect.triggered {
			s.triggerZeroPowerWorldEffect(&effect)
			effect.triggered = true
		}
		s.updateZeroPowerTankDissolve(effect)
		effect.age++
		if effect.age < effect.duration {
			active = append(active, effect)
		}
	}
	s.zeroPowerEffects = active
}

func (s *GameScene) triggerZeroPowerWorldEffect(effect *zeroPowerAnimation) {
	if effect == nil || effect.tank == nil || effect.tank.body == nil {
		return
	}
	switch effect.kind {
	case zeroPowerEffectGrenadeImpact, zeroPowerEffectLargeGrenadeImpact, zeroPowerEffectAtomImpact:
		s.startZeroPowerImpactAtTank(effect.tank, effect.weapon)
	case zeroPowerEffectScatterProjectiles:
		s.fireZeroPowerScatterProjectiles(effect.tank)
	}
}

func (s *GameScene) updateCloudSearchEffects() {
	if len(s.cloudSearchEffects) == 0 {
		return
	}
	pause := secondsToFrames(0.25)
	active := s.cloudSearchEffects[:0]
	for _, effect := range s.cloudSearchEffects {
		if effect == nil || effect.cloud == nil || effect.cloud.sprite == nil {
			continue
		}
		if effect.delay > 0 {
			effect.delay--
			active = append(active, effect)
			continue
		}
		effect.age++
		if effect.age < effect.duration+pause {
			active = append(active, effect)
		}
	}
	s.cloudSearchEffects = active
}

func (s *GameScene) updateTurnAdvanceDelay() {
	if s.turnAdvanceDelay > 0 {
		s.updateCloudSearchEffects()
	}
	s.turnAdvanceDelay--
	if s.turnAdvanceDelay > 0 {
		if s.updateAirStrikeCamera() {
			return
		}
		if s.updateLaserCamera() {
			return
		}
		if s.updateShockwaveCamera() {
			return
		}
		if s.updateMoskitoCamera() {
			return
		}
		if s.updateWaterBlotchCamera() {
			return
		}
		if s.updateWaterCamera() {
			return
		}
		if s.updateCloudSearchCamera() {
			return
		}
		if s.updatePalmCamera() {
			return
		}
		if s.updateCrumblerCamera() {
			return
		}
		s.updateZeroPowerCamera()
		return
	}
	s.turnAdvanceDelay = 0
	s.removePendingPalmRevengeTank()
	if s.endRoundIfOnlyOneTankRemains() {
		return
	}
	s.advanceActivePlayer()
}

func (s *GameScene) removePendingPalmRevengeTank() {
	if s.palmRevengeRemoval == nil {
		return
	}
	tank := s.palmRevengeRemoval
	tank.zeroPowerGone = true
	s.removeTankSprites(tank)
	s.palmRevengeRemoval = nil
}

func (s *GameScene) updateBattleCamera() {
	if s.updateAirStrikeCamera() {
		return
	}
	if s.updateLaserCamera() {
		return
	}
	if s.updateShockwaveCamera() {
		return
	}
	if s.updateMoskitoCamera() {
		return
	}
	if s.updateWaterBlotchCamera() {
		return
	}
	if s.updateWaterCamera() {
		return
	}
	if s.updateCloudSearchCamera() {
		return
	}
	if s.updatePalmCamera() {
		return
	}
	if s.updateCrumblerCamera() {
		return
	}
	if s.updateZeroPowerCamera() {
		return
	}
	if s.scrollOMatActive() || s.scrollBarDragging {
		return
	}
	if s.projectilesActive() || s.activePlayerIndex < 0 {
		return
	}
	s.cameraGoal = s.cameraTargetForTank(s.activePlayerIndex)
	s.cameraGoalY = 0
	s.cameraX = approach(s.cameraX, s.cameraGoal, 0.08, 0.35)
	s.cameraY = approach(s.cameraY, s.cameraGoalY, 0.08, 0.35)
}

func (s *GameScene) focusPalmCamera(palm *battlePalm) {
	s.palmCameraFocus = palm
	s.updatePalmCamera()
}

func (s *GameScene) updatePalmCamera() bool {
	if s.palmCameraFocus == nil || s.palmCameraFocus.sprite == nil {
		return false
	}
	s.cameraGoal = s.cameraTargetForPalm(s.palmCameraFocus)
	s.cameraGoalY = 0
	s.cameraX = approach(s.cameraX, s.cameraGoal, 0.10, 0.45)
	s.cameraY = approach(s.cameraY, s.cameraGoalY, 0.10, 0.45)
	return true
}

func (s *GameScene) updateCloudSearchCamera() bool {
	effect := s.activeCloudSearchEffect()
	if effect == nil || effect.cloud == nil || effect.cloud.sprite == nil {
		return false
	}
	s.cameraGoal = s.cameraTargetForCloud(effect.cloud)
	s.cameraGoalY = s.cameraTargetForWorldY(effect.cloud.sprite.Bounds().Center().Y)
	s.cameraX = approach(s.cameraX, s.cameraGoal, 0.10, 0.45)
	s.cameraY = approach(s.cameraY, s.cameraGoalY, 0.10, 0.45)
	return true
}

func (s *GameScene) activeCloudSearchEffect() *cloudSearchEffect {
	for _, effect := range s.cloudSearchEffects {
		if effect != nil && effect.delay <= 0 && effect.cloud != nil && effect.cloud.sprite != nil {
			return effect
		}
	}
	return nil
}

func (s *GameScene) updateWaterBlotchCamera() bool {
	if s.waterBlotchFocus == nil || s.waterBlotchFocus.age >= s.waterBlotchFocus.duration {
		return false
	}
	screenWidth := core.Config().Screen.Width
	s.cameraGoal = math.Max(0, math.Min(s.worldWidth-screenWidth, s.waterBlotchFocus.pos.X-screenWidth/2))
	s.cameraGoalY = 0
	s.cameraX = approach(s.cameraX, s.cameraGoal, 0.10, 0.45)
	s.cameraY = approach(s.cameraY, s.cameraGoalY, 0.10, 0.45)
	return true
}

func (s *GameScene) cameraTargetForPalm(palm *battlePalm) float64 {
	if palm == nil || palm.sprite == nil {
		return s.cameraX
	}
	screenWidth := core.Config().Screen.Width
	centerX := palm.sprite.Bounds().Center().X
	return math.Max(0, math.Min(s.worldWidth-screenWidth, centerX-screenWidth/2))
}

func (s *GameScene) cameraTargetForCloud(cloud *battleCloud) float64 {
	if cloud == nil || cloud.sprite == nil {
		return s.cameraX
	}
	screenWidth := core.Config().Screen.Width
	centerX := cloud.sprite.Bounds().Center().X
	return math.Max(0, math.Min(s.worldWidth-screenWidth, centerX-screenWidth/2))
}

func (s *GameScene) updateWaterCamera() bool {
	if s.waterCameraFocus == nil {
		return false
	}
	if s.waterCameraFocus.age >= s.waterCameraFocus.duration {
		return false
	}
	pos := s.waterCameraFocus.position()
	screenWidth := core.Config().Screen.Width
	s.cameraGoal = math.Max(0, math.Min(s.worldWidth-screenWidth, pos.X-screenWidth/2))
	s.cameraGoalY = 0
	s.cameraX = approach(s.cameraX, s.cameraGoal, 0.10, 0.45)
	s.cameraY = approach(s.cameraY, s.cameraGoalY, 0.10, 0.45)
	return true
}

func (s *GameScene) updateAirStrikeCamera() bool {
	if len(s.airStrikeImpacts) == 0 || s.airStrikeImpacts[0] == nil {
		return false
	}
	pos := s.airStrikeFocus(s.airStrikeImpacts[0])
	screenWidth := core.Config().Screen.Width
	s.cameraGoal = math.Max(0, math.Min(s.worldWidth-screenWidth, pos.X-screenWidth/2))
	s.cameraGoalY = s.cameraTargetForWorldY(pos.Y)
	s.cameraX = approach(s.cameraX, s.cameraGoal, 0.14, 0.5)
	s.cameraY = approach(s.cameraY, s.cameraGoalY, 0.14, 0.5)
	return true
}

func (s *GameScene) airStrikeFocus(impact *airStrikeImpact) engine.Vec {
	if impact == nil {
		return engine.Vec{}
	}
	return impact.pos
}

func (s *GameScene) updateLaserCamera() bool {
	if len(s.laserEffects) == 0 || s.laserEffects[0] == nil {
		return false
	}
	pos := s.laserEffects[0].tip
	screenWidth := core.Config().Screen.Width
	s.cameraGoal = math.Max(0, math.Min(s.worldWidth-screenWidth, pos.X-screenWidth/2))
	s.cameraGoalY = s.cameraTargetForWorldY(pos.Y)
	s.cameraX = approach(s.cameraX, s.cameraGoal, 0.12, 0.55)
	s.cameraY = approach(s.cameraY, s.cameraGoalY, 0.12, 0.55)
	return true
}

func (s *GameScene) updateShockwaveCamera() bool {
	if len(s.shockwaveImpacts) == 0 || s.shockwaveImpacts[0] == nil {
		return false
	}
	impact := s.shockwaveImpacts[0]
	angle := float64(impact.age) * shockwaveCameraAngular
	screenWidth := core.Config().Screen.Width
	baseX := s.cameraTargetForWorldX(impact.pos.X)
	baseY := s.cameraTargetForWorldY(impact.pos.Y) - shockwaveCameraOrbit
	s.cameraGoal = math.Max(0, math.Min(s.worldWidth-screenWidth, baseX+math.Cos(angle)*shockwaveCameraOrbit))
	s.cameraGoalY = math.Max(-s.skyExtraHeight(), math.Min(0, baseY+math.Sin(angle)*shockwaveCameraOrbit))
	s.cameraX = approach(s.cameraX, s.cameraGoal, 0.82, 8.0)
	s.cameraY = approach(s.cameraY, s.cameraGoalY, 0.82, 8.0)
	return true
}

func (s *GameScene) updateMoskitoCamera() bool {
	var pos *engine.Vec
	if len(s.moskitoEffects) > 0 && s.moskitoEffects[0] != nil {
		pos = &s.moskitoEffects[0].pos
	} else if s.moskitoCameraFocus != nil && s.turnAdvanceDelay > 0 {
		pos = s.moskitoCameraFocus
	}
	if pos == nil {
		return false
	}
	screenWidth := core.Config().Screen.Width
	s.cameraGoal = math.Max(0, math.Min(s.worldWidth-screenWidth, pos.X-screenWidth/2))
	s.cameraGoalY = s.cameraTargetForWorldY(pos.Y)
	s.cameraX = approach(s.cameraX, s.cameraGoal, 0.10, 0.45)
	s.cameraY = approach(s.cameraY, s.cameraGoalY, 0.10, 0.45)
	return true
}

func (s *GameScene) updateCrumblerCamera() bool {
	if s.crumblerCameraFocus == nil {
		return false
	}
	if len(s.smallCrumblerImpacts) == 0 && len(s.sandFalls) == 0 && s.turnAdvanceDelay <= 0 {
		s.crumblerCameraFocus = nil
		return false
	}
	screenWidth := core.Config().Screen.Width
	s.cameraGoal = math.Max(0, math.Min(s.worldWidth-screenWidth, s.crumblerCameraFocus.X-screenWidth/2))
	s.cameraGoalY = s.cameraTargetForWorldY(s.crumblerCameraFocus.Y)
	s.cameraX = approach(s.cameraX, s.cameraGoal, 0.10, 0.45)
	s.cameraY = approach(s.cameraY, s.cameraGoalY, 0.10, 0.45)
	return true
}

func (s *GameScene) updateZeroPowerCamera() bool {
	if len(s.zeroPowerEffects) == 0 {
		return false
	}
	effect := s.zeroPowerEffects[0]
	for index, tank := range s.tanks {
		if tank == effect.tank {
			s.cameraGoal = s.cameraTargetForTank(index)
			s.cameraGoalY = 0
			s.cameraX = approach(s.cameraX, s.cameraGoal, 0.12, 0.4)
			s.cameraY = approach(s.cameraY, s.cameraGoalY, 0.12, 0.4)
			return true
		}
	}
	return false
}

func (s *GameScene) activeTank() *battleTank {
	if s.activePlayerIndex < 0 || s.activePlayerIndex >= len(s.tanks) {
		return nil
	}
	return s.tanks[s.activePlayerIndex]
}

func (s *GameScene) hudTank() *battleTank {
	if !s.allTanksLanded() && s.spawnIndex >= 0 && s.spawnIndex < len(s.tanks) {
		return s.tanks[s.spawnIndex]
	}
	return s.activeTank()
}

func (s *GameScene) maxShotStrength() int {
	tank := s.activeTank()
	if tank == nil {
		return 0
	}
	return maxInt(0, tank.power)
}

func (s *GameScene) minShotStrength() int {
	if s.maxShotStrength() < 0 {
		return 0
	}
	return 0
}

func (s *GameScene) clampActiveShotStrength() {
	tank := s.activeTank()
	if tank == nil {
		return
	}
	tank.shotStrength = minInt(tank.shotStrength, s.maxShotStrength())
	tank.shotStrength = maxInt(s.minShotStrength(), tank.shotStrength)
	if !s.canSelectWeaponSlot(tank, tank.selectedWeapon) {
		tank.selectedWeapon = 0
	}
}

func (s *GameScene) damageTank(tank *battleTank, damage int, attacker *battleTank, cause damageCause) {
	if tank == nil || tank.power <= 0 || damage <= 0 {
		return
	}
	previousPower := tank.power
	nominalDamage := damage
	s.awardDamageCredits(tank, attacker, nominalDamage)
	appliedDamage := s.applyEnergyShieldDamage(tank, nominalDamage)
	tank.power = maxInt(0, tank.power-appliedDamage)
	tank.shotStrength = minInt(tank.shotStrength, maxInt(0, tank.power))
	if previousPower > 0 && tank.power == 0 {
		s.awardZeroPowerScore(tank, attacker, cause)
		s.startZeroPowerAnimation(tank)
	} else if previousPower > tank.power {
		s.playEventSound(soundEventTankHit)
	}
}

func (s *GameScene) damageTankAsTerrain(tank *battleTank, damage int, attacker *battleTank, cause damageCause) {
	if tank == nil || tank.power <= 0 || damage <= 0 {
		return
	}
	previousPower := tank.power
	nominalDamage := damage
	s.awardDamageCredits(tank, attacker, nominalDamage)
	appliedDamage := s.applyEnergyShieldDamage(tank, nominalDamage)
	tank.power = maxInt(0, tank.power-appliedDamage)
	tank.shotStrength = minInt(tank.shotStrength, maxInt(0, tank.power))
	if previousPower > 0 && tank.power == 0 {
		s.awardZeroPowerScore(tank, attacker, cause)
		tank.zeroPowerShown = true
		tank.zeroPowerGone = false
		tank.terrainLocked = true
		s.removeZeroPowerEffectsForTank(tank)
	}
}

func (s *GameScene) awardDamageCredits(victim, attacker *battleTank, nominalDamage int) {
	if victim == nil || attacker == nil || victim == attacker || nominalDamage <= 0 {
		return
	}
	s.addCredits(victim.playerIndex, nominalDamage*core.Config().Gameplay.Scoring.DamageReceivedCreditMultiplier)
}

func (s *GameScene) awardZeroPowerScore(defeated, attacker *battleTank, cause damageCause) {
	if defeated == nil || attacker == nil {
		return
	}
	if defeated == attacker {
		s.applySuicidePenalty(defeated)
		return
	}
	s.rewardKill(attacker)
}

func (s *GameScene) rewardKill(attacker *battleTank) {
	if attacker == nil {
		return
	}
	scoring := core.Config().Gameplay.Scoring
	s.addScore(attacker.playerIndex, scoring.Kill.Points)
	s.addCredits(attacker.playerIndex, scoring.Kill.Credits)
}

func (s *GameScene) applySuicidePenalty(player *battleTank) {
	if player == nil {
		return
	}
	scoring := core.Config().Gameplay.Scoring
	s.addScore(player.playerIndex, -scoring.Suicide.PointsPenalty)
	s.addCredits(player.playerIndex, -scoring.Suicide.CreditsPenalty)
}

func (s *GameScene) addScore(playerIndex, points int) {
	if playerIndex < 0 {
		return
	}
	if len(s.scores) <= playerIndex {
		next := make([]int, playerIndex+1)
		copy(next, s.scores)
		s.scores = next
	}
	if len(s.roundScores) <= playerIndex {
		next := make([]int, playerIndex+1)
		copy(next, s.roundScores)
		s.roundScores = next
	}
	s.scores[playerIndex] += points
	s.roundScores[playerIndex] += points
	if playerIndex < len(s.tanks) && s.tanks[playerIndex] != nil {
		s.tanks[playerIndex].score = s.scores[playerIndex]
	}
}

func (s *GameScene) addCredits(playerIndex, credits int) {
	if playerIndex < 0 || credits == 0 {
		return
	}
	if len(s.credits) <= playerIndex {
		next := make([]int, playerIndex+1)
		copy(next, s.credits)
		s.credits = next
	}
	s.credits[playerIndex] = maxInt(0, s.credits[playerIndex]+credits)
}

func (s *GameScene) scoreForPlayer(playerIndex int) int {
	if playerIndex < 0 || playerIndex >= len(s.scores) {
		return 0
	}
	return s.scores[playerIndex]
}

func (s *GameScene) startZeroPowerAnimation(tank *battleTank) {
	if tank == nil || tank.zeroPowerShown {
		return
	}
	tank.zeroPowerShown = true
	tank.tint = color.RGBA{A: 255}
	models.RecolorTankBody(tank.body, tank.tint)
	models.RecolorCannon(tank.cannon, tank.tint)
	choices := s.zeroPowerChoices()
	if len(choices) == 0 {
		return
	}
	choice := choices[s.rng.Intn(len(choices))]
	s.playZeroPowerSound(zeroPowerSound(choice.name))
	delay := maxInt(0, s.zeroPowerStartDelay)
	effect := zeroPowerAnimation{
		tank:  tank,
		delay: delay,
	}
	if choice.kind == zeroPowerEffectSprite {
		animation := s.zeroPowerAnimations[choice.spriteIdx]
		effect.kind = zeroPowerEffectSprite
		effect.animation = animation
		effect.duration = maxInt(1, animation.totalTicks)
	} else {
		effect.kind = choice.kind
		effect.weapon = zeroPowerImpactWeapon(effect.kind)
		effect.duration = s.zeroPowerWorldEffectDuration(effect.kind, effect.weapon)
	}
	s.zeroPowerEffects = append(s.zeroPowerEffects, effect)
	duration := maxInt(1, effect.duration)
	if s.turnAdvanceDelay < duration+delay {
		s.turnAdvanceDelay = duration + delay
	}
}

func (s *GameScene) zeroPowerChoices() []zeroPowerChoice {
	all := s.allZeroPowerChoices()
	if !core.Config().Debug.Enabled || len(core.Config().Debug.ZeroPowerAnimations) == 0 {
		return all
	}
	allowed := make(map[string]bool)
	for _, name := range core.Config().Debug.ZeroPowerAnimations {
		allowed[canonicalZeroPowerName(name)] = true
	}
	choices := make([]zeroPowerChoice, 0, len(all))
	for _, choice := range all {
		if allowed[choice.name] {
			choices = append(choices, choice)
		}
	}
	if len(choices) == 0 {
		return all
	}
	return choices
}

func (s *GameScene) allZeroPowerChoices() []zeroPowerChoice {
	choices := make([]zeroPowerChoice, 0, len(s.zeroPowerAnimations)+int(zeroPowerEffectScatterProjectiles))
	for i := range s.zeroPowerAnimations {
		name := "sprite_" + strconv.Itoa(i)
		if i < len(s.zeroPowerNames) && s.zeroPowerNames[i] != "" {
			name = canonicalZeroPowerName(s.zeroPowerNames[i])
		}
		choices = append(choices, zeroPowerChoice{name: name, spriteIdx: i, kind: zeroPowerEffectSprite})
	}
	for kind := zeroPowerEffectGrenadeImpact; kind <= zeroPowerEffectScatterProjectiles; kind++ {
		choices = append(choices, zeroPowerChoice{name: zeroPowerEffectName(kind), kind: kind})
	}
	return choices
}

func zeroPowerEffectName(kind zeroPowerEffectKind) string {
	switch kind {
	case zeroPowerEffectGrenadeImpact:
		return "grenade_impact"
	case zeroPowerEffectLargeGrenadeImpact:
		return "large_grenade_impact"
	case zeroPowerEffectAtomImpact:
		return "atom_impact"
	case zeroPowerEffectScatterProjectiles:
		return "scatter_projectiles"
	default:
		return "sprite"
	}
}

func canonicalZeroPowerName(name string) string {
	value := strings.ToLower(strings.TrimSpace(name))
	value = strings.ReplaceAll(value, "-", "_")
	value = strings.ReplaceAll(value, " ", "_")
	switch value {
	case "zero_power_dust_explosion", "dust_explosion", "dust":
		return "dust"
	case "zero_power_explosion", "explosion":
		return "explosion"
	case "zero_power_mushroom_explosion", "mushroom_explosion", "mushroom", "pilz":
		return "mushroom"
	case "zero_power_player_smoke", "player_smoke", "smoke", "rauch":
		return "smoke"
	case "grenade", "grenade_impact", "granate":
		return "grenade_impact"
	case "large_grenade", "large_grenade_impact", "grosse_granate", "große_granate":
		return "large_grenade_impact"
	case "atom", "atom_bomb", "atom_impact", "atombombe":
		return "atom_impact"
	case "scatter", "scatter_projectiles", "three_projectiles", "drei_geschosse":
		return "scatter_projectiles"
	default:
		return value
	}
}

func zeroPowerImpactWeapon(kind zeroPowerEffectKind) weaponspkg.Weapon {
	switch kind {
	case zeroPowerEffectGrenadeImpact:
		return weaponspkg.Grenade()
	case zeroPowerEffectLargeGrenadeImpact:
		return weaponspkg.LargeGrenade()
	case zeroPowerEffectAtomImpact:
		return weaponspkg.AtomBomb()
	default:
		return weaponspkg.Grenade()
	}
}

func (s *GameScene) zeroPowerWorldEffectDuration(kind zeroPowerEffectKind, weapon weaponspkg.Weapon) int {
	switch kind {
	case zeroPowerEffectScatterProjectiles:
		return 420
	case zeroPowerEffectGrenadeImpact, zeroPowerEffectLargeGrenadeImpact, zeroPowerEffectAtomImpact:
		return s.impactAnimationFramesForWeapon(weapon) + s.impactPauseFrames()
	default:
		return zeroPowerFrames
	}
}

func (s *GameScene) updateZeroPowerTankDissolve(effect zeroPowerAnimation) {
	tank := effect.tank
	if tank == nil || tank.zeroPowerGone {
		return
	}
	dissolveStart := (effect.duration * 2) / 3
	if effect.age < dissolveStart {
		return
	}
	progress := float64(effect.age-dissolveStart+1) / float64(maxInt(1, zeroPowerDissolveFrames))
	if progress >= 1 {
		tank.zeroPowerGone = true
		tank.tint.A = 0
	} else {
		tank.tint.A = uint8(math.Round(255 * (1 - progress)))
	}
	models.RecolorTankBody(tank.body, tank.tint)
	models.RecolorCannon(tank.cannon, tank.tint)
}

func (s *GameScene) dropUnsupportedTanks() {
	for _, tank := range s.tanks {
		if tank == nil || tank.body == nil || !tank.landed || tank.falling || tank.terrainLocked {
			continue
		}
		targetY, stable := s.tankSupportState(tank.body)
		currentBottom := tank.body.Pos.Y + tank.body.Size.Y
		if stable {
			targetY = s.alignedTankBottomY(tank.body)
		}
		if targetY <= currentBottom+1 {
			startY := tank.body.Pos.Y
			s.alignTankBodyToSurface(tank.body)
			s.applyFallDamage(tank, tank.body.Pos.Y-startY)
			continue
		}

		tank.landed = false
		tank.falling = true
		tank.fallDamage = true
		tank.fallStartY = tank.body.Pos.Y
		tank.fallTargetY = targetY
		tank.body.Velocity = engine.Vec{}
		tank.body.Steps = engine.MakeBehaviors(
			s.behaviorFallOntoGround(s.ground, tank, false),
		)
	}
}

func (s *GameScene) alignedTankBottomY(tank *engine.Sprite) float64 {
	if tank == nil {
		return 0
	}
	if models.IsSmallTankBody(tank) {
		return s.uprightTankBottomY(tank)
	}
	copy := *tank
	pos := *tank.Pos
	copy.Pos = &pos
	if s.isXMV12TankBody(tank) {
		s.alignXMV12TankBodyToSurface(&copy)
	} else {
		s.ground.AlignSpriteToSurface(&copy)
	}
	return copy.Pos.Y + copy.Size.Y
}

func (s *GameScene) uprightTankBottomY(tank *engine.Sprite) float64 {
	if tank == nil {
		return 0
	}
	minX := int(math.Floor(tank.Pos.X))
	maxX := int(math.Ceil(tank.Pos.X + tank.Size.X))
	bottomY := s.ground.SurfaceY(tank.Pos.X + tank.Size.X/2)
	for x := minX; x <= maxX; x++ {
		bottomY = math.Max(bottomY, s.ground.SurfaceY(float64(x)))
	}
	return bottomY
}

func (s *GameScene) alignTankBodyToSurface(tank *engine.Sprite) {
	if models.IsSmallTankBody(tank) {
		s.ground.AlignSpriteUprightToSurface(tank)
		return
	}
	if s.isXMV12TankBody(tank) {
		s.alignXMV12TankBodyToSurface(tank)
		return
	}
	s.ground.AlignSpriteToSurface(tank)
}

func (s *GameScene) isXMV12TankBody(tank *engine.Sprite) bool {
	if tank == nil {
		return false
	}
	meta, ok := tank.Meta.(*models.TankBodyMeta)
	return ok && meta.Kind == models.TankBodyKindXMV12
}

func (s *GameScene) alignXMV12TankBodyToSurface(tank *engine.Sprite) {
	if tank == nil || tank.Pos == nil || tank.Size == nil || s.ground.Size == nil {
		return
	}
	const treadContact = 0.34

	centerX := tank.Pos.X + tank.Size.X/2
	leftX := centerX - tank.Size.X*treadContact
	rightX := centerX + tank.Size.X*treadContact
	leftY := s.ground.SurfaceY(leftX)
	rightY := s.ground.SurfaceY(rightX)

	tank.Rot = math.Atan2(rightY-leftY, rightX-leftX)
	centerY := (leftY+rightY)/2 - math.Cos(tank.Rot)*tank.Size.Y/2
	tank.Pos = &engine.Vec{
		X: tank.Pos.X,
		Y: centerY - tank.Size.Y/2,
	}
}

func (s *GameScene) tankSupportState(tank *engine.Sprite) (float64, bool) {
	if tank == nil {
		return 0, true
	}
	const (
		samples           = 9
		requiredSupport   = 4
		footprintCoverage = 0.76
		supportTolerance  = 2
	)

	centerX := tank.Pos.X + tank.Size.X/2
	leftX := centerX - tank.Size.X*footprintCoverage/2
	step := tank.Size.X * footprintCoverage / float64(samples-1)
	currentBottom := tank.Pos.Y + tank.Size.Y
	targetY := currentBottom
	supported := 0
	centerSupported := false

	for i := 0; i < samples; i++ {
		x := leftX + float64(i)*step
		surfaceY := s.ground.SurfaceY(x)
		if surfaceY > targetY {
			targetY = surfaceY
		}
		if surfaceY <= currentBottom+supportTolerance {
			supported++
			if i == samples/2 {
				centerSupported = true
			}
		}
	}

	return targetY, centerSupported && supported >= requiredSupport
}

func (s *GameScene) darkenTank(tank *battleTank, amount float64) {
	if tank == nil {
		return
	}
	factor := math.Max(0, 1-amount)
	tank.tint = color.RGBA{
		R: uint8(float64(tank.tint.R) * factor),
		G: uint8(float64(tank.tint.G) * factor),
		B: uint8(float64(tank.tint.B) * factor),
		A: tank.tint.A,
	}
	models.RecolorTankBody(tank.body, tank.tint)
	models.RecolorCannon(tank.cannon, tank.tint)
}

func (s *GameScene) drawProjectile(screen *ebiten.Image, camera *ebiten.GeoM) {
	if !s.projectilesActive() {
		return
	}

	for _, p := range s.projectiles {
		if p == nil {
			continue
		}
		s.drawSingleProjectile(screen, camera, p)
	}
}

func (s *GameScene) drawSingleProjectile(screen *ebiten.Image, camera *ebiten.GeoM, p *projectile) {
	weapon := s.weaponForProjectile(p)
	if weapon.Mosquitos && p.mosquitoPreview && len(s.moskitosAnimation.frames) > 0 {
		drawAnimationCenteredLooping(screen, camera, s.moskitosAnimation, p.pos, int(s.time))
		return
	}
	c := weapon.Color
	if p.zeroPowerScatter {
		s.drawScatterProjectileTrail(screen, camera, p)
	}
	if weapon.ShowTrail {
		s.drawProjectileTail(screen, camera, p)
	}
	projected := p.pos.Project(camera)
	radius := projectileRadiusForWeapon(weapon)
	if weapon.RoundProjectile {
		vector.DrawFilledCircle(screen, float32(projected.X), float32(projected.Y), float32(radius), c, true)
		return
	}
	drawFilledRect(screen, image.Rect(int(projected.X-radius), int(projected.Y-radius), int(projected.X+radius), int(projected.Y+radius)), c)
}

func (s *GameScene) drawScatterProjectileTrail(screen *ebiten.Image, camera *ebiten.GeoM, p *projectile) {
	if p == nil || len(p.trail) < 2 {
		return
	}
	c := p.scatterImpactColor
	if c.A == 0 {
		c = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	}
	for i := 1; i < len(p.trail); i++ {
		from := p.trail[i-1].Project(camera)
		to := p.trail[i].Project(camera)
		vector.StrokeLine(screen, float32(from.X), float32(from.Y), float32(to.X), float32(to.Y), 1, c, true)
	}
}

func (s *GameScene) drawProjectileTail(screen *ebiten.Image, camera *ebiten.GeoM, p *projectile) {
	if p == nil || len(p.trail) < 2 {
		return
	}
	const (
		redAtDistance  = 70.0
		transparentAt  = 250.0
		tailLineWidth  = 1.0
		tailStartAlpha = 200.0
		tailMidAlpha   = 175.0
	)

	distanceFromProjectile := 0.0
	for i := len(p.trail) - 1; i > 0 && distanceFromProjectile < transparentAt; i-- {
		current := p.trail[i]
		previous := p.trail[i-1]
		segment := current.Sub(previous)
		segmentLength := segment.Len()
		if segmentLength <= 0 {
			continue
		}

		remaining := transparentAt - distanceFromProjectile
		if segmentLength > remaining {
			keep := remaining / segmentLength
			previous = engine.V(
				current.X-(current.X-previous.X)*keep,
				current.Y-(current.Y-previous.Y)*keep,
			)
			segmentLength = remaining
		}

		midDistance := distanceFromProjectile + segmentLength/2
		redProgress := math.Min(1, midDistance/redAtDistance)
		alpha := tailStartAlpha + (tailMidAlpha-tailStartAlpha)*math.Min(1, midDistance/redAtDistance)
		if midDistance > redAtDistance {
			alpha = tailMidAlpha * math.Max(0, 1-(midDistance-redAtDistance)/(transparentAt-redAtDistance))
		}
		if alpha <= 0 {
			break
		}

		c := color.RGBA{
			R: 255,
			G: uint8(math.Round(255 * (1 - redProgress))),
			B: uint8(math.Round(255 * (1 - redProgress))),
			A: uint8(math.Round(alpha)),
		}
		from := previous.Project(camera)
		to := current.Project(camera)
		vector.StrokeLine(screen, float32(from.X), float32(from.Y), float32(to.X), float32(to.Y), tailLineWidth, c, true)
		distanceFromProjectile += segmentLength
	}
}

func (s *GameScene) drawProjectileReentryAnimation(screen *ebiten.Image) {
	anim := s.reentryAnimation
	if anim == nil || anim.duration <= 0 || s.reentryEarth == nil {
		return
	}

	screen.Fill(colornames.Black)
	screenBounds := screen.Bounds()
	screenW := float64(screenBounds.Dx())
	screenH := float64(screenBounds.Dy())
	center := engine.V(screenW/2, screenH/2)
	earthSize := math.Min(screenW, screenH) * 0.42
	earthRect := image.Rect(
		int(center.X-earthSize/2),
		int(center.Y-earthSize/2),
		int(center.X+earthSize/2),
		int(center.Y+earthSize/2),
	)

	progress := float64(anim.age) / math.Max(1, float64(anim.duration-1))
	progress = math.Max(0, math.Min(1, progress))
	x, y, projectileBehindEarth := reentryProjectilePosition(center, earthSize, anim.direction, progress)
	if progress >= 0.95 {
		drawScaledImage(screen, s.reentryEarth, earthRect)
		s.drawReentryTank(screen, anim, center, earthSize)
		return
	}
	if projectileBehindEarth {
		s.drawReentryProjectile(screen, anim, x, y, progress, earthSize)
	}
	drawScaledImage(screen, s.reentryEarth, earthRect)
	s.drawReentryTank(screen, anim, center, earthSize)
	if !projectileBehindEarth {
		s.drawReentryProjectile(screen, anim, x, y, progress, earthSize)
	}
}

func reentryProjectilePosition(center engine.Vec, earthSize float64, direction int, progress float64) (float64, float64, bool) {
	if direction == 0 {
		direction = 1
	}
	radiusX := earthSize * 0.58
	startSide := -1.0
	if direction < 0 {
		startSide = 1
	}
	endSide := -startSide

	if progress < 0.08 {
		t := progress / 0.08
		t = t * t * (3 - 2*t)
		return center.X + startSide*radiusX*t, center.Y, false
	}

	if progress < 0.58 {
		t := (progress - 0.08) / 0.50
		return center.X + (startSide+(endSide-startSide)*t)*radiusX, center.Y, true
	}

	t := (progress - 0.58) / 0.37
	t = math.Max(0, math.Min(1, t))
	t = t * t * (3 - 2*t)
	stopOffset := endSide * radiusX * 0.16
	return center.X + endSide*radiusX + (stopOffset-endSide*radiusX)*t, center.Y, false
}

func (s *GameScene) drawReentryProjectile(screen *ebiten.Image, anim *projectileReentryAnimation, x, y, progress, earthSize float64) {
	if anim == nil || anim.projectile == nil {
		return
	}
	weapon := s.weaponForProjectile(anim.projectile)
	radius := projectileRadiusForWeapon(weapon) * math.Max(1.1, earthSize/230)
	alpha := uint8(255)
	if progress > 0.18 && progress < 0.82 {
		alpha = 185
	}
	c := color.RGBA{R: weapon.Color.R, G: weapon.Color.G, B: weapon.Color.B, A: alpha}
	if weapon.RoundProjectile {
		vector.DrawFilledCircle(screen, float32(x), float32(y), float32(radius), c, true)
		return
	}
	drawFilledRect(screen, image.Rect(int(x-radius), int(y-radius), int(x+radius), int(y+radius)), c)
}

func (s *GameScene) drawReentryTank(screen *ebiten.Image, anim *projectileReentryAnimation, center engine.Vec, earthSize float64) {
	if anim == nil || anim.shooter == nil || anim.shooter.body == nil {
		return
	}
	body := anim.shooter.body
	scale := 1.0
	bodyW := body.Size.X * scale
	bodyH := body.Size.Y * scale
	bodyPos := engine.V(center.X-bodyW/2, center.Y-bodyH/2)
	drawScaledImage(screen, body.Image, image.Rect(int(bodyPos.X), int(bodyPos.Y), int(bodyPos.X+bodyW), int(bodyPos.Y+bodyH)))

	if anim.shooter.cannon == nil || anim.shooter.cannon.Drawable == nil {
		return
	}
	cannon := *anim.shooter.cannon
	mount := engine.V(body.Size.X, body.Size.Y+20)
	anchor := engine.V(cannon.Size.X, cannon.Size.Y)
	if cannon.RotAnchor != nil {
		mount = models.TankCannonMount(body)
		anchor = *cannon.RotAnchor
	}
	cannon.Pos = &engine.Vec{
		X: bodyPos.X + mount.X*scale - anchor.X*scale,
		Y: bodyPos.Y + mount.Y*scale - anchor.Y*scale,
	}
	cannon.Size = &engine.Vec{X: anim.shooter.cannon.Size.X * scale, Y: anim.shooter.cannon.Size.Y * scale}
	cannon.Rot = anim.cannonRot
	cannon.Draw(nil, screen)
}

func (s *GameScene) drawImpacts(screen *ebiten.Image, camera *ebiten.GeoM) {
	for _, impact := range s.impacts {
		projected := impact.pos.Project(camera)
		progress := float64(impact.age) / math.Max(1, float64(impact.duration))
		if impact.style == weaponspkg.ImpactAnimationPlasma {
			drawPlasmaImpact(screen, projected, impact.radius, progress)
			continue
		}
		if impact.style == weaponspkg.ImpactAnimationHBomb {
			drawHBombImpact(screen, projected, impact.radius, progress)
			continue
		}
		cycles := impact.cycles
		if cycles <= 0 {
			cycles = 2
		}
		if impact.outward {
			fillProgress := math.Max(0, math.Min(1, progress))
			vector.DrawFilledCircle(screen, float32(projected.X), float32(projected.Y), float32(impact.radius), color.RGBA{R: 255, G: 0, B: 0, A: 220}, true)
			steps := 8
			for i := steps; i >= 1; i-- {
				t := float64(i) / float64(steps)
				radius := float32(impact.radius * fillProgress * t)
				alpha := uint8(255 * math.Pow(t, 0.7))
				vector.DrawFilledCircle(screen, float32(projected.X), float32(projected.Y), radius, color.RGBA{R: 0, G: 0, B: 0, A: alpha}, true)
			}
			continue
		}
		impactColor := impact.color
		if impactColor.A != 0 {
			vector.DrawFilledCircle(screen, float32(projected.X), float32(projected.Y), float32(impact.radius), impactColor, true)
			continue
		}
		cycleProgress := math.Mod(progress*float64(cycles), 1)
		steps := 8
		if impactColor.A == 0 {
			impactColor = color.RGBA{R: 255, A: 220}
		}
		for i := steps; i >= 1; i-- {
			t := float64(i) / float64(steps)
			radius := float32(impact.radius * t)
			intensity := math.Pow(t, 0.7) * cycleProgress
			vector.DrawFilledCircle(screen, float32(projected.X), float32(projected.Y), radius, color.RGBA{
				R: uint8(float64(impactColor.R) * intensity),
				G: uint8(float64(impactColor.G) * intensity),
				B: uint8(float64(impactColor.B) * intensity),
				A: impactColor.A,
			}, true)
		}
	}
}

func drawPlasmaImpact(screen *ebiten.Image, center engine.Vec, radius, progress float64) {
	progress = math.Max(0, math.Min(1, progress))
	black := color.RGBA{R: 0, G: 0, B: 0, A: 245}
	green := color.RGBA{R: 148, G: 255, B: 78, A: 240}

	switch {
	case progress < plasmaBuildProgress:
		t := easeOut(progress / plasmaBuildProgress)
		visibleRadius := radius * t
		drawPlasmaRings(screen, center, 0, visibleRadius, black)
	case progress < plasmaGreenProgress:
		drawPlasmaRings(screen, center, 0, radius, black)
		t := easeOut((progress - plasmaBuildProgress) / (plasmaGreenProgress - plasmaBuildProgress))
		vector.StrokeCircle(screen, float32(center.X), float32(center.Y), float32(radius*t), 5, green, true)
	default:
		t := easeIn((progress - plasmaGreenProgress) / (1 - plasmaGreenProgress))
		drawPlasmaRings(screen, center, radius*t, radius, black)
	}
}

func drawPlasmaRings(screen *ebiten.Image, center engine.Vec, minRadius, maxRadius float64, c color.RGBA) {
	if maxRadius <= 0 {
		return
	}
	if minRadius < 0 {
		minRadius = 0
	}
	for r := math.Max(plasmaRingSpacing, math.Ceil(minRadius/plasmaRingSpacing)*plasmaRingSpacing); r <= maxRadius; r += plasmaRingSpacing {
		vector.StrokeCircle(screen, float32(center.X), float32(center.Y), float32(r), plasmaRingThickness, c, true)
	}
}

func drawHBombImpact(screen *ebiten.Image, center engine.Vec, radius, progress float64) {
	progress = math.Max(0, math.Min(1, progress))
	red := color.RGBA{R: 255, G: 0, B: 0, A: 225}
	black := color.RGBA{R: 0, G: 0, B: 0, A: 245}
	deepPurple := color.RGBA{R: 17, G: 0, B: 34, A: 230}

	switch {
	case progress < 0.12:
		vector.DrawFilledCircle(screen, float32(center.X), float32(center.Y), float32(radius), red, true)
	case progress < 0.34:
		t := easeOut((progress - 0.12) / 0.22)
		drawRadialFill(screen, center, radius, red, black, t)
	case progress < 0.68:
		t := 1 - easeIn((progress-0.34)/0.34)
		drawRadialFill(screen, center, radius, red, black, t)
	default:
		t := easeInOut((progress - 0.68) / 0.32)
		drawPurpleDissolve(screen, center, radius, deepPurple, t)
	}
}

func drawRadialFill(screen *ebiten.Image, center engine.Vec, radius float64, base, fill color.RGBA, fillProgress float64) {
	fillProgress = math.Max(0, math.Min(1, fillProgress))
	vector.DrawFilledCircle(screen, float32(center.X), float32(center.Y), float32(radius), base, true)
	steps := 10
	for i := steps; i >= 1; i-- {
		t := float64(i) / float64(steps)
		r := radius * fillProgress * t
		alpha := uint8(float64(fill.A) * math.Pow(t, 0.6))
		vector.DrawFilledCircle(screen, float32(center.X), float32(center.Y), float32(r), color.RGBA{R: fill.R, G: fill.G, B: fill.B, A: alpha}, true)
	}
}

func drawPurpleDissolve(screen *ebiten.Image, center engine.Vec, radius float64, c color.RGBA, progress float64) {
	progress = math.Max(0, math.Min(1, progress))
	steps := 12
	start := radius * progress
	for i := steps; i >= 1; i-- {
		t := float64(i) / float64(steps)
		r := start + (radius-start)*t
		if r <= 0 {
			continue
		}
		alpha := uint8(float64(c.A) * (1 - progress) * math.Pow(t, 0.45))
		vector.DrawFilledCircle(screen, float32(center.X), float32(center.Y), float32(r), color.RGBA{R: c.R, G: c.G, B: c.B, A: alpha}, true)
	}
}

func easeIn(t float64) float64 {
	t = math.Max(0, math.Min(1, t))
	return t * t
}

func easeOut(t float64) float64 {
	t = math.Max(0, math.Min(1, t))
	return 1 - (1-t)*(1-t)
}

func easeInOut(t float64) float64 {
	t = math.Max(0, math.Min(1, t))
	if t < 0.5 {
		return 2 * t * t
	}
	return 1 - math.Pow(-2*t+2, 2)/2
}

func (s *GameScene) drawAnimatedImpacts(screen *ebiten.Image, camera *ebiten.GeoM) {
	for _, impact := range s.animatedImpacts {
		drawAnimationBottomCentered(screen, camera, impact.animation, impact.pos, impact.age)
	}
}

func (s *GameScene) drawWaterBlubbers(screen *ebiten.Image, camera *ebiten.GeoM) {
	for _, effect := range s.waterBlubbers {
		if effect == nil {
			continue
		}
		drawAnimationCenteredLooping(screen, camera, effect.animation, effect.position(), effect.age)
	}
}

func (s *GameScene) drawWaterBlotches(screen *ebiten.Image, camera *ebiten.GeoM) {
	for _, effect := range s.waterBlotches {
		if effect == nil {
			continue
		}
		drawAnimationBottomCentered(screen, camera, effect.animation, effect.pos, effect.age)
	}
}

func (s *GameScene) drawMoskitoEffects(screen *ebiten.Image, camera *ebiten.GeoM) {
	for _, effect := range s.moskitoEffects {
		if effect == nil {
			continue
		}
		if len(s.moskitosAnimation.frames) > 0 {
			drawAnimationCenteredLooping(screen, camera, s.moskitosAnimation, effect.pos, int(s.time))
		}
		if effect.phase == moskitoPhaseQuestion && len(s.questionAnimation.frames) > 0 {
			drawAnimationCentered(screen, camera, s.questionAnimation, *effect.pos.Add(engine.V(4, -28)), effect.age)
		}
	}
}

func (s *GameScene) drawShockwaveImpacts(screen *ebiten.Image, camera *ebiten.GeoM) {
	yellow := color.RGBA{R: 255, G: 235, B: 0, A: 245}
	black := color.RGBA{R: 0, G: 0, B: 0, A: 245}
	for _, impact := range s.shockwaveImpacts {
		if impact == nil || impact.radius <= 0 || shockwavePulseFrames <= 0 {
			continue
		}
		pulseAge := impact.age % shockwavePulseFrames
		progress := easeOut(float64(pulseAge) / float64(shockwavePulseFrames-1))
		radius := impact.radius * progress
		if radius <= 0 {
			continue
		}
		projected := impact.pos.Project(camera)
		vector.StrokeCircle(screen, float32(projected.X), float32(projected.Y), float32(radius), float32(shockwaveLineWidth), yellow, true)
		innerRadius := math.Max(0, radius-shockwaveLineWidth/2)
		vector.StrokeCircle(screen, float32(projected.X), float32(projected.Y), float32(innerRadius), float32(shockwaveInnerLineWidth), black, true)
	}
}

func (s *GameScene) drawAirStrikeImpacts(screen *ebiten.Image, camera *ebiten.GeoM) {
	for _, impact := range s.airStrikeImpacts {
		if impact == nil {
			continue
		}
		if !impact.bojeHidden && len(s.blinkBojeAnimation.frames) > 0 {
			drawAnimationBottomCenteredLooping(screen, camera, s.blinkBojeAnimation, impact.pos, impact.age)
		}
		for i := range impact.bombs {
			bomb := &impact.bombs[i]
			if bomb.impacted || impact.age < bomb.delay || s.bulletBombImage == nil {
				continue
			}
			s.drawAirStrikeBomb(screen, camera, bomb.pos)
		}
	}
}

func (s *GameScene) drawAirStrikeBomb(screen *ebiten.Image, camera *ebiten.GeoM, pos engine.Vec) {
	projected := pos.Project(camera)
	bounds := s.bulletBombImage.Bounds()
	w := float64(bounds.Dx())
	h := float64(bounds.Dy())
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-w/2, -h/2)
	op.GeoM.Rotate(airStrikeBombAngle)
	op.GeoM.Translate(projected.X, projected.Y)
	screen.DrawImage(s.bulletBombImage, op)
}

func (s *GameScene) drawLaserEffects(screen *ebiten.Image, camera *ebiten.GeoM) {
	laserColor := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	for _, effect := range s.laserEffects {
		if effect == nil {
			continue
		}
		start := effect.start.Project(camera)
		end := effect.end.Project(camera)
		vector.StrokeLine(screen, float32(start.X), float32(start.Y), float32(end.X), float32(end.Y), float32(laserLineThickness), laserColor, true)
		for _, smoke := range effect.smokes {
			if smoke.age >= s.laserSmokeAnimation.totalTicks {
				continue
			}
			drawAnimationCentered(screen, camera, s.laserSmokeAnimation, smoke.pos, smoke.age)
		}
	}
}

func (s *GameScene) drawCloudSearchEffects(screen *ebiten.Image, camera *ebiten.GeoM) {
	for _, effect := range s.cloudSearchEffects {
		if effect == nil || effect.delay > 0 || effect.age >= effect.duration || effect.cloud == nil || effect.cloud.sprite == nil {
			continue
		}
		frame := effect.animation.frameAt(effect.age)
		if frame == nil {
			continue
		}
		center := effect.cloud.sprite.Bounds().Center().Project(camera)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(
			center.X-float64(frame.Bounds().Dx())/2,
			center.Y-float64(frame.Bounds().Dy())/2,
		)
		screen.DrawImage(frame, op)
	}
}

func (s *GameScene) drawMoleImpacts(screen *ebiten.Image, camera *ebiten.GeoM) {
	for _, impact := range s.moleImpacts {
		if impact == nil {
			continue
		}
		switch impact.phase {
		case molePhaseTunnel:
			drawMoleArrow(screen, camera, impact.pos)
		}
	}
}

func drawMoleArrow(screen *ebiten.Image, camera *ebiten.GeoM, pos engine.Vec) {
	projected := pos.Project(camera)
	width := float32(19)
	height := float32(12)
	x := float32(projected.X)
	y := float32(projected.Y)
	vertices := []ebiten.Vertex{
		{DstX: x, DstY: y + height*0.55, SrcX: 0, SrcY: 0, ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1},
		{DstX: x - width/2, DstY: y - height*0.45, SrcX: 0, SrcY: 0, ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1},
		{DstX: x + width/2, DstY: y - height*0.45, SrcX: 0, SrcY: 0, ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1},
	}
	screen.DrawTriangles(vertices, []uint16{0, 1, 2}, getSolidWhiteImage(), nil)
}

func (s *GameScene) drawSmallCrumblerImpacts(screen *ebiten.Image, camera *ebiten.GeoM) {
	crumbColor := color.RGBA{R: 226, G: 185, B: 103, A: 255}
	for _, impact := range s.smallCrumblerImpacts {
		if impact == nil {
			continue
		}
		for i := range impact.crumbs {
			crumb := &impact.crumbs[i]
			if crumb.pos.Y < impact.start.Y {
				continue
			}
			projected := crumb.pos.Project(camera)
			drawFilledRect(
				screen,
				image.Rect(int(projected.X), int(projected.Y), int(projected.X)+impact.cfg.width, int(projected.Y)+impact.cfg.height),
				crumbColor,
			)
		}
	}
}

func (s *GameScene) drawWaterFills(screen *ebiten.Image, camera *ebiten.GeoM) {
	if len(s.waterFills) == 0 || len(s.waterAnimation.pixels) == 0 {
		return
	}
	frameIndex := (int(s.time) / 8) % len(s.waterAnimation.pixels)
	screenWidth := int(core.Config().Screen.Width)
	visibleLeft := maxInt(0, int(math.Floor(s.cameraX))-1)
	visibleRight := int(math.Ceil(s.cameraX)) + screenWidth + 1

	for i := range s.waterFills {
		fill := &s.waterFills[i]
		if fill.rightX < visibleLeft || fill.leftX > visibleRight {
			continue
		}
		img := s.waterFillImage(fill, frameIndex)
		if img == nil {
			continue
		}
		projected := engine.V(float64(fill.leftX), math.Floor(fill.topY)).Project(camera)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(math.Round(projected.X), math.Round(projected.Y))
		screen.DrawImage(img, op)
	}
}

func (s *GameScene) waterFillImage(fill *waterFill, frameIndex int) *ebiten.Image {
	if fill == nil || len(fill.surfaceY) == 0 || len(s.waterAnimation.pixels) == 0 {
		return nil
	}
	if fill.image != nil && fill.cacheAge == fill.age && fill.cacheFrame == frameIndex {
		return fill.image
	}
	frame := s.waterAnimation.pixels[frameIndex%len(s.waterAnimation.pixels)]
	if frame == nil {
		return nil
	}
	frameBounds := frame.Bounds()
	frameW := frameBounds.Dx()
	frameH := frameBounds.Dy()
	if frameW <= 0 || frameH <= 0 {
		return nil
	}

	y0 := int(math.Floor(fill.topY))
	y1 := int(math.Ceil(deepestWaterBottom(fill)))
	width := fill.rightX - fill.leftX + 1
	height := y1 - y0
	if width <= 0 || height <= 0 {
		return nil
	}

	base := color.RGBA{R: 6, G: 48, B: 164, A: 255}
	pixels := make([]byte, width*height*4)
	top := int(math.Floor(s.currentWaterTop(fill)))
	for x := 0; x < width; x++ {
		bottom := int(math.Ceil(fill.surfaceY[x]))
		if bottom <= top {
			continue
		}
		for y := top; y < bottom; y++ {
			if y < y0 || y >= y1 {
				continue
			}
			textureX := positiveMod(fill.leftX+x, frameW)
			textureY := positiveMod(y, frameH)
			c := frame.RGBAAt(frameBounds.Min.X+textureX, frameBounds.Min.Y+textureY)
			if c.A == 0 {
				c = base
			} else {
				c.A = 255
			}
			offset := ((y-y0)*width + x) * 4
			pixels[offset] = c.R
			pixels[offset+1] = c.G
			pixels[offset+2] = c.B
			pixels[offset+3] = c.A
		}
	}

	if fill.image == nil || fill.image.Bounds().Dx() != width || fill.image.Bounds().Dy() != height {
		fill.image = ebiten.NewImage(width, height)
	}
	fill.image.WritePixels(pixels)
	fill.cacheAge = fill.age
	fill.cacheFrame = frameIndex
	return fill.image
}

func deepestWaterBottom(fill *waterFill) float64 {
	if fill == nil || len(fill.surfaceY) == 0 {
		return 0
	}
	deepest := fill.topY
	for _, y := range fill.surfaceY {
		if y > deepest {
			deepest = y
		}
	}
	return deepest
}

func (s *GameScene) drawPalmEffects(screen *ebiten.Image, camera *ebiten.GeoM) {
	for _, palm := range s.palms {
		if palm == nil || palm.sprite == nil {
			continue
		}
		bounds := palm.sprite.Bounds()
		if palm.screaming && s.palmScreamImage != nil {
			pos := engine.V(bounds.Center().X, bounds.Min.Y+bounds.H()*0.24)
			projected := pos.Project(camera)
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(
				projected.X-float64(s.palmScreamImage.Bounds().Dx())/2,
				projected.Y-float64(s.palmScreamImage.Bounds().Dy())/2,
			)
			screen.DrawImage(s.palmScreamImage, op)
		}
		if palm.grinning && s.palmGrinImage != nil {
			pos := engine.V(bounds.Center().X, bounds.Min.Y+bounds.H()*0.24)
			projected := pos.Project(camera)
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(
				projected.X-float64(s.palmGrinImage.Bounds().Dx())/2,
				projected.Y-float64(s.palmGrinImage.Bounds().Dy())/2,
			)
			screen.DrawImage(s.palmGrinImage, op)
		}
		if palm.eyesOn {
			frame := s.palmEyesFrame(palm.eyeAge)
			if frame != nil {
				pos := engine.V(bounds.Center().X, bounds.Min.Y+bounds.H()*0.24)
				projected := pos.Project(camera)
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Translate(
					projected.X-float64(frame.Bounds().Dx())/2,
					projected.Y-float64(frame.Bounds().Dy())/2,
				)
				screen.DrawImage(frame, op)
			}
		}
		switch palm.state {
		case palmStateBurning:
			pos := engine.V(bounds.Center().X, bounds.Min.Y+bounds.H()*0.24)
			drawAnimationCentered(screen, camera, s.palmFireAnimation, pos, palm.age)
		case palmStateSkeletonSmoking:
			pos := engine.V(bounds.Center().X, bounds.Min.Y+bounds.H()*0.24)
			drawAnimationCentered(screen, camera, s.palmSmokeAnimation, pos, palm.age)
		case palmStateCrumbling:
			pos := engine.V(bounds.Center().X, bounds.Max.Y)
			drawAnimationBottomCentered(screen, camera, s.palmCrumbleAnimation, pos, palm.age)
		}
	}
}

func (s *GameScene) drawPalmRevenge(screen *ebiten.Image, camera *ebiten.GeoM) {
	event := s.palmRevenge
	if event == nil || event.cloud == nil || event.cloud.sprite == nil || event.cloud.sprite.Pos == nil || event.cloud.sprite.Size == nil {
		return
	}
	cloudImg := event.cloud.image
	if cloudImg == nil {
		cloudImg = s.cloudImageForKind(event.cloud.kind)
	}
	if cloudImg == nil {
		return
	}
	bounds := event.cloud.sprite.Bounds()
	projected := engine.V(bounds.Min.X, bounds.Min.Y).Project(camera)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(bounds.W()/float64(cloudImg.Bounds().Dx()), bounds.H()/float64(cloudImg.Bounds().Dy()))
	op.GeoM.Translate(projected.X, projected.Y)
	brightness := s.palmRevengeCloudBrightness(event)
	op.ColorScale.Scale(brightness, brightness, brightness, 1)
	screen.DrawImage(cloudImg, op)

	face := s.palmRevengeCloudFace(event)
	if face != nil {
		center := bounds.Center().Project(camera)
		faceOp := &ebiten.DrawImageOptions{}
		faceOp.GeoM.Translate(center.X-float64(face.Bounds().Dx())/2, center.Y-float64(face.Bounds().Dy())/2)
		screen.DrawImage(face, faceOp)
	}

	if s.palmRevengeLightningVisible(event) && event.target != nil && event.target.body != nil && s.lightningImage != nil {
		cloudCenter := bounds.Center().Project(camera)
		tankBounds := event.target.body.Bounds()
		tankTop := engine.V(tankBounds.Center().X, tankBounds.Min.Y).Project(camera)
		lightningOp := &ebiten.DrawImageOptions{}
		scaleY := math.Max(0.2, (tankTop.Y-cloudCenter.Y)/float64(s.lightningImage.Bounds().Dy()))
		lightningOp.GeoM.Scale(1, scaleY)
		lightningOp.GeoM.Translate(cloudCenter.X-float64(s.lightningImage.Bounds().Dx())/2, cloudCenter.Y)
		screen.DrawImage(s.lightningImage, lightningOp)
	}

	if event.smokeDone && event.target != nil && event.target.body != nil {
		body := event.target.body.Bounds()
		drawAnimationCenteredLooping(screen, camera, s.zeroPowerSmoke, engine.V(body.Min.X+body.W()/2, body.Min.Y-body.H()+12), event.smokeAge)
	}
}

func (s *GameScene) palmRevengeCloudBrightness(event *palmRevengeEvent) float32 {
	if event == nil {
		return 1
	}
	minBrightness := 0.38
	switch event.phase {
	case palmRevengeCloudFocus:
		progress := math.Min(1, float64(event.age)/float64(palmRevengeFocusFrames))
		return float32(1 - (1-minBrightness)*progress)
	case palmRevengeCloudAttack, palmRevengeLightning:
		return float32(minBrightness)
	case palmRevengeRecover:
		progress := math.Min(1, float64(event.age)/float64(palmRevengeRecoverFrames))
		return float32(minBrightness + (1-minBrightness)*progress)
	default:
		return 1
	}
}

func (s *GameScene) palmRevengeCloudFace(event *palmRevengeEvent) *ebiten.Image {
	if event == nil {
		return nil
	}
	switch event.phase {
	case palmRevengeCloudFocus, palmRevengeCloudAttack, palmRevengeLightning:
		return s.cloudAngryImage
	case palmRevengeRecover:
		frame := s.cloudGrinAnimation.frameAt(event.age)
		if frame != nil {
			return frame
		}
		return s.cloudGrinImage
	default:
		return nil
	}
}

func (s *GameScene) palmRevengeLightningVisible(event *palmRevengeEvent) bool {
	if event == nil || event.phase != palmRevengeLightning {
		return false
	}
	return event.age < 18 || (event.age >= 48 && event.age < 66)
}

func (s *GameScene) drawPalmRevengeScreenFlash(screen *ebiten.Image) {
	event := s.palmRevenge
	if event == nil || event.phase != palmRevengeLightning {
		return
	}
	alpha := s.palmRevengeScreenFlashAlpha(event.age)
	if alpha <= 0 {
		return
	}
	drawFilledRect(screen, screen.Bounds(), color.RGBA{R: 255, G: 255, B: 255, A: alpha})
}

func (s *GameScene) drawAbortRoundFlash(screen *ebiten.Image) {
	if !s.abortRoundFlashing {
		return
	}
	battlefield := image.Rect(0, 0, int(core.Config().Screen.Width), int(s.battlefieldHeight()))
	s.drawAbortRoundFlashOverlay(screen, battlefield)
}

func (s *GameScene) drawAbortRoundFlashOverlay(screen *ebiten.Image, r image.Rectangle) {
	if screen == nil || r.Empty() {
		return
	}
	age := maxInt(0, minInt(s.abortRoundFlashAge, abortRoundFlashFrames))
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(float64(r.Dx()), float64(r.Dy()))
	op.GeoM.Translate(float64(r.Min.X), float64(r.Min.Y))
	if age <= abortRoundFlashWhiteFrame {
		progress := float64(age) / float64(maxInt(1, abortRoundFlashWhiteFrame))
		if progress <= 0 {
			return
		}
		v := float32(progress)
		op.ColorScale.Scale(v, v, v, v)
		screen.DrawImage(getSolidWhiteImage(), op)
		return
	}
	progress := float64(age-abortRoundFlashWhiteFrame) / float64(maxInt(1, abortRoundFlashFrames-abortRoundFlashWhiteFrame))
	value := float32(1 - progress)
	op.ColorScale.Scale(value, value, value, 1)
	screen.DrawImage(getSolidWhiteImage(), op)
}

func (s *GameScene) palmRevengeScreenFlashAlpha(age int) uint8 {
	localAge, ok := lightningVisibleLocalAge(age)
	if !ok {
		return 0
	}
	pulseLength := 9
	pulse := localAge % pulseLength
	half := pulseLength / 2
	var strength float64
	if pulse <= half {
		strength = float64(pulse) / float64(maxInt(1, half))
	} else {
		strength = float64(pulseLength-pulse) / float64(maxInt(1, pulseLength-half))
	}
	return uint8(math.Round(165 * math.Max(0, math.Min(1, strength))))
}

func lightningVisibleLocalAge(age int) (int, bool) {
	switch {
	case age >= 0 && age < 18:
		return age, true
	case age >= 48 && age < 66:
		return age - 48, true
	default:
		return 0, false
	}
}

func (s *GameScene) drawPalmLeafFalls(screen *ebiten.Image, camera *ebiten.GeoM) {
	for _, leaf := range s.palmLeafFalls {
		if leaf.image == nil {
			continue
		}
		projected := leaf.pos.Project(camera)
		op := &ebiten.DrawImageOptions{}
		w := float64(leaf.image.Bounds().Dx())
		h := float64(leaf.image.Bounds().Dy())
		op.GeoM.Translate(-w/2, -h/2)
		op.GeoM.Rotate(math.Sin(float64(leaf.age)*0.19+leaf.pos.X) * 0.35)
		op.GeoM.Translate(projected.X, projected.Y)
		screen.DrawImage(leaf.image, op)
	}
}

func drawAnimationCentered(screen *ebiten.Image, camera *ebiten.GeoM, animation spriteAnimation, center engine.Vec, tick int) {
	drawAnimationFrame(screen, camera, animation, center, tick, 0.5, 0.5)
}

func drawAnimationCenteredLooping(screen *ebiten.Image, camera *ebiten.GeoM, animation spriteAnimation, center engine.Vec, tick int) {
	if animation.totalTicks > 0 {
		tick %= animation.totalTicks
	}
	drawAnimationFrame(screen, camera, animation, center, tick, 0.5, 0.5)
}

func drawAnimationBottomCentered(screen *ebiten.Image, camera *ebiten.GeoM, animation spriteAnimation, bottomCenter engine.Vec, tick int) {
	drawAnimationFrame(screen, camera, animation, bottomCenter, tick, 0.5, 1)
}

func drawAnimationBottomCenteredLooping(screen *ebiten.Image, camera *ebiten.GeoM, animation spriteAnimation, bottomCenter engine.Vec, tick int) {
	if animation.totalTicks > 0 {
		tick %= animation.totalTicks
	}
	drawAnimationFrame(screen, camera, animation, bottomCenter, tick, 0.5, 1)
}

func drawAnimationFrame(screen *ebiten.Image, camera *ebiten.GeoM, animation spriteAnimation, anchor engine.Vec, tick int, anchorX, anchorY float64) {
	frame := animation.frameAt(tick)
	if frame == nil {
		return
	}
	projected := anchor.Project(camera)
	width := float64(animation.width) * animation.scaleX
	height := float64(animation.height) * animation.scaleY
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(animation.scaleX, animation.scaleY)
	op.GeoM.Translate(projected.X-width*anchorX, projected.Y-height*anchorY)
	screen.DrawImage(frame, op)
}

func (s *GameScene) drawSandFalls(screen *ebiten.Image, camera *ebiten.GeoM) {
	for _, fall := range s.sandFalls {
		progress := float64(fall.age) / math.Max(1, float64(fall.duration))
		for _, pixel := range fall.pixels {
			y := float64(pixel.FromY) + (float64(pixel.ToY)-float64(pixel.FromY))*progress
			projected := engine.V(float64(pixel.X), y).Project(camera)
			drawFilledRect(screen, image.Rect(int(projected.X), int(projected.Y), int(projected.X)+1, int(projected.Y)+1), pixel.Color)
		}
	}
}

func (s *GameScene) drawZeroPowerEffects(screen *ebiten.Image, camera *ebiten.GeoM) {
	for _, effect := range s.zeroPowerEffects {
		if effect.tank == nil || effect.tank.body == nil || len(effect.animation.frames) == 0 {
			continue
		}
		body := effect.tank.body.Bounds()
		center := body.Center()
		if effect.delay > 0 {
			continue
		}
		frame := effect.animation.frameAt(effect.age)
		if frame == nil {
			continue
		}
		anchor := engine.V(center.X, center.Y).Project(camera)
		op := &ebiten.DrawImageOptions{}
		width := float64(effect.animation.width) * effect.animation.scaleX
		height := float64(effect.animation.height) * effect.animation.scaleY
		x := anchor.X - width/2
		y := anchor.Y - height/2
		if effect.animation.anchor == zeroPowerAnchorTankBottom {
			tankBottom := engine.V(center.X, body.Max.Y).Project(camera)
			y = tankBottom.Y - height
		}
		op.GeoM.Scale(effect.animation.scaleX, effect.animation.scaleY)
		op.GeoM.Translate(x, y)
		screen.DrawImage(frame, op)
	}
}

func (s *GameScene) drawGameHUD(screen *ebiten.Image) {
	t := texts()
	screenCfg := core.Config().Screen
	hud := image.Rect(0, int(s.battlefieldHeight()), int(screenCfg.Width), int(screenCfg.Height))
	drawFilledRect(screen, hud, color.RGBA{R: 5, G: 7, B: 10, A: 242})
	drawFilledRect(screen, image.Rect(hud.Min.X, hud.Min.Y, hud.Max.X, hud.Min.Y+2), color.RGBA{R: 245, G: 246, B: 214, A: 255})

	active := s.hudTank()
	if s.xmV12DriveMode {
		s.drawXMV12HUD(screen, hud, active)
		return
	}
	playerName := t.GameDefaultPlayerName
	playerColor := color.RGBA{R: 255, G: 160, B: 28, A: 255}
	power := 100
	shotStrength := 20
	if active != nil {
		playerName = active.player.Name
		playerColor = active.player.Color
		power = active.power
		shotStrength = active.shotStrength
	}

	s.drawHUDStepper(screen, image.Rect(10, hud.Min.Y+14, 112, hud.Min.Y+44), t.GameHUDStrength, shotStrength)
	s.drawHUDStepper(screen, image.Rect(10, hud.Min.Y+50, 122, hud.Min.Y+80), t.GameHUDAngle, int(math.Round(s.cannonDisplayAngleDegreesForTank(active))))

	centerX := int(screenCfg.Width) / 2
	drawText(screen, playerName, centerX-42, hud.Min.Y+30, playerColor)
	drawButton(screen, s.hudFireButtonRect(), t.GameHUDFire)
	if active != nil && s.playerHasXMV12(active.playerIndex) {
		s.drawIgnitionButton(screen, s.hudIgnitionRect(), s.mousePressedInRect(s.hudIgnitionRect()))
		s.drawIgnitionLabel(screen, s.hudIgnitionRect(), t.GameHUDIgnition)
	}

	windArrow := "->"
	if s.windDirection < 0 {
		windArrow = "<-"
	}
	rightX := int(screenCfg.Width) - 170
	if s.projectileReentry && s.reentrySymbol != nil {
		drawScaledImage(screen, s.reentrySymbol, image.Rect(rightX-60, hud.Min.Y+24, rightX-12, hud.Min.Y+56))
	}
	drawText(screen, t.GameHUDWind+": "+strconv.Itoa(s.wind)+" ("+windArrow+")", rightX, hud.Min.Y+30, colornames.White)
	drawText(screen, t.GameHUDPower+": "+strconv.Itoa(power), rightX, hud.Min.Y+55, colornames.White)

	for i := 0; i < s.weaponSlotCount(); i++ {
		s.drawWeaponSlot(screen, i)
	}
}

func (s *GameScene) DrawMobileOverlay(screen *ebiten.Image, viewport image.Rectangle) {
	s.drawMobileSideControls(screen, viewport)
}

func (s *GameScene) drawScoreTable(screen *ebiten.Image) {
	if !s.showScoreTable {
		return
	}
	t := texts()

	screenCfg := core.Config().Screen
	tableW := int(math.Min(620, screenCfg.Width-80))
	if tableW < 360 {
		tableW = int(screenCfg.Width) - 32
	}
	rowH := 30
	tableH := 112 + rowH*len(s.tanks)
	left := int(screenCfg.Width)/2 - tableW/2
	top := int(s.battlefieldHeight())/2 - tableH/2
	if top < 42 {
		top = 42
	}
	right := left + tableW
	bottom := top + tableH

	drawFilledRect(screen, image.Rect(left-14, top-16, right+14, bottom+12), color.RGBA{R: 22, G: 10, B: 38, A: 118})

	title := t.GameScoreRound + " " + strconv.Itoa(maxInt(1, s.roundNumber)) + " " + t.GameScoreOf + " " + strconv.Itoa(maxInt(1, s.g.rounds))
	drawCenteredText(screen, title, image.Rect(left, top, right, top+24), colornames.Yellow)

	headerY := top + 60
	nameX := left + 80
	scoreX := left + tableW/2 + 60
	statusX := right - 110
	lineColor := color.RGBA{R: 250, G: 246, B: 230, A: 230}
	textColor := color.RGBA{R: 250, G: 246, B: 255, A: 255}

	drawText(screen, t.GameScorePlayer, nameX, headerY, textColor)
	drawText(screen, t.GameScoreSuccess, scoreX, headerY, textColor)
	drawText(screen, t.GameScoreStatus, statusX, headerY, textColor)

	separatorY := headerY + 24
	drawFilledRect(screen, image.Rect(left+18, separatorY, right-18, separatorY+2), lineColor)
	drawFilledRect(screen, image.Rect(left+tableW/2-8, headerY-22, left+tableW/2-6, bottom-18), lineColor)
	drawFilledRect(screen, image.Rect(right-170, headerY-22, right-168, bottom-18), lineColor)

	rows := s.scoreTableTanks()
	for i, tank := range rows {
		if tank == nil {
			continue
		}
		y := separatorY + 32 + i*rowH
		status := t.GameStatusActive
		if tank.power <= 0 {
			status = t.GameStatusOut
		}
		drawText(screen, tank.player.Name, nameX, y, textColor)
		drawText(screen, strconv.Itoa(s.scoreForPlayer(tank.playerIndex)), scoreX+28, y, textColor)
		drawText(screen, status, statusX+18, y, textColor)
	}
}

func (s *GameScene) drawPlayerNames(screen *ebiten.Image, camera *ebiten.GeoM) {
	if !s.showPlayerNames {
		return
	}
	for _, tank := range s.tanks {
		if tank == nil || tank.body == nil || tank.power <= 0 || tank.zeroPowerGone {
			continue
		}
		body := tank.body.Bounds()
		center := engine.V(body.Min.X+body.W()/2, body.Min.Y-26).Project(camera)
		label := tank.player.Name
		b := text.BoundString(dialogTextFace, label)
		drawTextFace(screen, label, dialogTextFace, int(center.X)-b.Dx()/2, int(center.Y), tank.player.Color)
	}
}

func (s *GameScene) drawGameDialogs(screen *ebiten.Image) {
	t := texts()
	if s.gamePaused {
		drawFilledRect(screen, image.Rect(0, 0, int(core.Config().Screen.Width), int(core.Config().Screen.Height)), color.RGBA{A: 70})
		drawCenteredText(screen, t.GamePause, image.Rect(0, int(core.Config().Screen.Height)/2-24, int(core.Config().Screen.Width), int(core.Config().Screen.Height)/2+24), colornames.White)
	}
	if s.confirmDialog != gameConfirmNone {
		s.drawConfirmDialog(screen)
		return
	}
	if s.gameHelpOpen {
		s.drawGameHelpDialog(screen)
	}
	if s.playerInfoOpen {
		s.drawPlayerInfoDialog(screen)
	}
}

func gameHelpRect() image.Rectangle {
	screen := core.Config().Screen
	w := minInt(860, maxInt(620, int(screen.Width)-40))
	h := minInt(460, maxInt(440, int(screen.Height)-40))
	return centerDialogRect(w, h)
}

func gameHelpOKRect() image.Rectangle {
	r := gameHelpRect()
	return image.Rect(r.Max.X-75, r.Max.Y-31, r.Max.X-15, r.Max.Y-9)
}

func playerInfoRect() image.Rectangle {
	return centerDialogRect(477, 318)
}

func playerInfoOKRect() image.Rectangle {
	r := playerInfoRect()
	return image.Rect(r.Max.X-86, r.Max.Y-47, r.Max.X-15, r.Max.Y-23)
}

func confirmDialogRect() image.Rectangle {
	return centerDialogRect(350, 136)
}

func confirmDialogYesRect() image.Rectangle {
	r := confirmDialogRect()
	return image.Rect(r.Max.X-166, r.Max.Y-44, r.Max.X-96, r.Max.Y-20)
}

func confirmDialogNoRect() image.Rectangle {
	r := confirmDialogRect()
	return image.Rect(r.Max.X-84, r.Max.Y-44, r.Max.X-14, r.Max.Y-20)
}

func centerDialogRect(w, h int) image.Rectangle {
	screen := core.Config().Screen
	return image.Rect(0, 0, w, h).Add(image.Pt((int(screen.Width)-w)/2, (int(screen.Height)-h)/2))
}

func (s *GameScene) drawGameHelpDialog(screen *ebiten.Image) {
	t := texts()
	r := gameHelpRect()
	drawDialogWindow(screen, r, t.GameHelpTitle)

	logo := image.Rect(r.Max.X-88, r.Min.Y+67, r.Max.X-17, r.Max.Y-90)
	keysBox := image.Rect(r.Min.X+12, r.Min.Y+63, logo.Min.X-10, r.Max.Y-14)
	drawGroupBox(screen, keysBox, t.GameHelpKeys)

	lines := []struct {
		key  string
		desc string
	}{
		{"Enter:", t.GameHelpFire},
		{"Tab (+ Shift):", t.GameHelpNextWeapon},
		{"Pfeil links (+ Shift):", t.GameHelpRotateClockwise},
		{"Pfeil rechts (+ Shift):", t.GameHelpRotateCounterClockwise},
		{"Pfeil hoch (+ Shift):", t.GameHelpIncreaseStrength},
		{"Pfeil runter (+ Shift):", t.GameHelpDecreaseStrength},
		{"m:", t.GameHelpIgnition},
		{"i:", t.GameHelpInvertAngle},
		{"s:", t.GameHelpScrollOMat},
		{"+:", t.GameHelpIncreaseStrengthByTen},
		{"-:", t.GameHelpDecreaseStrengthByTen},
		{"1, 2, ..., 0:", t.GameHelpPlayerInfo},
		{"Leertaste:", t.GameHelpScoreTable},
		{"", ""},
		{"Strg Q:", t.GameHelpQuit},
		{"Strg E:", t.GameHelpAbortRound},
		{"Strg P:", t.GameHelpPause},
		{"Strg N:", t.GameHelpTogglePlayerNames},
	}
	y := keysBox.Min.Y + 28
	for _, line := range lines {
		if line.key != "" {
			drawTextFace(screen, line.key, dialogTextFace, keysBox.Min.X+12, y, colornames.Black)
			drawTextFace(screen, line.desc, dialogTextFace, keysBox.Min.X+200, y, colornames.Black)
		}
		y += 18
	}

	drawFilledRect(screen, logo, color.RGBA{R: 0, G: 105, B: 100, A: 255})
	drawVerticalLogo(screen, logo)
	drawDialogButton(screen, gameHelpOKRect(), t.DialogOK)
}

func (s *GameScene) drawConfirmDialog(screen *ebiten.Image) {
	title := ""
	switch s.confirmDialog {
	case gameConfirmAbortRound:
		title = "Laufende Runde abbrechen"
	case gameConfirmQuit:
		title = "Tank Blaster sofort beenden?"
	default:
		return
	}
	r := confirmDialogRect()
	iconRect := image.Rect(r.Min.X+38, r.Min.Y+48, r.Min.X+70, r.Min.Y+80)
	textRect := image.Rect(iconRect.Max.X+14, r.Min.Y+50, r.Max.X-20, r.Min.Y+78)
	drawFilledRect(screen, image.Rect(0, 0, int(core.Config().Screen.Width), int(core.Config().Screen.Height)), color.RGBA{A: 70})
	drawDialogWindow(screen, r, title)
	if s.questionIconImage != nil {
		drawScaledImage(screen, s.questionIconImage, iconRect)
	}
	drawTextFace(screen, title, dialogTextFace, textRect.Min.X, textRect.Min.Y+19, colornames.Black)
	drawDialogButton(screen, confirmDialogYesRect(), "Ja")
	drawDialogButton(screen, confirmDialogNoRect(), "Nein")
}

func drawVerticalLogo(screen *ebiten.Image, r image.Rectangle) {
	chars := []string{"TANK", "BLASTER"}
	x := r.Min.X + r.Dx()/2 - 8
	y := r.Min.Y + 38
	for _, word := range chars {
		for _, ch := range word {
			drawTextFace(screen, string(ch), uiTextFace, x+2, y+2, colornames.Black)
			drawTextFace(screen, string(ch), uiTextFace, x, y, color.RGBA{R: 0, G: 220, B: 220, A: 255})
			y += 18
		}
		y += 10
	}
}

func (s *GameScene) drawPlayerInfoDialog(screen *ebiten.Image) {
	t := texts()
	r := playerInfoRect()
	drawDialogWindow(screen, r, t.PlayerInfoTitle)

	tank := s.tankByPlayerIndex(s.playerInfoIndex)
	if tank == nil {
		drawCenteredTextFace(screen, t.PlayerInfoMissing, r, dialogTextFace, colornames.Black)
		drawDialogButton(screen, playerInfoOKRect(), t.DialogOK)
		return
	}

	left := image.Rect(r.Min.X+12, r.Min.Y+61, r.Min.X+213, r.Max.Y-23)
	drawGroupBox(screen, left, tank.player.Name)
	portraitRect := image.Rect(left.Min.X+10, left.Min.Y+15, left.Min.X+122, left.Min.Y+124)
	drawFrame(screen, portraitRect, colornames.White, color.RGBA{R: 170, G: 170, B: 170, A: 255})
	if portrait := s.portraitForPlayer(tank.player); portrait != nil {
		drawScaledImage(screen, portrait, insetRect(portraitRect, 2))
	}
	drawCenteredTextFace(screen, t.PlayerInfoTankModel, image.Rect(left.Min.X+122, left.Min.Y+20, left.Max.X-8, left.Min.Y+38), dialogTextFace, colornames.Black)
	modelRect := image.Rect(left.Min.X+132, left.Min.Y+42, left.Max.X-18, left.Min.Y+86)
	drawFilledRect(screen, modelRect, colornames.Black)
	if tank.body != nil && tank.body.Image != nil {
		drawScaledImage(screen, tank.body.Image, insetRect(modelRect, 4))
	}
	modelName := t.PlayerInfoStandardTank
	if s.playerHasXMV12(tank.playerIndex) {
		modelName = "XM-V12"
	}
	drawCenteredTextFace(screen, modelName, image.Rect(left.Min.X+126, left.Min.Y+88, left.Max.X-8, left.Min.Y+108), dialogTextFace, colornames.Black)

	status := t.GameStatusActive
	if tank.power <= 0 || tank.zeroPowerGone {
		status = t.GameStatusEliminated
	}
	infoY := left.Min.Y + 146
	infoRows := []struct {
		label string
		value string
	}{
		{t.PlayerInfoStatus, status},
		{t.PlayerInfoEnergy, strconv.Itoa(tank.power)},
		{t.PlayerInfoMoney, "$" + strconv.Itoa(s.creditForPlayer(tank.playerIndex))},
		{t.PlayerInfoEnergyShield, strconv.Itoa(s.energyShieldPercentForPlayer(tank.playerIndex)) + "%"},
	}
	for _, row := range infoRows {
		drawTextFace(screen, row.label, dialogTextFace, left.Min.X+10, infoY, colornames.Black)
		drawTextFace(screen, row.value, dialogTextFace, left.Min.X+112, infoY, colornames.Black)
		infoY += 24
	}

	arsenalLabelX := r.Min.X + 227
	drawTextFace(screen, t.PlayerInfoArsenal, dialogTextFace, arsenalLabelX, r.Min.Y+56, colornames.Black)
	arsenal := image.Rect(arsenalLabelX, r.Min.Y+63, r.Max.X-13, r.Max.Y-62)
	drawFrame(screen, arsenal, colornames.White, color.RGBA{R: 150, G: 150, B: 150, A: 255})
	s.drawPlayerArsenal(screen, tank, insetRect(arsenal, 4))
	drawDialogButton(screen, playerInfoOKRect(), t.DialogOK)
}

func (s *GameScene) portraitForPlayer(player PlayerConfig) *ebiten.Image {
	if player.Kind == PlayerComputer {
		if portrait := s.computerPortraits[player.ComputerID]; portrait != nil {
			return portrait
		}
		return s.computerPortraits[computerplayers.DoedelID]
	}
	return s.humanPortrait
}

func (s *GameScene) creditForPlayer(playerIndex int) int {
	if playerIndex < 0 || playerIndex >= len(s.credits) {
		return 0
	}
	return s.credits[playerIndex]
}

func (s *GameScene) drawPlayerArsenal(screen *ebiten.Image, tank *battleTank, r image.Rectangle) {
	if tank == nil {
		return
	}
	t := texts()
	y := r.Min.Y + 12
	drawArsenalRow := func(name string, count int) bool {
		if y > r.Max.Y-8 {
			return false
		}
		drawTextFace(screen, name, dialogTextFace, r.Min.X, y, colornames.Black)
		drawTextFace(screen, strconv.Itoa(count), dialogTextFace, r.Max.X-82, y, colornames.Black)
		y += 16
		return true
	}
	drawArsenalRow(t.ItemTrainingAmmo, 10000)
	drawPriorityWeaponSlot := func(slot int) {
		weaponList := weaponspkg.List()
		if slot <= 0 || slot >= s.weaponSlotCount() || slot >= len(weaponList) {
			return
		}
		count := s.ammoForWeaponSlot(tank.playerIndex, slot)
		if count <= 0 {
			return
		}
		drawArsenalRow(localizedWeaponSlotName(slot, weaponList[slot].Name), count)
	}
	drawPriorityWeaponSlot(10)
	if count := s.mfsBoosterCountForPlayer(tank.playerIndex); count > 0 {
		drawArsenalRow(t.ItemMFSBooster, count)
	}
	if shield := s.energyShieldPercentForPlayer(tank.playerIndex); shield > 0 {
		drawArsenalRow(t.ItemEnergyShield, shield)
	}
	if s.playerHasXMV12(tank.playerIndex) {
		drawArsenalRow(t.ItemXMV12Tank, 1)
		if diesel := s.dieselForPlayer(tank.playerIndex); diesel > 0 {
			drawArsenalRow(t.ItemDiesel, diesel)
		}
	}

	weaponList := weaponspkg.List()
	for slot := 1; slot < s.weaponSlotCount() && slot < len(weaponList); slot++ {
		if slot == 10 {
			continue
		}
		count := s.ammoForWeaponSlot(tank.playerIndex, slot)
		if count <= 0 {
			continue
		}
		if !drawArsenalRow(localizedWeaponSlotName(slot, weaponList[slot].Name), count) {
			break
		}
	}
}

func (s *GameScene) scoreTableTanks() []*battleTank {
	if len(s.tanks) == 0 {
		return nil
	}
	rows := make([]*battleTank, 0, len(s.tanks))
	for _, tank := range s.tanks {
		rows = append(rows, tank)
	}
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i] == nil {
			return false
		}
		if rows[j] == nil {
			return true
		}
		leftScore := s.scoreForPlayer(rows[i].playerIndex)
		rightScore := s.scoreForPlayer(rows[j].playerIndex)
		if leftScore != rightScore {
			return leftScore > rightScore
		}
		return rows[i].playerIndex < rows[j].playerIndex
	})
	return rows
}

func (s *GameScene) drawRoundTransitionBanner(screen *ebiten.Image) {
	if s.roundTransitionDelay <= 0 && !s.roundSeriesComplete {
		return
	}

	screenCfg := core.Config().Screen
	bannerW := int(math.Min(760, screenCfg.Width-140))
	if bannerW < 320 {
		bannerW = int(screenCfg.Width) - 32
	}
	bannerH := 56
	left := int(screenCfg.Width)/2 - bannerW/2
	top := int(s.battlefieldHeight()*0.53) - bannerH/2
	r := image.Rect(left, top, left+bannerW, top+bannerH)

	drawFrame(screen, r, color.RGBA{R: 4, G: 4, B: 4, A: 232}, colornames.Yellow)

	played := maxInt(0, minInt(s.roundNumber, maxInt(1, s.g.rounds)))
	remaining := maxInt(0, maxInt(1, s.g.rounds)-played)
	t := texts()
	textValue := strconv.Itoa(played) + " " + t.RoundTransitionPlayed + " " + strconv.Itoa(remaining) + " " + t.RoundTransitionRemainingRounds
	drawCenteredText(screen, textValue, r, colornames.White)
}

func (s *GameScene) updateRoundTransition() error {
	if s.roundSeriesComplete {
		s.reportOnlineMatchComplete()
		return s.g.SetNewScene(NewHallOfFameScene(s.hallOfFameScores()))
	}
	s.roundTransitionDelay--
	if s.roundTransitionDelay > 0 {
		return nil
	}
	s.roundTransitionDelay = 0
	if s.roundNumber >= maxInt(1, s.g.rounds) {
		s.roundSeriesComplete = true
		s.reportOnlineMatchComplete()
		return s.g.SetNewScene(NewHallOfFameScene(s.hallOfFameScores()))
	}
	s.roundNumber++
	s.beginShop()
	return nil
}

func (s *GameScene) hallOfFameScores() []hallOfFameScore {
	rows := make([]hallOfFameScore, 0, len(s.players))
	for i, player := range s.players {
		rows = append(rows, hallOfFameScore{
			Name:  player.Name,
			Score: s.scoreForPlayer(i),
			Color: player.Color,
			Index: i,
		})
	}
	return rows
}

func drawScaledImage(screen, img *ebiten.Image, r image.Rectangle) {
	if img == nil || r.Empty() {
		return
	}
	bounds := img.Bounds()
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(float64(r.Dx())/float64(bounds.Dx()), float64(r.Dy())/float64(bounds.Dy()))
	op.GeoM.Translate(float64(r.Min.X), float64(r.Min.Y))
	screen.DrawImage(img, op)
}

func splitImageFrames(img *ebiten.Image, frameWidth int) []*ebiten.Image {
	if img == nil || frameWidth <= 0 {
		return nil
	}
	bounds := img.Bounds()
	frames := make([]*ebiten.Image, 0, bounds.Dx()/frameWidth)
	for x := bounds.Min.X; x+frameWidth <= bounds.Max.X; x += frameWidth {
		if frame, ok := img.SubImage(image.Rect(x, bounds.Min.Y, x+frameWidth, bounds.Max.Y)).(*ebiten.Image); ok {
			frames = append(frames, frame)
		}
	}
	return frames
}

func insetRect(r image.Rectangle, inset int) image.Rectangle {
	return image.Rect(r.Min.X+inset, r.Min.Y+inset, r.Max.X-inset, r.Max.Y-inset)
}

func (s *GameScene) drawHUDStepper(screen *ebiten.Image, r image.Rectangle, label string, value int) {
	buttonW := 24
	drawButton(screen, image.Rect(r.Min.X, r.Min.Y, r.Min.X+buttonW, r.Min.Y+22), "-")
	drawButton(screen, image.Rect(r.Min.X+buttonW+4, r.Min.Y, r.Min.X+buttonW*2+4, r.Min.Y+22), "+")
	drawText(screen, label+": "+strconv.Itoa(value), r.Min.X+buttonW*2+12, r.Min.Y+17, colornames.White)
}

func (s *GameScene) drawXMV12HUD(screen *ebiten.Image, hud image.Rectangle, tank *battleTank) {
	if tank == nil {
		return
	}
	hudCenterY := hud.Min.Y + hud.Dy()/2
	gaugeHeight := 69
	gaugeRect := image.Rect(hud.Min.X+28, hudCenterY-gaugeHeight/2, hud.Min.X+125, hudCenterY-gaugeHeight/2+gaugeHeight)
	drawScaledImage(screen, s.fuelGaugeImage, gaugeRect)
	s.drawFuelNeedle(screen, gaugeRect, float64(s.dieselForPlayer(tank.playerIndex))/float64(xmV12MaxDiesel))

	slopeHeight := 33
	slopeRect := image.Rect(gaugeRect.Max.X+89, hudCenterY-slopeHeight/2, gaugeRect.Max.X+200, hudCenterY-slopeHeight/2+slopeHeight)
	drawScaledImage(screen, s.slopeMeterImage, slopeRect)
	s.drawSlopeMarker(screen, slopeRect, tank)

	centerX := int(core.Config().Screen.Width) / 2
	drawText(screen, tank.player.Name, centerX-38, hudCenterY-18, tank.player.Color)
	drawButton(screen, s.xmV12LeftButtonRect(), "<")
	drawButton(screen, s.xmV12StopButtonRect(), "STOP")
	drawButton(screen, s.xmV12RightButtonRect(), ">")

	s.drawIgnitionButton(screen, s.xmV12MotorOffRect(), s.mousePressedInRect(s.xmV12MotorOffRect()))
	s.drawIgnitionLabel(screen, s.xmV12MotorOffRect(), texts().XMV12MotorOff)
}

func (s *GameScene) drawFuelNeedle(screen *ebiten.Image, r image.Rectangle, ratio float64) {
	ratio = math.Max(0, math.Min(1, ratio))
	anchor := engine.V(float64(r.Min.X)+37*float64(r.Dx())/97, float64(r.Min.Y)+65*float64(r.Dy())/69)
	length := float32(56 * float64(r.Dx()) / 97)
	angle := engine.DegToRad(-4 - ratio*76)
	end := anchor.Add(engine.V(float64(length), 0).Rotated(angle))
	vector.StrokeLine(screen, float32(anchor.X), float32(anchor.Y), float32(end.X), float32(end.Y), 1, colornames.White, false)
}

func (s *GameScene) drawSlopeMarker(screen *ebiten.Image, r image.Rectangle, tank *battleTank) {
	if tank == nil || tank.body == nil {
		return
	}
	degrees := math.Max(-60, math.Min(60, engine.RadToDeg(tank.body.Rot)))
	x := float64(r.Min.X) + float64(r.Dx())*(degrees+60)/120
	centerX := float64(r.Min.X) + float64(r.Dx())/2
	x1 := int(math.Min(centerX, x))
	x2 := int(math.Max(centerX, x))
	if x2 <= x1 {
		x2 = x1 + 1
	}
	bar := image.Rect(x1, r.Min.Y+13, x2, r.Min.Y+30)
	c := colornames.White
	absSlope := math.Abs(degrees)
	switch {
	case absSlope >= 50:
		c = color.RGBA{R: 220, G: 0, B: 0, A: 255}
	case absSlope >= 40:
		c = color.RGBA{R: 245, G: 118, B: 0, A: 255}
	case absSlope >= 30:
		c = color.RGBA{R: 245, G: 232, B: 0, A: 255}
	}
	drawFilledRect(screen, bar, c)
}

func (s *GameScene) drawIgnitionButton(screen *ebiten.Image, r image.Rectangle, pressed bool) {
	if len(s.ignitionFrames) == 0 {
		drawButton(screen, r, "")
		return
	}
	frame := s.ignitionFrames[0]
	if pressed && len(s.ignitionFrames) > 1 {
		frame = s.ignitionFrames[1]
	}
	drawScaledImage(screen, frame, r)
}

func (s *GameScene) drawIgnitionLabel(screen *ebiten.Image, button image.Rectangle, label string) {
	labelRect := image.Rect(button.Min.X-20, button.Max.Y+4, button.Max.X+20, button.Max.Y+20)
	drawCenteredTextFace(screen, label, labelRect, dialogTextFace, colornames.White)
}

func (s *GameScene) drawWeaponSlot(screen *ebiten.Image, index int) {
	r := s.weaponSlotRect(index)
	shopItemIndex := s.itemIndexForWeaponSlot(index)
	active := s.activeTank()
	inStock := false
	if index == 0 {
		inStock = true
	} else if active != nil && shopItemIndex >= 0 {
		inStock = s.shopItemCountForPlayer(active.playerIndex, shopItemIndex) > 0
	}
	img := s.shop.weaponbarOutOfStock
	if inStock {
		img = s.shop.weaponbarOnStock
	}
	if active != nil && index == active.selectedWeapon {
		img = s.shop.weaponbarActive
	}
	if img == nil {
		return
	}
	src := weaponbarSlotSourceRect(img, index)
	slot, ok := img.SubImage(src).(*ebiten.Image)
	if !ok {
		return
	}
	drawScaledImage(screen, slot, r)
}

func (s *GameScene) weaponSlotRect(index int) image.Rectangle {
	screenCfg := core.Config().Screen
	screenW := int(screenCfg.Width)
	sourceW := 640
	sourceH := 34
	if s.shop.weaponbarOnStock != nil {
		bounds := s.shop.weaponbarOnStock.Bounds()
		sourceW = bounds.Dx()
		sourceH = bounds.Dy()
	}
	x1 := index * screenW / weaponbarSlotCount
	x2 := (index + 1) * screenW / weaponbarSlotCount
	height := int(math.Round(float64(sourceH) * float64(screenW) / float64(sourceW)))
	y := int(screenCfg.Height) - height
	return image.Rect(x1, y, x2, y+height)
}

func (s *GameScene) weaponSlotCount() int {
	return weaponbarSlotCount
}

func weaponbarSlotSourceRect(img *ebiten.Image, index int) image.Rectangle {
	bounds := img.Bounds()
	x1 := bounds.Min.X + index*bounds.Dx()/weaponbarSlotCount
	x2 := bounds.Min.X + (index+1)*bounds.Dx()/weaponbarSlotCount
	return image.Rect(x1, bounds.Min.Y, x2, bounds.Max.Y)
}

func (s *GameScene) shopItemCountForActivePlayer(itemIndex int) int {
	tank := s.activeTank()
	if tank == nil {
		return 0
	}
	return s.shopItemCountForPlayer(tank.playerIndex, itemIndex)
}

func (s *GameScene) shopItemCountForPlayer(playerIndex, itemIndex int) int {
	if playerIndex < 0 || playerIndex >= len(s.inventories) || itemIndex < 0 {
		return 0
	}
	inventory := s.inventories[playerIndex]
	count := 0
	if itemIndex < len(inventory.classA) {
		count += inventory.classA[itemIndex]
	}
	if itemIndex < len(inventory.classB) {
		count += inventory.classB[itemIndex]
	}
	return count
}

func (s *GameScene) cannonAngleDegrees() float64 {
	return s.cannonDisplayAngleDegreesForTank(s.activeTank())
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func debugScrollBarRect() image.Rectangle {
	screen := core.Config().Screen
	y := int(screen.Height) - gameHUDHeight - 20
	return image.Rect(12, y, int(screen.Width)-12, y+16)
}

func (s *GameScene) hudStrengthMinusRect() image.Rectangle {
	hudY := int(s.battlefieldHeight())
	return image.Rect(10, hudY+14, 34, hudY+36)
}

func (s *GameScene) hudStrengthPlusRect() image.Rectangle {
	hudY := int(s.battlefieldHeight())
	return image.Rect(38, hudY+14, 62, hudY+36)
}

func (s *GameScene) hudAngleMinusRect() image.Rectangle {
	hudY := int(s.battlefieldHeight())
	return image.Rect(10, hudY+50, 34, hudY+72)
}

func (s *GameScene) hudAnglePlusRect() image.Rectangle {
	hudY := int(s.battlefieldHeight())
	return image.Rect(38, hudY+50, 62, hudY+72)
}

func (s *GameScene) hudFireButtonRect() image.Rectangle {
	centerX := int(core.Config().Screen.Width) / 2
	hudY := int(s.battlefieldHeight())
	return image.Rect(centerX-64, hudY+44, centerX+64, hudY+76)
}

func (s *GameScene) hudIgnitionRect() image.Rectangle {
	hudY := int(s.battlefieldHeight())
	return image.Rect(210, hudY+19, 241, hudY+49)
}

func (s *GameScene) xmV12LeftButtonRect() image.Rectangle {
	centerX := int(core.Config().Screen.Width) / 2
	hudY := int(s.battlefieldHeight())
	return image.Rect(centerX-80, hudY+56, centerX-58, hudY+78)
}

func (s *GameScene) xmV12StopButtonRect() image.Rectangle {
	centerX := int(core.Config().Screen.Width) / 2
	hudY := int(s.battlefieldHeight())
	return image.Rect(centerX-54, hudY+56, centerX+54, hudY+78)
}

func (s *GameScene) xmV12RightButtonRect() image.Rectangle {
	centerX := int(core.Config().Screen.Width) / 2
	hudY := int(s.battlefieldHeight())
	return image.Rect(centerX+58, hudY+56, centerX+80, hudY+78)
}

func (s *GameScene) xmV12MotorOffRect() image.Rectangle {
	screen := core.Config().Screen
	hudY := int(s.battlefieldHeight())
	return image.Rect(int(screen.Width)-78, hudY+40, int(screen.Width)-47, hudY+70)
}

func (s *GameScene) mousePressedInRect(r image.Rectangle) bool {
	if !primaryPointerPressed() {
		return false
	}
	x, y := primaryPointerPosition()
	return image.Pt(x, y).In(r)
}

func (s *GameScene) battlefieldHeight() float64 {
	return math.Max(120, core.Config().Screen.Height-gameHUDHeight)
}

func (s *GameScene) skyExtraHeight() float64 {
	return s.battlefieldHeight()*0.10 + additionalVisibleSkyHeight
}

func (s *GameScene) behaviorMoveOnButton(source *engine.Sprite) {
	if MoveLeft() {
		moveLeft := engine.Vec{X: source.Pos.X - float64(source.MovementSpeed), Y: source.Pos.Y}
		source.Pos = &moveLeft
	} else if MoveRight() {
		moveRight := engine.Vec{X: source.Pos.X + float64(source.MovementSpeed), Y: source.Pos.Y}
		source.Pos = &moveRight
	}
}

func (s *GameScene) behaviorFallOntoGround(ground models.Ground, tank *battleTank, spawnPause bool) engine.Behavior {
	return func(source *engine.Sprite) {
		const gravity = 0.38
		const maxFallSpeed = 12.0

		source.Velocity.Y = math.Min(maxFallSpeed, source.Velocity.Y+gravity)
		source.Pos = source.Pos.Add(source.Velocity)

		centerX := source.Pos.X + source.Size.X/2
		landingY := ground.SurfaceY(centerX)
		if tank != nil && models.IsSmallTankBody(source) {
			landingY = s.uprightTankBottomY(source)
		}
		if tank != nil && tank.fallDamage {
			landingY = math.Max(landingY, tank.fallTargetY)
		}
		if source.Pos.Y+source.Size.Y < landingY {
			return
		}

		source.Velocity = engine.Vec{}
		if tank != nil {
			if tank.fallDamage {
				s.alignTankBodyToSurface(source)
			} else {
				s.alignTankBodyToSurface(source)
			}
			tank.landed = true
			tank.falling = false
			if tank.fallDamage {
				fallDistance := source.Pos.Y - tank.fallStartY
				s.applyFallDamage(tank, fallDistance)
				tank.fallDamage = false
				tank.fallTargetY = 0
			}
			if spawnPause {
				s.spawnPauseFrames = s.spawnLandingPauseFrames()
			}
		} else {
			ground.AlignSpriteToSurface(source)
		}
		source.Steps = nil
	}
}

func (s *GameScene) applyFallDamage(tank *battleTank, fallDistance float64) {
	if tank == nil || fallDistance <= 0 {
		return
	}
	damage := (2 * int(fallDistance)) / 3
	damage = maxInt(1, damage)
	s.damageTank(tank, damage, s.lastDamageSource, damageCauseFall)
}

func (s *GameScene) spawnLandingPauseFrames() int {
	return secondsToFrames(core.Config().Gameplay.SpawnLandingPauseSeconds)
}

func (s *GameScene) impactAnimationFramesForWeapon(weapon weaponspkg.Weapon) int {
	return secondsToFrames(weaponspkg.ImpactAnimationSeconds(weapon))
}

func (s *GameScene) impactPauseFrames() int {
	return secondsToFrames(core.Config().Gameplay.ImpactPauseSeconds)
}

func (s *GameScene) tankHitPauseFrames() int {
	return secondsToFrames(core.Config().Gameplay.TankHitPauseSeconds)
}

func (s *GameScene) palmHitPauseFrames() int {
	return secondsToFrames(core.Config().Gameplay.PalmHitPauseSeconds)
}

func secondsToFrames(seconds float64) int {
	return maxInt(0, int(math.Round(seconds*60)))
}

func (s *GameScene) chooseStartingPlayerAfterLanding() {
	if s.activePlayerIndex >= 0 || !s.allTanksLanded() {
		return
	}
	living := s.livingTankIndexes()
	if len(living) == 0 {
		return
	}
	if s.g.online != nil {
		if index := s.g.online.state.CurrentPlayerIndex; index >= 0 && index < len(s.tanks) && s.tankCanAct(s.tanks[index]) {
			s.activePlayerIndex = index
			s.resetComputerTurnPlans()
			return
		}
	}
	s.activePlayerIndex = living[s.rng.Intn(len(living))]
	s.resetComputerTurnPlans()
}

func (s *GameScene) onlineCanControlActivePlayer() bool {
	return s.onlineCanControlPlayer(s.activePlayerIndex)
}

func (s *GameScene) onlineCanControlPlayer(playerIndex int) bool {
	if s.g.online == nil {
		return true
	}
	if playerIndex < 0 || playerIndex >= len(s.g.online.state.Players) {
		return false
	}
	return s.g.online.state.Players[playerIndex].ID == s.g.online.playerID
}

func (s *GameScene) allTanksLanded() bool {
	if len(s.tanks) == 0 {
		return false
	}
	for _, tank := range s.tanks {
		if tank == nil {
			return false
		}
		if tank.power > 0 && !tank.landed {
			return false
		}
	}
	return true
}

func (s *GameScene) tankCanAct(tank *battleTank) bool {
	return tank != nil && tank.power > 0 && tank.landed && !tank.falling
}

func (s *GameScene) ensureActivePlayerCanAct() bool {
	if s.activePlayerIndex >= 0 && s.activePlayerIndex < len(s.tanks) && s.tankCanAct(s.tanks[s.activePlayerIndex]) {
		return true
	}
	if s.endRoundIfOnlyOneTankRemains() {
		return false
	}
	next := s.nextActivePlayerIndex()
	if next < 0 {
		return false
	}
	s.activePlayerIndex = next
	s.resetComputerTurnPlans()
	s.cameraGoal = s.cameraTargetForTank(s.activePlayerIndex)
	s.cameraGoalY = 0
	return true
}

func (s *GameScene) livingTankIndexes() []int {
	indexes := make([]int, 0, len(s.tanks))
	for index, tank := range s.tanks {
		if s.tankCanAct(tank) {
			indexes = append(indexes, index)
		}
	}
	return indexes
}

func (s *GameScene) livingTankCount() int {
	count := 0
	for _, tank := range s.tanks {
		if tank != nil && tank.power > 0 {
			count++
		}
	}
	return count
}

func (s *GameScene) endRoundIfOnlyOneTankRemains() bool {
	if s.roundTransitionDelay > 0 || s.roundSeriesComplete {
		return true
	}
	if len(s.tanks) <= 1 || s.livingTankCount() > 1 {
		return false
	}
	s.playEventSound(soundEventRoundEnd)
	if winner := s.roundWinner(); winner != nil {
		s.rewardRoundWinner(winner)
	}
	s.roundTransitionDelay = secondsToFrames(roundTransitionSeconds)
	return true
}

func (s *GameScene) roundWinner() *battleTank {
	var winner *battleTank
	for _, tank := range s.tanks {
		if tank == nil || tank.power <= 0 {
			continue
		}
		if winner != nil {
			return nil
		}
		winner = tank
	}
	return winner
}

func (s *GameScene) rewardRoundWinner(winner *battleTank) {
	if winner == nil || winner.power <= 0 {
		return
	}
	scoring := core.Config().Gameplay.Scoring
	s.addScore(winner.playerIndex, scoring.RoundWin.Points)
	s.addCredits(winner.playerIndex, winner.power*scoring.RoundWin.CreditPerRemainingEnergy)
}

func (s *GameScene) behaviorAttachCannonToTank(tank *engine.Sprite) engine.Behavior {
	return func(source *engine.Sprite) {
		if tank == nil {
			return
		}
		if source.RotAnchor != nil {
			mount := models.TankCannonMount(tank)
			center := engine.V(tank.Size.X/2, tank.Size.Y/2)
			offset := mount.Sub(center).Rotated(tank.Rot)
			anchorWorld := engine.V(tank.Pos.X+center.X, tank.Pos.Y+center.Y).Add(offset)
			source.Pos = &engine.Vec{
				X: anchorWorld.X - source.RotAnchor.X,
				Y: anchorWorld.Y - source.RotAnchor.Y,
			}
			return
		}
		source.Pos = &engine.Vec{
			X: tank.Pos.X + tank.Size.X/2 - source.Size.X/2,
			Y: tank.Pos.Y + tank.Size.Y*0.28 - source.Size.Y/2,
		}
	}
}

func (s *GameScene) behaviorRotateOnButton(source *engine.Sprite) bool {
	if shouldAdjustCannon(ebiten.KeyArrowLeft) {
		s.playEventSound(soundEventCannonRotateLeft)
		source.Rot -= engine.DegToRad(s.humanCannonStep())
		return true
	}
	if shouldAdjustCannon(ebiten.KeyArrowRight) {
		s.playEventSound(soundEventCannonRotateRight)
		source.Rot += engine.DegToRad(s.humanCannonStep())
		return true
	}
	return false
}

func (s *GameScene) humanCannonStep() float64 {
	if shiftPressed() {
		return humanCannonStepDegrees * 10
	}
	return humanCannonStepDegrees
}

func (s *GameScene) adjustTankCannon(tank *battleTank, degrees float64) {
	if tank == nil || tank.cannon == nil {
		return
	}
	if degrees < 0 {
		s.playEventSound(soundEventCannonRotateLeft)
	} else if degrees > 0 {
		s.playEventSound(soundEventCannonRotateRight)
	}
	tank.cannon.Rot += engine.DegToRad(degrees)
	s.wrapCannonRotationToTank(tank.cannon, tank.body)
	s.updateXMV12FacingFromCannon(tank)
	s.syncOnlineAim(tank)
}

func (s *GameScene) behaviorRotateActiveCannon(source *engine.Sprite) {
	if s.activePlayerIndex < 0 || s.activePlayerIndex >= len(s.tanks) {
		return
	}
	tank := s.tanks[s.activePlayerIndex]
	if tank == nil || tank.cannon != source {
		return
	}
	if tank.player.Kind == PlayerComputer {
		s.clampCannonRotationToTank(source, tank.body)
		return
	}
	if !s.onlineCanControlPlayer(tank.playerIndex) {
		s.clampCannonRotationToTank(source, tank.body)
		return
	}
	if !s.activePlayerCanAdjustShot() {
		return
	}
	if s.scrollOMatActive() {
		return
	}
	if s.behaviorRotateOnButton(source) {
		s.wrapCannonRotationToTank(source, tank.body)
	} else {
		s.clampCannonRotationToTank(source, tank.body)
	}
	s.updateXMV12FacingFromCannon(tank)
	s.syncOnlineAim(tank)
}

func (s *GameScene) resetComputerTurnPlans() {
	for _, tank := range s.tanks {
		if tank != nil {
			tank.computerPlan = nil
		}
	}
}

func (s *GameScene) clampCannonRotationToTank(cannon, tank *engine.Sprite) {
	if cannon == nil || tank == nil {
		return
	}
	minRot := tank.Rot - math.Pi
	maxRot := tank.Rot
	cannon.Rot = math.Max(minRot, math.Min(maxRot, cannon.Rot))
}

func (s *GameScene) wrapCannonRotationToTank(cannon, tank *engine.Sprite) {
	if cannon == nil || tank == nil {
		return
	}
	minRot := tank.Rot - math.Pi
	maxRot := tank.Rot
	if cannon.Rot < minRot {
		cannon.Rot = maxRot
		return
	}
	if cannon.Rot > maxRot {
		cannon.Rot = minRot
	}
}
