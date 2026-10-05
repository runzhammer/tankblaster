package tankblaster

import (
	"image"
	"os"
	"path/filepath"
	"testing"

	"github.com/runzhammer/tankblaster/pkg/buildinfo"
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

func TestPlayerSelectionVersionLabelUsesVPrefix(t *testing.T) {
	previous := buildinfo.Version
	defer func() {
		buildinfo.Version = previous
	}()

	buildinfo.Version = "1.0.2"
	if got, want := playerSelectionVersionLabel(), "v1.0.2"; got != want {
		t.Fatalf("version label = %q, want %q", got, want)
	}

	buildinfo.Version = "v1.0.2"
	if got, want := playerSelectionVersionLabel(), "v1.0.2"; got != want {
		t.Fatalf("version label = %q, want %q", got, want)
	}

	tempDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tempDir, "VERSION"), []byte("1.0.9\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(tempDir)
	buildinfo.Version = "dev"
	if got, want := playerSelectionVersionLabel(), "v1.0.9 (dev)"; got != want {
		t.Fatalf("version label = %q, want %q", got, want)
	}
}
