package view

import (
	"reflect"

	"github.com/weqqr/rio/pkg/ui/react"
)

type nodeKind uint8

const (
	kindGroup nodeKind = iota
	kindText
	kindSpacer
	kindCustom
	kindColumn
	kindRow
	kindBox
)

type disposable struct {
	key     any
	started bool
	cleanup func()
}

type keyedSlot struct {
	key   any
	value any
}

type node struct {
	owner  *Composition
	parent *node

	kind   nodeKind
	typ    reflect.Type
	ident  uintptr
	key    any
	hasKey bool

	base   View
	render View
	mods   Modifier

	body  func(*Scope)
	scope *Scope
	rs    *react.EffectScope
	eff   *react.Effect

	slots []any
	slot  int

	align    Alignment
	alignSet bool
	weight   float32

	children []*node
	flat     []*node
	flatOK   bool

	runNodes []*node
	runIndex int

	measured Size
	claimed  bool
	pending  bool
	mounted  bool

	sideEffects []func()
	disposables []*disposable
}

func (n *node) view() View { return n.render }

func (n *node) setMods(mods Modifier) {
	if n.mods.same(mods) && n.render != nil {
		return
	}
	n.mods = mods
	n.render = mods.wrap(n.base)
	n.align, n.alignSet = mods.alignment()
	n.weight = mods.flex()
}

func (n *node) mount() {
	n.mounted = true
	if n.body == nil {
		return
	}
	n.rs = react.NewEffectScope()
	n.scope = &Scope{node: n, comp: n.owner}
	owner := n.owner
	n.rs.Run(func() {
		react.OnScopeDispose(n.runDisposables)
		n.eff = react.NewEffect(n.runBody, &react.EffectConfig{
			Scheduler: func() { owner.request(n) },
		})
	})
}

func (n *node) run() {
	if n.mounted {
		n.eff.Run()
		return
	}
	n.mount()
}

func (n *node) unmount() {
	if !n.mounted {
		return
	}
	n.mounted = false
	n.pending = false
	if n.rs != nil {
		n.rs.Stop()
		n.rs = nil
		n.eff = nil
	}
	n.children = nil
	n.flat = nil
	n.flatOK = false
	n.slots = nil
	n.sideEffects = nil
	n.disposables = nil
	n.claimed = false
	n.invalidateAncestors()
}

func (n *node) runBody() {
	n.runIndex = 0
	n.runNodes = n.runNodes[:0]
	n.slot = 0
	n.sideEffects = n.sideEffects[:0]
	for _, child := range n.children {
		child.claimed = false
	}
	n.rs.Run(func() { n.body(n.scope) })
	for _, child := range n.children {
		if !child.claimed {
			child.unmount()
		}
	}
	n.children = append(n.children[:0], n.runNodes...)
	n.flatOK = false
	n.invalidateAncestors()
	n.owner.measured = false
	if len(n.sideEffects) > 0 {
		effects := append([]func(){}, n.sideEffects...)
		n.sideEffects = n.sideEffects[:0]
		for _, effect := range effects {
			effect()
		}
	}
}

func (n *node) invalidateAncestors() {
	for parent := n.parent; parent != nil; parent = parent.parent {
		parent.flatOK = false
	}
}

func (n *node) layoutChildren() []*node {
	if n.flatOK {
		return n.flat
	}
	n.flat = flatten(n.children, n.flat[:0])
	n.flatOK = true
	return n.flat
}

func flatten(children []*node, out []*node) []*node {
	for _, child := range children {
		if child.kind == kindGroup {
			out = flatten(child.children, out)
			continue
		}
		out = append(out, child)
	}
	return out
}

func (n *node) runDisposables() {
	for i := len(n.disposables) - 1; i >= 0; i-- {
		d := n.disposables[i]
		if d.cleanup == nil {
			continue
		}
		cleanup := d.cleanup
		d.cleanup = nil
		cleanup()
	}
}

func (n *node) forget(d *disposable) {
	for i, item := range n.disposables {
		if item == d {
			n.disposables = append(n.disposables[:i], n.disposables[i+1:]...)
			return
		}
	}
}

func (n *node) match(kind nodeKind, typ reflect.Type, ident uintptr, key any, hasKey bool, index int) *node {
	if hasKey {
		for _, child := range n.children {
			if child.claimed || !child.hasKey || child.kind != kind || child.typ != typ || child.ident != ident {
				continue
			}
			if sameKey(child.key, key) {
				return child
			}
		}
		return nil
	}
	if index >= len(n.children) {
		return nil
	}
	child := n.children[index]
	if child.claimed || child.hasKey || child.kind != kind || child.typ != typ || child.ident != ident {
		return nil
	}
	return child
}

func sameKey(a, b any) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	av, bv := reflect.ValueOf(a), reflect.ValueOf(b)
	if av.Type() != bv.Type() || !av.Comparable() {
		return false
	}
	return av.Equal(bv)
}
