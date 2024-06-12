package machine

import "reflect"

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
	val := reflect.New(reflect.TypeOf((*T)(nil)).Elem()).Elem()

	field := val.FieldByName("_self")
	if field.IsValid() && field.CanSet() && field.Kind() == reflect.Ptr {
		field.Set(val.Addr())
	}

	return val.Addr().Interface().(*T)
}
