package weapons

import "github.com/runzhammer/gamedemo/pkg/tankblaster/soundpaths"

const (
	splitterBombFragmentScale = grenadeScale * 0.65
)

func SplitterBomb() Weapon {
	return withDefaultSounds(Weapon{
		Name:                   "Splitterbombe",
		Color:                  whiteProjectileColor(),
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
	fragment.RadialDamage = RadialDamageProfile{InnerRadius: 10, OuterRadius: 15, MaxDamage: 100}
	fragment.ProjectileScale = splitterBombFragmentScale
	fragment.ImpactSound = soundpaths.SoundClusterexplo
	return fragment
}
