package weapons

func Training() Weapon {
	return Weapon{
		Name:      "Spurgeschoß",
		Color:     whiteProjectileColor(),
		Damage:    DirectHitDamage,
		Unlocked:  true,
		ShowTrail: true,
	}
}
