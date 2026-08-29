package weapons

func AirStrike() Weapon {
	return Weapon{
		Name:                   "Luftschlag",
		Color:                  whiteProjectileColor(),
		Damage:                 DirectHitDamage,
		Unlocked:               true,
		RoundProjectile:        true,
		ImpactAnimationSeconds: DefaultImpactAnimationSeconds,
		AirStrike:              true,
	}
}
