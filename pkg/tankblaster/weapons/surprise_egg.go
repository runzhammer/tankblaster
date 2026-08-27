package weapons

func SurpriseEgg() Weapon {
	return Weapon{
		Name:            "Überraschungsei",
		Color:           whiteProjectileColor(),
		Damage:          DirectHitDamage,
		Unlocked:        true,
		RoundProjectile: true,
		SurpriseEgg:     true,
	}
}
