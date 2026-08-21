package tankblaster

const (
	scrollOMatItemName     = "Scroll-o-Mat"
	scrollOMatKeyboardStep = 16
)

func (s *GameScene) scrollOMatItemIndex() int {
	for index, item := range shopItems() {
		if item.name == scrollOMatItemName {
			return index
		}
	}
	return -1
}

func (s *GameScene) isScrollOMatItem(itemIndex int) bool {
	return itemIndex >= 0 && itemIndex == s.scrollOMatItemIndex()
}

func (s *GameScene) scrollOMatActive() bool {
	tank := s.activeTank()
	if tank == nil {
		return false
	}
	return s.isScrollOMatItem(s.itemIndexForWeaponSlot(tank.selectedWeapon))
}
