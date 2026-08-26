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
	if p != nil && s.itemIndexForWeaponSlot(p.weaponIndex) == 3 {
		if len(weapons) > 4 {
			return weapons[4]
		}
	}
	if p != nil && s.itemIndexForWeaponSlot(p.weaponIndex) == 4 {
		if len(weapons) > 5 {
			return weapons[5]
		}
	}
	if p != nil && s.itemIndexForWeaponSlot(p.weaponIndex) == 5 {
		if len(weapons) > 6 {
			return weapons[6]
		}
	}
	if p != nil && s.itemIndexForWeaponSlot(p.weaponIndex) == 6 {
		if len(weapons) > 7 {
			return weapons[7]
		}
	}
	if p != nil && s.itemIndexForWeaponSlot(p.weaponIndex) == 7 {
		if len(weapons) > 8 {
			return weapons[8]
		}
	}
	if p != nil && s.itemIndexForWeaponSlot(p.weaponIndex) == 8 {
		if len(weapons) > 9 {
			return weapons[9]
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
