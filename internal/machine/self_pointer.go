package machine

// SelfPointer gives virtual dispatch to types built by embedding: Go calls a
// method of an embedded type with the embedded value as receiver, so
// overridable methods are called through Self(), the outermost value.
type SelfPointer[T any] interface {
	Self() T
}

type SelfPointing[T any] struct {
	_self T
}

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
