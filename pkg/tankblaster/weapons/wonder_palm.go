package weapons

func WonderPalm() Weapon {
	return Weapon{
		Name:            "Wunderpalme",
		Color:           whiteProjectileColor(),
		Damage:          DirectHitDamage,
		Unlocked:        true,
		RoundProjectile: true,
		PlantsPalm:      true,
	}
}
