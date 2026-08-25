package weapons

const (
	fireballImpactDamage       = 40
	fireballImpactExtraSeconds = 0.4
)

func Fireball() Weapon {
	return Weapon{
		Name:                        "Feuerkugel",
		Color:                       whiteProjectileColor(),
		Damage:                      DirectHitDamage,
		Unlocked:                    true,
		RoundProjectile:             true,
		ImpactAnimationExtraSeconds: fireballImpactExtraSeconds,
		ImpactAnimationStyle:        ImpactAnimationFireball,
		ImpactDamage:                fireballImpactDamage,
	}
}
