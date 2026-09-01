package weapons

func AirStrike() Weapon {
	return withDefaultSounds(Weapon{
		Name:                   "Luftschlag",
		Color:                  whiteProjectileColor(),
		RadialDamage:           RadialDamageProfile{InnerRadius: 60, OuterRadius: 90, MaxDamage: 100},
		Unlocked:               true,
		RoundProjectile:        true,
		ImpactAnimationSeconds: DefaultImpactAnimationSeconds,
		AirStrike:              true,
	})
}
