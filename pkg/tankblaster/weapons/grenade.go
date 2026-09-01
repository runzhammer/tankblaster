package weapons

const (
	grenadeScale = 0.6
)

func Grenade() Weapon {
	return withDefaultSounds(Weapon{
		Name:                   "Granate",
		Color:                  whiteProjectileColor(),
		RadialDamage:           RadialDamageProfile{InnerRadius: 17, OuterRadius: 33, MaxDamage: 100},
		ImpactCycles:           1,
		Unlocked:               true,
		RoundProjectile:        true,
		DamagesTerrain:         true,
		ProjectileScale:        grenadeScale,
		ImpactAnimationSeconds: DefaultImpactAnimationSeconds,
	})
}
