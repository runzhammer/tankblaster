package engine

type Movable interface {
	SetPosition(vec Vec)
	GetSprites() []*Sprite
}
