package engine

import (
	"github.com/cevaris/ordered_map"
	"github.com/hajimehoshi/ebiten/v2"
)

// SpriteContainer is a simple interface for Sprite containers
type SpriteContainer interface {
	Len() int
	Contains(obj *Sprite) bool
	Iterator() SpriteIterator
}

// SpriteContainer is a simple interface for Sprite containers that support tags
type TaggedSpriteContainer interface {
	SpriteContainer
	TagIterator(tags ...string) SpriteIterator
}

// SpriteIterator is a function used for iterating over collections of Sprites.
// It is used as follows:
//
// iter := set.Iterator()
// for obj, ok := iter(); ok; obj, ok = iter() {
//   ..use obj..
// }
// Removing an sprite during iteration is undefined.
type SpriteIterator func() (next *Sprite, ok bool)

// Layers is a container for multiple Sprites collections such that
// a particular drawing order can be preserved. Updates and Draws will
// happen from the lowest layer to the highest layer.
type Layers []*Sprites

// NewLayers creates a new container of Sprites with a given amount of layers.
func NewLayers(n int) Layers {
	layers := make(Layers, 0, n)
	for i := 0; i < n; i++ {
		layers = append(layers, NewSprites())
	}
	return layers
}

// Update updates all Sprites. Updates happen in the first layer forward.
func (ly Layers) Update() {
	for _, layer := range ly {
		layer.Update()
	}
}

// Draw draws all Sprites Draws happen in the first layer forward.
func (ly Layers) Draw(camera *ebiten.GeoM, image *ebiten.Image) {
	for _, layer := range ly {
		layer.Draw(camera, image)
	}
}

// Len returns the total number of Sprites in all layers
func (ly Layers) Len() int {
	sum := 0
	for _, layer := range ly {
		sum += layer.Len()
	}
	return sum
}

// Contains checks to see if an Sprite is in any layer
func (ly Layers) Contains(obj *Sprite) bool {
	for _, layer := range ly {
		if layer.Contains(obj) {
			return true
		}
	}
	return false
}

// Iterator returns an SpriteIterator for all sprites in all layers
// from the lowest layer to highest
func (ly Layers) Iterator() SpriteIterator {
	iters := make([]SpriteIterator, len(ly))
	for index, layer := range ly {
		iters[index] = layer.All().Iterator()
	}
	return chainIterators(iters)
}

// Iterator returns an SpriteIterator for all sprites in all layers
// from the highest layer to lowest
func (ly Layers) IteratorTop() SpriteIterator {
	iters := make([]SpriteIterator, len(ly))
	for index := len(ly) - 1; index >= 0; index-- {
		iters[index] = ly[index].All().Iterator()
	}
	return chainIterators(iters)
}

// Iterator returns an SpriteIterator for all sprites
// with the given tags in all layers from the lowest layer to highest
func (ly Layers) TagIterator(tags ...string) SpriteIterator {
	if len(tags) == 0 {
		return emptySpriteIterator
	}
	iters := make([]SpriteIterator, len(ly)*len(tags))
	index := 0
	for _, layer := range ly {
		for _, tag := range tags {
			iters[index] = layer.Tagged(tag).Iterator()
			index++
		}
	}
	return chainIterators(iters)
}

// Iterator returns an SpriteIterator for all sprites
// with the given tags in all layers from the highest layer to lowest
func (ly Layers) TagIteratorTop(tags ...string) SpriteIterator {
	if len(tags) == 0 {
		return emptySpriteIterator
	}
	iters := make([]SpriteIterator, len(ly)*len(tags))
	index := 0
	for layerIndex := len(ly) - 1; layerIndex >= 0; layerIndex-- {
		for _, tag := range tags {
			iters[index] = ly[layerIndex].Tagged(tag).Iterator()
			index++
		}
	}
	return chainIterators(iters)
}

// Sprites is a container of Sprite so that Sprites can be quickly added
// and removed from a single source. Sprites are also retrievable by tag
// allowing for quick access for a particular subset of Sprite.
//
// The Tag of an Sprite should not be modified after being added to this
// container.
type Sprites struct {
	all    *SpriteSet
	tagged spriteTagMap
}

// NewSprites makes a new Sprites container.
func NewSprites() *Sprites {
	return &Sprites{
		all:    NewSpriteSet(),
		tagged: make(spriteTagMap),
	}
}

// Len returns the amount of Sprites in this container
func (o *Sprites) Len() int {
	return o.all.Len()
}

// All returns the SpriteSet containing all Sprites in this container
func (o *Sprites) All() *SpriteSet {
	return o.all
}

// Tagged returns an SpriteSet containing all Sprites in this container
// that have a particular tag. Tags with empty strings are not recorded
// and Sprites whose tags were modified after being added are not
// considered.
func (o *Sprites) Tagged(tag string) *SpriteSet {
	return o.tagged[tag]
}

// Add adds an sprite to this container. If the Sprite has a Tag, that
// tag is used to quickly access a particular subset of Sprite.
func (o *Sprites) Add(obj *Sprite) {
	o.all.Add(obj)
	if obj.Tag != "" {
		o.tagged.add(obj.Tag, obj)
	}
}

// Add all *Sprites to left side *Sprites
func AddSprites(objs1 *Sprites, objs2 *Sprites) *Sprites {
	iter := objs2.All().Iterator()
	for obj, ok := iter(); ok; obj, ok = iter() {
		objs1.Add(obj)
	}
	return objs1
}

// Remove removes an sprite from this container.
func (o *Sprites) Remove(obj *Sprite) {
	o.all.Remove(obj)
	if obj.Tag != "" {
		o.tagged.remove(obj.Tag, obj)
	}
}

// Contains tests to see if an sprite is contained in this container.
func (o *Sprites) Contains(obj *Sprite) bool {
	return o.all.Contains(obj)
}

// Update performs all PreSteps, then all Steps, then all PostSteps
// of Sprite in this container.
func (o *Sprites) Update() {
	o.all.Update()
}

// Draw draws all Sprite in this container.
func (o *Sprites) Draw(camera *ebiten.GeoM, image *ebiten.Image) {
	o.all.Draw(camera, image)
}

// Iterator gets an SpriteIterator for all Sprite in this container
func (o *Sprites) Iterator() SpriteIterator {
	return o.All().Iterator()
}

// TagIterator gets an SpriteIterator for all Sprite in this
// container with the given tags
func (o *Sprites) TagIterator(tags ...string) SpriteIterator {
	if len(tags) == 0 {
		return emptySpriteIterator
	}
	iters := make([]SpriteIterator, len(tags))
	index := 0
	for _, tag := range tags {
		iters[index] = o.Tagged(tag).Iterator()
		index++
	}
	return chainIterators(iters)
}

// SpriteSet is an ordered set of Sprite.
type SpriteSet struct {
	set *ordered_map.OrderedMap
}

// NewSpriteSet creates a new empty set
func NewSpriteSet() *SpriteSet {
	return &SpriteSet{
		set: ordered_map.NewOrderedMap(),
	}
}

// emptySpriteIterator is an SpriteIterator that always
// returns the ultimate result.
func emptySpriteIterator() (*Sprite, bool) {
	return nil, false
}

// Iterator returns an iterator function that can be used
// to iterate over all sprites in this set.
func (os *SpriteSet) Iterator() SpriteIterator {
	if os == nil {
		return emptySpriteIterator
	}
	iter := os.set.IterFunc()
	return func() (*Sprite, bool) {
		next, ok := iter()
		if ok {
			return next.Key.(*Sprite), true
		}
		return nil, false
	}
}

// Len returns the size of this set
func (os *SpriteSet) Len() int {
	if os == nil {
		return 0
	}
	return os.set.Len()
}

// Contains tests if an Sprite is contained in this set
func (os *SpriteSet) Contains(obj *Sprite) bool {
	if os == nil {
		return false
	}
	_, ok := os.set.Get(obj)
	return ok
}

// Add adds sprites to this set
func (os *SpriteSet) Add(obj *Sprite) {
	os.set.Set(obj, struct{}{})
}

// Remove removes sprites from this set
func (os *SpriteSet) Remove(obj *Sprite) {
	os.set.Delete(obj)
}

// Update performs all PreSteps, then all Steps, then all PostSteps
// of Sprite in this container.
func (os *SpriteSet) Update() {
	iter := os.Iterator()
	for sprite, ok := iter(); ok; sprite, ok = iter() {
		sprite.PreSteps.Execute(sprite)
	}
	iter = os.Iterator()
	for sprite, ok := iter(); ok; sprite, ok = iter() {
		sprite.Steps.Execute(sprite)
	}
	iter = os.Iterator()
	for sprite, ok := iter(); ok; sprite, ok = iter() {
		sprite.PostSteps.Execute(sprite)
	}
}

// Draw draws all Sprite in this container.
func (os *SpriteSet) Draw(camera *ebiten.GeoM, image *ebiten.Image) {
	iter := os.Iterator()
	for sprite, ok := iter(); ok; sprite, ok = iter() {
		sprite.Draw(camera, image)
	}
}

// spriteTagMap is a defaultdict-like map for adding and removing
// sprites from an SpriteSet by tag
type spriteTagMap map[string]*SpriteSet

func (m spriteTagMap) add(tag string, obj *Sprite) {
	set := m[tag]
	if set == nil {
		set = NewSpriteSet()
		m[tag] = set
	}
	set.Add(obj)
}

func (m spriteTagMap) remove(tag string, obj *Sprite) {
	set := m[tag]
	if set != nil {
		set.Remove(obj)
	}
}

// chainIterators will iterate through a slice of SpriteIterator consecutively
func chainIterators(iters []SpriteIterator) SpriteIterator {
	index := 0
	return func() (*Sprite, bool) {
		for index < len(iters) {
			next, ok := iters[index]()
			if !ok {
				index++
				continue
			}
			return next, ok
		}
		return nil, false
	}
}
