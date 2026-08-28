package weapons

func Mosquitos() Weapon {
	return Weapon{
		Name:            "Moskitos",
		Color:           whiteProjectileColor(),
		Damage:          DirectHitDamage,
		Unlocked:        true,
		RoundProjectile: true,
		ProjectileScale: grenadeScale,
		Mosquitos:       true,
	}
}
