package weapons

func Laser() Weapon {
	return Weapon{
		Name:            "Laser",
		Color:           whiteProjectileColor(),
		Damage:          DirectHitDamage,
		Unlocked:        true,
		RoundProjectile: true,
		ProjectileScale: grenadeScale,
		Laser:           true,
	}
}
