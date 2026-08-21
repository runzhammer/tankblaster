package tankblaster

type shopItem struct {
	name        string
	price       int
	stock       int
	screenIndex int
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
		{name: scrollOMatItemName, price: 1000, stock: 1, screenIndex: 19},
		{name: "Energieschild", price: 15000, stock: 1, screenIndex: 20},
		{name: "MFS Verstärker", price: 12000, stock: 1, screenIndex: 21},
		{name: "XM-V12 Panzer", price: 9890, stock: 1, screenIndex: 22},
		{name: "Diesel (F54)", price: 400, stock: 100, screenIndex: 23},
	}
}
