package weapons

func MFSTriple() Weapon {
	return withDefaultSounds(Weapon{
		Name:                   "MFS 3-fach",
		Color:                  whiteProjectileColor(),
		RadialDamage:           RadialDamageProfile{InnerRadius: 32, OuterRadius: 48, MaxDamage: 100},
		Unlocked:               true,
		RoundProjectile:        true,
		DamagesTerrain:         true,
		ProjectileScale:        largeGrenadeScale,
		ImpactScale:            largeGrenadeImpactScale,
		ImpactAnimationSeconds: DefaultImpactAnimationSeconds,
		TripleShot:             true,
	})
}
