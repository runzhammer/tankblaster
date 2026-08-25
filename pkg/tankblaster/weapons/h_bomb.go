package weapons

const hBombImpactScale = atomBombImpactScale * 2
const hBombImpactAnimationExtraSeconds = 2.5

func HBomb() Weapon {
	return Weapon{
		Name:                        "H-Bombe",
		Color:                       whiteProjectileColor(),
		Damage:                      DirectHitDamage,
		Unlocked:                    true,
		RoundProjectile:             true,
		DamagesTerrain:              true,
		ImpactAnimationExtraSeconds: hBombImpactAnimationExtraSeconds,
		ImpactScale:                 hBombImpactScale,
		ImpactCycles:                1,
		ImpactAnimationStyle:        ImpactAnimationHBomb,
	}
}
