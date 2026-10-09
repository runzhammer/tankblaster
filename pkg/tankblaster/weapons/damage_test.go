package weapons

import "testing"

func TestCalculateRadialDamageGrenadeBoundaries(t *testing.T) {
	profile := Grenade().RadialDamage
	tests := map[int]int{
		0:  100,
		17: 100,
		18: 93,
		32: 6,
		33: 0,
		34: 0,
	}
	for distance, want := range tests {
		if got := CalculateRadialDamage(distance, profile); got != want {
			t.Fatalf("distance %d damage = %d, want %d", distance, got, want)
		}
	}
}

func TestDirectWeaponDamageTable(t *testing.T) {
	tests := []struct {
		name   string
		weapon Weapon
		damage int
	}{
		{"training", Training(), 0},
		{"wonder palm", WonderPalm(), 0},
		{"water", Water(), 0},
		{"small crumblers", SmallCrumblers(), 0},
		{"large crumblers", LargeCrumblers(), 0},
		{"mosquitos", Mosquitos(), 100},
		{"laser", Laser(), 100},
	}
	for _, tt := range tests {
		if got := tt.weapon.Damage; got != tt.damage {
			t.Fatalf("%s direct damage = %d, want %d", tt.name, got, tt.damage)
		}
	}
}

func TestOriginalRadialDamageProfiles(t *testing.T) {
	tests := []struct {
		name    string
		profile RadialDamageProfile
		inner   int
		outer   int
		max     int
	}{
		{"grenade", Grenade().RadialDamage, 17, 33, 100},
		{"large grenade", LargeGrenade().RadialDamage, 32, 48, 100},
		{"atom bomb", AtomBomb().RadialDamage, 60, 90, 100},
		{"h bomb", HBomb().RadialDamage, 110, 165, 100},
		{"plasma melter", PlasmaMelter().RadialDamage, 198, 297, 100},
		{"fireball", Fireball().RadialDamage, 19, 57, 100},
		{"moles", Moles().RadialDamage, 53, 80, 100},
		{"mfs state 0", MFSTriple().RadialDamage, 32, 48, 100},
		{"shockwave", Shockwave().RadialDamage, 100, 225, 100},
		{"airstrike bomb", AirStrike().RadialDamage, 60, 90, 100},
		{"splitter fragment", SplitterBombFragment().RadialDamage, 10, 15, 100},
	}
	for _, tt := range tests {
		if got := tt.profile; got.InnerRadius != tt.inner || got.OuterRadius != tt.outer || got.MaxDamage != tt.max {
			t.Fatalf("%s profile = %+v, want inner=%d outer=%d max=%d", tt.name, got, tt.inner, tt.outer, tt.max)
		}
	}
}
