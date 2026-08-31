package weapons

func WonderPalm() Weapon {
	return withDefaultSounds(Weapon{
		Name:                   "Wunderpalme",
		Color:                  whiteProjectileColor(),
		Damage:                 DirectHitDamage,
		Unlocked:               true,
		RoundProjectile:        true,
		ImpactAnimationSeconds: DefaultImpactAnimationSeconds,
		PlantsPalm:             true,
	})
}
