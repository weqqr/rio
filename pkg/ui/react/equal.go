package react

import "reflect"

type sameFunc[T any] func(a, b T) bool

func same[T any]() sameFunc[T] {
	t := reflect.TypeOf((*T)(nil)).Elem()
	switch t.Kind() {
	case reflect.Slice, reflect.Map:

		return func(a, b T) bool {
			return reflect.ValueOf(any(a)).UnsafePointer() == reflect.ValueOf(any(b)).UnsafePointer()
		}
	case reflect.Func:

		return func(a, b T) bool { return false }
	case reflect.Interface, reflect.Float32, reflect.Float64:
		return sameAny[T]
	case reflect.Struct, reflect.Array:
		if containsInterface(t) || !t.Comparable() {
			return sameAny[T]
		}
	}
	return func(a, b T) bool { return any(a) == any(b) }
}

func sameAny[T any](a, b T) bool { return sameDynamic(any(a), any(b)) }

func sameDynamic(a, b any) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	av, bv := reflect.ValueOf(a), reflect.ValueOf(b)
	if av.Type() != bv.Type() {
		return false
	}
	if !av.Comparable() {

		switch av.Kind() {
		case reflect.Slice, reflect.Map:
			return av.UnsafePointer() == bv.UnsafePointer()
		}
		return false
	}
	switch av.Kind() {
	case reflect.Float32, reflect.Float64:
		x, y := av.Float(), bv.Float()
		return x == y || x != x && y != y
	}
	return av.Equal(bv)
}

func containsInterface(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Interface:
		return true
	case reflect.Array:
		return containsInterface(t.Elem())
	case reflect.Struct:
		for i := range t.NumField() {
			if containsInterface(t.Field(i).Type) {
				return true
			}
		}
	}
	return false
}
