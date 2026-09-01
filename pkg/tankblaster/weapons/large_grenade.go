package weapons

const (
	largeGrenadeScale       = 1.2
	largeGrenadeImpactScale = 3.0
)

func LargeGrenade() Weapon {
	return withDefaultSounds(Weapon{
		Name:                   "Große Granate",
		Color:                  whiteProjectileColor(),
		RadialDamage:           RadialDamageProfile{InnerRadius: 32, OuterRadius: 48, MaxDamage: 100},
		Unlocked:               true,
		RoundProjectile:        true,
		DamagesTerrain:         true,
		ProjectileScale:        largeGrenadeScale,
		ImpactScale:            largeGrenadeImpactScale,
		ImpactAnimationSeconds: DefaultImpactAnimationSeconds,
	})
}
