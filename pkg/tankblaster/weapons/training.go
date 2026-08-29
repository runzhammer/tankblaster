package weapons

func Training() Weapon {
	return Weapon{
		Name:                   "Spurgeschoß",
		Color:                  whiteProjectileColor(),
		Damage:                 0,
		Unlocked:               true,
		RoundProjectile:        true,
		ShowTrail:              true,
		ImpactAnimationSeconds: DefaultImpactAnimationSeconds,
	}
}
