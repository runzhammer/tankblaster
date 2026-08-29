package weapons

const (
	grenadeScale = 0.6
)

func Grenade() Weapon {
	return Weapon{
		Name:                   "Granate",
		Color:                  whiteProjectileColor(),
		Damage:                 DirectHitDamage,
		ImpactCycles:           1,
		Unlocked:               true,
		RoundProjectile:        true,
		DamagesTerrain:         true,
		ProjectileScale:        grenadeScale,
		ImpactAnimationSeconds: DefaultImpactAnimationSeconds,
	}
}
