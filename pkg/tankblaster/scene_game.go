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
	layerGround
	layerTanks
	layerProjectiles
	numLayers
)

const gameHUDHeight = 132

const weaponbarSlotCount = 20

const (
	projectileRadius         = 4
	impactRadiusMultiplier   = 4
	directHitCreditBonus     = 4000
	impactSplashMinDamage    = 10
	impactSplashMaxDamage    = 40
	sandFallFrames           = 12
	fallDamageStepPixels     = 20
	fallDamagePerStep        = 10
	zeroPowerFrames          = 216
	zeroPowerDissolveFrames  = 12
	creditsPerScorePoint     = 500
	debugShopStartingCredits = 20000
	roundTransitionSeconds   = 5
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

type computerTurnPlan struct {
	phase          computerTurnPhase
	delay          int
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
	pos      engine.Vec
	radius   float64
	age      int
	duration int
	cycles   int
	outward  bool
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

	layers     engine.Layers
	ground     models.Ground
	worldWidth float64
	cameraX    float64
	cameraGoal float64
	rng        *rand.Rand

	tanks                []*battleTank
	players              []PlayerConfig
	scores               []int
	roundScores          []int
	credits              []int
	inventories          []shopInventory
	computerMemories     []computerplayers.Memory
	roundNumber          int
	spawnIndex           int
	spawnPauseFrames     int
	activePlayerIndex    int
	wind                 int
	windDirection        int
	projectile           *projectile
	impacts              []impactAnimation
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
		computer:            mustImageFromPNG(r.PlayerComputer),
		storeBg:             mustImageFromPNG(r.StoreBackground),
		storeMainLeft:       mustImageFromPNG(r.StoreMainLeft),
		storeMainRight:      mustImageFromPNG(r.StoreMainRight),
		storeRoll:           mustImageFromPNG(r.StoreRoll),
		storeIcons:          mustImageFromPNG(r.StoreIcons),
		weaponbarActive:     mustImageFromPNG(r.WeaponbarActive),
		weaponbarOnStock:    mustImageFromPNG(r.WeaponbarOnStock),
		weaponbarOutOfStock: mustImageFromPNG(r.WeaponbarOutOfStock),
	}
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
	s.cameraGoal = s.cameraTargetForTank(0)
	s.layers[layerGround] = engine.AddSprites(s.layers[layerGround], gr.Sprites)
	s.layers[layerBackground] = engine.AddSprites(s.layers[layerBackground], b.Sprites)
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
		s.updateSandFalls()
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
		decision := computerplayers.Decide(tank.player.ComputerID, s.computerPlayerState(tank), s.rng)
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
		if plan.delay > 0 {
			plan.delay--
			return
		}
		if tank.shotStrength < plan.targetStrength {
			tank.shotStrength++
			plan.delay = 3
			return
		}
		if tank.shotStrength > plan.targetStrength {
			tank.shotStrength--
			plan.delay = 3
			return
		}
		plan.phase = computerTurnAdjustAngle
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
			Index: tank.playerIndex,
			X:     center.X,
			Y:     center.Y,
			Power: tank.power,
			Alive: tank.power > 0 && tank.landed && !tank.falling,
		})
	}
	return state
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
		step := math.Max(1, float64(tank.cannon.RotationSpeed))
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
		computerID:  tank.player.ComputerID,
		targetIndex: tank.computerPlan.decision.TargetIndex,
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
		TargetX: record.targetX,
		ImpactX: impact.X,
		Hit:     directHit && hitPlayerIndex == record.targetIndex,
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
	if p.pos.X < -80 || p.pos.X > s.worldWidth+80 || p.pos.Y > battlefieldHeight+80 || p.pos.Y < -battlefieldHeight {
		s.reportComputerShot(p.pos, -1, false)
		s.finishProjectile()
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
	if !weapon.DamagesTerrain {
		s.reportComputerShot(p.pos, -1, false)
		return false
	}

	radius := impactRadiusForWeapon(weapon)
	duration := s.impactAnimationFramesForWeapon(weapon)
	s.zeroPowerStartDelay = (duration * 2) / 3
	defer func() {
		s.zeroPowerStartDelay = 0
	}()
	s.damageTanksInImpactRadius(p.pos, radius)
	s.reportComputerShot(p.pos, -1, false)
	falls := s.ground.ApplyCrater(p.pos.X, p.pos.Y, radius)
	if len(falls) > 0 {
		s.sandFalls = append(s.sandFalls, sandFallAnimation{
			pixels:   falls,
			duration: sandFallFrames,
		})
	}
	s.dropUnsupportedTanks()
	s.impacts = append(s.impacts, impactAnimation{
		pos:      p.pos,
		radius:   radius,
		duration: duration,
		cycles:   impactCyclesForWeapon(weapon),
		outward:  weapon.ImpactGradientOutward,
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
		if impact.age < impact.duration {
			active = append(active, impact)
		}
	}
	s.impacts = active
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

	active := s.activeTank()
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
	s.drawHUDStepper(screen, image.Rect(10, hud.Min.Y+50, 122, hud.Min.Y+80), "Winkel", int(math.Round(s.cannonAngleDegrees())))

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
