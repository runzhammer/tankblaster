package models

import "github.com/cevaris/ordered_map"

// TankIterator is a function used for iterating over collections of Tanks.
// It is used as follows:
//
// iter := set.Iterator()
// for obj, ok := iter(); ok; obj, ok = iter() {
//   ..use obj..
// }
// Removing an Tank during iteration is undefined.
type TankIterator func() (next *Tank, ok bool)

// TankSet is an ordered set of Tanks.
type TankSet struct {
	set *ordered_map.OrderedMap
}

// NewTankSet creates a new empty set
func NewTankSet() *TankSet {
	return &TankSet{
		set: ordered_map.NewOrderedMap(),
	}
}

// Add adds Tanks to this set
func (os *TankSet) Add(obj *Tank) {
	os.set.Set(obj, struct{}{})
}

// Remove removes Tanks from this set
func (os *TankSet) Remove(obj *Tank) {
	os.set.Delete(obj)
}

// emptyTankIterator is an TankIterator that always
// returns the ultimate result.
func emptyTankIterator() (*Tank, bool) {
	return nil, false
}

// Iterator returns an iterator function that can be used
// to iterate over all Tanks in this set.
func (os *TankSet) Iterator() TankIterator {
	if os == nil {
		return emptyTankIterator
	}
	iter := os.set.IterFunc()
	return func() (*Tank, bool) {
		next, ok := iter()
		if ok {
			return next.Key.(*Tank), true
		}
		return nil, false
	}
}

// Update performs all PreSteps, then all Steps, then all PostSteps
// of Tanks in this container.
func (ts *TankSet) Update() {
	iter := ts.Iterator()
	for Tank, ok := iter(); ok; Tank, ok = iter() {
		Tank.PreSteps.Execute(Tank)
	}
	iter = ts.Iterator()
	for Tank, ok := iter(); ok; Tank, ok = iter() {
		Tank.Steps.Execute(Tank)
	}
	iter = ts.Iterator()
	for Tank, ok := iter(); ok; Tank, ok = iter() {
		Tank.PostSteps.Execute(Tank)
	}
}
