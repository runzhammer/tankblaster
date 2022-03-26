package title

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/runzhammer/gamedemo/pkg/core"
	"github.com/runzhammer/gamedemo/pkg/engine"
	"github.com/runzhammer/gamedemo/pkg/models"
)

var _ core.Scene = (*titleScene)(nil)

const (
	layerBackground = iota
	layerMenu
	layerForeground
	numLayers
)

type titleScene struct {
	g *Game

	layers engine.Layers
}

func NewTitleScene(game *Game) (core.Scene, error) {
	// loader := game.context.Loader()

	// tanksImage, err := loader.EbitenImage("images/ic_tanks.png", nil)
	// if err != nil {
	// 	return nil, err
	// }
	// shipwreckImage, err := loader.EbitenImage("images/ship_blue.png")
	// if err != nil {
	// 	return nil, err
	// }

	layers := engine.NewLayers(numLayers)

	titleScene := &titleScene{
		g:      game,
		layers: layers,
	}

	menu := []struct {
		name  string
		image *ebiten.Image
	}{
		{"tanks", models.NewTank("Tanks").Sprite.Image},
		{"shipwreck", models.NewTank("Shipwreck").Sprite.Image},
	}

	w := core.Config().Screen.Width / float64(len(menu)) * 0.75
	h := w

	dx := w * 1.10

	x := (core.Config().Screen.Width - (float64(len(menu)) * dx)) * 0.5
	y := core.Config().Screen.Height*0.5 - h*0.5

	for _, m := range menu {
		layers[layerMenu].Add(&engine.Sprite{
			Tag:      m.name,
			Drawable: engine.NewImageDrawable(m.image),
			Pos:      engine.V(x, y),
			Size:     engine.V(w, h),
		})
		x += dx
	}

	return titleScene, nil
}

func (s *titleScene) Update() error {
	if obj := s.menuTouch(); obj != nil {
		return &core.ChangeSceneError{Game: obj.Tag}
	}
	return nil
}

func (s *titleScene) Draw(image *ebiten.Image) {
	s.layers.Draw(nil, image)
}

func objectBoundsContainsPoint(obj *engine.Sprite, x, y float64) bool {
	return obj.Pos.X <= x &&
		obj.Pos.X+obj.Size.X >= x &&
		obj.Pos.Y <= y &&
		obj.Pos.Y+obj.Size.Y >= y
}

func (s *titleScene) menuTouch() *engine.Sprite {
	for _, touch := range ebiten.TouchIDs() {
		x, y := ebiten.TouchPosition(touch)
		if obj := s.menuTouchAt(float64(x), float64(y)); obj != nil {
			return obj
		}
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		return s.menuTouchAt(float64(x), float64(y))
	}
	return nil
}

func (s *titleScene) menuTouchAt(x, y float64) *engine.Sprite {
	xf, yf := float64(x), float64(y)
	iter := s.layers[layerMenu].Iterator()
	for obj, ok := iter(); ok; obj, ok = iter() {
		if objectBoundsContainsPoint(obj, xf, yf) {
			return obj
		}
	}
	return nil
}
