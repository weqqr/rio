package view

import (
	"reflect"

	"github.com/weqqr/rio/pkg/ui/react"
)

type Scope struct {
	node   *node
	comp   *Composition
	key    any
	hasKey bool
}

func (s *Scope) Key(key any) {
	s.key = key
	s.hasKey = true
}

func (s *Scope) Text(value string, mods ...Modifier) {
	m := merge(mods)
	n := s.emit(kindText, nil, 0)
	text, ok := n.base.(*textView)
	if !ok {
		text = &textView{measurer: s.comp.measurer}
		n.base = text
	}
	style := TextStyle{Size: DefaultTextSize}
	if size := m.font(); size > 0 {
		style.Size = size
	}
	if text.value != value || text.style != style {
		text.value = value
		text.style = style
		text.measured = false
	}
	n.setMods(m)
	n.mounted = true
}

func (s *Scope) Spacer(mods ...Modifier) {
	m := merge(mods)
	n := s.emit(kindSpacer, nil, 0)
	if n.base == nil {
		n.base = &spacerView{}
	}
	n.setMods(m)
	n.mounted = true
}

func (s *Scope) Custom(v View, mods ...Modifier) {
	m := merge(mods)
	n := s.emit(kindCustom, reflect.TypeOf(v), 0)
	n.base = v
	n.mods = m
	n.render = m.wrap(v)
	n.align, n.alignSet = m.alignment()
	n.weight = m.flex()
	n.mounted = true
}

func (s *Scope) Column(content func(*Scope), mods ...Modifier) {
	m := merge(mods)
	n := s.emit(kindColumn, nil, 0)
	column, ok := n.base.(*columnView)
	if !ok {
		column = &columnView{owner: n}
		n.base = column
	}
	column.owner = n
	column.gap = m.spacing()
	column.align, _ = m.alignment()
	n.setMods(m)
	n.body = func(c *Scope) { content(c) }
	n.run()
}

func (s *Scope) Row(content func(*Scope), mods ...Modifier) {
	m := merge(mods)
	n := s.emit(kindRow, nil, 0)
	row, ok := n.base.(*rowView)
	if !ok {
		row = &rowView{owner: n}
		n.base = row
	}
	row.owner = n
	row.gap = m.spacing()
	row.align, _ = m.alignment()
	n.setMods(m)
	n.body = func(c *Scope) { content(c) }
	n.run()
}

func (s *Scope) Box(content func(*Scope), mods ...Modifier) {
	m := merge(mods)
	n := s.emit(kindBox, nil, 0)
	box, ok := n.base.(*boxView)
	if !ok {
		box = &boxView{owner: n}
		n.base = box
	}
	box.owner = n
	box.align, _ = m.alignment()
	n.setMods(m)
	n.body = func(c *Scope) { content(c) }
	n.run()
}

func (s *Scope) Component(body func(*Scope)) {
	n := s.emit(kindGroup, nil, reflect.ValueOf(body).Pointer())
	n.body = body
	n.run()
}

func (s *Scope) Items[T any](items []T, key func(T) any, item func(*Scope, T)) {
	for _, value := range items {
		if key != nil {
			s.Key(key(value))
		}
		s.Component(func(c *Scope) { item(c, value) })
	}
}

func (s *Scope) State[T any](initial T) *react.Ref[T] {
	return s.Remember(func() *react.Ref[T] { return react.NewRef(initial) })
}

func (s *Scope) Derived[T any](compute func() T) *react.Computed[T] {
	return s.Remember(func() *react.Computed[T] { return react.NewComputed(compute) })
}

func (s *Scope) Remember[T any](init func() T) T {
	index := s.nextSlot()
	if index < len(s.node.slots) {
		if value, ok := s.node.slots[index].(T); ok {
			return value
		}
	}
	value := init()
	s.putSlot(index, value)
	return value
}

func (s *Scope) RememberKey[T any](key any, init func() T) T {
	index := s.nextSlot()
	if index < len(s.node.slots) {
		if entry, ok := s.node.slots[index].(keyedSlot); ok && sameKey(entry.key, key) {
			if value, ok := entry.value.(T); ok {
				return value
			}
		}
	}
	value := init()
	s.putSlot(index, keyedSlot{key: key, value: value})
	return value
}

func (s *Scope) SideEffect(effect func()) {
	s.node.sideEffects = append(s.node.sideEffects, effect)
}

func (s *Scope) DisposableEffect(key any, effect func() func()) {
	n := s.node
	index := s.nextSlot()
	var previous *disposable
	if index < len(n.slots) {
		previous, _ = n.slots[index].(*disposable)
	}
	if previous != nil && previous.started && sameKey(previous.key, key) {
		return
	}
	if previous != nil {
		n.forget(previous)
		if previous.cleanup != nil {
			cleanup := previous.cleanup
			previous.cleanup = nil
			cleanup()
		}
	}
	d := &disposable{key: key, started: true, cleanup: effect()}
	s.putSlot(index, d)
	n.disposables = append(n.disposables, d)
}

func (s *Scope) emit(kind nodeKind, typ reflect.Type, ident uintptr) *node {
	parent := s.node
	key, hasKey := s.key, s.hasKey
	s.key, s.hasKey = nil, false
	index := parent.runIndex
	parent.runIndex++
	if reuse := parent.match(kind, typ, ident, key, hasKey, index); reuse != nil {
		reuse.claimed = true
		parent.runNodes = append(parent.runNodes, reuse)
		return reuse
	}
	emitted := &node{
		owner:   s.comp,
		parent:  parent,
		kind:    kind,
		typ:     typ,
		ident:   ident,
		key:     key,
		hasKey:  hasKey,
		claimed: true,
	}
	parent.runNodes = append(parent.runNodes, emitted)
	return emitted
}

func (s *Scope) nextSlot() int {
	index := s.node.slot
	s.node.slot++
	return index
}

func (s *Scope) putSlot(index int, value any) {
	if index < len(s.node.slots) {
		s.node.slots[index] = value
		return
	}
	s.node.slots = append(s.node.slots, value)
}
