package weapons

import "github.com/runzhammer/tankblaster/pkg/tankblaster/soundpaths"

const hBombImpactScale = atomBombImpactScale * 2
const hBombImpactAnimationSeconds = 2.0

func HBomb() Weapon {
	weapon := withDefaultSounds(Weapon{
		Name:                   "H-Bombe",
		Color:                  whiteProjectileColor(),
		RadialDamage:           RadialDamageProfile{InnerRadius: 110, OuterRadius: 165, MaxDamage: 100},
		Unlocked:               true,
		RoundProjectile:        true,
		DamagesTerrain:         true,
		ImpactAnimationSeconds: hBombImpactAnimationSeconds,
		ImpactScale:            hBombImpactScale,
		ImpactCycles:           1,
		ImpactAnimationStyle:   ImpactAnimationHBomb,
	})
	weapon.ImpactSound = soundpaths.SoundWasserstoff
	return weapon
}
