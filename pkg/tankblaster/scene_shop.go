package tankblaster

import (
	"image"
	"image/color"
	"math"
	"sort"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/runzhammer/gamedemo/pkg/core"
	"golang.org/x/image/colornames"
)

type shopMode uint8

const (
	shopModeEntry shopMode = iota
	shopModeClassA
	shopModeClassB
)

const classBDiscountMultiplier = 2105.0 / 3175.0

const shopListKeyRepeatFrames = 9

const (
	shopComputerEntryDelay    = 35
	shopComputerNavigateDelay = 6
	shopComputerBuyDelay      = 28
	shopComputerBackDelay     = 24
	shopComputerContinueDelay = 35
)

type shopComputerPhase uint8

const (
	shopComputerEnterList shopComputerPhase = iota
	shopComputerNavigate
	shopComputerBuy
	shopComputerBack
	shopComputerContinue
)

type shopComputerPlan struct {
	playerIndex int
	phase       shopComputerPhase
	delay       int
	mode        shopMode
	targetIndex int
	purchases   int
}

var (
	shopMainLeftOverlayRect  = image.Rect(130, 110, 400, 240)
	shopMainRightOverlayRect = image.Rect(270, 100, 499, 240)
)

type shopInventory struct {
	classA []int
	classB []int
}

type shopAssets struct {
	human               *ebiten.Image
	computer            *ebiten.Image
	storeBg             *ebiten.Image
	storeMainLeft       *ebiten.Image
	storeMainRight      *ebiten.Image
	storeRoll           *ebiten.Image
	storeIcons          *ebiten.Image
	weaponbarActive     *ebiten.Image
	weaponbarOnStock    *ebiten.Image
	weaponbarOutOfStock *ebiten.Image
}

func makeShopInventories(playerCount int) []shopInventory {
	items := shopItems()
	inventories := make([]shopInventory, playerCount)
	for i := range inventories {
		inventories[i] = shopInventory{
			classA: make([]int, len(items)),
			classB: make([]int, len(items)),
		}
		if len(inventories[i].classA) > 0 {
			inventories[i].classA[0] = items[0].stock
		}
		if core.Config().Debug.Enabled && len(inventories[i].classA) > 1 {
			inventories[i].classA[1] = 50
		}
		if core.Config().Debug.Enabled && len(inventories[i].classA) > 2 {
			inventories[i].classA[2] = 50
		}
		if core.Config().Debug.Enabled && len(inventories[i].classA) > 3 {
			inventories[i].classA[3] = 50
		}
		if core.Config().Debug.Enabled && len(inventories[i].classA) > 4 {
			inventories[i].classA[4] = 50
		}
		if core.Config().Debug.Enabled && len(inventories[i].classA) > 5 {
			inventories[i].classA[5] = 50
		}
		if core.Config().Debug.Enabled && len(inventories[i].classA) > 6 {
			inventories[i].classA[6] = 50
		}
		if core.Config().Debug.Enabled && len(inventories[i].classA) > 7 {
			inventories[i].classA[7] = 50
		}
		if core.Config().Debug.Enabled && len(inventories[i].classA) > 8 {
			inventories[i].classA[8] = 50
		}
		if core.Config().Debug.Enabled && len(inventories[i].classA) > 9 {
			inventories[i].classA[9] = 50
		}
		if core.Config().Debug.Enabled {
			for itemIndex, item := range items {
				if item.name == scrollOMatItemName {
					inventories[i].classA[itemIndex] = 1
					break
				}
			}
		}
	}
	return inventories
}

func (s *GameScene) beginShop() {
	s.phase = phaseShop
	s.shopPlayerCursor = 0
	s.shopHoverClass = 0
	s.shopMode = shopModeEntry
	s.shopSelectedIndex = 0
	s.shopComputerPlan = nil
	s.shopPlayerOrder = make([]int, len(s.players))
	for i := range s.shopPlayerOrder {
		s.shopPlayerOrder[i] = i
	}
	s.resetShopStock()
	s.pickClassBItems()
	sort.SliceStable(s.shopPlayerOrder, func(i, j int) bool {
		left := s.shopPlayerOrder[i]
		right := s.shopPlayerOrder[j]
		return s.roundScoreForPlayer(left) > s.roundScoreForPlayer(right)
	})
}

func (s *GameScene) handleShopInput() {
	if s.currentShopPlayerIsComputer() {
		s.handleComputerShopInput()
		return
	}

	x, y := ebiten.CursorPosition()
	cursor := image.Pt(x, y)
	if s.shopMode != shopModeEntry {
		s.handleShopListInput(cursor)
		return
	}

	switch {
	case cursor.In(s.shopClassARect()):
		s.shopHoverClass = 1
	case cursor.In(s.shopClassBRect()):
		s.shopHoverClass = 2
	default:
		s.shopHoverClass = 0
	}

	if !inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		return
	}
	if cursor.In(s.shopClassARect()) {
		s.shopMode = shopModeClassA
		s.shopSelectedIndex = 0
		return
	}
	if cursor.In(s.shopClassBRect()) {
		s.shopMode = shopModeClassB
		s.shopSelectedIndex = 0
		return
	}
	if !cursor.In(s.shopContinueRect()) {
		return
	}
	if s.shopPlayerCursor < len(s.shopPlayerOrder)-1 {
		s.advanceShopPlayer()
		return
	}
	s.startRound()
}

func (s *GameScene) handleComputerShopInput() {
	playerIndex := s.currentShopPlayerIndex()
	if playerIndex < 0 {
		return
	}
	if s.shopComputerPlan == nil || s.shopComputerPlan.playerIndex != playerIndex {
		s.shopComputerPlan = s.newComputerShopPlan(playerIndex)
	}
	plan := s.shopComputerPlan
	if plan == nil {
		return
	}
	if plan.delay > 0 {
		plan.delay--
		return
	}

	switch plan.phase {
	case shopComputerEnterList:
		s.shopMode = plan.mode
		s.shopSelectedIndex = 0
		s.shopHoverClass = 0
		plan.phase = shopComputerNavigate
		plan.delay = shopComputerNavigateDelay
	case shopComputerNavigate:
		items := s.visibleShopItemIndexes()
		if len(items) == 0 {
			plan.phase = shopComputerBack
			plan.delay = shopComputerBackDelay
			return
		}
		plan.targetIndex = maxInt(0, minInt(plan.targetIndex, len(items)-1))
		if s.shopSelectedIndex < plan.targetIndex {
			s.shopSelectedIndex++
			plan.delay = shopComputerNavigateDelay
			return
		}
		if s.shopSelectedIndex > plan.targetIndex {
			s.shopSelectedIndex--
			plan.delay = shopComputerNavigateDelay
			return
		}
		plan.phase = shopComputerBuy
		plan.delay = shopComputerBuyDelay
	case shopComputerBuy:
		before := s.credits[playerIndex]
		s.buySelectedShopItem()
		if s.credits[playerIndex] < before {
			plan.purchases++
		}
		next, ok := s.nextComputerShopChoiceInMode(playerIndex, plan.mode, plan.purchases)
		if !ok {
			plan.phase = shopComputerBack
			plan.delay = shopComputerBackDelay
			return
		}
		plan.targetIndex = next.targetIndex
		plan.phase = shopComputerNavigate
		plan.delay = shopComputerNavigateDelay
	case shopComputerBack:
		s.shopMode = shopModeEntry
		s.shopHoverClass = 0
		plan.phase = shopComputerContinue
		plan.delay = shopComputerContinueDelay
	case shopComputerContinue:
		s.shopComputerPlan = nil
		if s.shopPlayerCursor < len(s.shopPlayerOrder)-1 {
			s.advanceShopPlayer()
			return
		}
		s.startRound()
	}
}

type computerShopChoice struct {
	mode        shopMode
	targetIndex int
}

func (s *GameScene) newComputerShopPlan(playerIndex int) *shopComputerPlan {
	choice, ok := s.nextComputerShopChoice(playerIndex, 0)
	if !ok {
		return &shopComputerPlan{
			playerIndex: playerIndex,
			phase:       shopComputerContinue,
			delay:       shopComputerContinueDelay,
		}
	}
	if choice.mode == shopModeClassB {
		s.shopHoverClass = 2
	} else {
		s.shopHoverClass = 1
	}
	return &shopComputerPlan{
		playerIndex: playerIndex,
		phase:       shopComputerEnterList,
		delay:       shopComputerEntryDelay,
		mode:        choice.mode,
		targetIndex: choice.targetIndex,
	}
}

func (s *GameScene) nextComputerShopChoice(playerIndex, purchases int) (computerShopChoice, bool) {
	if playerIndex < 0 || playerIndex >= len(s.credits) || s.credits[playerIndex] <= 0 {
		return computerShopChoice{}, false
	}
	if purchases > 0 && s.rng.Float64() < 0.35+float64(purchases)*0.12 {
		return computerShopChoice{}, false
	}
	choices := s.affordableComputerShopChoices(playerIndex, shopModeClassB)
	if len(choices) > 0 && s.rng.Float64() < 0.68 {
		return choices[s.rng.Intn(len(choices))], true
	}
	choices = s.affordableComputerShopChoices(playerIndex, shopModeClassA)
	if len(choices) > 0 {
		return choices[s.rng.Intn(len(choices))], true
	}
	choices = s.affordableComputerShopChoices(playerIndex, shopModeClassB)
	if len(choices) > 0 {
		return choices[s.rng.Intn(len(choices))], true
	}
	return computerShopChoice{}, false
}

func (s *GameScene) nextComputerShopChoiceInMode(playerIndex int, mode shopMode, purchases int) (computerShopChoice, bool) {
	if playerIndex < 0 || playerIndex >= len(s.credits) || s.credits[playerIndex] <= 0 {
		return computerShopChoice{}, false
	}
	if s.rng.Float64() < 0.35+float64(purchases)*0.12 {
		return computerShopChoice{}, false
	}
	choices := s.affordableComputerShopChoices(playerIndex, mode)
	if len(choices) == 0 {
		return computerShopChoice{}, false
	}
	return choices[s.rng.Intn(len(choices))], true
}

func (s *GameScene) affordableComputerShopChoices(playerIndex int, mode shopMode) []computerShopChoice {
	previousMode := s.shopMode
	previousSelection := s.shopSelectedIndex
	s.shopMode = mode
	defer func() {
		s.shopMode = previousMode
		s.shopSelectedIndex = previousSelection
	}()

	items := s.visibleShopItemIndexes()
	choices := make([]computerShopChoice, 0, len(items))
	for listIndex, itemIndex := range items {
		s.shopSelectedIndex = listIndex
		price := s.shopPriceForItem(itemIndex)
		if price <= 0 || s.credits[playerIndex] < price || s.shopStockForSelected(itemIndex) <= 0 {
			continue
		}
		if s.isScrollOMatItem(itemIndex) && s.shopItemCountForPlayer(playerIndex, itemIndex) > 0 {
			continue
		}
		choices = append(choices, computerShopChoice{mode: mode, targetIndex: listIndex})
	}
	return choices
}

func (s *GameScene) advanceShopPlayer() {
	s.shopPlayerCursor++
	s.shopHoverClass = 0
	s.shopMode = shopModeEntry
	s.shopSelectedIndex = 0
	s.shopComputerPlan = nil
}

func (s *GameScene) currentShopPlayerIsComputer() bool {
	playerIndex := s.currentShopPlayerIndex()
	return playerIndex >= 0 && playerIndex < len(s.players) && s.players[playerIndex].Kind == PlayerComputer
}

func (s *GameScene) drawShop(screen *ebiten.Image) {
	if s.shopMode != shopModeEntry {
		s.drawShopList(screen)
		return
	}

	s.drawStoreImage(screen, s.shop.storeBg)
	switch s.shopHoverClass {
	case 1:
		s.drawStoreOverlay(screen, s.shop.storeMainLeft, shopMainLeftOverlayRect)
	case 2:
		s.drawStoreOverlay(screen, s.shop.storeMainRight, shopMainRightOverlayRect)
	}

	playerIndex := s.currentShopPlayerIndex()
	player := PlayerConfig{Name: "Spieler"}
	if playerIndex >= 0 && playerIndex < len(s.players) {
		player = s.players[playerIndex]
	}

	s.drawShopPlayerPanel(screen, playerIndex, player)
	drawButton(screen, s.shopContinueRect(), "Weiter")
}

func (s *GameScene) handleShopListInput(cursor image.Point) {
	items := s.visibleShopItemIndexes()
	if len(items) == 0 {
		return
	}
	if shouldNavigateShopList(ebiten.KeyArrowUp) {
		s.shopSelectedIndex = maxInt(0, s.shopSelectedIndex-1)
	}
	if shouldNavigateShopList(ebiten.KeyArrowDown) {
		s.shopSelectedIndex = minInt(len(items)-1, s.shopSelectedIndex+1)
	}
	_, wheelY := ebiten.Wheel()
	if wheelY > 0 {
		s.shopSelectedIndex = maxInt(0, s.shopSelectedIndex-int(math.Ceil(wheelY)))
	}
	if wheelY < 0 {
		s.shopSelectedIndex = minInt(len(items)-1, s.shopSelectedIndex+int(math.Ceil(-wheelY)))
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeyKPEnter) {
		s.buySelectedShopItem()
	}
	if !inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		return
	}
	if cursor.In(s.shopBackRect()) {
		s.shopMode = shopModeEntry
		s.shopHoverClass = 0
		return
	}
	if cursor.In(s.shopBuyRect()) {
		s.buySelectedShopItem()
		return
	}
	if index, ok := s.shopListIndexAt(cursor, len(items)); ok {
		s.shopSelectedIndex = index
	}
}

func shouldNavigateShopList(key ebiten.Key) bool {
	if inpututil.IsKeyJustPressed(key) {
		return true
	}
	held := inpututil.KeyPressDuration(key)
	return held >= shopListKeyRepeatFrames && held%shopListKeyRepeatFrames == 0
}

func (s *GameScene) drawShopList(screen *ebiten.Image) {
	s.drawStoreImage(screen, s.shop.storeRoll)

	items := s.visibleShopItemIndexes()
	if len(items) == 0 {
		return
	}
	s.shopSelectedIndex = maxInt(0, minInt(s.shopSelectedIndex, len(items)-1))
	itemIndex := items[s.shopSelectedIndex]
	item := shopItems()[itemIndex]

	s.drawDynamicShopList(screen, items)
	s.drawShopItemDetails(screen, itemIndex, item)
	playerIndex := s.currentShopPlayerIndex()
	player := PlayerConfig{Name: "Spieler"}
	if playerIndex >= 0 && playerIndex < len(s.players) {
		player = s.players[playerIndex]
	}
	s.drawShopPlayerPanel(screen, playerIndex, player)
	drawButton(screen, s.shopBackRect(), "Zurueck >>")
}

func (s *GameScene) drawDynamicShopList(screen *ebiten.Image, indexes []int) {
	titleRect := s.shopStoreScaledRect(image.Rect(210, 36, 430, 66))
	title := "- Klasse A -"
	if s.shopMode == shopModeClassB {
		title = "- Schnäppchen -"
	}
	drawCenteredText(screen, title, titleRect, colornames.Black)

	listRect := s.shopListRect()
	rowH := s.shopListRowHeight()
	visibleRows := s.shopVisibleRows(len(indexes))
	offset := s.shopListScrollOffset(len(indexes))
	for row := 0; row < visibleRows; row++ {
		listIndex := offset + row
		if listIndex < 0 || listIndex >= len(indexes) {
			continue
		}
		itemIndex := indexes[listIndex]
		if itemIndex < 0 || itemIndex >= len(shopItems()) {
			continue
		}
		r := image.Rect(listRect.Min.X, listRect.Min.Y+row*rowH, listRect.Max.X, listRect.Min.Y+(row+1)*rowH)
		if listIndex == s.shopSelectedIndex {
			drawFilledRect(screen, insetRect(r, maxInt(4, r.Dx()/40)), color.RGBA{R: 0, G: 0, B: 55, A: 255})
			drawCenteredText(screen, shopItems()[itemIndex].name, r, colornames.White)
			continue
		}
		drawCenteredText(screen, shopItems()[itemIndex].name, r, colornames.Black)
	}
}

func (s *GameScene) drawShopItemDetails(screen *ebiten.Image, itemIndex int, item shopItem) {
	detailRect := s.shopScaledRect(image.Rect(1085, 52, 1586, 428))
	drawFrame(screen, detailRect, colornames.Black, color.RGBA{R: 0, G: 38, B: 255, A: 255})
	drawText(screen, item.name, detailRect.Min.X+26, detailRect.Min.Y+46, colornames.White)
	drawText(screen, "Anzahl: "+strconv.Itoa(item.stock), detailRect.Min.X+26, detailRect.Min.Y+116, colornames.White)
	drawText(screen, "Preis: "+strconv.Itoa(s.shopPriceForItem(itemIndex)), detailRect.Min.X+26, detailRect.Min.Y+166, colornames.White)
	drawFrame(screen, s.shopDetailIconRect(), colornames.Black, colornames.Red)
	s.drawShopItemIcon(screen, itemIndex, insetRect(s.shopDetailIconRect(), 4), false)
	if s.shopMode == shopModeClassB {
		drawButton(screen, s.shopBuyRect(), "Zuschlagen!")
		return
	}
	drawButton(screen, s.shopBuyRect(), "Kaufen")
}

func (s *GameScene) buySelectedShopItem() {
	playerIndex := s.currentShopPlayerIndex()
	itemIndex := s.selectedShopItemIndex()
	if playerIndex < 0 || playerIndex >= len(s.credits) || itemIndex < 0 {
		return
	}
	price := s.shopPriceForItem(itemIndex)
	if price <= 0 || s.credits[playerIndex] < price || s.shopStockForSelected(itemIndex) <= 0 {
		return
	}
	if s.isScrollOMatItem(itemIndex) && s.shopItemCountForPlayer(playerIndex, itemIndex) > 0 {
		return
	}
	s.credits[playerIndex] -= price
	quantity := shopItems()[itemIndex].stock
	if quantity <= 0 {
		quantity = 1
	}
	if s.shopMode == shopModeClassB {
		s.shopClassBStock[s.shopSelectedIndex]--
		s.ensureInventory(playerIndex)
		s.inventories[playerIndex].classB[itemIndex] += quantity
		return
	}
	s.shopClassAStock[itemIndex]--
	s.ensureInventory(playerIndex)
	s.inventories[playerIndex].classA[itemIndex] += quantity
}

func (s *GameScene) ensureInventory(playerIndex int) {
	if playerIndex < 0 {
		return
	}
	items := shopItems()
	for len(s.inventories) <= playerIndex {
		s.inventories = append(s.inventories, shopInventory{})
	}
	if len(s.inventories[playerIndex].classA) < len(items) {
		next := make([]int, len(items))
		copy(next, s.inventories[playerIndex].classA)
		s.inventories[playerIndex].classA = next
	}
	if len(s.inventories[playerIndex].classB) < len(items) {
		next := make([]int, len(items))
		copy(next, s.inventories[playerIndex].classB)
		s.inventories[playerIndex].classB = next
	}
}

func (s *GameScene) selectedShopItemIndex() int {
	items := s.visibleShopItemIndexes()
	if s.shopSelectedIndex < 0 || s.shopSelectedIndex >= len(items) {
		return -1
	}
	return items[s.shopSelectedIndex]
}

func (s *GameScene) visibleShopItemIndexes() []int {
	if s.shopMode == shopModeClassB {
		return s.shopClassBItems
	}
	items := shopItems()
	indexes := make([]int, len(items))
	for i := range indexes {
		indexes[i] = i
	}
	return indexes
}

func (s *GameScene) shopStockForSelected(itemIndex int) int {
	if s.shopMode == shopModeClassB {
		if s.shopSelectedIndex >= 0 && s.shopSelectedIndex < len(s.shopClassBStock) {
			return s.shopClassBStock[s.shopSelectedIndex]
		}
		return 0
	}
	if itemIndex >= 0 && itemIndex < len(s.shopClassAStock) {
		return s.shopClassAStock[itemIndex]
	}
	return 0
}

func (s *GameScene) shopPriceForItem(itemIndex int) int {
	items := shopItems()
	if itemIndex < 0 || itemIndex >= len(items) {
		return 0
	}
	if s.shopMode == shopModeClassB {
		return maxInt(1, int(math.Round(float64(items[itemIndex].price)*classBDiscountMultiplier)))
	}
	return items[itemIndex].price
}

func (s *GameScene) resetShopStock() {
	items := shopItems()
	s.shopClassAStock = make([]int, len(items))
	for i := range items {
		s.shopClassAStock[i] = 1
	}
	s.shopClassBStock = nil
}

func (s *GameScene) pickClassBItems() {
	items := shopItems()
	indexes := s.rng.Perm(len(items))
	s.shopClassBItems = append([]int(nil), indexes[:minInt(3, len(indexes))]...)
	s.shopClassBStock = make([]int, len(s.shopClassBItems))
	for i := range s.shopClassBStock {
		s.shopClassBStock[i] = 1
	}
}

func (s *GameScene) shopListIndexAt(cursor image.Point, count int) (int, bool) {
	list := s.shopListRect()
	if !cursor.In(list) || count <= 0 {
		return 0, false
	}
	rowH := s.shopListRowHeight()
	offset := s.shopListScrollOffset(count)
	index := offset + (cursor.Y-list.Min.Y)/rowH
	if index < 0 || index >= count || index >= offset+s.shopVisibleRows(count) {
		return 0, false
	}
	return index, true
}

func (s *GameScene) shopListRect() image.Rectangle {
	if s.shopMode == shopModeClassB {
		return s.shopStoreScaledRect(image.Rect(220, 70, 386, 128))
	}
	return s.shopStoreScaledRect(image.Rect(218, 68, 388, 350))
}

func (s *GameScene) shopListRowHeight() int {
	return maxInt(22, int(math.Round(19*float64(s.shopStoreRect().Dy())/400)))
}

func (s *GameScene) shopVisibleRows(count int) int {
	if count <= 0 {
		return 0
	}
	rows := s.shopListRect().Dy() / s.shopListRowHeight()
	return minInt(count, maxInt(1, rows))
}

func (s *GameScene) shopListScrollOffset(count int) int {
	visibleRows := s.shopVisibleRows(count)
	if visibleRows <= 0 || count <= visibleRows {
		return 0
	}
	return minInt(count-visibleRows, maxInt(0, s.shopSelectedIndex-visibleRows+1))
}

func (s *GameScene) coverShopRoundLabel(screen *ebiten.Image) {
	fill := color.RGBA{R: 0, G: 18, B: 190, A: 255}
	if s.shopMode != shopModeEntry {
		fill = color.RGBA{R: 31, G: 49, B: 96, A: 255}
	}
	drawFilledRect(screen, s.shopScaledRect(image.Rect(0, 0, 190, 36)), fill)
}

func (s *GameScene) shopBuyRect() image.Rectangle {
	return s.shopScaledRect(image.Rect(1224, 350, 1444, 402))
}

func (s *GameScene) shopDetailIconRect() image.Rectangle {
	return s.shopScaledRect(image.Rect(1124, 324, 1204, 404))
}

func (s *GameScene) shopBackRect() image.Rectangle {
	return s.shopScaledRect(image.Rect(1358, 932, 1582, 987))
}

func (s *GameScene) drawShopPlayerPanel(screen *ebiten.Image, playerIndex int, player PlayerConfig) {
	portraitRect := s.shopScaledRect(image.Rect(15, 1032, 184, 1193))
	nameRect := s.shopScaledRect(image.Rect(200, 1035, 829, 1184))
	moneyRect := s.shopScaledRect(image.Rect(856, 1035, 1182, 1184))
	statusRect := s.shopScaledRect(image.Rect(1208, 1035, 1584, 1184))

	drawFrame(screen, portraitRect, colornames.White, color.RGBA{R: 255, G: 140, B: 0, A: 255})
	drawFrame(screen, nameRect, color.RGBA{R: 148, G: 148, B: 148, A: 255}, color.RGBA{R: 0, G: 38, B: 255, A: 255})
	drawFrame(screen, moneyRect, color.RGBA{R: 190, G: 190, B: 190, A: 255}, color.RGBA{R: 0, G: 38, B: 255, A: 255})
	drawFrame(screen, statusRect, color.RGBA{R: 224, G: 224, B: 224, A: 255}, color.RGBA{R: 0, G: 38, B: 255, A: 255})

	portrait := s.shop.human
	if player.Kind == PlayerComputer {
		portrait = s.shop.computer
	}
	innerPortrait := insetRect(portraitRect, 8)
	drawScaledImage(screen, portrait, innerPortrait)
	drawPaintSwatch(screen, image.Rect(portraitRect.Max.X-30, portraitRect.Min.Y+10, portraitRect.Max.X-10, portraitRect.Min.Y+32), player.Color)

	drawCenteredText(screen, player.Name, nameRect, colornames.White)
	money := 0
	if playerIndex >= 0 && playerIndex < len(s.credits) {
		money = s.credits[playerIndex]
	}
	drawCenteredText(screen, "Geld: $"+strconv.Itoa(money), moneyRect, colornames.Black)
}

func (s *GameScene) currentShopPlayerIndex() int {
	if s.shopPlayerCursor < 0 || s.shopPlayerCursor >= len(s.shopPlayerOrder) {
		return -1
	}
	return s.shopPlayerOrder[s.shopPlayerCursor]
}

func (s *GameScene) roundScoreForPlayer(playerIndex int) int {
	if playerIndex < 0 || playerIndex >= len(s.roundScores) {
		return 0
	}
	return s.roundScores[playerIndex]
}

func (s *GameScene) shopClassARect() image.Rectangle {
	return s.shopStoreScaledRect(image.Rect(146, 86, 210, 184))
}

func (s *GameScene) shopClassBRect() image.Rectangle {
	return s.shopStoreScaledRect(image.Rect(432, 78, 496, 180))
}

func (s *GameScene) shopContinueRect() image.Rectangle {
	return s.shopScaledRect(image.Rect(1304, 932, 1580, 987))
}

func (s *GameScene) shopStoreRect() image.Rectangle {
	screenCfg := core.Config().Screen
	panelTop := s.shopScaledRect(image.Rect(0, 1032, 1, 1032)).Min.Y
	if panelTop <= 0 || panelTop > int(screenCfg.Height) {
		panelTop = int(screenCfg.Height * 0.84)
	}
	return image.Rect(0, 0, int(screenCfg.Width), panelTop)
}

func (s *GameScene) shopStoreScaledRect(r image.Rectangle) image.Rectangle {
	store := s.shopStoreRect()
	scaleX := float64(store.Dx()) / 640
	scaleY := float64(store.Dy()) / 400
	return image.Rect(
		store.Min.X+int(math.Round(float64(r.Min.X)*scaleX)),
		store.Min.Y+int(math.Round(float64(r.Min.Y)*scaleY)),
		store.Min.X+int(math.Round(float64(r.Max.X)*scaleX)),
		store.Min.Y+int(math.Round(float64(r.Max.Y)*scaleY)),
	)
}

func (s *GameScene) shopScaledRect(r image.Rectangle) image.Rectangle {
	screenCfg := core.Config().Screen
	scaleX := screenCfg.Width / 1612
	scaleY := screenCfg.Height / 1209
	return image.Rect(
		int(math.Round(float64(r.Min.X)*scaleX)),
		int(math.Round(float64(r.Min.Y)*scaleY)),
		int(math.Round(float64(r.Max.X)*scaleX)),
		int(math.Round(float64(r.Max.Y)*scaleY)),
	)
}

func (s *GameScene) drawStoreImage(screen, img *ebiten.Image) {
	if img == nil {
		return
	}
	drawScaledImage(screen, img, s.shopStoreRect())
}

func (s *GameScene) drawStoreOverlay(screen, img *ebiten.Image, r image.Rectangle) {
	if img == nil {
		return
	}
	srcRect := img.Bounds()
	if srcRect.Dx() > 1 {
		srcRect.Max.X--
	}
	if trimmed, ok := img.SubImage(srcRect).(*ebiten.Image); ok {
		img = trimmed
	}
	drawScaledImage(screen, img, s.shopStoreScaledRect(r))
}

func (s *GameScene) drawShopItemIcon(screen *ebiten.Image, itemIndex int, r image.Rectangle, disabled bool) {
	if itemIndex < 0 {
		return
	}
	src := s.shop.storeIcons
	if src == nil {
		return
	}
	iconRect := image.Rect(1124, 324, 1204, 404)
	if src == s.shop.storeIcons {
		cell := 32
		col := itemIndex % 16
		row := itemIndex / 16
		iconRect = image.Rect(col*cell, row*cell, col*cell+cell, row*cell+cell)
		if iconRect.Max.Y > src.Bounds().Dy() {
			return
		}
	}
	icon, ok := src.SubImage(iconRect).(*ebiten.Image)
	if !ok {
		return
	}
	drawScaledImageWithDisabled(screen, icon, r, disabled)
}

func drawScaledImageWithDisabled(screen, img *ebiten.Image, r image.Rectangle, disabled bool) {
	if img == nil || r.Empty() {
		return
	}
	bounds := img.Bounds()
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(float64(r.Dx())/float64(bounds.Dx()), float64(r.Dy())/float64(bounds.Dy()))
	op.GeoM.Translate(float64(r.Min.X), float64(r.Min.Y))
	if disabled {
		op.ColorScale.Scale(0.28, 0.28, 0.28, 0.72)
	}
	screen.DrawImage(img, op)
}
