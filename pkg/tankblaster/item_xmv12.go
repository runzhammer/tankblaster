package tankblaster

const (
	xmV12ItemName          = "XM-V12 Panzer"
	dieselItemName         = "Diesel (F54)"
	xmV12InitialDiesel     = 300
	xmV12DebugDiesel       = 1000
	xmV12MaxDiesel         = 1400
	xmV12DieselPerPurchase = 100
	xmV12DriveSpeed        = 1.35
	xmV12DieselPerFrame    = 0.35
)

func (s *GameScene) xmV12ItemIndex() int {
	return shopItemIndexByName(xmV12ItemName)
}

func (s *GameScene) dieselItemIndex() int {
	return shopItemIndexByName(dieselItemName)
}

func (s *GameScene) isXMV12Item(itemIndex int) bool {
	return itemIndex >= 0 && itemIndex == s.xmV12ItemIndex()
}

func (s *GameScene) isDieselItem(itemIndex int) bool {
	return itemIndex >= 0 && itemIndex == s.dieselItemIndex()
}

func (s *GameScene) playerHasXMV12(playerIndex int) bool {
	return playerIndex >= 0 && playerIndex < len(s.inventories) && s.inventories[playerIndex].hasXMV12
}

func (s *GameScene) dieselForPlayer(playerIndex int) int {
	if playerIndex < 0 || playerIndex >= len(s.inventories) {
		return 0
	}
	return int(s.inventories[playerIndex].diesel)
}

func (s *GameScene) addDiesel(playerIndex, amount int) {
	if playerIndex < 0 || amount <= 0 {
		return
	}
	s.ensureInventory(playerIndex)
	s.inventories[playerIndex].diesel = minFloat(float64(xmV12MaxDiesel), s.inventories[playerIndex].diesel+float64(amount))
}

func (s *GameScene) buyXMV12(playerIndex int) {
	if playerIndex < 0 {
		return
	}
	s.ensureInventory(playerIndex)
	if !s.inventories[playerIndex].hasXMV12 {
		s.inventories[playerIndex].hasXMV12 = true
		s.addDiesel(playerIndex, xmV12InitialDiesel)
		s.replaceTankModel(playerIndex)
	}
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
