package weapons

const (
	splitterBombFragmentScale = grenadeScale * 0.65
)

func SplitterBomb() Weapon {
	return withDefaultSounds(Weapon{
		Name:                   "Splitterbombe",
		Color:                  whiteProjectileColor(),
		Damage:                 DirectHitDamage,
		Unlocked:               true,
		RoundProjectile:        true,
		DamagesTerrain:         true,
		ProjectileScale:        grenadeScale,
		ImpactAnimationSeconds: DefaultImpactAnimationSeconds,
		SplitterBomb:           true,
	})
}

func SplitterBombFragment() Weapon {
	fragment := Grenade()
	fragment.ProjectileScale = splitterBombFragmentScale
	return fragment
}
