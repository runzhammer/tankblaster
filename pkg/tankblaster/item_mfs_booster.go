package tankblaster

const (
	mfsBoosterItemName        = "MFS Verstärker"
	mfsBoosterUsesPerPurchase = 6
)

func (s *GameScene) mfsBoosterItemIndex() int {
	return shopItemIndexByName(mfsBoosterItemName)
}

func (s *GameScene) isMFSBoosterItem(itemIndex int) bool {
	return itemIndex >= 0 && itemIndex == s.mfsBoosterItemIndex()
}

func shopItemIndexByName(name string) int {
	for index, item := range shopItems() {
		if item.name == name {
			return index
		}
	}
	return -1
}

func (s *GameScene) consumeMFSBoosterCharge(playerIndex int) bool {
	if playerIndex < 0 || playerIndex >= len(s.inventories) {
		return false
	}
	s.ensureInventory(playerIndex)
	if s.inventories[playerIndex].mfsBoosterCharges <= 0 {
		return false
	}
	s.inventories[playerIndex].mfsBoosterCharges--
	return true
}

func (s *GameScene) mfsBoosterCountForPlayer(playerIndex int) int {
	if playerIndex < 0 || playerIndex >= len(s.inventories) {
		return 0
	}
	charges := s.inventories[playerIndex].mfsBoosterCharges
	if charges <= 0 {
		return 0
	}
	return (charges + mfsBoosterUsesPerPurchase - 1) / mfsBoosterUsesPerPurchase
}
