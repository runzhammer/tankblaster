package tankblaster

import (
	"github.com/hajimehoshi/ebiten/v2"

	_ "image/png"

	"image/color"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/runzhammer/gamedemo/pkg/core"
	"github.com/runzhammer/gamedemo/pkg/engine"
	"golang.org/x/image/colornames"
)

var _ core.Scene = (*titleScene)(nil)

var (
	textColors = []color.Color{
		colornames.Red,
		colornames.Lightblue,
		colornames.Coral,
		colornames.Cornflowerblue,
		colornames.White,
	}
)

type titleScene struct {
	g    *GameLoop
	time float64

	title        engine.Text
	instructions engine.Texts
	scoreboard   *engine.Texts
}

func NewTitleScene(game *GameLoop) (core.Scene, error) {
	// loader := game.context.Loader()

	// titleFace, err := loader.Face("fonts/DampfPlatz.ttf", 240)
	// if err != nil {
	// 	return nil, err
	// }

	// instructionsFace, err := loader.Face("fonts/Lekton-Regular.ttf", 12)
	// if err != nil {
	// 	return nil, err
	// }

	// var scoreboard *engine.Texts
	// if game.redScore != 0 || game.blueScore != 0 {
	// 	face, err := loader.Face("fonts/BlackKnightFLF.ttf", 36)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	texts := make(engine.Texts, 0, 3)
	// 	texts.Pushf(face, colornames.Blue, "Blue: %d", game.blueScore)
	// 	texts.Push(face, colornames.White, " - ")
	// 	texts.Pushf(face, colornames.Red, "Red: %d", game.redScore)
	// 	scoreboard = &texts
	// }

	s := &titleScene{
		g: game,
		// title:        engine.NewText(titleFace, colornames.White, Title),
		// instructions: engine.NewTexts(instructionsFace, colornames.White, instructionsStrings),
		// scoreboard:   scoreboard,
	}

	return s, nil
}

func (s *titleScene) Update() error {
	s.time += 1

	if Begin() {
		return s.g.SetNewScene(NewGameScene)
	}

	return nil
}

func (s *titleScene) Draw(image *ebiten.Image) {
	s.drawTitle(image)
	s.drawScoreboard(image)
	s.drawInstructions(image)
}

func (s *titleScene) drawScoreboard(image *ebiten.Image) {
	const (
		vpad = 36
		jit  = 1
	)

	if s.scoreboard == nil {
		return
	}

	width := s.scoreboard.SingleLineWidth()
	x := int(core.Config().Screen.Width)/2 - width/2

	for i, t := range *s.scoreboard {
		dx, dy := x, vpad
		switch {
		case i == 0 && s.g.blueScore > s.g.redScore:
			dx += jitter(jit)
			dy += jitter(jit)
		case i == 2 && s.g.redScore > s.g.blueScore:
			dx += jitter(jit)
			dy += jitter(jit)
		}
		t.Draw(image, dx, dy, engine.AlignLeft)
		x += t.Advance
	}
}

func (s *titleScene) drawInstructions(image *ebiten.Image) {
	const (
		space = 4
		vpad  = 4
	)
	if s.instructions == nil {
		return
	}
	height := s.instructions.MultiLineHeight(space)
	s.instructions.DrawLines(image, space, int(core.Config().Screen.Width)/2, int(core.Config().Screen.Height)-height-vpad, engine.AlignCenter)
}

func (s *titleScene) drawTitle(image *ebiten.Image) {
	const jit = 3
	for _, textColor := range textColors {
		dx, dy := jitter(jit), jitter(jit)
		text.Draw(image, s.title.Text, s.title.Face, int(core.Config().Screen.Width)/2-s.title.W/2+dx, int(core.Config().Screen.Height)/2+s.title.H/2+dy, textColor)
	}
}

func jitter(amount int) int {
	return rand.Intn(2*amount+1) - amount
}
