package weapons

const (
	largeGrenadeScale       = 1.2
	largeGrenadeImpactScale = 3.0
)

func LargeGrenade() Weapon {
	return withDefaultSounds(Weapon{
		Name:                   "Große Granate",
		Color:                  whiteProjectileColor(),
		Damage:                 DirectHitDamage,
		Unlocked:               true,
		RoundProjectile:        true,
		DamagesTerrain:         true,
		ProjectileScale:        largeGrenadeScale,
		ImpactScale:            largeGrenadeImpactScale,
		ImpactAnimationSeconds: DefaultImpactAnimationSeconds,
	})
}
