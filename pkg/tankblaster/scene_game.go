package tankblaster

import (
	"image"
	"image/color"
	"math"
	"math/rand"
	"strconv"
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	_ "image/jpeg"
	_ "image/png"

	"github.com/runzhammer/gamedemo/pkg/core"
	"github.com/runzhammer/gamedemo/pkg/engine"
	"github.com/runzhammer/gamedemo/pkg/models"
	"golang.org/x/image/colornames"
)

var _ core.Scene = (*GameScene)(nil)

type Phase uint8

const (
	phaseBattle Phase = iota
	phaseBlueVictory
	phaseRedVictory
)

const (
	layerBackground = iota
	layerGround
	layerTanks
	// layerBullets
	numLayers
)

const debugScrollMode = true

type battleTank struct {
	player  PlayerConfig
	body    *engine.Sprite
	cannon  *engine.Sprite
	landed  bool
	falling bool
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

	tanks             []*battleTank
	spawnIndex        int
	activePlayerIndex int
}

func NewGameScene(game *GameLoop) (core.Scene, error) {
	// loader := game.context.Loader()

	s := &GameScene{
		g:                 game,
		phase:             phaseBattle,
		rng:               rand.New(rand.NewSource(time.Now().UnixNano())),
		spawnIndex:        0,
		activePlayerIndex: -1,
		// cannonSFX:   cannonSFX,
		// messageFace: messageFace,
		// shot:        shotDrawable,
		layers: engine.NewLayers(numLayers),
	}

	players := s.playersForRound()
	s.worldWidth = worldWidthForPlayers(len(players))

	b := models.NewBackgroundWithWidth(s.worldWidth)
	gr := models.NewGroundWithWidth(s.worldWidth)
	s.ground = gr

	for tankIndex, player := range players {
		tank := models.NewTank(player.Name, player.Color)
		battleTank := &battleTank{player: player}
		tankBody := tank.Body()
		if tankBody != nil {
			tankBody.Pos = randomTankDropPosition(s.rng, tankIndex, len(players), tankBody.Size, s.worldWidth)
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

	// iter = t1.Sprites.All().Iterator()
	// for obj, ok := iter(); ok; obj, ok = iter() {
	// 	s.layers[layerTanks].Add(obj)
	// }

	// iter = gr.Sprites.All().Iterator()
	// for obj, ok := iter(); ok; obj, ok = iter() {
	// 	s.layers[layerGround].Add(obj)
	// }

	// iter = b.Sprites.All().Iterator()
	// for obj, ok := iter(); ok; obj, ok = iter() {
	// 	s.layers[layerBackground].Add(obj)
	// }

	// log.Printf("t1.Sprite: %v", t1.Sprite)
	// log.Printf("t2.Sprite: %v", t2.Sprite)
	// log.Printf("gr.Sprite: %v", gr.Sprite)
	// log.Printf("b.Sprite: %v", b.Sprite)

	return s, nil
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

func randomTankDropPosition(rng *rand.Rand, index, count int, size *engine.Vec, worldWidth float64) *engine.Vec {
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
		Y: -size.Y - rng.Float64()*screen.Height*0.55,
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
		s.layers.Update()
		s.advanceSpawnSequence()
		s.chooseStartingPlayerAfterLanding()
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
	if tank == nil || tank.body == nil || tank.falling {
		return
	}
	if math.Abs(s.cameraX-s.cameraGoal) > 1 {
		return
	}

	tank.falling = true
	tank.body.Steps = engine.MakeBehaviors(
		s.behaviorFallOntoGround(s.ground, tank),
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

func (s *GameScene) Draw(screen *ebiten.Image) {
	camera := ebiten.GeoM{}
	camera.Translate(-s.cameraX, 0)

	s.layers.Draw(&camera, screen)
	s.drawDebugScrollBar(screen)

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

func debugScrollBarRect() image.Rectangle {
	screen := core.Config().Screen
	return image.Rect(12, int(screen.Height)-22, int(screen.Width)-12, int(screen.Height)-6)
}

// func (s *GameScene) reflectInBounds(source *engine.Sprite, dt float64) {
// 	objBounds := source.Bounds()
// 	switch {
// 	case objBounds.Min.X <= 0:
// 		source.Velocity = engine.V(-source.Velocity.X, source.Velocity.Y)
// 		source.Rot = source.Velocity.Angle()
// 		source.Pos = &engine.V(0, source.Pos.Y)
// 	case objBounds.Max.X >= core.Config().Screen.Width:
// 		source.Velocity = engine.V(-source.Velocity.X, source.Velocity.Y)
// 		source.Rot = source.Velocity.Angle()
// 		source.Pos = &engine.V(core.Config().Screen.Width-source.Size.X, source.Pos.Y)
// 	}
// 	switch {
// 	case objBounds.Min.Y <= 0:
// 		source.Velocity = engine.V(source.Velocity.X, -source.Velocity.Y)
// 		source.Rot = source.Velocity.Angle()
// 		source.Pos = &engine.V(source.Pos.X, 0)
// 	case objBounds.Max.Y >= core.Config().Screen.Height:
// 		source.Velocity = engine.V(source.Velocity.X, -source.Velocity.Y)
// 		source.Rot = source.Velocity.Angle()
// 		source.Pos = &engine.V(source.Pos.X, core.Config().Screen.Height-source.Size.Y)
// 	}
// }

func (s *GameScene) behaviorMoveOnButton(source *engine.Sprite) {
	if MoveLeft() {
		moveLeft := engine.Vec{X: source.Pos.X - float64(source.MovementSpeed), Y: source.Pos.Y}
		source.Pos = &moveLeft
	} else if MoveRight() {
		moveRight := engine.Vec{X: source.Pos.X + float64(source.MovementSpeed), Y: source.Pos.Y}
		source.Pos = &moveRight
	}
}

func (s *GameScene) behaviorFallOntoGround(ground models.Ground, tank *battleTank) engine.Behavior {
	return func(source *engine.Sprite) {
		const gravity = 0.38
		const maxFallSpeed = 12.0

		source.Velocity.Y = math.Min(maxFallSpeed, source.Velocity.Y+gravity)
		source.Pos = source.Pos.Add(source.Velocity)

		centerX := source.Pos.X + source.Size.X/2
		landingY := ground.SurfaceY(centerX)
		if source.Pos.Y+source.Size.Y < landingY {
			return
		}

		source.Velocity = engine.Vec{}
		ground.AlignSpriteToSurface(source)
		if tank != nil {
			tank.landed = true
		}
		source.Steps = nil
	}
}

func (s *GameScene) chooseStartingPlayerAfterLanding() {
	if s.activePlayerIndex >= 0 || !s.allTanksLanded() {
		return
	}
	s.activePlayerIndex = s.rng.Intn(len(s.tanks))
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
	if s.tanks[s.activePlayerIndex].cannon != source {
		return
	}
	s.behaviorRotateOnButton(source)
}

func (s *GameScene) behaviorRedRotateOnButton(source *engine.Sprite, dt float64) {
	// if RedRotate() {
	// 	// rotate
	// 	source.Rot += engine.DegToRad(-tankRotatesPerSecond*360) * dt
	// 	s.redShotDelay = 0
	// } else {
	// 	source.Velocity = engine.V(tankSpeed, 0).Rotated(source.Rot)
	// 	engine.Movement(source, dt)
	// 	if s.redShotDelay > 1.0/autoShotPerSecond {
	// 		s.spawnRedShots()
	// 		s.redShotDelay = 0
	// 	}
	// }
}

func (s *GameScene) spawnBlueShots() {

	// bounds := s.bluePlayer.Bounds()
	// pos1 := bounds.Center().Add(engine.V(bounds.W()/2, 2).Rotated(s.bluePlayer.Rot))
	// pos2 := bounds.Center().Add(engine.V(bounds.W()/2, -8).Rotated(s.bluePlayer.Rot))

	// blueBullet1 := &engine.Sprite{
	// 	Tag:      tagBlueBullet,
	// 	Pos:      pos1,
	// 	Size:     engine.V(8, 8),
	// 	Drawable: s.shot,
	// 	Velocity: engine.V(bulletSpeed, 0).Rotated(s.bluePlayer.Rot),
	// 	Steps: engine.MakeBehaviors(
	// 		engine.Movement,
	// 	),
	// 	PostSteps: engine.MakeBehaviors(
	// 		s.behaviorRemoveOutOfBounds,
	// 	),
	// }
	// blueBullet2 := &engine.Sprite{
	// 	Tag:      tagBlueBullet,
	// 	Pos:      pos2,
	// 	Size:     engine.V(8, 8),
	// 	Drawable: s.shot,
	// 	Velocity: engine.V(bulletSpeed, 0).Rotated(s.bluePlayer.Rot),
	// 	Steps: engine.MakeBehaviors(
	// 		engine.Movement,
	// 	),
	// 	PostSteps: engine.MakeBehaviors(
	// 		s.behaviorRemoveOutOfBounds,
	// 	),
	// }
	// s.layers[layerBullets].Add(blueBullet1)
	// s.layers[layerBullets].Add(blueBullet2)

	// if !s.g.context.Muted() {
	// 	s.cannonSFX.Play()
	// }
}

func (s *GameScene) spawnRedShots() {

	// bounds := s.redPlayer.Bounds()
	// offset := engine.V(bounds.H()/2, -8).Rotated(s.redPlayer.Rot)
	// pos := bounds.Center().Add(offset)

	// redBullet := &engine.Sprite{
	// 	Tag:      tagRedBullet,
	// 	Pos:      pos,
	// 	Size:     engine.V(14, 14),
	// 	Drawable: s.shot,
	// 	Velocity: engine.V(bulletSpeed, 0).Rotated(s.redPlayer.Rot),
	// 	Steps: engine.MakeBehaviors(
	// 		engine.Movement,
	// 	),
	// 	PostSteps: engine.MakeBehaviors(
	// 		s.behaviorRemoveOutOfBounds,
	// 	),
	// }
	// s.layers[layerBullets].Add(redBullet)

	// if !s.g.context.Muted() {
	// 	s.cannonSFX.Play()
	// }
}

// func (s *GameScene) behaviorRemoveOutOfBounds(source *engine.Sprite, dt float64) {
// 	if !engine.Collision(source.Bounds(), core.Config().Screen.Bounds) {
// 		s.layers[layerBullets].Remove(source)
// 	}
// }

// func (s *gameScene) behaviorRedHitsBlueBullet(source *engine.Sprite, dt float64) {
// 	if s.phase != phaseBattle {
// 		return
// 	}
// 	sourceBounds := source.Bounds().ScaledAtCenter(tankCollisionScale)
// 	iter := s.layers.TagIterator(tagBlueBullet)
// 	for bullet, ok := iter(); ok; bullet, ok = iter() {
// 		if engine.Collision(sourceBounds, bullet.Bounds()) {
// 			s.g.blueScore++
// 			s.phase = phaseBlueVictory
// 			s.onVictory("Blue", colornames.Cadetblue)
// 			break
// 		}
// 	}
// }

// func (s *gameScene) behaviorBlueHitsRedBullet(source *engine.Sprite, dt float64) {
// 	if s.phase != phaseBattle {
// 		return
// 	}
// 	sourceBounds := source.Bounds().ScaledAtCenter(tankCollisionScale)
// 	iter := s.layers.TagIterator(tagRedBullet)
// 	for bullet, ok := iter(); ok; bullet, ok = iter() {
// 		if engine.Collision(sourceBounds, bullet.Bounds()) {
// 			s.g.redScore++
// 			s.phase = phaseRedVictory
// 			s.onVictory("Red", colornames.Indianred)
// 			break
// 		}
// 	}
// }

// func (s *gameScene) onVictory(winner string, textColor color.Color) {
// 	s.victoryTime = victoryMessageDuration

// 	saying := winningMessages[rand.Intn(len(winningMessages))]
// 	victoryMessage := fmt.Sprintf(saying, winner)

// 	text := engine.NewText(s.messageFace, textColor, victoryMessage)
// 	s.message = &text
// }
