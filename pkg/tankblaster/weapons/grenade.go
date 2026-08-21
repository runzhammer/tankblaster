package weapons

func Grenade() Weapon {
	return Weapon{
		Name:            "Granate",
		Color:           whiteProjectileColor(),
		Damage:          DirectHitDamage,
		Unlocked:        true,
		RoundProjectile: true,
		DamagesTerrain:  true,
	}
}
