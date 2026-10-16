package react

type Ref[T any] struct {
	dep       dep
	value     T
	unchanged sameFunc[T]
}

func NewRef[T any](value T) *Ref[T] {
	return &Ref[T]{value: value, unchanged: same[T]()}
}

func NewRefEqual[T any](value T, equal func(newValue, oldValue T) bool) *Ref[T] {
	r := NewRef(value)
	if equal != nil {
		r.unchanged = equal
	}
	return r
}

func (r *Ref[T]) Get() T {
	r.dep.track()
	return r.value
}

func (r *Ref[T]) Peek() T { return r.value }

func (r *Ref[T]) Set(value T) {
	if r.unchanged(r.value, value) {
		return
	}
	r.value = value
	r.dep.trigger()
}
