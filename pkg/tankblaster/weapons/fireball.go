package weapons

const (
	fireballImpactDamage           = 40
	fireballImpactAnimationSeconds = 0.6
)

func Fireball() Weapon {
	return Weapon{
		Name:                   "Feuerkugel",
		Color:                  whiteProjectileColor(),
		Damage:                 DirectHitDamage,
		Unlocked:               true,
		RoundProjectile:        true,
		ImpactAnimationSeconds: fireballImpactAnimationSeconds,
		ImpactAnimationStyle:   ImpactAnimationFireball,
		ImpactDamage:           fireballImpactDamage,
	}
}
