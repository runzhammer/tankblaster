package weapons

func Water() Weapon {
	return Weapon{
		Name:                   "Wasser",
		Color:                  whiteProjectileColor(),
		Damage:                 DirectHitDamage,
		Unlocked:               true,
		RoundProjectile:        true,
		ImpactAnimationSeconds: DefaultImpactAnimationSeconds,
		FillsWater:             true,
	}
}
