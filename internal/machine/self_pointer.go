package machine

import (
	"reflect"
	"unsafe"
)

type SelfPointer[T any] interface {
	Self() T
}

type SelfPointing[T any] struct {
	_self T
}

func (sp *SelfPointing[T]) Self() T {
	return sp._self
}

func MakeSelf[T any]() *T {
	t := new(T)
	v := reflect.ValueOf(t).Elem()
	f := v.FieldByName("_self")
	reflect.NewAt(f.Type(), unsafe.Pointer(f.UnsafeAddr())).Elem().Set(reflect.ValueOf(t))

	return t
}

func ReSelf[T any](t *T) *T {
	v := reflect.ValueOf(t).Elem()
	f := v.FieldByName("_self")
	reflect.NewAt(f.Type(), unsafe.Pointer(f.UnsafeAddr())).Elem().Set(reflect.ValueOf(t))

	return t
}
