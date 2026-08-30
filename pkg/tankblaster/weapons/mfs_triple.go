package weapons

func MFSTriple() Weapon {
	return withDefaultSounds(Weapon{
		Name:                   "MFS 3-fach",
		Color:                  whiteProjectileColor(),
		Damage:                 DirectHitDamage,
		Unlocked:               true,
		RoundProjectile:        true,
		DamagesTerrain:         true,
		ProjectileScale:        largeGrenadeScale,
		ImpactScale:            largeGrenadeImpactScale,
		ImpactAnimationSeconds: DefaultImpactAnimationSeconds,
		TripleShot:             true,
	})
}
