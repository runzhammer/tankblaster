package weapons

func Moles() Weapon {
	return withDefaultSounds(Weapon{
		Name:                   "Maulwürfe",
		Color:                  whiteProjectileColor(),
		Damage:                 DirectHitDamage,
		Unlocked:               true,
		RoundProjectile:        true,
		ImpactAnimationSeconds: DefaultImpactAnimationSeconds,
		Moles:                  true,
	})
}
