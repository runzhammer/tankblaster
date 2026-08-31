package weapons

import "github.com/runzhammer/gamedemo/pkg/tankblaster/soundpaths"

const (
	plasmaImpactScale            = hBombImpactScale * 2.5
	plasmaImpactAnimationSeconds = 6.5
)

func PlasmaMelter() Weapon {
	weapon := withDefaultSounds(Weapon{
		Name:                   "Plasmaschmelzer",
		Color:                  whiteProjectileColor(),
		Damage:                 DirectHitDamage,
		Unlocked:               true,
		RoundProjectile:        true,
		DamagesTerrain:         true,
		ImpactScale:            plasmaImpactScale,
		ImpactAnimationSeconds: plasmaImpactAnimationSeconds,
		ImpactCycles:           1,
		ImpactAnimationStyle:   ImpactAnimationPlasma,
	})
	weapon.ImpactSound = soundpaths.SoundPlasma
	return weapon
}
