package gamecore

import "testing"

func TestNewMatchIncludesWorldSeeds(t *testing.T) {
	state := NewEngine(123).NewMatch("match-1", []Player{
		{ID: "p1"},
		{ID: "p2"},
	})

	if state.TerrainSeed == 0 {
		t.Fatal("terrain seed is zero")
	}
	if state.TankSeed == 0 {
		t.Fatal("tank seed is zero")
	}
	if state.PalmSeed == 0 {
		t.Fatal("palm seed is zero")
	}
}
