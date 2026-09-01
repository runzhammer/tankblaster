package weapons

func SurpriseEgg() Weapon {
	return withDefaultSounds(Weapon{
		Name:                   "Überraschungsei",
		Color:                  whiteProjectileColor(),
		Unlocked:               true,
		RoundProjectile:        true,
		ImpactAnimationSeconds: DefaultImpactAnimationSeconds,
		SurpriseEgg:            true,
	})
}
