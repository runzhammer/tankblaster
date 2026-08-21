package weapons

const (
	largeGrenadeScale       = 1.8
	largeGrenadeImpactScale = 3.0
)

func LargeGrenade() Weapon {
	return Weapon{
		Name:            "Große Granate",
		Color:           whiteProjectileColor(),
		Damage:          DirectHitDamage,
		Unlocked:        true,
		RoundProjectile: true,
		DamagesTerrain:  true,
		ProjectileScale: largeGrenadeScale,
		ImpactScale:     largeGrenadeImpactScale,
	}
}
