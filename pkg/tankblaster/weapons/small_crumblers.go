package weapons

func SmallCrumblers() Weapon {
	return withDefaultSounds(Weapon{
		Name:                   "Brösler, klein",
		Color:                  whiteProjectileColor(),
		Unlocked:               true,
		RoundProjectile:        true,
		ImpactAnimationSeconds: DefaultImpactAnimationSeconds,
		SmallCrumblers:         true,
	})
}
