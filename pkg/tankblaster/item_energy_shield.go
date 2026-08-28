package tankblaster

import "math"

const (
	energyShieldItemName            = "Energieschild"
	energyShieldMaxPercent          = 100.0
	energyShieldPercentPerPurchase  = 100.0
	energyShieldDamageMultiplier    = 0.25
	energyShieldCostPerDamagePoint  = 0.5
	energyShieldDamageAbsorbPerUnit = (1.0 - energyShieldDamageMultiplier) / energyShieldCostPerDamagePoint
)

func (s *GameScene) energyShieldItemIndex() int {
	return shopItemIndexByName(energyShieldItemName)
}

func (s *GameScene) isEnergyShieldItem(itemIndex int) bool {
	return itemIndex >= 0 && itemIndex == s.energyShieldItemIndex()
}

func (s *GameScene) addEnergyShield(playerIndex, purchases int) {
	if playerIndex < 0 || purchases <= 0 {
		return
	}
	s.ensureInventory(playerIndex)
	s.inventories[playerIndex].energyShield = math.Min(
		energyShieldMaxPercent,
		s.inventories[playerIndex].energyShield+float64(purchases)*energyShieldPercentPerPurchase,
	)
}

func (s *GameScene) energyShieldPercentForPlayer(playerIndex int) int {
	if playerIndex < 0 || playerIndex >= len(s.inventories) {
		return 0
	}
	return maxInt(0, minInt(int(math.Round(s.inventories[playerIndex].energyShield)), int(energyShieldMaxPercent)))
}

func (s *GameScene) applyEnergyShieldDamage(tank *battleTank, damage int) int {
	if tank == nil || damage <= 0 || tank.playerIndex < 0 || tank.playerIndex >= len(s.inventories) {
		return damage
	}
	s.ensureInventory(tank.playerIndex)
	shield := s.inventories[tank.playerIndex].energyShield
	if shield <= 0 {
		return damage
	}

	requiredShield := float64(damage) * energyShieldCostPerDamagePoint
	if shield >= requiredShield {
		s.inventories[tank.playerIndex].energyShield = math.Max(0, shield-requiredShield)
		return maxInt(1, int(math.Round(float64(damage)*energyShieldDamageMultiplier)))
	}

	absorbedDamage := shield * energyShieldDamageAbsorbPerUnit
	s.inventories[tank.playerIndex].energyShield = 0
	return maxInt(1, int(math.Round(float64(damage)-absorbedDamage)))
}
