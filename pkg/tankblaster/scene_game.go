package tankblaster

import (
	"github.com/hajimehoshi/ebiten/v2"

	_ "image/jpeg"
	_ "image/png"

	"github.com/runzhammer/gamedemo/pkg/core"
	"github.com/runzhammer/gamedemo/pkg/engine"
	"github.com/runzhammer/gamedemo/pkg/models"
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

type GameScene struct {
	g    *GameLoop
	time float64

	phase Phase

	// shot engine.Drawable

	layers engine.Layers
}

func NewGameScene(game *GameLoop) (core.Scene, error) {
	// loader := game.context.Loader()

	s := &GameScene{
		g:     game,
		phase: phaseBattle,
		// cannonSFX:   cannonSFX,
		// messageFace: messageFace,
		// shot:        shotDrawable,
		layers: engine.NewLayers(numLayers),
	}

	b := models.NewBackground()
	gr := models.NewGround()

	t1 := models.NewTank("Player 1")
	t1.SetPosition(engine.Vec{X: 0, Y: gr.Position.Y - t1.Size.Y})

	// t2 := models.NewTank("Player 2")

	t1.TankCannon.Sprite.Steps = engine.MakeBehaviors(
		s.behaviorRotateCannonOnButton,
		s.behaviorMoveTankOnButton,
	)

	t1.TankChassis.Sprite.Steps = engine.MakeBehaviors(
		s.behaviorMoveTankOnButton,
	)

	s.layers[layerTanks].AddSprites(t1.GetSprites())
	s.layers[layerGround].AddSprites(gr.GetSprites())
	s.layers[layerBackground].AddSprites(b.GetSprites())

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

// func (g *GameScene) Movement(source *engine.Sprite) {
// 	if MoveLeft() {
// 		// move left
// 		source.Velocity = source.Velocity.Rotated(source.Rot)
// 		engine.Movement(source)
// 	} else {
// 		source.Velocity = source.Velocity.Rotated(source.Rot)
// 		engine.Movement(source)
// 	}
// }

func (s *GameScene) Update() error {
	s.time += 1

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
		s.layers.Update()
	case phaseBlueVictory:
		fallthrough
	case phaseRedVictory:
		// if s.victoryTime <= 0 {
		// 	return s.g.SetNewScene(NewTitleScene)
		// }
	}

	return nil
}

func (s *GameScene) Draw(screen *ebiten.Image) {

	s.layers.Draw(nil, screen)

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

func (s *GameScene) behaviorMoveTankOnButton(source *engine.Sprite) {
	if MoveLeft() {
		moveLeft := engine.Vec{X: source.Pos.X - source.MovementSpeed, Y: source.Pos.Y}
		source.Pos = &moveLeft
	} else if MoveRight() {
		moveRight := engine.Vec{X: source.Pos.X + source.MovementSpeed, Y: source.Pos.Y}
		source.Pos = &moveRight
	}
}

// func (s *GameScene) behaviorMoveOnButton(source *engine.Sprite) {
// 	if MoveLeft() {
// 		moveLeft := engine.Vec{X: source.Pos.X - float64(source.MovementSpeed), Y: source.Pos.Y}
// 		source.Pos = moveLeft
// 	} else if MoveRight() {
// 		moveRight := engine.Vec{X: source.Pos.X + float64(source.MovementSpeed), Y: source.Pos.Y}
// 		source.Pos = moveRight
// 		engine.Movement(source)
// 	}
// }

func (s *GameScene) behaviorRotateCannonOnButton(source *engine.Sprite) {

	// var rotate float64

	// if RadToDeg() >= source.MaxRange[0] {
	// 	rotate = DegToRad(source.MaxRange[0])
	// }

	// if engine.RadToDeg(source.Rot+source.RotNormal) <= source.MaxRange[1] {
	// 	rotate = DegToRad(source.MaxRange[1])
	// }

	if RotateLeft() {
		degL := engine.RadToDeg(source.Rot) + source.RotationSpeedPerSecond/60*360*-1
		if degL > source.MaxRange[0] {
			source.Rot = engine.DegToRad(degL)
		}
	} else if RotateRight() {
		degR := engine.RadToDeg(source.Rot) + source.RotationSpeedPerSecond/60*360
		if degR < source.MaxRange[1] {
			source.Rot = engine.DegToRad(degR)
		}
	}
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
