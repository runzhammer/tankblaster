package weapons

func Laser() Weapon {
	return withDefaultSounds(Weapon{
		Name:                   "Laser",
		Color:                  whiteProjectileColor(),
		Damage:                 DirectHitDamage,
		Unlocked:               true,
		RoundProjectile:        true,
		ProjectileScale:        grenadeScale,
		ImpactAnimationSeconds: DefaultImpactAnimationSeconds,
		Laser:                  true,
	})
}
