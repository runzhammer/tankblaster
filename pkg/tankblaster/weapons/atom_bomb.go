package weapons

import "github.com/runzhammer/gamedemo/pkg/tankblaster/soundpaths"

const (
	atomBombImpactScale            = 4.0
	atomBombImpactAnimationSeconds = 0.8
)

func AtomBomb() Weapon {
	weapon := withDefaultSounds(Weapon{
		Name:                   "Atombombe",
		Color:                  whiteProjectileColor(),
		RadialDamage:           RadialDamageProfile{InnerRadius: 60, OuterRadius: 90, MaxDamage: 100},
		Unlocked:               true,
		RoundProjectile:        true,
		DamagesTerrain:         true,
		ImpactScale:            atomBombImpactScale,
		ImpactAnimationSeconds: atomBombImpactAnimationSeconds,
		ImpactCycles:           1,
		ImpactGradientOutward:  true,
	})
	weapon.ImpactSound = soundpaths.SoundAtom
	return weapon
}
