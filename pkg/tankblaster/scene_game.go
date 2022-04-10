package tankblaster

import (
	"github.com/hajimehoshi/ebiten/v2"

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

	layers  engine.Layers
	tankSet models.TankSet
}

func NewGameScene(game *GameLoop) (core.Scene, error) {
	// loader := game.context.Loader()

	s := &GameScene{
		g:     game,
		phase: phaseBattle,
		// cannonSFX:   cannonSFX,
		// messageFace: messageFace,
		// shot:        shotDrawable,
		layers:  engine.NewLayers(numLayers),
		tankSet: *models.NewTankSet(),
	}

	b := models.NewBackground()
	gr := models.NewGround()

	t1 := models.NewTank("Player 1", &gr)

	randomX := engine.IntRand(core.Config().Screen.Width)
	t1.SetPosition(&engine.Vec{X: randomX, Y: t1.Ground.GetGroundY(randomX)})
	// t1.SetPosition(&engine.Vec{X: 0, Y: gr.Coords[0]})

	t1.Steps = models.MakeBehaviors(
		s.behaviorMoveTankOnButton,
		s.behaviorRotateCannonOnButton,
	)

	s.layers[layerTanks].AddSprites(t1.GetSprites())
	s.layers[layerGround].AddSprites(gr.GetSprites())
	s.layers[layerBackground].AddSprites(b.GetSprites())

	s.tankSet.Add(t1)

	return s, nil
}

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
		s.tankSet.Update()
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

func (s *GameScene) behaviorMoveTankOnButton(t *models.Tank) {

	if MoveLeft() {
		toLeft := t.Position.X - t.TankChassis.Sprite.MovementSpeed
		if toLeft > 0 {
			moveLeft := engine.Vec{X: toLeft, Y: t.Ground.GetGroundY(toLeft)}
			t.SetPosition(&moveLeft)
		}
	} else if MoveRight() {
		toRight := t.Position.X + t.TankChassis.Sprite.MovementSpeed
		if toRight+t.Size.X < core.Config().Screen.Width {
			moveRight := engine.Vec{X: toRight, Y: t.Ground.GetGroundY(toRight)}
			t.SetPosition(&moveRight)
		}
	}
}

func (s *GameScene) behaviorRotateCannonOnButton(t *models.Tank) {
	if RotateLeft() {
		// log.Println("ROT LEFT")
		degL := engine.RadToDeg(t.TankCannon.Sprite.Rot) + t.TankCannon.Sprite.RotationSpeedPerSecond/60*360*-1
		if degL > t.TankCannon.Sprite.MaxRange[0] {
			t.TankCannon.Sprite.Rot = engine.DegToRad(degL)
		}
	} else if RotateRight() {
		degR := engine.RadToDeg(t.TankCannon.Sprite.Rot) + t.TankCannon.Sprite.RotationSpeedPerSecond/60*360
		if degR < t.TankCannon.Sprite.MaxRange[1] {
			t.TankCannon.Sprite.Rot = engine.DegToRad(degR)
		}
	}
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
