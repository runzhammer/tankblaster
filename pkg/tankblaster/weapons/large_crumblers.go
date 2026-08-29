package weapons

func LargeCrumblers() Weapon {
	return Weapon{
		Name:                   "Brösler, groß",
		Color:                  whiteProjectileColor(),
		Damage:                 DirectHitDamage,
		Unlocked:               true,
		RoundProjectile:        true,
		ImpactAnimationSeconds: DefaultImpactAnimationSeconds,
		LargeCrumblers:         true,
	}
}
