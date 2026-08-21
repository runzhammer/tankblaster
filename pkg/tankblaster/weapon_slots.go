package tankblaster

import weaponspkg "github.com/runzhammer/gamedemo/pkg/tankblaster/weapons"

func gameWeapons() []weaponspkg.Weapon {
	return weaponspkg.List()
}

func (s *GameScene) weaponForProjectile(p *projectile) weaponspkg.Weapon {
	weapons := gameWeapons()
	if p != nil && p.weaponIndex == 0 {
		return weapons[0]
	}
	if p != nil && s.itemIndexForWeaponSlot(p.weaponIndex) == 1 {
		return weapons[2]
	}
	if p != nil && s.itemIndexForWeaponSlot(p.weaponIndex) == 2 {
		return weapons[3]
	}
	return weapons[1]
}

func projectileRadiusForWeapon(weapon weaponspkg.Weapon) float64 {
	return weaponspkg.ProjectileRadius(weapon, projectileRadius)
}

func impactRadiusForWeapon(weapon weaponspkg.Weapon) float64 {
	return weaponspkg.ImpactRadius(weapon, projectileRadius*impactRadiusMultiplier)
}

func impactCyclesForWeapon(weapon weaponspkg.Weapon) int {
	return weaponspkg.ImpactCycles(weapon)
}
