package weapons

const (
	atomBombImpactScale            = 4.0
	atomBombImpactAnimationSeconds = 0.8
)

func AtomBomb() Weapon {
	return withDefaultSounds(Weapon{
		Name:                   "Atombombe",
		Color:                  whiteProjectileColor(),
		Damage:                 DirectHitDamage,
		Unlocked:               true,
		RoundProjectile:        true,
		DamagesTerrain:         true,
		ImpactScale:            atomBombImpactScale,
		ImpactAnimationSeconds: atomBombImpactAnimationSeconds,
		ImpactCycles:           1,
		ImpactGradientOutward:  true,
	})
}
