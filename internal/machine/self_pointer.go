package machine

// selfPointer gives virtual dispatch to types built by embedding: Go calls a
// method of an embedded type with the embedded value as receiver, so
// overridable methods are called through Self(), the outermost value.
//
// Only the elements use it. Their defaults in genericElement (and in Symbol,
// Primitive, Var, which the element types embed in turn) call other methods
// that a type may override, as the original's class hierarchy did, and about
// a hundred types rely on that. Modes, contexts and input handlers do not need
// it and are plain structs.
type selfPointer[T any] interface {
	self() T
}

// selfPointing holds the self pointer; embed it and build the value with
// makeSelf or reSelf.
type selfPointing[T any] struct {
	_self T
}

// Self is the outermost value that embeds sp.
func (sp *selfPointing[T]) self() T {
	return sp._self
}

func (sp *selfPointing[T]) setSelf(x any) {
	sp._self = x.(T)
}

// selfSetter is implemented by every type that embeds a selfPointing.
type selfSetter interface {
	setSelf(any)
}

// makeSelf allocates a T whose self pointer points to itself.
func makeSelf[T any]() *T {
	return reSelf(new(T))
}

// reSelf points t's self pointer at t; constructors call it after copying an
// embedded value, whose self pointer still refers to the copy's source.
func reSelf[T any](t *T) *T {
	any(t).(selfSetter).setSelf(t)
	return t
}
