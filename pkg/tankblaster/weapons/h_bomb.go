package weapons

const hBombImpactScale = atomBombImpactScale * 2
const hBombImpactAnimationSeconds = 2.0

func HBomb() Weapon {
	return Weapon{
		Name:                   "H-Bombe",
		Color:                  whiteProjectileColor(),
		Damage:                 DirectHitDamage,
		Unlocked:               true,
		RoundProjectile:        true,
		DamagesTerrain:         true,
		ImpactAnimationSeconds: hBombImpactAnimationSeconds,
		ImpactScale:            hBombImpactScale,
		ImpactCycles:           1,
		ImpactAnimationStyle:   ImpactAnimationHBomb,
	}
}
