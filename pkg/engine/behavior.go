package engine

// Behavior is what happens when an object meets a condition for a given time delta
type Behavior func(source Movable)

// Reaction is what happens when two objects meet a condition for a given time delta.
// Source is the Sprite performing a behavior "with" is the object the source is
// reacting with.
type Reaction func(source, with Movable)

// Behaviors is a slice of Behavior that should happen in succession
type Behaviors []Behavior

// MakeBehaviors is a convenience function for turning a sequence
// of Behaviors or Behavior functions into Behaviors
func MakeBehaviors(behaviors ...Behavior) Behaviors {
	return Behaviors(behaviors)
}

// Execute executes all behaviors for an object with a time delta
func (b Behaviors) Execute(source Movable) {
	for _, behavior := range b {
		behavior(source)
	}
}

// Movement is a Behavior that will move a source an object
// by its velocity scaled by time delta
var Movement = Behavior(func(source Movable) {
	for _, s := range source.GetSprites() {
		v := s.Velocity.Scaled(1)
		s.Pos = s.Pos.Add(v)
	}
})

// FaceDirection is a behavior that adjusts an Sprite's
// Rot (rotation) to face the same angle as its Velocity.
func FaceDirection(source Movable) {
	for _, s := range source.GetSprites() {
		s.Rot = s.Velocity.Angle()
	}
}
