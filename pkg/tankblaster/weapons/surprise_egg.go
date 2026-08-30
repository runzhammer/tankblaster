package weapons

func SurpriseEgg() Weapon {
	return withDefaultSounds(Weapon{
		Name:                   "Überraschungsei",
		Color:                  whiteProjectileColor(),
		Damage:                 DirectHitDamage,
		Unlocked:               true,
		RoundProjectile:        true,
		ImpactAnimationSeconds: DefaultImpactAnimationSeconds,
		SurpriseEgg:            true,
	})
}
