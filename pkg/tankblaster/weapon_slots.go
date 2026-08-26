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
	if p != nil {
		weaponIndex := s.itemIndexForWeaponSlot(p.weaponIndex) + 1
		if weaponIndex > 0 && weaponIndex < len(weapons) {
			return weapons[weaponIndex]
		}
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
