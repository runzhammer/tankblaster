package weapons

func Training() Weapon {
	return Weapon{
		Name:      "Training",
		Color:     whiteProjectileColor(),
		Damage:    DirectHitDamage,
		Unlocked:  true,
		ShowTrail: true,
	}
}
