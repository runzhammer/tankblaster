package weapons

const (
	plasmaImpactScale               = hBombImpactScale * 2.5
	plasmaImpactAnimationSeconds    = 6.5
	plasmaImpactBaseAnimationOffset = plasmaImpactAnimationSeconds - 0.6
)

func PlasmaMelter() Weapon {
	return Weapon{
		Name:                        "Plasmaschmelzer",
		Color:                       whiteProjectileColor(),
		Damage:                      DirectHitDamage,
		Unlocked:                    true,
		RoundProjectile:             true,
		DamagesTerrain:              true,
		ImpactScale:                 plasmaImpactScale,
		ImpactAnimationExtraSeconds: plasmaImpactBaseAnimationOffset,
		ImpactCycles:                1,
		ImpactAnimationStyle:        ImpactAnimationPlasma,
	}
}
