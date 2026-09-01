package weapons

const (
	fireballImpactAnimationSeconds = 0.6
)

func Fireball() Weapon {
	return withDefaultSounds(Weapon{
		Name:                   "Feuerkugel",
		Color:                  whiteProjectileColor(),
		RadialDamage:           RadialDamageProfile{InnerRadius: 19, OuterRadius: 57, MaxDamage: 100},
		Unlocked:               true,
		RoundProjectile:        true,
		ImpactAnimationSeconds: fireballImpactAnimationSeconds,
		ImpactAnimationStyle:   ImpactAnimationFireball,
	})
}
