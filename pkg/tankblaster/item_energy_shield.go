package tankblaster

const (
	energyShieldItemName           = "Energieschild"
	energyShieldPercentPerPurchase = 100
	energyShieldDamageNumerator    = 3
	energyShieldDamageDenominator  = 7
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
	s.inventories[playerIndex].energyShield += purchases * energyShieldPercentPerPurchase
}

func (s *GameScene) energyShieldPercentForPlayer(playerIndex int) int {
	if playerIndex < 0 || playerIndex >= len(s.inventories) {
		return 0
	}
	return maxInt(0, s.inventories[playerIndex].energyShield)
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

	shield -= (energyShieldDamageNumerator * damage) / energyShieldDamageDenominator
	s.inventories[tank.playerIndex].energyShield = shield
	if shield < 0 {
		s.inventories[tank.playerIndex].energyShield = 0
		return -shield
	}
	return 0
}
