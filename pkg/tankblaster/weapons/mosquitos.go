package weapons

func Mosquitos() Weapon {
	return withDefaultSounds(Weapon{
		Name:                   "Moskitos",
		Color:                  whiteProjectileColor(),
		Damage:                 DirectHitDamage,
		Unlocked:               true,
		RoundProjectile:        true,
		ProjectileScale:        grenadeScale,
		ImpactAnimationSeconds: DefaultImpactAnimationSeconds,
		Mosquitos:              true,
	})
}
