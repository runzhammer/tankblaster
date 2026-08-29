package weapons

func Shockwave() Weapon {
	return Weapon{
		Name:                   "Schockwelle",
		Color:                  whiteProjectileColor(),
		Damage:                 DirectHitDamage,
		Unlocked:               true,
		RoundProjectile:        true,
		ImpactScale:            plasmaImpactScale,
		ImpactAnimationSeconds: DefaultImpactAnimationSeconds,
		Shockwave:              true,
	}
}
