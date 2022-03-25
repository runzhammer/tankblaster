package tankblaster

import (
	"github.com/hajimehoshi/ebiten/v2"

	"image/color"
	_ "image/jpeg"
	_ "image/png"

	"math/rand"

	"fmt"

	"strconv"

	"github.com/runzhammer/gamedemo/pkg/core"
	"github.com/runzhammer/gamedemo/pkg/engine"
	"github.com/runzhammer/gamedemo/pkg/models"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font"
)

var _ core.Scene = (*gameScene)(nil)

type Phase uint8

const (
	phaseCountdown Phase = iota
	phaseBattle
	phaseBlueVictory
	phaseRedVictory
)

const (
	layerBackground = iota
	layerTanks
	numLayers
	ground
	background
)

type gameScene struct {
	g    *Game
	time float64

	phase Phase

	messageFace font.Face
	message     *engine.Text

	// cannonSFX engine.AudioPlayer

	bluePlayer *engine.Sprite
	redPlayer  *engine.Sprite

	victoryTime float64

	shot engine.Drawable

	blueShotDelay float64
	redShotDelay  float64

	layers engine.Layers
}

func NewGameScene(game *Game) (core.Scene, error) {
	// loader := game.context.Loader()

	b := models.NewBackground()
	gr := models.NewGround()

	t1 := models.NewTank(engine.Vec{X: 128, Y: float64(ScreenHeight) - 160})
	t1.Name = "Player 1"

	t2 := models.NewTank(engine.Vec{X: float64(ScreenWidth) - 64 - 128, Y: float64(ScreenHeight) - 160})
	t2.Name = "Player 2"

	s := &gameScene{
		g:     game,
		phase: phaseCountdown,
		// cannonSFX:   cannonSFX,
		// messageFace: messageFace,
		// shot:        shotDrawable,
		layers: engine.NewLayers(numLayers),
	}
	s.layers[layerTanks].Add(&t1.Sprite)
	s.layers[layerTanks].Add(&t2.Sprite)
	s.layers[layerBackground].Add(&b.Sprite)
	s.layers[layerBackground].Add(gr.Sprite)

	return s, nil
}

func (s *gameScene) Update(dt float64) error {
	s.time += dt

	switch s.phase {
	case phaseCountdown:

		countdownTime := s.time * 2
		if countdownTime >= 3 {
			s.phase = phaseBattle
			break
		}
		seconds := 3 - int(countdownTime)

		countdownColorIndex := 3 - seconds
		if countdownColorIndex < 0 {
			countdownColorIndex = 0
		}
		text := engine.NewText(s.messageFace, countdownColors[countdownColorIndex], strconv.Itoa(seconds))
		s.message = &text
	case phaseBattle:
		s.blueShotDelay += dt
		s.redShotDelay += dt
		s.layers.Update(dt)
	case phaseBlueVictory:
		fallthrough
	case phaseRedVictory:
		s.victoryTime -= dt
		if s.victoryTime <= 0 {
			return s.g.SetNewScene(NewTitleScene)
		}
	}

	return nil
}

func (s *gameScene) Draw(image *ebiten.Image) {
	s.layers.Draw(nil, image)

	switch s.phase {
	case phaseBattle:
	case phaseBlueVictory:
		fallthrough
	case phaseRedVictory:
		fallthrough
	case phaseCountdown:
		if s.message == nil {
			return
		}
		s.message.Draw(image, core.ScreenWidth/2, core.ScreenHeight/2+s.message.H/2, engine.AlignCenter)
	}
}

func (s *gameScene) reflectInBounds(source *engine.Sprite, dt float64) {
	objBounds := source.Bounds()
	switch {
	case objBounds.Min.X <= 0:
		source.Velocity = engine.V(-source.Velocity.X, source.Velocity.Y)
		source.Rot = source.Velocity.Angle()
		source.Pos = engine.V(0, source.Pos.Y)
	case objBounds.Max.X >= core.ScreenWidth:
		source.Velocity = engine.V(-source.Velocity.X, source.Velocity.Y)
		source.Rot = source.Velocity.Angle()
		source.Pos = engine.V(core.ScreenWidth-source.Size.X, source.Pos.Y)
	}
	switch {
	case objBounds.Min.Y <= 0:
		source.Velocity = engine.V(source.Velocity.X, -source.Velocity.Y)
		source.Rot = source.Velocity.Angle()
		source.Pos = engine.V(source.Pos.X, 0)
	case objBounds.Max.Y >= core.ScreenHeight:
		source.Velocity = engine.V(source.Velocity.X, -source.Velocity.Y)
		source.Rot = source.Velocity.Angle()
		source.Pos = engine.V(source.Pos.X, core.ScreenHeight-source.Size.Y)
	}
}

func (s *gameScene) behaviorBlueRotateOnButton(source *engine.Sprite, dt float64) {
	if BlueRotate() {
		// rotate
		source.Rot += engine.DegToRad(-tankRotatesPerSecond*360) * dt
		s.blueShotDelay = 0
	} else {
		source.Velocity = engine.V(tankSpeed, 0).Rotated(source.Rot)
		engine.Movement(source, dt)
		if s.blueShotDelay > 1.0/autoShotPerSecond {
			s.spawnBlueShots()
			s.blueShotDelay = 0
		}
	}
}

func (s *gameScene) behaviorRedRotateOnButton(source *engine.Sprite, dt float64) {
	if RedRotate() {
		// rotate
		source.Rot += engine.DegToRad(-tankRotatesPerSecond*360) * dt
		s.redShotDelay = 0
	} else {
		source.Velocity = engine.V(tankSpeed, 0).Rotated(source.Rot)
		engine.Movement(source, dt)
		if s.redShotDelay > 1.0/autoShotPerSecond {
			s.spawnRedShots()
			s.redShotDelay = 0
		}
	}
}

func (s *gameScene) spawnBlueShots() {

	bounds := s.bluePlayer.Bounds()
	pos1 := bounds.Center().Add(engine.V(bounds.W()/2, 2).Rotated(s.bluePlayer.Rot))
	pos2 := bounds.Center().Add(engine.V(bounds.W()/2, -8).Rotated(s.bluePlayer.Rot))

	blueBullet1 := &engine.Sprite{
		Tag:      tagBlueBullet,
		Pos:      pos1,
		Size:     engine.V(8, 8),
		Drawable: s.shot,
		Velocity: engine.V(bulletSpeed, 0).Rotated(s.bluePlayer.Rot),
		Steps: engine.MakeBehaviors(
			engine.Movement,
		),
		PostSteps: engine.MakeBehaviors(
			s.behaviorRemoveOutOfBounds,
		),
	}
	blueBullet2 := &engine.Sprite{
		Tag:      tagBlueBullet,
		Pos:      pos2,
		Size:     engine.V(8, 8),
		Drawable: s.shot,
		Velocity: engine.V(bulletSpeed, 0).Rotated(s.bluePlayer.Rot),
		Steps: engine.MakeBehaviors(
			engine.Movement,
		),
		PostSteps: engine.MakeBehaviors(
			s.behaviorRemoveOutOfBounds,
		),
	}
	s.layers[layerBullets].Add(blueBullet1)
	s.layers[layerBullets].Add(blueBullet2)

	if !s.g.context.Muted() {
		s.cannonSFX.Play()
	}
}

func (s *gameScene) spawnRedShots() {

	bounds := s.redPlayer.Bounds()
	offset := engine.V(bounds.H()/2, -8).Rotated(s.redPlayer.Rot)
	pos := bounds.Center().Add(offset)

	redBullet := &engine.Sprite{
		Tag:      tagRedBullet,
		Pos:      pos,
		Size:     engine.V(14, 14),
		Drawable: s.shot,
		Velocity: engine.V(bulletSpeed, 0).Rotated(s.redPlayer.Rot),
		Steps: engine.MakeBehaviors(
			engine.Movement,
		),
		PostSteps: engine.MakeBehaviors(
			s.behaviorRemoveOutOfBounds,
		),
	}
	s.layers[layerBullets].Add(redBullet)

	if !s.g.context.Muted() {
		s.cannonSFX.Play()
	}
}

func (s *gameScene) behaviorRemoveOutOfBounds(source *engine.Sprite, dt float64) {
	if !engine.Collision(source.Bounds(), core.ScreenBounds) {
		s.layers[layerBullets].Remove(source)
	}
}

func (s *gameScene) behaviorRedHitsBlueBullet(source *engine.Sprite, dt float64) {
	if s.phase != phaseBattle {
		return
	}
	sourceBounds := source.Bounds().ScaledAtCenter(tankCollisionScale)
	iter := s.layers.TagIterator(tagBlueBullet)
	for bullet, ok := iter(); ok; bullet, ok = iter() {
		if engine.Collision(sourceBounds, bullet.Bounds()) {
			s.g.blueScore++
			s.phase = phaseBlueVictory
			s.onVictory("Blue", colornames.Cadetblue)
			break
		}
	}
}

func (s *gameScene) behaviorBlueHitsRedBullet(source *engine.Sprite, dt float64) {
	if s.phase != phaseBattle {
		return
	}
	sourceBounds := source.Bounds().ScaledAtCenter(tankCollisionScale)
	iter := s.layers.TagIterator(tagRedBullet)
	for bullet, ok := iter(); ok; bullet, ok = iter() {
		if engine.Collision(sourceBounds, bullet.Bounds()) {
			s.g.redScore++
			s.phase = phaseRedVictory
			s.onVictory("Red", colornames.Indianred)
			break
		}
	}
}

func (s *gameScene) onVictory(winner string, textColor color.Color) {
	s.victoryTime = victoryMessageDuration

	saying := winningMessages[rand.Intn(len(winningMessages))]
	victoryMessage := fmt.Sprintf(saying, winner)

	text := engine.NewText(s.messageFace, textColor, victoryMessage)
	s.message = &text
}
