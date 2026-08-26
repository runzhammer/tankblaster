package weapons

func MFSTriple() Weapon {
	return Weapon{
		Name:            "MFS 3-fach",
		Color:           whiteProjectileColor(),
		Damage:          DirectHitDamage,
		Unlocked:        true,
		RoundProjectile: true,
		DamagesTerrain:  true,
		ProjectileScale: largeGrenadeScale,
		ImpactScale:     largeGrenadeImpactScale,
		TripleShot:      true,
	}
}
