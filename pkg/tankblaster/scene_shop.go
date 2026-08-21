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
	r "github.com/runzhammer/gamedemo/resources"
	"golang.org/x/image/colornames"
)

type shopMode uint8

const (
	shopModeEntry shopMode = iota
	shopModeClassA
	shopModeClassB
)

const classBDiscountMultiplier = 2105.0 / 3175.0

type shopItem struct {
	name        string
	price       int
	stock       int
	screenIndex int
}

type shopInventory struct {
	classA []int
	classB []int
}

type shopAssets struct {
	entry         *ebiten.Image
	classA        *ebiten.Image
	classB        *ebiten.Image
	human         *ebiten.Image
	computer      *ebiten.Image
	trainingOn    *ebiten.Image
	trainingOff   *ebiten.Image
	classAScreens []*ebiten.Image
	classBScreen  *ebiten.Image
}

func shopItems() []shopItem {
	return []shopItem{
		{name: "Granate", price: 1250, stock: 50, screenIndex: 1},
		{name: "große Granate", price: 2000, stock: 10, screenIndex: 2},
		{name: "Atombombe", price: 3175, stock: 2, screenIndex: 3},
		{name: "H-Bombe", price: 3500, stock: 1, screenIndex: 4},
		{name: "Plasmaschmelzer", price: 10500, stock: 1, screenIndex: 5},
		{name: "Wunderpalme", price: 2300, stock: 1, screenIndex: 6},
		{name: "Feuerkugel", price: 1500, stock: 2, screenIndex: 7},
		{name: "Wasser", price: 4000, stock: 2, screenIndex: 8},
		{name: "Maulwürfe", price: 1450, stock: 3, screenIndex: 9},
		{name: "MFS 3-fach", price: 3000, stock: 3, screenIndex: 10},
		{name: "Brösler, klein", price: 300, stock: 8, screenIndex: 11},
		{name: "Brösler, groß", price: 900, stock: 2, screenIndex: 12},
		{name: "Überraschungsei", price: 1000, stock: 1, screenIndex: 13},
		{name: "Moskitos", price: 3750, stock: 1, screenIndex: 14},
		{name: "Schockwelle", price: 7600, stock: 2, screenIndex: 15},
		{name: "Luftschlag", price: 13300, stock: 1, screenIndex: 16},
		{name: "Splitterbombe", price: 1300, stock: 2, screenIndex: 17},
		{name: "Laser", price: 500, stock: 1, screenIndex: 18},
		{name: "Scroll-o-Mat", price: 1000, stock: 1, screenIndex: 19},
		{name: "Energieschild", price: 15000, stock: 1, screenIndex: 20},
		{name: "MFS Verstärker", price: 12000, stock: 1, screenIndex: 21},
		{name: "XM-V12 Panzer", price: 9890, stock: 1, screenIndex: 22},
		{name: "Diesel (F54)", price: 400, stock: 100, screenIndex: 23},
	}
}

func loadShopClassAScreens() []*ebiten.Image {
	items := shopItems()
	screens := make([]*ebiten.Image, len(items))
	for index, item := range items {
		screens[index] = mustImageFromPNG(mustReadShopScreen("class-a-" + strconv.Itoa(item.screenIndex) + ".png"))
	}
	return screens
}

func mustReadShopScreen(name string) []byte {
	data, err := r.ShopListScreens.ReadFile(name)
	if err != nil {
		panic(err)
	}
	return data
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
	}
	return inventories
}

func (s *GameScene) beginShop() {
	s.phase = phaseShop
	s.shopPlayerCursor = 0
	s.shopHoverClass = 0
	s.shopMode = shopModeEntry
	s.shopSelectedIndex = 0
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
		s.shopPlayerCursor++
		s.shopHoverClass = 0
		return
	}
	s.startRound()
}

func (s *GameScene) drawShop(screen *ebiten.Image) {
	if s.shopMode != shopModeEntry {
		s.drawShopList(screen)
		return
	}

	bg := s.shop.entry
	switch s.shopHoverClass {
	case 1:
		bg = s.shop.classA
	case 2:
		bg = s.shop.classB
	}
	drawScaledImage(screen, bg, image.Rect(0, 0, int(core.Config().Screen.Width), int(core.Config().Screen.Height)))
	s.coverShopRoundLabel(screen)

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
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) {
		s.shopSelectedIndex = maxInt(0, s.shopSelectedIndex-1)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) {
		s.shopSelectedIndex = minInt(len(items)-1, s.shopSelectedIndex+1)
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

func (s *GameScene) drawShopList(screen *ebiten.Image) {
	bg := s.shop.classBScreen
	if s.shopMode == shopModeClassA && s.shopSelectedIndex >= 0 && s.shopSelectedIndex < len(s.shop.classAScreens) {
		bg = s.shop.classAScreens[s.shopSelectedIndex]
	}
	drawScaledImage(screen, bg, image.Rect(0, 0, int(core.Config().Screen.Width), int(core.Config().Screen.Height)))
	s.coverShopRoundLabel(screen)

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
	if s.shopMode == shopModeClassA {
		return
	}

	titleRect := s.shopScaledRect(image.Rect(490, 84, 970, 136))
	drawCenteredText(screen, "- Schnäppchen -", titleRect, colornames.Black)

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
			drawFilledRect(screen, r, color.RGBA{R: 0, G: 0, B: 55, A: 255})
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
	drawText(screen, "Anzahl: "+strconv.Itoa(s.shopStockForSelected(itemIndex)), detailRect.Min.X+26, detailRect.Min.Y+116, colornames.White)
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
	s.credits[playerIndex] -= price
	if s.shopMode == shopModeClassB {
		s.shopClassBStock[s.shopSelectedIndex]--
		s.ensureInventory(playerIndex)
		s.inventories[playerIndex].classB[itemIndex]++
		return
	}
	s.shopClassAStock[itemIndex]--
	s.ensureInventory(playerIndex)
	s.inventories[playerIndex].classA[itemIndex]++
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
	for i, item := range items {
		s.shopClassAStock[i] = item.stock
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
		return s.shopScaledRect(image.Rect(490, 164, 970, 326))
	}
	return s.shopScaledRect(image.Rect(490, 150, 1055, 936))
}

func (s *GameScene) shopListRowHeight() int {
	screenCfg := core.Config().Screen
	return maxInt(22, int(math.Round(48*screenCfg.Height/1209)))
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
	return s.shopScaledRect(image.Rect(340, 285, 520, 485))
}

func (s *GameScene) shopClassBRect() image.Rectangle {
	return s.shopScaledRect(image.Rect(1085, 260, 1245, 460))
}

func (s *GameScene) shopContinueRect() image.Rectangle {
	return s.shopScaledRect(image.Rect(1304, 932, 1580, 987))
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

func (s *GameScene) drawShopItemIcon(screen *ebiten.Image, itemIndex int, r image.Rectangle, disabled bool) {
	if itemIndex < 0 || itemIndex >= len(s.shop.classAScreens) {
		return
	}
	src := s.shop.classAScreens[itemIndex]
	if src == nil {
		return
	}
	icon, ok := src.SubImage(image.Rect(1124, 324, 1204, 404)).(*ebiten.Image)
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
