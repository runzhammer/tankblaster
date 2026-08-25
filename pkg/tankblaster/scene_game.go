package tankblaster

import (
	"bytes"
	"image"
	"image/color"
	"math"
	"math/rand"
	"strconv"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	_ "image/jpeg"
	_ "image/png"

	"github.com/runzhammer/gamedemo/pkg/core"
	"github.com/runzhammer/gamedemo/pkg/engine"
	"github.com/runzhammer/gamedemo/pkg/models"
	"github.com/runzhammer/gamedemo/pkg/tankblaster/computerplayers"
	weaponspkg "github.com/runzhammer/gamedemo/pkg/tankblaster/weapons"
	r "github.com/runzhammer/gamedemo/resources"
	"golang.org/x/image/colornames"
)

var _ core.Scene = (*GameScene)(nil)

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
	layerPalms
	layerTanks
	layerProjectiles
	numLayers
)

const gameHUDHeight = 132

const weaponbarSlotCount = 20

const (
	projectileRadius          = 4
	impactRadiusMultiplier    = 4
	directHitCreditBonus      = 4000
	impactSplashMinDamage     = 10
	impactSplashMaxDamage     = 40
	sandFallFrames            = 12
	fallDamageStepPixels      = 20
	fallDamagePerStep         = 10
	zeroPowerFrames           = 216
	zeroPowerDissolveFrames   = 12
	creditsPerScorePoint      = 500
	debugShopStartingCredits  = 20000
	roundTransitionSeconds    = 5
	computerAdjustSpeedFactor = 1.6
)

const (
	plasmaImpactVisualYOffset = 38
	plasmaRingSpacing         = 2
	plasmaRingThickness       = 1
	plasmaBuildProgress       = 2.5 / 6.5
	plasmaGreenProgress       = 5.0 / 6.5
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
	selectedWeapon int
	shotStrength   int
	computerPlan   *computerTurnPlan
}

type battlePalm struct {
	sprite *engine.Sprite
	pixels *image.RGBA
	state  palmState
	age    int
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
	sprite *engine.Sprite
	speed  float64
	kind   cloudKind
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
	pos         engine.Vec
	prev        engine.Vec
	velocity    engine.Vec
	weaponIndex int
	trail       []engine.Vec
}

type impactAnimation struct {
	pos            engine.Vec
	radius         float64
	age            int
	duration       int
	cycles         int
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
}

type sandFallAnimation struct {
	pixels   []models.SandFallPixel
	age      int
	duration int
}

type zeroPowerAnimation struct {
	tank      *battleTank
	animation spriteAnimation
	delay     int
	age       int
	duration  int
}

type spriteAnimation struct {
	frames     []*ebiten.Image
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

	layers               engine.Layers
	ground               models.Ground
	worldWidth           float64
	cameraX              float64
	cameraGoal           float64
	rng                  *rand.Rand
	palmImage            *ebiten.Image
	palmPixels           *image.RGBA
	palmSkeletonImage    *ebiten.Image
	palmSkeletonPixels   *image.RGBA
	palmFireAnimation    spriteAnimation
	palmSmokeAnimation   spriteAnimation
	palmCrumbleAnimation spriteAnimation
	fireballAnimation    spriteAnimation
	clouds               []*battleCloud
	cloudAssets          []cloudAsset

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
	projectile           *projectile
	impacts              []impactAnimation
	animatedImpacts      []animatedImpact
	sandFalls            []sandFallAnimation
	zeroPowerEffects     []zeroPowerAnimation
	zeroPowerAnimations  []spriteAnimation
	shop                 shopAssets
	turnAdvanceDelay     int
	roundTransitionDelay int
	roundSeriesComplete  bool
	showScoreTable       bool
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
	scrollBarDragging    bool
	zeroPowerStartDelay  int
	lastComputerShot     computerShotRecord
	palmCameraFocus      *battlePalm
}

func NewGameScene(game *GameLoop) (core.Scene, error) {
	// loader := game.context.Loader()

	s := &GameScene{
		g:                 game,
		phase:             phaseBattle,
		rng:               rand.New(rand.NewSource(time.Now().UnixNano())),
		activePlayerIndex: -1,
		roundNumber:       1,
	}
	zeroPowerAnimations, err := loadZeroPowerAnimations()
	if err != nil {
		return nil, err
	}
	s.zeroPowerAnimations = zeroPowerAnimations
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
	s.cloudAssets = loadCloudAssets()
	s.players = s.playersForRound()
	s.scores = make([]int, len(s.players))
	s.roundScores = make([]int, len(s.players))
	s.credits = make([]int, len(s.players))
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
	for i := range s.credits {
		if s.credits[i] < debugShopStartingCredits {
			s.credits[i] = debugShopStartingCredits
		}
	}
	s.beginShop()
	return s, nil
}

func (s *GameScene) startRound() {
	s.phase = phaseBattle
	s.layers = engine.NewLayers(numLayers)
	s.spawnIndex = 0
	s.spawnPauseFrames = 0
	s.activePlayerIndex = -1
	s.projectile = nil
	s.impacts = nil
	s.sandFalls = nil
	s.zeroPowerEffects = nil
	s.turnAdvanceDelay = 0
	s.roundTransitionDelay = 0
	s.roundSeriesComplete = false
	s.lastDamageSource = nil
	s.lastComputerShot = computerShotRecord{}
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

	s.worldWidth = worldWidthForPlayers(len(s.players))

	battlefieldHeight := s.battlefieldHeight()
	b := models.NewBackgroundWithSize(s.worldWidth, battlefieldHeight)
	gr := models.NewRandomGroundWithSize(s.worldWidth, battlefieldHeight, s.rng.Int63())
	s.ground = gr
	s.tanks = nil
	s.palms = nil
	s.clouds = nil
	s.animatedImpacts = nil

	for tankIndex, player := range s.players {
		tank := models.NewTank(player.Name, player.Color)
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
	s.cameraGoal = s.cameraTargetForTank(0)
	s.layers[layerGround] = engine.AddSprites(s.layers[layerGround], gr.Sprites)
	s.layers[layerPalms] = engine.AddSprites(s.layers[layerPalms], s.palmSprites())
	s.layers[layerClouds] = engine.AddSprites(s.layers[layerClouds], s.cloudSprites())
	s.layers[layerBackground] = engine.AddSprites(s.layers[layerBackground], b.Sprites)
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
	minY := 20.0
	maxY := math.Max(minY, battlefieldHeight*0.28)
	laneWidth := s.worldWidth / float64(count)

	for i := 0; i < count; i++ {
		asset := s.cloudAssets[s.rng.Intn(len(s.cloudAssets))]
		scale := 0.55 + s.rng.Float64()*0.45
		size := engine.V(asset.size.X*scale, asset.size.Y*scale)
		laneCenter := laneWidth*float64(i) + laneWidth/2
		x := laneCenter - size.X/2 + (s.rng.Float64()-0.5)*laneWidth*0.5
		x = math.Max(-size.X, math.Min(s.worldWidth, x))
		y := minY + s.rng.Float64()*(maxY-minY)
		speed := (cfg.MinSpeed + s.rng.Float64()*speedRange) * windBoost
		cloud := &battleCloud{
			kind:  asset.kind,
			speed: speed,
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
		if cloud != nil && cloud.sprite != nil {
			sprites.Add(cloud.sprite)
		}
	}
	return sprites
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
		pixels: s.palmPixels,
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

func (s *GameScene) plantPalmAtImpact(pos engine.Vec) {
	if s.palmImage == nil || s.palmPixels == nil {
		return
	}
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
	palm.state = palmStateBurning
	palm.age = 0
}

func (s *GameScene) crumblePalm(palm *battlePalm) {
	if palm == nil || palm.state == palmStateCrumbling {
		return
	}
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

func (s *GameScene) palmEffectDelayFrames(palm *battlePalm) int {
	if palm == nil {
		return s.palmHitPauseFrames()
	}
	pause := secondsToFrames(0.5)
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
		s.updateSandFalls()
		s.updatePalms()
		s.updateZeroPowerEffects()
		if s.roundTransitionDelay > 0 || s.roundSeriesComplete {
			s.updateRoundTransition()
			return nil
		}
		if s.allTanksLanded() {
			if s.turnAdvanceDelay > 0 {
				s.updateTurnAdvanceDelay()
			} else {
				s.clampActiveShotStrength()
				s.handleBattleInput()
				s.handleComputerTurn()
				s.updateProjectile()
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
	s.cameraX = approach(s.cameraX, s.cameraGoal, 0.08, 0.35)

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
	}
}

func (s *GameScene) cameraTargetForTank(index int) float64 {
	if index < 0 || index >= len(s.tanks) || s.tanks[index] == nil || s.tanks[index].body == nil {
		return s.cameraX
	}

	screenWidth := core.Config().Screen.Width
	body := s.tanks[index].body
	centerX := body.Pos.X + body.Size.X/2
	return math.Max(0, math.Min(s.worldWidth-screenWidth, centerX-screenWidth/2))
}

func approach(current, target, smoothing, minStep float64) float64 {
	delta := target - current
	if math.Abs(delta) <= minStep {
		return target
	}
	return current + delta*smoothing
}

type zeroPowerAnimationSheet struct {
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
			animation.delays = append(animation.delays, ticks)
			animation.totalTicks += ticks
		}
	}
	if animation.totalTicks <= 0 {
		animation.totalTicks = zeroPowerFrames
	}
	return animation, nil
}

func loadZeroPowerAnimations() ([]spriteAnimation, error) {
	sources := []zeroPowerAnimationSheet{
		{data: r.ZeroPowerDustExplosionPNG, frameWidth: 20, delay: 6, scaleX: 0.5, scaleY: 0.5},
		{data: r.ZeroPowerExplosionPNG, frameWidth: 67, frameHeight: 64, delay: 6},
		{data: r.ZeroPowerMushroomExplosionPNG, frameWidth: 51, frameHeight: 57, delay: 6, anchor: zeroPowerAnchorTankBottom},
		{data: r.ZeroPowerPlayerSmokePNG, frameWidth: 21, delay: 6, scaleX: 1.0, scaleY: 1.3},
	}
	animations := make([]spriteAnimation, 0, len(sources))
	for _, source := range sources {
		animation, err := loadSpriteAnimation(source)
		if err != nil {
			return nil, err
		}
		if len(animation.frames) > 0 {
			animations = append(animations, animation)
		}
	}
	return animations, nil
}

func loadPalmAsset() (*ebiten.Image, *image.RGBA, error) {
	return loadImageWithPixels(r.Palm)
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
	animation.frames = make([]*ebiten.Image, 0, len(frames)*repeats)
	for i := 0; i < repeats; i++ {
		animation.frames = append(animation.frames, frames...)
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
	camera.Translate(-s.cameraX, 0)

	s.layers.Draw(&camera, screen)
	s.drawImpacts(screen, &camera)
	s.drawAnimatedImpacts(screen, &camera)
	s.drawPalmEffects(screen, &camera)
	s.drawSandFalls(screen, &camera)
	s.drawZeroPowerEffects(screen, &camera)
	s.drawProjectile(screen, &camera)
	s.drawGameHUD(screen)
	s.drawDebugScrollBar(screen)
	s.drawScoreTable(screen)
	s.drawRoundTransitionBanner(screen)

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
	if !s.scrollBarAvailable() {
		return
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		s.scrollBarDragging = image.Pt(x, y).In(debugScrollBarRect())
	}
	if !ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		s.scrollBarDragging = false
	}
	if s.scrollBarDragging {
		x, _ := ebiten.CursorPosition()
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
	if s.projectile != nil || s.activePlayerIndex < 0 || s.activePlayerIndex >= len(s.tanks) {
		return
	}
	tank := s.activeTank()
	if tank == nil {
		return
	}
	if tank.player.Kind == PlayerComputer {
		return
	}

	if shouldAdjustStrength(ebiten.KeyArrowUp) {
		tank.shotStrength = minInt(s.maxShotStrength(), tank.shotStrength+1)
	}
	if shouldAdjustStrength(ebiten.KeyArrowDown) {
		tank.shotStrength = maxInt(s.minShotStrength(), tank.shotStrength-1)
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		for i := 0; i < s.weaponSlotCount(); i++ {
			if image.Pt(x, y).In(s.weaponSlotRect(i)) && s.canSelectWeaponSlot(tank, i) {
				tank.selectedWeapon = i
				break
			}
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeyKPEnter) {
		s.fireActiveWeapon()
	}
}

func (s *GameScene) handleComputerTurn() {
	if s.projectile != nil || s.activePlayerIndex < 0 || s.activePlayerIndex >= len(s.tanks) {
		return
	}
	tank := s.activeTank()
	if tank == nil || tank.player.Kind != PlayerComputer || !s.tankCanAct(tank) {
		return
	}
	if tank.computerPlan == nil {
		decision := computerplayers.Decide(s.effectiveComputerID(tank), s.computerPlayerState(tank), s.rng)
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
		if math.Abs(s.cameraX-s.cameraGoal) > 2 {
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
			tank.shotStrength++
			return
		}
		if tank.shotStrength > plan.targetStrength {
			tank.shotStrength--
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
		s.fireActiveWeapon()
	}
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
	const maxComputerProjectileSlot = 3
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
			return false
		}
		if delta < 0 {
			step = -step
		}
		tank.cannon.Rot += engine.DegToRad(step)
		s.clampCannonRotationToTank(tank.cannon, tank.body)
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

func (s *GameScene) canSelectWeaponSlot(tank *battleTank, slot int) bool {
	if tank == nil || slot < 0 {
		return false
	}
	if slot == 0 {
		return true
	}
	return s.ammoForWeaponSlot(tank.playerIndex, slot) > 0
}

func (s *GameScene) consumeSelectedWeaponAmmo(tank *battleTank) bool {
	if tank == nil {
		return false
	}
	slot := tank.selectedWeapon
	if slot == 0 {
		return true
	}
	itemIndex := s.itemIndexForWeaponSlot(slot)
	if itemIndex < 0 || s.shopItemCountForPlayer(tank.playerIndex, itemIndex) <= 0 {
		return false
	}
	if s.isScrollOMatItem(itemIndex) {
		return true
	}
	s.ensureInventory(tank.playerIndex)
	if itemIndex < len(s.inventories[tank.playerIndex].classA) && s.inventories[tank.playerIndex].classA[itemIndex] > 0 {
		s.inventories[tank.playerIndex].classA[itemIndex]--
		return true
	}
	if itemIndex < len(s.inventories[tank.playerIndex].classB) && s.inventories[tank.playerIndex].classB[itemIndex] > 0 {
		s.inventories[tank.playerIndex].classB[itemIndex]--
		return true
	}
	return false
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
	if !s.consumeSelectedWeaponAmmo(tank) {
		return
	}

	bounds := tank.cannon.Bounds()
	muzzle := bounds.Center().Add(engine.V(bounds.W()/2+7, 0).Rotated(tank.cannon.Rot))
	speed := 1.4 + float64(tank.shotStrength)*0.32
	s.projectile = &projectile{
		pos:         *muzzle,
		prev:        *muzzle,
		velocity:    engine.V(speed, 0).Rotated(tank.cannon.Rot),
		weaponIndex: tank.selectedWeapon,
		trail:       []engine.Vec{*muzzle},
	}
	s.lastDamageSource = tank
	if tank.player.Kind == PlayerComputer {
		s.lastComputerShot = s.computerShotRecordFor(tank)
	}
	tank.computerPlan = nil
}

func (s *GameScene) updateProjectile() {
	if s.projectile == nil {
		return
	}

	const gravity = 0.16
	windAcceleration := float64(s.windDirection*s.wind) * 0.00065

	p := s.projectile
	p.prev = p.pos
	p.velocity.X += windAcceleration
	p.velocity.Y += gravity
	p.pos = *p.pos.Add(p.velocity)
	p.trail = append(p.trail, p.pos)
	if len(p.trail) > 260 {
		p.trail = p.trail[len(p.trail)-260:]
	}

	if p.pos.X >= 0 && p.pos.X <= s.worldWidth {
		s.cameraGoal = math.Max(0, math.Min(s.worldWidth-core.Config().Screen.Width, p.pos.X-core.Config().Screen.Width/2))
		s.cameraX = approach(s.cameraX, s.cameraGoal, 0.12, 0.4)
	}

	battlefieldHeight := s.battlefieldHeight()
	if p.pos.X < -80 || p.pos.X > s.worldWidth+80 || p.pos.Y > battlefieldHeight+80 {
		s.reportComputerShot(p.pos, -1, false)
		s.finishProjectile()
		return
	}

	weapon := s.weaponForProjectile(p)
	if palm := s.projectileHitsPalm(p, projectileRadiusForWeapon(weapon)); palm != nil {
		s.reportComputerShot(p.pos, -1, false)
		if weapon.ImpactAnimationStyle == weaponspkg.ImpactAnimationFireball && palm.state == palmStateAlive {
			s.ignitePalm(palm)
		} else if palm.state == palmStateSkeleton || palm.state == palmStateSkeletonSmoking {
			s.crumblePalm(palm)
		}
		s.projectile = nil
		s.focusPalmCamera(palm)
		s.delayTurnAdvance(s.palmEffectDelayFrames(palm))
		return
	}

	if p.pos.Y >= s.ground.SurfaceY(p.pos.X) {
		if s.onGroundImpact(p) {
			s.projectile = nil
			return
		}
		s.finishProjectile()
		return
	}

	hitRadius := projectileRadiusForWeapon(s.weaponForProjectile(p))
	hitBounds := engine.R(p.pos.X-hitRadius, p.pos.Y-hitRadius, p.pos.X+hitRadius, p.pos.Y+hitRadius)
	for _, tank := range s.tanks {
		if tank == nil || tank.body == nil {
			continue
		}
		if engine.Collision(hitBounds, tank.body.Bounds().ScaledAtCenter(0.78)) {
			if damage := s.weaponForProjectile(p).Damage; damage > 0 {
				s.damageTank(tank, damage, s.lastDamageSource, damageCauseDirect)
				s.awardDirectHitCredits(tank, s.lastDamageSource)
				s.darkenTank(tank, 0.10)
			}
			s.reportComputerShot(p.pos, tank.playerIndex, true)
			s.projectile = nil
			s.delayTurnAdvance(s.tankHitPauseFrames())
			return
		}
	}
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

func (s *GameScene) finishProjectile() {
	s.projectile = nil
	if len(s.tanks) == 0 {
		return
	}
	s.advanceActivePlayer()
}

func (s *GameScene) advanceActivePlayer() {
	if len(s.tanks) == 0 {
		return
	}
	if s.endRoundIfOnlyOneTankRemains() {
		return
	}
	s.lastDamageSource = nil
	s.palmCameraFocus = nil
	next := s.nextActivePlayerIndex()
	if next < 0 {
		return
	}
	s.activePlayerIndex = next
	s.resetComputerTurnPlans()
	s.clampActiveShotStrength()
	s.cameraGoal = s.cameraTargetForTank(s.activePlayerIndex)
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
	if !weapon.DamagesTerrain && !weapon.PlantsPalm && weapon.ImpactAnimationStyle != weaponspkg.ImpactAnimationFireball {
		s.reportComputerShot(p.pos, -1, false)
		return false
	}

	if weapon.ImpactAnimationStyle == weaponspkg.ImpactAnimationFireball {
		s.reportComputerShot(p.pos, -1, false)
		s.startFireballImpact(p.pos, weapon)
		return true
	}

	if weapon.PlantsPalm {
		s.reportComputerShot(p.pos, -1, false)
		s.plantPalmAtImpact(p.pos)
		s.delayTurnAdvance(s.palmHitPauseFrames())
		return true
	}

	radius := impactRadiusForWeapon(weapon)
	duration := s.impactAnimationFramesForWeapon(weapon)
	s.zeroPowerStartDelay = (duration * 2) / 3
	defer func() {
		s.zeroPowerStartDelay = 0
	}()
	s.reportComputerShot(p.pos, -1, false)

	impactPos := p.pos
	terrainApplied := true
	if weapon.ImpactAnimationStyle == weaponspkg.ImpactAnimationPlasma {
		impactPos.Y -= plasmaImpactVisualYOffset
		terrainApplied = false
		s.damageTanksInImpactRadiusFixed(impactPos, radius, weapon.Damage)
	} else {
		s.damageTanksInImpactRadius(p.pos, radius)
		falls := s.ground.ApplyCrater(p.pos.X, p.pos.Y, radius)
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
		outward:        weapon.ImpactGradientOutward,
		style:          weapon.ImpactAnimationStyle,
		terrainApplied: terrainApplied,
	})
	if minimumDelay := duration + s.impactPauseFrames(); s.turnAdvanceDelay < minimumDelay {
		s.turnAdvanceDelay = minimumDelay
	}
	return true
}

func (s *GameScene) damageTanksInImpactRadius(center engine.Vec, radius float64) {
	if radius <= 0 {
		return
	}
	for _, tank := range s.tanks {
		if tank == nil || tank.body == nil || tank.power <= 0 {
			continue
		}
		distance := distancePointToRect(center, tank.body.Bounds().ScaledAtCenter(0.78))
		if distance > radius {
			continue
		}
		damage := impactSplashMinDamage + int(math.Round(float64(impactSplashMaxDamage-impactSplashMinDamage)*(1-distance/radius)))
		damage = maxInt(impactSplashMinDamage, minInt(impactSplashMaxDamage, damage))
		s.damageTank(tank, damage, s.lastDamageSource, damageCauseDirect)
	}
}

func (s *GameScene) damageTanksInImpactRadiusFixed(center engine.Vec, radius float64, damage int) {
	if radius <= 0 || damage <= 0 {
		return
	}
	for _, tank := range s.tanks {
		if tank == nil || tank.body == nil || tank.power <= 0 {
			continue
		}
		distance := distancePointToRect(center, tank.body.Bounds().ScaledAtCenter(0.78))
		if distance > radius {
			continue
		}
		s.damageTank(tank, damage, s.lastDamageSource, damageCauseDirect)
	}
}

func (s *GameScene) startFireballImpact(pos engine.Vec, weapon weaponspkg.Weapon) {
	animation := s.fireballAnimation
	if len(animation.frames) == 0 {
		return
	}
	duration := maxInt(animation.totalTicks, s.impactAnimationFramesForWeapon(weapon))
	damage := weapon.ImpactDamage
	if damage <= 0 {
		damage = 40
	}
	effect := animatedImpact{
		pos:       pos,
		duration:  duration,
		animation: animation,
		damage:    damage,
	}
	s.animatedImpacts = append(s.animatedImpacts, effect)
	s.damageTanksInRectFixed(s.animationWorldRect(pos, animation), damage)
	if minimumDelay := duration + s.impactPauseFrames(); s.turnAdvanceDelay < minimumDelay {
		s.turnAdvanceDelay = minimumDelay
	}
}

func (s *GameScene) damageTanksInRectFixed(rect engine.Rect, damage int) {
	if damage <= 0 {
		return
	}
	for _, tank := range s.tanks {
		if tank == nil || tank.body == nil || tank.power <= 0 {
			continue
		}
		if !rectsIntersect(tank.body.Bounds().ScaledAtCenter(0.78), rect) {
			continue
		}
		s.damageTank(tank, damage, s.lastDamageSource, damageCauseDirect)
	}
}

func rectsIntersect(a, b engine.Rect) bool {
	return a.Min.X <= b.Max.X && a.Max.X >= b.Min.X && a.Min.Y <= b.Max.Y && a.Max.Y >= b.Min.Y
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
		if impact.style == weaponspkg.ImpactAnimationPlasma && !impact.terrainApplied {
			progress := float64(impact.age) / math.Max(1, float64(impact.duration))
			if progress >= plasmaGreenProgress {
				falls := s.ground.ApplyRingCrater(impact.pos.X, impact.pos.Y, impact.radius, plasmaRingSpacing, plasmaRingThickness)
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
		if impact.age < impact.duration {
			active = append(active, impact)
		}
	}
	s.impacts = active
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
		}
	}
	s.animatedImpacts = active
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
		s.updateZeroPowerTankDissolve(effect)
		effect.age++
		if effect.age < effect.duration {
			active = append(active, effect)
		}
	}
	s.zeroPowerEffects = active
}

func (s *GameScene) updateTurnAdvanceDelay() {
	s.turnAdvanceDelay--
	if s.turnAdvanceDelay > 0 {
		if s.updatePalmCamera() {
			return
		}
		s.updateZeroPowerCamera()
		return
	}
	s.turnAdvanceDelay = 0
	if s.endRoundIfOnlyOneTankRemains() {
		return
	}
	s.advanceActivePlayer()
}

func (s *GameScene) updateBattleCamera() {
	if s.updatePalmCamera() {
		return
	}
	if s.updateZeroPowerCamera() {
		return
	}
	if s.scrollOMatActive() || s.scrollBarDragging {
		return
	}
	if s.projectile != nil || s.activePlayerIndex < 0 {
		return
	}
	s.cameraGoal = s.cameraTargetForTank(s.activePlayerIndex)
	s.cameraX = approach(s.cameraX, s.cameraGoal, 0.08, 0.35)
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
	s.cameraX = approach(s.cameraX, s.cameraGoal, 0.10, 0.45)
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

func (s *GameScene) updateZeroPowerCamera() bool {
	if len(s.zeroPowerEffects) == 0 {
		return false
	}
	effect := s.zeroPowerEffects[0]
	for index, tank := range s.tanks {
		if tank == effect.tank {
			s.cameraGoal = s.cameraTargetForTank(index)
			s.cameraX = approach(s.cameraX, s.cameraGoal, 0.12, 0.4)
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
	if s.maxShotStrength() <= 0 {
		return 0
	}
	return 1
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
	if tank == nil || damage <= 0 {
		return
	}
	previousPower := tank.power
	tank.power = maxInt(0, tank.power-damage)
	tank.shotStrength = minInt(tank.shotStrength, maxInt(0, tank.power))
	if previousPower > 0 && tank.power == 0 {
		s.awardZeroPowerScore(tank, attacker, cause)
		s.startZeroPowerAnimation(tank)
	}
}

func (s *GameScene) awardDirectHitCredits(target, attacker *battleTank) {
	if target == nil || attacker == nil || target == attacker {
		return
	}
	if attacker.playerIndex < 0 || attacker.playerIndex >= len(s.credits) {
		return
	}
	s.credits[attacker.playerIndex] += directHitCreditBonus
}

func (s *GameScene) awardZeroPowerScore(defeated, attacker *battleTank, cause damageCause) {
	if defeated == nil || attacker == nil {
		return
	}
	if defeated == attacker {
		s.addScore(defeated.playerIndex, -3)
		for _, tank := range s.tanks {
			if tank != nil && tank != defeated {
				s.addScore(tank.playerIndex, 1)
			}
		}
		return
	}
	switch cause {
	case damageCauseFall:
		s.addScore(attacker.playerIndex, 1)
	default:
		s.addScore(attacker.playerIndex, 3)
	}
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
	if len(s.zeroPowerAnimations) == 0 {
		return
	}
	animation := s.zeroPowerAnimations[s.rng.Intn(len(s.zeroPowerAnimations))]
	duration := maxInt(1, animation.totalTicks)
	delay := maxInt(0, s.zeroPowerStartDelay)
	s.zeroPowerEffects = append(s.zeroPowerEffects, zeroPowerAnimation{
		tank:      tank,
		animation: animation,
		delay:     delay,
		duration:  duration,
	})
	if s.turnAdvanceDelay < duration+delay {
		s.turnAdvanceDelay = duration + delay
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
		if tank == nil || tank.body == nil || !tank.landed || tank.falling {
			continue
		}
		targetY, stable := s.tankSupportState(tank.body)
		currentBottom := tank.body.Pos.Y + tank.body.Size.Y
		if stable {
			targetY = s.alignedTankBottomY(tank.body)
		}
		if targetY <= currentBottom+1 {
			startY := tank.body.Pos.Y
			s.ground.AlignSpriteToSurface(tank.body)
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
	copy := *tank
	pos := *tank.Pos
	copy.Pos = &pos
	s.ground.AlignSpriteToSurface(&copy)
	return copy.Pos.Y + copy.Size.Y
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
	if s.projectile == nil {
		return
	}

	weapon := s.weaponForProjectile(s.projectile)
	c := weapon.Color
	if weapon.ShowTrail {
		for i, point := range s.projectile.trail {
			if i%2 != 0 {
				continue
			}
			projected := point.Project(camera)
			alpha := uint8(70 + minInt(185, i*4))
			drawFilledRect(screen, image.Rect(int(projected.X)-2, int(projected.Y)-2, int(projected.X)+2, int(projected.Y)+2), color.RGBA{R: c.R, G: c.G, B: c.B, A: alpha})
		}
	}
	projected := s.projectile.pos.Project(camera)
	radius := projectileRadiusForWeapon(weapon)
	if weapon.RoundProjectile {
		vector.DrawFilledCircle(screen, float32(projected.X), float32(projected.Y), float32(radius), c, true)
		return
	}
	drawFilledRect(screen, image.Rect(int(projected.X-radius), int(projected.Y-radius), int(projected.X+radius), int(projected.Y+radius)), c)
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
		cycleProgress := math.Mod(progress*float64(cycles), 1)
		steps := 8
		for i := steps; i >= 1; i-- {
			t := float64(i) / float64(steps)
			radius := float32(impact.radius * t)
			red := uint8(255 * math.Pow(t, 0.7) * cycleProgress)
			vector.DrawFilledCircle(screen, float32(projected.X), float32(projected.Y), radius, color.RGBA{R: red, G: 0, B: 0, A: 220}, true)
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

func (s *GameScene) drawPalmEffects(screen *ebiten.Image, camera *ebiten.GeoM) {
	for _, palm := range s.palms {
		if palm == nil || palm.sprite == nil {
			continue
		}
		bounds := palm.sprite.Bounds()
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

func drawAnimationCentered(screen *ebiten.Image, camera *ebiten.GeoM, animation spriteAnimation, center engine.Vec, tick int) {
	drawAnimationFrame(screen, camera, animation, center, tick, 0.5, 0.5)
}

func drawAnimationBottomCentered(screen *ebiten.Image, camera *ebiten.GeoM, animation spriteAnimation, bottomCenter engine.Vec, tick int) {
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
	screenCfg := core.Config().Screen
	hud := image.Rect(0, int(s.battlefieldHeight()), int(screenCfg.Width), int(screenCfg.Height))
	drawFilledRect(screen, hud, color.RGBA{R: 5, G: 7, B: 10, A: 242})
	drawFilledRect(screen, image.Rect(hud.Min.X, hud.Min.Y, hud.Max.X, hud.Min.Y+2), color.RGBA{R: 245, G: 246, B: 214, A: 255})

	active := s.hudTank()
	playerName := "Spieler"
	playerColor := color.RGBA{R: 255, G: 160, B: 28, A: 255}
	power := 100
	shotStrength := 20
	if active != nil {
		playerName = active.player.Name
		playerColor = active.player.Color
		power = active.power
		shotStrength = active.shotStrength
	}

	s.drawHUDStepper(screen, image.Rect(10, hud.Min.Y+14, 112, hud.Min.Y+44), "Stärke", shotStrength)
	s.drawHUDStepper(screen, image.Rect(10, hud.Min.Y+50, 122, hud.Min.Y+80), "Winkel", int(math.Round(s.cannonAngleDegreesForTank(active))))

	centerX := int(screenCfg.Width) / 2
	drawText(screen, playerName, centerX-42, hud.Min.Y+30, playerColor)
	drawButton(screen, image.Rect(centerX-64, hud.Min.Y+44, centerX+64, hud.Min.Y+76), "Feuer!")

	windArrow := "->"
	if s.windDirection < 0 {
		windArrow = "<-"
	}
	rightX := int(screenCfg.Width) - 170
	drawText(screen, "Wind: "+strconv.Itoa(s.wind)+" ("+windArrow+")", rightX, hud.Min.Y+30, colornames.White)
	drawText(screen, "Power: "+strconv.Itoa(power), rightX, hud.Min.Y+55, colornames.White)

	for i := 0; i < s.weaponSlotCount(); i++ {
		s.drawWeaponSlot(screen, i)
	}
}

func (s *GameScene) drawScoreTable(screen *ebiten.Image) {
	if !s.showScoreTable {
		return
	}

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

	title := "Runde " + strconv.Itoa(maxInt(1, s.roundNumber)) + " von " + strconv.Itoa(maxInt(1, s.g.rounds))
	drawCenteredText(screen, title, image.Rect(left, top, right, top+24), colornames.Yellow)

	headerY := top + 60
	nameX := left + 80
	scoreX := left + tableW/2 + 60
	statusX := right - 110
	lineColor := color.RGBA{R: 250, G: 246, B: 230, A: 230}
	textColor := color.RGBA{R: 250, G: 246, B: 255, A: 255}

	drawText(screen, "Spieler", nameX, headerY, textColor)
	drawText(screen, "Erfolg", scoreX, headerY, textColor)
	drawText(screen, "Status", statusX, headerY, textColor)

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
		status := "aktiv"
		if tank.power <= 0 {
			status = "aus"
		}
		drawText(screen, tank.player.Name, nameX, y, textColor)
		drawText(screen, strconv.Itoa(s.scoreForPlayer(tank.playerIndex)), scoreX+28, y, textColor)
		drawText(screen, status, statusX+18, y, textColor)
	}
}

func (s *GameScene) scoreTableTanks() []*battleTank {
	if len(s.tanks) == 0 {
		return nil
	}
	rows := make([]*battleTank, 0, len(s.tanks))
	if s.activePlayerIndex >= 0 && s.activePlayerIndex < len(s.tanks) {
		rows = append(rows, s.tanks[s.activePlayerIndex])
	}
	for index, tank := range s.tanks {
		if index == s.activePlayerIndex {
			continue
		}
		rows = append(rows, tank)
	}
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
	textValue := strconv.Itoa(played) + " gespielt, noch " + strconv.Itoa(remaining) + " Runden."
	drawCenteredText(screen, textValue, r, colornames.White)
}

func (s *GameScene) updateRoundTransition() {
	if s.roundSeriesComplete {
		return
	}
	s.roundTransitionDelay--
	if s.roundTransitionDelay > 0 {
		return
	}
	s.roundTransitionDelay = 0
	if s.roundNumber >= maxInt(1, s.g.rounds) {
		s.roundSeriesComplete = true
		return
	}
	s.roundNumber++
	s.beginShop()
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

func insetRect(r image.Rectangle, inset int) image.Rectangle {
	return image.Rect(r.Min.X+inset, r.Min.Y+inset, r.Max.X-inset, r.Max.Y-inset)
}

func (s *GameScene) drawHUDStepper(screen *ebiten.Image, r image.Rectangle, label string, value int) {
	buttonW := 24
	drawButton(screen, image.Rect(r.Min.X, r.Min.Y, r.Min.X+buttonW, r.Min.Y+22), "-")
	drawButton(screen, image.Rect(r.Min.X+buttonW+4, r.Min.Y, r.Min.X+buttonW*2+4, r.Min.Y+22), "+")
	drawText(screen, label+": "+strconv.Itoa(value), r.Min.X+buttonW*2+12, r.Min.Y+17, colornames.White)
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
	return s.cannonAngleDegreesForTank(s.activeTank())
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

func (s *GameScene) battlefieldHeight() float64 {
	return math.Max(120, core.Config().Screen.Height-gameHUDHeight)
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
		if tank != nil && tank.fallDamage {
			landingY = math.Max(landingY, tank.fallTargetY)
		}
		if source.Pos.Y+source.Size.Y < landingY {
			return
		}

		source.Velocity = engine.Vec{}
		if tank != nil {
			if tank.fallDamage {
				source.Pos = &engine.Vec{X: source.Pos.X, Y: landingY - source.Size.Y}
				source.Rot = 0
			} else {
				ground.AlignSpriteToSurface(source)
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
	damage := int(fallDistance * float64(fallDamagePerStep) / float64(fallDamageStepPixels))
	damage = maxInt(1, damage)
	s.damageTank(tank, damage, s.lastDamageSource, damageCauseFall)
}

func (s *GameScene) spawnLandingPauseFrames() int {
	return secondsToFrames(core.Config().Gameplay.SpawnLandingPauseSeconds)
}

func (s *GameScene) impactAnimationFrames() int {
	return secondsToFrames(core.Config().Gameplay.ImpactAnimationSeconds)
}

func (s *GameScene) impactAnimationFramesForWeapon(weapon weaponspkg.Weapon) int {
	return secondsToFrames(core.Config().Gameplay.ImpactAnimationSeconds + weapon.ImpactAnimationExtraSeconds)
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
	s.activePlayerIndex = living[s.rng.Intn(len(living))]
	s.resetComputerTurnPlans()
}

func (s *GameScene) allTanksLanded() bool {
	if len(s.tanks) == 0 {
		return false
	}
	for _, tank := range s.tanks {
		if tank == nil || !tank.landed {
			return false
		}
	}
	return true
}

func (s *GameScene) tankCanAct(tank *battleTank) bool {
	return tank != nil && tank.power > 0 && tank.landed && !tank.falling
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
	s.creditRoundScores()
	s.roundTransitionDelay = secondsToFrames(roundTransitionSeconds)
	return true
}

func (s *GameScene) creditRoundScores() {
	if len(s.credits) < len(s.roundScores) {
		next := make([]int, len(s.roundScores))
		copy(next, s.credits)
		s.credits = next
	}
	for index, score := range s.roundScores {
		if score <= 0 {
			continue
		}
		s.credits[index] += score * creditsPerScorePoint
	}
}

func (s *GameScene) behaviorAttachCannonToTank(tank *engine.Sprite) engine.Behavior {
	return func(source *engine.Sprite) {
		if tank == nil {
			return
		}
		source.Pos = &engine.Vec{
			X: tank.Pos.X + tank.Size.X/2 - source.Size.X/2,
			Y: tank.Pos.Y + tank.Size.Y*0.28 - source.Size.Y/2,
		}
	}
}

func (s *GameScene) behaviorRotateOnButton(source *engine.Sprite) {
	if RotateLeft() {
		source.Rot -= engine.DegToRad(float64(source.RotationSpeed))
	} else if RotateRight() {
		source.Rot += engine.DegToRad(float64(source.RotationSpeed))
	}
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
	if s.projectile != nil || s.roundTransitionDelay > 0 || s.roundSeriesComplete {
		return
	}
	if s.scrollOMatActive() {
		return
	}
	s.behaviorRotateOnButton(source)
	s.clampCannonRotationToTank(source, tank.body)
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
