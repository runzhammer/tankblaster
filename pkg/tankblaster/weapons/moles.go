package weapons

func Moles() Weapon {
	return Weapon{
		Name:            "Maulwürfe",
		Color:           whiteProjectileColor(),
		Damage:          DirectHitDamage,
		Unlocked:        true,
		RoundProjectile: true,
		Moles:           true,
	}
}
