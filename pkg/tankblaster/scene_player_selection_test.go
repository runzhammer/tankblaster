package tankblaster

import (
	"image"
	"testing"

	"github.com/runzhammer/tankblaster/pkg/tankblaster/computerplayers"
)

func TestColorSwatchClickOpensPaletteForComputerSlot(t *testing.T) {
	scene := &playerSelectionScene{
		focusedName:    3,
		openPaletteFor: -1,
	}
	scene.slots[0] = playerSelectionSlot{
		Kind:       PlayerComputer,
		ComputerID: computerplayers.DoedelID,
	}
	p := colorSwatchRectForSlot(0).Min.Add(image.Pt(1, 1))

	if !scene.handleSlotClick(p.X, p.Y) {
		t.Fatal("handleSlotClick returned false")
	}
	if got, want := scene.openPaletteFor, 0; got != want {
		t.Fatalf("openPaletteFor = %d, want %d", got, want)
	}
	if got, want := scene.focusedName, -1; got != want {
		t.Fatalf("focusedName = %d, want %d", got, want)
	}
	if got, want := scene.slots[0].ComputerID, computerplayers.DoedelID; got != want {
		t.Fatalf("computer id = %v, want %v", got, want)
	}
}
