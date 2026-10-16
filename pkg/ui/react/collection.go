package react

import "slices"

type ReactiveMap[K comparable, V any] struct {
	entries   map[K]V
	keys      map[K]*dep
	iter      dep
	unchanged sameFunc[V]
}

func NewReactiveMap[K comparable, V any]() *ReactiveMap[K, V] {
	return &ReactiveMap[K, V]{
		entries:   make(map[K]V),
		keys:      make(map[K]*dep),
		unchanged: same[V](),
	}
}

func (m *ReactiveMap[K, V]) Get(key K) V {
	if isTracking() {
		m.dep(key).track()
	}
	return m.entries[key]
}

func (m *ReactiveMap[K, V]) Has(key K) bool {
	if isTracking() {
		m.dep(key).track()
	}
	_, ok := m.entries[key]
	return ok
}

func (m *ReactiveMap[K, V]) Len() int {
	m.iter.track()
	return len(m.entries)
}

func (m *ReactiveMap[K, V]) Set(key K, value V) {
	previous, existed := m.entries[key]
	if existed && m.unchanged(previous, value) {
		return
	}
	m.entries[key] = value
	pass := newPass()
	d := m.keys[key]
	if !existed {
		m.iter.triggerPass(pass)
	}
	if d != nil {
		d.triggerPass(pass)
	}
}

func (m *ReactiveMap[K, V]) Delete(key K) bool {
	if _, existed := m.entries[key]; !existed {
		return false
	}
	delete(m.entries, key)
	d := m.keys[key]
	delete(m.keys, key)
	pass := newPass()
	m.iter.triggerPass(pass)
	if d != nil {
		d.triggerPass(pass)
	}
	return true
}

func (m *ReactiveMap[K, V]) Clear() {
	if len(m.entries) == 0 {
		return
	}
	clear(m.entries)
	deps := m.keys
	m.keys = make(map[K]*dep)
	pass := newPass()
	m.iter.triggerPass(pass)
	for _, d := range deps {
		d.triggerPass(pass)
	}
}

func (m *ReactiveMap[K, V]) Keys() []K {
	m.iter.track()
	keys := make([]K, 0, len(m.entries))
	for key := range m.entries {
		keys = append(keys, key)
	}
	return keys
}

func (m *ReactiveMap[K, V]) Values() []V {
	m.iter.track()
	values := make([]V, 0, len(m.entries))
	for key, value := range m.entries {
		if isTracking() {
			m.dep(key).track()
		}
		values = append(values, value)
	}
	return values
}

func (m *ReactiveMap[K, V]) Range(fn func(key K, value V) bool) {
	m.iter.track()
	for key, value := range m.entries {
		if isTracking() {
			m.dep(key).track()
		}
		if !fn(key, value) {
			return
		}
	}
}

func (m *ReactiveMap[K, V]) dep(key K) *dep {
	d, ok := m.keys[key]
	if !ok {
		d = new(dep)
		m.keys[key] = d
	}
	return d
}

type ReactiveSlice[T any] struct {
	items []T

	elements  []*dep
	iter      dep
	unchanged sameFunc[T]
}

func NewReactiveSlice[T any](items ...T) *ReactiveSlice[T] {
	s := &ReactiveSlice[T]{unchanged: same[T]()}
	s.items = append(s.items, items...)
	s.elements = make([]*dep, len(s.items))
	return s
}

func (s *ReactiveSlice[T]) Len() int {
	s.iter.track()
	return len(s.items)
}

func (s *ReactiveSlice[T]) Get(index int) T {
	value := s.items[index]
	if isTracking() {
		s.dep(index).track()
	}
	return value
}

func (s *ReactiveSlice[T]) Set(index int, value T) {
	if s.unchanged(s.items[index], value) {
		return
	}
	s.items[index] = value
	if d := s.elements[index]; d != nil {
		d.trigger()
	}
}

func (s *ReactiveSlice[T]) Append(values ...T) {
	if len(values) == 0 {
		return
	}
	start := len(s.items)
	s.items = append(s.items, values...)
	for len(s.elements) < len(s.items) {
		s.elements = append(s.elements, nil)
	}
	s.notify(start, len(s.items))
}

func (s *ReactiveSlice[T]) Insert(index int, values ...T) {
	if len(values) == 0 {
		return
	}
	s.items = slices.Insert(s.items, index, values...)
	for range values {
		s.elements = append(s.elements, nil)
	}
	s.notify(index, len(s.items))
}

func (s *ReactiveSlice[T]) Remove(index int) {
	s.items = slices.Delete(s.items, index, index+1)
	s.notify(index, len(s.elements))
}

func (s *ReactiveSlice[T]) Clear() {
	if len(s.items) == 0 {
		return
	}
	s.items = nil
	s.notify(0, len(s.elements))
}

func (s *ReactiveSlice[T]) Slice() []T {
	s.iter.track()
	out := make([]T, len(s.items))
	copy(out, s.items)
	for i := range out {
		if isTracking() {
			s.dep(i).track()
		}
	}
	return out
}

func (s *ReactiveSlice[T]) Range(fn func(index int, value T) bool) {
	s.iter.track()
	for i, value := range s.items {
		if isTracking() {
			s.dep(i).track()
		}
		if !fn(i, value) {
			return
		}
	}
}

func (s *ReactiveSlice[T]) notify(from, to int) {
	pass := newPass()
	s.iter.triggerPass(pass)

	pending := append([]*dep(nil), s.elements[from:to]...)
	for _, d := range pending {
		if d != nil {
			d.triggerPass(pass)
		}
	}
}

func (s *ReactiveSlice[T]) dep(index int) *dep {
	d := s.elements[index]
	if d == nil {
		d = new(dep)
		s.elements[index] = d
	}
	return d
}
