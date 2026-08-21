package weapons

const (
	atomBombImpactScale        = 4.0
	atomBombImpactExtraSeconds = 0.5
)

func AtomBomb() Weapon {
	return Weapon{
		Name:                        "Atombombe",
		Color:                       whiteProjectileColor(),
		Damage:                      DirectHitDamage,
		Unlocked:                    true,
		RoundProjectile:             true,
		DamagesTerrain:              true,
		ImpactScale:                 atomBombImpactScale,
		ImpactAnimationExtraSeconds: atomBombImpactExtraSeconds,
		ImpactCycles:                1,
		ImpactGradientOutward:       true,
	}
}
