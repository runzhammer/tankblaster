package weapons

func Shockwave() Weapon {
	return withDefaultSounds(Weapon{
		Name:                   "Schockwelle",
		Color:                  whiteProjectileColor(),
		RadialDamage:           RadialDamageProfile{InnerRadius: 100, OuterRadius: 225, MaxDamage: 100},
		Unlocked:               true,
		RoundProjectile:        true,
		ImpactScale:            plasmaImpactScale,
		ImpactAnimationSeconds: DefaultImpactAnimationSeconds,
		Shockwave:              true,
	})
}
