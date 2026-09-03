package tankblaster

import (
	"testing"

	"github.com/runzhammer/gamedemo/pkg/tankblaster/soundpaths"
	weaponspkg "github.com/runzhammer/gamedemo/pkg/tankblaster/weapons"
	r "github.com/runzhammer/gamedemo/resources"
)

func TestOriginalSoundFilesExist(t *testing.T) {
	paths := []string{
		soundpaths.SoundFire,
		soundpaths.SoundIncinerator1,
		soundpaths.SoundIncinerator2,
		soundpaths.SoundAtom,
		soundpaths.SoundWasserstoff,
		soundpaths.SoundPlasma,
		soundpaths.SoundGrowing,
		soundpaths.SoundBaumschrei,
		soundpaths.SoundPock,
		soundpaths.SoundBurning,
		soundpaths.SoundKlirr,
		soundpaths.SoundBlubber,
		soundpaths.SoundBlotsch,
		soundpaths.SoundBroesler,
		soundpaths.SoundMole,
		soundpaths.SoundMoskitos,
		soundpaths.SoundScream,
		soundpaths.SoundShockwave,
		soundpaths.SoundBeep,
		soundpaths.SoundJet,
		soundpaths.SoundClusterblast,
		soundpaths.SoundClusterexplo,
		soundpaths.SoundColorblop,
		soundpaths.SoundLaser,
		soundpaths.SoundDampf,
		soundpaths.SoundSchnaeppchen,
		soundpaths.SoundNukeall,
		soundpaths.SoundFlyout,
		soundpaths.SoundFlyin,
		soundpaths.SoundAnlassen,
		soundpaths.SoundMotor,
		soundpaths.SoundKette,
		soundpaths.SoundAbstellen,
		soundpaths.SoundDrehenLeft,
		soundpaths.SoundDrehenRight,
		soundpaths.SoundDrehenUp,
		soundpaths.SoundDrehenDown,
		soundpaths.SoundDrehen,
		soundpaths.SoundKlonk,
		soundpaths.SoundScream2,
		soundpaths.SoundExplosion2,
		soundpaths.SoundExplosion3,
		soundpaths.SoundBroeselklirr,
		soundpaths.SoundKasse,
		soundpaths.SoundMoney,
		soundpaths.SoundBupp,
		soundpaths.SoundChoose,
		soundpaths.SoundUp,
		soundpaths.SoundDown,
	}
	for _, path := range paths {
		if _, err := r.SoundBytes(path); err != nil {
			t.Fatalf("sound file %s missing: %v", path, err)
		}
	}
}

func TestOriginalWeaponSoundMappings(t *testing.T) {
	tests := []struct {
		name   string
		weapon weaponspkg.Weapon
		fire   string
		impact string
	}{
		{"training", weaponspkg.Training(), soundpaths.SoundFire, soundpaths.SoundIncinerator1},
		{"grenade", weaponspkg.Grenade(), soundpaths.SoundFire, soundpaths.SoundIncinerator1},
		{"large grenade", weaponspkg.LargeGrenade(), soundpaths.SoundFire, soundpaths.SoundIncinerator2},
		{"atom", weaponspkg.AtomBomb(), soundpaths.SoundFire, soundpaths.SoundAtom},
		{"h bomb", weaponspkg.HBomb(), soundpaths.SoundFire, soundpaths.SoundWasserstoff},
		{"plasma", weaponspkg.PlasmaMelter(), soundpaths.SoundFire, soundpaths.SoundPlasma},
		{"splitter fragment", weaponspkg.SplitterBombFragment(), soundpaths.SoundFire, soundpaths.SoundClusterexplo},
	}
	for _, tt := range tests {
		if got := tt.weapon.FireSound; got != tt.fire {
			t.Fatalf("%s fire sound = %s, want %s", tt.name, got, tt.fire)
		}
		if got := tt.weapon.ImpactSound; got != tt.impact {
			t.Fatalf("%s impact sound = %s, want %s", tt.name, got, tt.impact)
		}
	}
}

func TestOriginalEventSoundMappings(t *testing.T) {
	tests := map[soundEvent]string{
		soundEventAirStrikeBeacon:    soundpaths.SoundBeep,
		soundEventAirStrikeBomb:      soundpaths.SoundAtom,
		soundEventButtonPress:        soundpaths.None,
		soundEventSplitterBombSplit:  soundpaths.SoundClusterblast,
		soundEventLaserSmoke:         soundpaths.SoundDampf,
		soundEventDudImpact:          soundpaths.SoundSchnaeppchen,
		soundEventPalmRevenge:        soundpaths.SoundBaumschrei,
		soundEventRevengeTankBroken:  soundpaths.SoundHahaha,
		soundEventShopBuy:            soundpaths.SoundKasse,
		soundEventShopNotEnoughMoney: soundpaths.SoundMoney,
	}
	for event, want := range tests {
		if got := tankBlasterSounds.Events[event]; got != want {
			t.Fatalf("%s sound = %s, want %s", event, got, want)
		}
	}
}

func TestOriginalZeroPowerDeathSoundMappings(t *testing.T) {
	tests := map[zeroPowerSound]string{
		zeroPowerSoundDust:               soundpaths.SoundBroeselklirr,
		zeroPowerSoundExplosion:          soundpaths.SoundExplosion3,
		zeroPowerSoundMushroom:           soundpaths.SoundExplosion2,
		zeroPowerSoundGrenadeImpact:      soundpaths.SoundIncinerator1,
		zeroPowerSoundLargeGrenadeImpact: soundpaths.SoundIncinerator2,
		zeroPowerSoundAtomImpact:         soundpaths.SoundAtom,
		zeroPowerSoundScatterProjectiles: soundpaths.SoundClusterblast,
		zeroPowerSoundScatterImpact:      soundpaths.SoundColorblop,
	}
	for event, want := range tests {
		if got := tankBlasterSounds.ZeroPower[event]; got != want {
			t.Fatalf("%s sound = %s, want %s", event, got, want)
		}
	}
}

func TestBeepIsConfiguredForLoopUse(t *testing.T) {
	opts := configuredSoundOptions(soundpaths.SoundBeep)
	if !opts.KeepSilenceForLoop {
		t.Fatal("airstrike beep loop should keep its source silence")
	}
}
