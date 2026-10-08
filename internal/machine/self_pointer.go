package machine

// SelfPointer gives virtual dispatch to types built by embedding: Go calls a
// method of an embedded type with the embedded value as receiver, so
// overridable methods are called through Self(), the outermost value.
//
// Only the elements use it. Their defaults in GenericElement (and in Symbol,
// Primitive, Var, which the element types embed in turn) call other methods
// that a type may override, as the original's class hierarchy did, and about
// a hundred types rely on that. Modes, contexts and input handlers do not need
// it and are plain structs.
type SelfPointer[T any] interface {
	Self() T
}

// SelfPointing holds the self pointer; embed it and build the value with
// MakeSelf or ReSelf.
type SelfPointing[T any] struct {
	_self T
}

// Self is the outermost value that embeds sp.
func (sp *SelfPointing[T]) Self() T {
	return sp._self
}

func (sp *SelfPointing[T]) setSelf(x any) {
	sp._self = x.(T)
}

// selfSetter is implemented by every type that embeds a SelfPointing.
type selfSetter interface {
	setSelf(any)
}

// MakeSelf allocates a T whose self pointer points to itself.
func MakeSelf[T any]() *T {
	return ReSelf(new(T))
}

// ReSelf points t's self pointer at t; constructors call it after copying an
// embedded value, whose self pointer still refers to the copy's source.
func ReSelf[T any](t *T) *T {
	any(t).(selfSetter).setSelf(t)
	return t
}
