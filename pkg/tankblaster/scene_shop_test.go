package tankblaster

import (
	"bytes"
	"image"
	_ "image/png"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	r "github.com/runzhammer/tankblaster/resources"
)

func TestShopItemIconSourceRectUsesOriginalLadenIconCells(t *testing.T) {
	sheet := ebiten.NewImage(512, 64)

	tests := []struct {
		itemIndex int
		want      image.Rectangle
	}{
		{itemIndex: 0, want: image.Rect(0, 0, 32, 32)},
		{itemIndex: 15, want: image.Rect(480, 0, 512, 32)},
		{itemIndex: 16, want: image.Rect(0, 32, 32, 64)},
		{itemIndex: 23, want: image.Rect(224, 32, 256, 64)},
	}

	for _, tt := range tests {
		got, ok := shopItemIconSourceRect(sheet, tt.itemIndex)
		if !ok {
			t.Fatalf("shopItemIconSourceRect(%d) returned !ok", tt.itemIndex)
		}
		if got != tt.want {
			t.Fatalf("shopItemIconSourceRect(%d) = %v, want %v", tt.itemIndex, got, tt.want)
		}
	}
}

func TestShopItemIconSourceRectFallsBackToLegacyCells(t *testing.T) {
	sheet := ebiten.NewImage(512, 64)
	got, ok := shopItemIconSourceRect(sheet.SubImage(image.Rect(0, 0, 256, 64)).(*ebiten.Image), 17)
	if !ok {
		t.Fatal("shopItemIconSourceRect returned !ok")
	}
	if want := image.Rect(32, 32, 64, 64); got != want {
		t.Fatalf("shopItemIconSourceRect = %v, want %v", got, want)
	}
}

func TestStoreIconAssetKeepsOriginalSheetDimensions(t *testing.T) {
	img, _, err := image.Decode(bytes.NewReader(r.StoreIcons))
	if err != nil {
		t.Fatalf("store icons decode: %v", err)
	}
	if got, want := img.Bounds().Size(), image.Pt(512, 64); got != want {
		t.Fatalf("store icons size = %v, want %v", got, want)
	}
	assertTransparentPixel(t, img, 32, 10)
	assertTransparentPixel(t, img, 10, 32)
	assertOpaquePixel(t, img, 16, 16)
	assertOpaquePixel(t, img, 80, 16)
}

func TestWeaponbarAssetsKeepOriginalDimensions(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{name: "active", data: r.WeaponbarActive},
		{name: "onstock", data: r.WeaponbarOnStock},
		{name: "outofstock", data: r.WeaponbarOutOfStock},
	}

	for _, tt := range tests {
		img, _, err := image.Decode(bytes.NewReader(tt.data))
		if err != nil {
			t.Fatalf("%s weaponbar decode: %v", tt.name, err)
		}
		if got, want := img.Bounds().Size(), image.Pt(640, 34); got != want {
			t.Fatalf("%s weaponbar size = %v, want %v", tt.name, got, want)
		}
	}
}

func TestWeaponbarSlotSourceRectKeepsTwentyExactSlots(t *testing.T) {
	sheet := ebiten.NewImage(640, 34)

	tests := []struct {
		index int
		want  image.Rectangle
	}{
		{index: 0, want: image.Rect(0, 0, 32, 34)},
		{index: 1, want: image.Rect(32, 0, 64, 34)},
		{index: 19, want: image.Rect(608, 0, 640, 34)},
	}

	for _, tt := range tests {
		got := weaponbarSlotSourceRect(sheet, tt.index)
		if got != tt.want {
			t.Fatalf("weaponbarSlotSourceRect(%d) = %v, want %v", tt.index, got, tt.want)
		}
	}
}

func TestEffectAnimationAssetsDoNotIncludeBottomGuideBars(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		size image.Point
	}{
		{name: "fireball", data: r.FireballImpactPNG, size: image.Pt(108, 57)},
		{name: "palm fire", data: r.PalmFirePNG, size: image.Pt(344, 73)},
		{name: "palm smoke", data: r.PalmSmokePNG, size: image.Pt(292, 65)},
		{name: "moskitos", data: r.MoskitosPNG, size: image.Pt(69, 22)},
		{name: "laser smoke", data: r.LaserSmokePNG, size: image.Pt(60, 13)},
		{name: "palm crumble", data: r.PalmCrumblePNG, size: image.Pt(363, 304)},
	}

	for _, tt := range tests {
		img, _, err := image.Decode(bytes.NewReader(tt.data))
		if err != nil {
			t.Fatalf("%s decode: %v", tt.name, err)
		}
		if got := img.Bounds().Size(); got != tt.size {
			t.Fatalf("%s size = %v, want %v", tt.name, got, tt.size)
		}
	}

	palmCrumble, _, err := image.Decode(bytes.NewReader(r.PalmCrumblePNG))
	if err != nil {
		t.Fatalf("palm crumble decode: %v", err)
	}
	assertTransparentPixel(t, palmCrumble, 20, 300)
}

func assertTransparentPixel(t *testing.T, img image.Image, x, y int) {
	t.Helper()
	_, _, _, alpha := img.At(x, y).RGBA()
	if alpha != 0 {
		t.Fatalf("pixel %d,%d alpha = %d, want transparent", x, y, alpha)
	}
}

func assertOpaquePixel(t *testing.T, img image.Image, x, y int) {
	t.Helper()
	_, _, _, alpha := img.At(x, y).RGBA()
	if alpha == 0 {
		t.Fatalf("pixel %d,%d alpha = %d, want visible", x, y, alpha)
	}
}
