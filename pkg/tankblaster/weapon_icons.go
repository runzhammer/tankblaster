package tankblaster

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	weaponspkg "github.com/runzhammer/gamedemo/pkg/tankblaster/weapons"
)

func drawWeaponIcon(screen *ebiten.Image, r image.Rectangle, weapon weaponspkg.Weapon) {
	switch weapon.Name {
	case "Training":
		drawFilledRect(screen, image.Rect(r.Min.X+11, r.Min.Y+11, r.Max.X-11, r.Max.Y-11), weapon.Color)
	case "Granate":
		drawFilledRect(screen, image.Rect(r.Min.X+4, r.Min.Y+18, r.Max.X-4, r.Min.Y+23), weapon.Color)
		drawFilledRect(screen, image.Rect(r.Max.X-12, r.Min.Y+13, r.Max.X-5, r.Min.Y+28), color.RGBA{R: 255, G: 222, B: 76, A: 255})
	default:
		drawFilledRect(screen, r, weapon.Color)
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
