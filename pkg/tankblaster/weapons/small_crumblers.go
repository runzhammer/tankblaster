package weapons

func SmallCrumblers() Weapon {
	return Weapon{
		Name:            "Brösler, klein",
		Color:           whiteProjectileColor(),
		Damage:          DirectHitDamage,
		Unlocked:        true,
		RoundProjectile: true,
		SmallCrumblers:  true,
	}
}
