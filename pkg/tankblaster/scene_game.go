package tankblaster

import (
	"bytes"
	"image"
	"image/color"
	"image/gif"
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

const debugScrollMode = true

const gameHUDHeight = 132

const (
	projectileRadius           = 4
	impactRadiusMultiplier     = 4
	directHitDamage            = 100
	directHitCreditBonus       = 4000
	impactSplashMinDamage      = 10
	impactSplashMaxDamage      = 40
	largeGrenadeScale          = 1.8
	largeGrenadeImpactScale    = 3.0
	atomBombImpactScale        = 4.0
	atomBombImpactExtraSeconds = 0.5
	sandFallFrames             = 12
	fallDamageStepPixels       = 20
	fallDamagePerStep          = 10
	zeroPowerFrames            = 216
	creditsPerScorePoint       = 500
	debugShopStartingCredits   = 20000
	roundTransitionSeconds     = 5
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
	selectedWeapon int
	shotStrength   int
}

type weapon struct {
	name                        string
	color                       color.RGBA
	damage                      int
	unlocked                    bool
	showTrail                   bool
	roundProjectile             bool
	damagesTerrain              bool
	projectileScale             float64
	impactScale                 float64
	impactAnimationExtraSeconds float64
	impactCycles                int
	impactGradientOutward       bool
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
	tank     *battleTank
	age      int
	duration int
}

type gifAnimation struct {
	frames     []*ebiten.Image
	delays     []int
	totalTicks int
	width      int
	height     int
}

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
	zeroPowerSmoke       gifAnimation
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
	smoke, err := loadGIFAnimation(r.ZeroPowerSmokeGIF)
	if err != nil {
		return nil, err
	}
	s.zeroPowerSmoke = smoke
	s.shop = shopAssets{
		human:          mustImageFromPNG(r.PlayerHuman),
		computer:       mustImageFromPNG(r.PlayerComputer),
		trainingOn:     mustImageFromPNG(r.TrainingAmmoSelected),
		trainingOff:    mustImageFromPNG(r.TrainingAmmoDeselected),
		storeBg:        mustImageFromPNG(r.StoreBackground),
		storeMainLeft:  mustImageFromPNG(r.StoreMainLeft),
		storeMainRight: mustImageFromPNG(r.StoreMainRight),
		storeRoll:      mustImageFromPNG(r.StoreRoll),
		storeIcons:     mustImageFromPNG(r.StoreIcons),
	}
	s.players = s.playersForRound()
	s.scores = make([]int, len(s.players))
	s.roundScores = make([]int, len(s.players))
	s.credits = make([]int, len(s.players))
	s.inventories = makeShopInventories(len(s.players))
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
		s.handleDebugScroll()
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

func loadGIFAnimation(data []byte) (gifAnimation, error) {
	decoded, err := gif.DecodeAll(bytes.NewReader(data))
	if err != nil {
		return gifAnimation{}, err
	}

	animation := gifAnimation{
		frames: make([]*ebiten.Image, 0, len(decoded.Image)),
		delays: make([]int, 0, len(decoded.Image)),
	}
	for i, frame := range decoded.Image {
		ticks := 4
		if i < len(decoded.Delay) && decoded.Delay[i] > 0 {
			ticks = maxInt(1, int(math.Round(float64(decoded.Delay[i])*60/100)))
		}
		animation.frames = append(animation.frames, ebiten.NewImageFromImage(frame))
		animation.delays = append(animation.delays, ticks)
		animation.totalTicks += ticks
		if frame.Bounds().Dx() > animation.width {
			animation.width = frame.Bounds().Dx()
		}
		if frame.Bounds().Dy() > animation.height {
			animation.height = frame.Bounds().Dy()
		}
	}
	if animation.totalTicks <= 0 {
		animation.totalTicks = zeroPowerFrames
	}
	return animation, nil
}

func (a gifAnimation) frameAt(tick int) *ebiten.Image {
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

func (s *GameScene) handleDebugScroll() {
	if !debugScrollMode || s.worldWidth <= core.Config().Screen.Width {
		return
	}
	if !ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		return
	}

	x, y := ebiten.CursorPosition()
	bar := debugScrollBarRect()
	if !image.Pt(x, y).In(bar) {
		return
	}

	t := float64(x-bar.Min.X) / float64(bar.Dx())
	s.cameraX = math.Max(0, math.Min(s.worldWidth-core.Config().Screen.Width, t*(s.worldWidth-core.Config().Screen.Width)))
}

func (s *GameScene) drawDebugScrollBar(screen *ebiten.Image) {
	if !debugScrollMode || s.worldWidth <= core.Config().Screen.Width {
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

func (s *GameScene) handleBattleInput() {
	if s.projectile != nil || s.activePlayerIndex < 0 || s.activePlayerIndex >= len(s.tanks) {
		return
	}
	tank := s.activeTank()
	if tank == nil {
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

func (s *GameScene) weaponForProjectile(p *projectile) weapon {
	weapons := gameWeapons()
	if p != nil && p.weaponIndex == 0 {
		return weapons[0]
	}
	if p != nil && s.itemIndexForWeaponSlot(p.weaponIndex) == 1 {
		return weapons[2]
	}
	if p != nil && s.itemIndexForWeaponSlot(p.weaponIndex) == 2 {
		return weapons[3]
	}
	return weapons[1]
}

func (s *GameScene) fireActiveWeapon() {
	tank := s.activeTank()
	if tank == nil || tank.cannon == nil {
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
			if damage := s.weaponForProjectile(p).damage; damage > 0 {
				s.damageTank(tank, damage, s.lastDamageSource, damageCauseDirect)
				s.awardDirectHitCredits(tank, s.lastDamageSource)
				s.darkenTank(tank, 0.10)
			}
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
		s.advanceActivePlayer()
		return
	}
	s.turnAdvanceDelay = frames
}

func (s *GameScene) onGroundImpact(p *projectile) bool {
	if p == nil {
		return false
	}
	weapon := s.weaponForProjectile(p)
	if !weapon.damagesTerrain {
		return false
	}

	radius := impactRadiusForWeapon(weapon)
	duration := s.impactAnimationFramesForWeapon(weapon)
	s.damageTanksInImpactRadius(p.pos, radius)
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
		outward:  weapon.impactGradientOutward,
	})
	s.turnAdvanceDelay = duration + s.impactPauseFrames()
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
	duration := maxInt(1, s.zeroPowerSmoke.totalTicks)
	s.zeroPowerEffects = append(s.zeroPowerEffects, zeroPowerAnimation{
		tank:     tank,
		duration: duration,
	})
	if s.turnAdvanceDelay < duration {
		s.turnAdvanceDelay = duration
	}
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
	c := weapon.color
	if weapon.showTrail {
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
	if weapon.roundProjectile {
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
	if len(s.zeroPowerSmoke.frames) == 0 {
		return
	}
	for _, effect := range s.zeroPowerEffects {
		if effect.tank == nil || effect.tank.body == nil {
			continue
		}
		body := effect.tank.body.Bounds()
		center := body.Center()
		frame := s.zeroPowerSmoke.frameAt(effect.age)
		if frame == nil {
			continue
		}
		anchor := engine.V(center.X, body.Min.Y-26).Project(camera)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(
			anchor.X-float64(s.zeroPowerSmoke.width)/2,
			anchor.Y-float64(s.zeroPowerSmoke.height)/2,
		)
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
	weapons := gameWeapons()
	shopItemIndex := s.itemIndexForWeaponSlot(index)
	unlocked := index < len(weapons) && weapons[index].unlocked
	active := s.activeTank()
	if index > 0 {
		unlocked = active != nil && s.shopItemCountForPlayer(active.playerIndex, shopItemIndex) > 0
	}
	fill := color.RGBA{R: 28, G: 32, B: 37, A: 255}
	border := color.RGBA{R: 94, G: 101, B: 110, A: 255}
	if !unlocked {
		fill = color.RGBA{R: 34, G: 35, B: 37, A: 255}
		border = color.RGBA{R: 52, G: 55, B: 60, A: 255}
	}
	if unlocked && active != nil && index == active.selectedWeapon {
		border = color.RGBA{R: 236, G: 58, B: 63, A: 255}
	}

	drawFrame(screen, r, fill, border)
	inner := image.Rect(r.Min.X+7, r.Min.Y+7, r.Max.X-7, r.Max.Y-7)
	if index == 0 {
		if active != nil && active.selectedWeapon == 0 {
			drawScaledImage(screen, s.shop.trainingOn, r)
		} else {
			drawScaledImage(screen, s.shop.trainingOff, r)
		}
		return
	}
	if shopItemIndex >= 0 {
		s.drawShopItemIcon(screen, shopItemIndex, inner, !unlocked)
		return
	}
	if unlocked {
		drawWeaponIcon(screen, inner, weapons[index])
		return
	}
	drawLockedWeaponIcon(screen, inner, index)
}

func (s *GameScene) weaponSlotRect(index int) image.Rectangle {
	screenCfg := core.Config().Screen
	slot := 32
	gap := 5
	count := s.weaponSlotCount()
	total := count*slot + (count-1)*gap
	x := int(screenCfg.Width)/2 - total/2 + index*(slot+gap)
	y := int(screenCfg.Height) - 44
	return image.Rect(x, y, x+slot, y+slot)
}

func (s *GameScene) weaponSlotCount() int {
	return 1 + len(shopItems())
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

func drawWeaponIcon(screen *ebiten.Image, r image.Rectangle, weapon weapon) {
	switch weapon.name {
	case "Training":
		drawFilledRect(screen, image.Rect(r.Min.X+11, r.Min.Y+11, r.Max.X-11, r.Max.Y-11), weapon.color)
	case "Granate":
		drawFilledRect(screen, image.Rect(r.Min.X+4, r.Min.Y+18, r.Max.X-4, r.Min.Y+23), weapon.color)
		drawFilledRect(screen, image.Rect(r.Max.X-12, r.Min.Y+13, r.Max.X-5, r.Min.Y+28), color.RGBA{R: 255, G: 222, B: 76, A: 255})
	default:
		drawFilledRect(screen, r, weapon.color)
	}
}

func drawLockedWeaponIcon(screen *ebiten.Image, r image.Rectangle, index int) {
	c := color.RGBA{R: 77, G: 80, B: 84, A: 255}
	switch index % 5 {
	case 0:
		drawFilledRect(screen, image.Rect(r.Min.X+5, r.Min.Y+16, r.Max.X-5, r.Min.Y+21), c)
	case 1:
		drawFilledRect(screen, image.Rect(r.Min.X+13, r.Min.Y+4, r.Min.X+19, r.Max.Y-4), c)
		drawFilledRect(screen, image.Rect(r.Min.X+4, r.Min.Y+13, r.Max.X-4, r.Min.Y+19), c)
	case 2:
		drawFilledRect(screen, image.Rect(r.Min.X+7, r.Min.Y+7, r.Max.X-7, r.Max.Y-7), c)
	case 3:
		drawFilledRect(screen, image.Rect(r.Min.X+3, r.Min.Y+22, r.Max.X-3, r.Min.Y+27), c)
		drawFilledRect(screen, image.Rect(r.Min.X+8, r.Min.Y+15, r.Max.X-8, r.Min.Y+20), c)
	default:
		drawFilledRect(screen, image.Rect(r.Min.X+15, r.Min.Y+3, r.Min.X+21, r.Max.Y-3), c)
	}
}

func (s *GameScene) cannonAngleDegrees() float64 {
	tank := s.activeTank()
	if tank == nil || tank.cannon == nil || tank.body == nil {
		return 0
	}
	leftLimit := tank.body.Rot - math.Pi
	displayAngle := engine.RadToDeg(tank.cannon.Rot - leftLimit)
	return math.Max(0, math.Min(180, displayAngle))
}

func gameWeapons() []weapon {
	return []weapon{
		{name: "Training", color: color.RGBA{R: 238, G: 238, B: 238, A: 255}, damage: directHitDamage, unlocked: true, showTrail: true},
		{name: "Granate", color: color.RGBA{R: 238, G: 238, B: 238, A: 255}, damage: directHitDamage, unlocked: true, showTrail: false, roundProjectile: true, damagesTerrain: true},
		{name: "Große Granate", color: color.RGBA{R: 238, G: 238, B: 238, A: 255}, damage: directHitDamage, unlocked: true, showTrail: false, roundProjectile: true, damagesTerrain: true, projectileScale: largeGrenadeScale, impactScale: largeGrenadeImpactScale},
		{name: "Atombombe", color: color.RGBA{R: 238, G: 238, B: 238, A: 255}, damage: directHitDamage, unlocked: true, showTrail: false, roundProjectile: true, damagesTerrain: true, impactScale: atomBombImpactScale, impactAnimationExtraSeconds: atomBombImpactExtraSeconds, impactCycles: 1, impactGradientOutward: true},
	}
}

func projectileRadiusForWeapon(weapon weapon) float64 {
	scale := weapon.projectileScale
	if scale <= 0 {
		scale = 1
	}
	return float64(projectileRadius) * scale
}

func impactRadiusForWeapon(weapon weapon) float64 {
	scale := weapon.impactScale
	if scale <= 0 {
		scale = 1
	}
	return float64(projectileRadius*impactRadiusMultiplier) * scale
}

func impactCyclesForWeapon(weapon weapon) int {
	if weapon.impactCycles <= 0 {
		return 2
	}
	return weapon.impactCycles
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

func (s *GameScene) impactAnimationFramesForWeapon(weapon weapon) int {
	return secondsToFrames(core.Config().Gameplay.ImpactAnimationSeconds + weapon.impactAnimationExtraSeconds)
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
	if s.projectile != nil || s.roundTransitionDelay > 0 || s.roundSeriesComplete {
		return
	}
	s.behaviorRotateOnButton(source)
	s.clampCannonRotationToTank(source, tank.body)
}

func (s *GameScene) clampCannonRotationToTank(cannon, tank *engine.Sprite) {
	if cannon == nil || tank == nil {
		return
	}
	minRot := tank.Rot - math.Pi
	maxRot := tank.Rot
	cannon.Rot = math.Max(minRot, math.Min(maxRot, cannon.Rot))
}
