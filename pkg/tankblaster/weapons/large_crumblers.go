package weapons

func LargeCrumblers() Weapon {
	return withDefaultSounds(Weapon{
		Name:                   "Brösler, groß",
		Color:                  whiteProjectileColor(),
		Unlocked:               true,
		RoundProjectile:        true,
		ImpactAnimationSeconds: DefaultImpactAnimationSeconds,
		LargeCrumblers:         true,
	})
}
