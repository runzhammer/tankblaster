package weapons

import "image/color"

const DirectHitDamage = 100

type Weapon struct {
	Name                        string
	Color                       color.RGBA
	Damage                      int
	Unlocked                    bool
	ShowTrail                   bool
	RoundProjectile             bool
	DamagesTerrain              bool
	ProjectileScale             float64
	ImpactScale                 float64
	ImpactAnimationExtraSeconds float64
	ImpactCycles                int
	ImpactGradientOutward       bool
	ImpactAnimationStyle        ImpactAnimationStyle
	PlantsPalm                  bool
	ImpactDamage                int
	FillsWater                  bool
	Moles                       bool
	TripleShot                  bool
	SmallCrumblers              bool
	LargeCrumblers              bool
	SurpriseEgg                 bool
	Utility                     bool
}

type ImpactAnimationStyle uint8

const (
	ImpactAnimationDefault ImpactAnimationStyle = iota
	ImpactAnimationHBomb
	ImpactAnimationPlasma
	ImpactAnimationFireball
)

func List() []Weapon {
	return []Weapon{
		Training(),
		Grenade(),
		LargeGrenade(),
		AtomBomb(),
		HBomb(),
		PlasmaMelter(),
		WonderPalm(),
		Fireball(),
		Water(),
		Moles(),
		MFSTriple(),
		SmallCrumblers(),
		LargeCrumblers(),
		SurpriseEgg(),
		FutureWeapon("Moskitos"),
		FutureWeapon("Schockwelle"),
		FutureWeapon("Luftschlag"),
		FutureWeapon("Splitterbombe"),
		FutureWeapon("Laser"),
		UtilityItem("Scroll-o-Mat"),
		UtilityItem("Energieschild"),
		UtilityItem("MFS Verstärker"),
		UtilityItem("XM-V12 Panzer"),
		UtilityItem("Diesel (F54)"),
	}
}

func FutureWeapon(name string) Weapon {
	return Weapon{
		Name:            name,
		Color:           whiteProjectileColor(),
		Damage:          DirectHitDamage,
		Unlocked:        true,
		RoundProjectile: true,
		DamagesTerrain:  true,
		ProjectileScale: grenadeScale,
	}
}

func UtilityItem(name string) Weapon {
	return Weapon{
		Name:     name,
		Unlocked: true,
		Utility:  true,
	}
}

func ProjectileRadius(weapon Weapon, baseRadius float64) float64 {
	scale := weapon.ProjectileScale
	if scale <= 0 {
		scale = 1
	}
	return baseRadius * scale
}

func ImpactRadius(weapon Weapon, baseRadius float64) float64 {
	scale := weapon.ImpactScale
	if scale <= 0 {
		scale = 1
	}
	return baseRadius * scale
}

func ImpactCycles(weapon Weapon) int {
	if weapon.ImpactCycles <= 0 {
		return 2
	}
	return weapon.ImpactCycles
}

func whiteProjectileColor() color.RGBA {
	return color.RGBA{R: 238, G: 238, B: 238, A: 255}
}
