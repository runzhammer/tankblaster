package weapons

func Moles() Weapon {
	return withDefaultSounds(Weapon{
		Name:                   "Maulwürfe",
		Color:                  whiteProjectileColor(),
		RadialDamage:           RadialDamageProfile{InnerRadius: 53, OuterRadius: 80, MaxDamage: 100},
		Unlocked:               true,
		RoundProjectile:        true,
		ImpactAnimationSeconds: DefaultImpactAnimationSeconds,
		Moles:                  true,
	})
}
