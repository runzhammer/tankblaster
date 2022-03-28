package engine

// Behavior is what happens when an object meets a condition for a given time delta
type SpriteBehavior func(source *Sprite)

// Reaction is what happens when two objects meet a condition for a given time delta.
// Source is the Sprite performing a behavior "with" is the object the source is
// reacting with.
type SpriteReaction func(source, with *Sprite)

// Behaviors is a slice of Behavior that should happen in succession
type SpriteBehaviors []SpriteBehavior

// MakeBehaviors is a convenience function for turning a sequence
// of Behaviors or Behavior functions into Behaviors
func MakeSpriteBehaviors(behaviors ...Behavior) Behaviors {
	return Behaviors(behaviors)
}

// Execute executes all behaviors for an object with a time delta
func (b SpriteBehaviors) Execute(source *Sprite) {
	for _, behavior := range b {
		behavior(source)
	}
}

// Movement is a Behavior that will move a source an object
// by its velocity scaled by time delta
var SpriteMovement = SpriteBehavior(func(source *Sprite) {
	v := source.Velocity.Scaled(1)
	source.Pos = source.Pos.Add(v)
})

// FaceDirection is a behavior that adjusts an Sprite's
// Rot (rotation) to face the same angle as its Velocity.
func SpriteFaceDirection(source *Sprite) {
	source.Rot = source.Velocity.Angle()
}
