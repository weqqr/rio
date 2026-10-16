package ecs

import (
	"iter"
	"reflect"
	"sort"
	"strconv"
)

type Registry struct {
	resources      map[reflect.Type]any
	resourceEvents []resourceEvent
	archetypes     []*archetype
	byTypeKey      map[string]*archetype
	typeIDs        map[reflect.Type]int

	entities []entityRecord
	free     []int

	epoch uint64
}

type entityRecord struct {
	gen       uint32
	alive     bool
	archetype int
	row       int
	id        EntityID
}

type resourceEvent struct {
	typ   reflect.Type
	added bool
	value any
}

func NewRegistry() *Registry {
	return &Registry{
		resources: make(map[reflect.Type]any),
		byTypeKey: make(map[string]*archetype),
		typeIDs:   make(map[reflect.Type]int),
		entities:  make([]entityRecord, 0, 16),
		free:      make([]int, 0),
	}
}

func (r *Registry) Get[T interface{ *TBase }, TBase any]() (T, bool) {
	ty := reflect.TypeFor[T]()
	resource, ok := r.resources[ty]
	if !ok {
		return *new(T), false
	}
	return resource.(T), ok
}

func (r *Registry) Set[T interface{ *TBase }, TBase any](value T) {
	ty := reflect.TypeFor[T]()
	if _, ok := r.resources[ty]; !ok {
		r.resourceEvents = append(r.resourceEvents, resourceEvent{typ: ty, added: true, value: value})
	}
	r.resources[ty] = value
}

func (r *Registry) Unset[T interface{ *TBase }, TBase any]() {
	ty := reflect.TypeFor[T]()
	if v, ok := r.resources[ty]; ok {
		r.resourceEvents = append(r.resourceEvents, resourceEvent{typ: ty, value: v})
		delete(r.resources, ty)
	}
}

func (r *Registry) Query[T any]() iter.Seq2[EntityID, T] {
	return Query[T]{reg: r}.Iter()
}

func Component[T any](r *Registry, id EntityID) *T {
	if !r.Alive(id) {
		return nil
	}
	rec := &r.entities[id.index()]
	arc := r.archetypes[rec.archetype]
	col, ok := arc.columnByType[reflect.TypeFor[T]()]
	if !ok {
		return nil
	}
	return arc.columns[col].rowPtr(rec.row).(*T)
}

func (c *column) rowPtr(row int) any {
	return c.slice.Index(row).Addr().Interface()
}

func (r *Registry) Spawn(components ...any) EntityID {
	comps := make(map[reflect.Type]any, len(components))
	for _, c := range components {
		if c == nil {
			panic("ecs: nil component")
		}
		comps[reflect.TypeOf(c)] = c
	}

	arc := r.archetypeFor(comps)

	id, rec := r.allocEntity()
	arc.appendRow(id, comps)
	rec.archetype = r.archetypeIndex(arc)
	rec.row = arc.count - 1
	r.epoch++
	return id
}

func (r *Registry) Despawn(id EntityID) {
	r.checkEntityIndex(id, "Despawn")
	rec := &r.entities[id.index()]
	if !rec.alive || rec.gen != id.generation() {
		return
	}

	r.epoch++

	moved, swapped := r.archetypes[rec.archetype].removeRow(rec.row)
	if swapped {
		r.entities[moved.index()].row = rec.row
	}

	rec.alive = false
	rec.archetype = -1
	rec.gen++
	r.free = append(r.free, id.index())
}

func (r *Registry) Add(id EntityID, components ...any) {
	r.checkEntityIndex(id, "Add")
	rec := &r.entities[id.index()]
	if !rec.alive || rec.gen != id.generation() {
		return
	}

	old := r.archetypes[rec.archetype]

	for _, c := range components {
		if c == nil {
			panic("ecs: nil component")
		}
	}
	hasNew := false
	for _, c := range components {
		if !old.typeSet[reflect.TypeOf(c)] {
			hasNew = true
			break
		}
	}
	if !hasNew {
		for _, c := range components {
			col := old.columns[old.columnByType[reflect.TypeOf(c)]]
			col.setRow(rec.row, c)
		}
		return
	}

	comps := make(map[reflect.Type]any, len(old.types)+len(components))
	for _, typ := range old.types {
		comps[typ] = old.columns[old.columnByType[typ]].copyRow(rec.row)
	}
	for _, c := range components {
		comps[reflect.TypeOf(c)] = c
	}

	r.epoch++
	r.migrate(rec, old, r.archetypeFor(comps), comps)
}

func (r *Registry) Remove[T any](id EntityID) {
	r.checkEntityIndex(id, "Remove")
	rec := &r.entities[id.index()]
	if !rec.alive || rec.gen != id.generation() {
		return
	}

	typ := reflect.TypeFor[T]()
	old := r.archetypes[rec.archetype]
	if !old.typeSet[typ] {
		return
	}

	comps := make(map[reflect.Type]any, len(old.types)-1)
	for _, t := range old.types {
		if t == typ {
			continue
		}
		comps[t] = old.columns[old.columnByType[t]].copyRow(rec.row)
	}

	r.epoch++
	r.migrate(rec, old, r.archetypeFor(comps), comps)
}

func (r *Registry) Alive(id EntityID) bool {
	idx := id.index()
	if idx < 0 || idx >= len(r.entities) {
		return false
	}
	rec := r.entities[idx]
	return rec.alive && rec.gen == id.generation()
}

func (r *Registry) Epoch() uint64 { return r.epoch }

func (r *Registry) migrate(rec *entityRecord, old, next *archetype, comps map[reflect.Type]any) {
	moved, swapped := old.removeRow(rec.row)
	if swapped {
		r.entities[moved.index()].row = rec.row
	}

	next.appendRow(rec.id, comps)
	rec.row = next.count - 1
	rec.archetype = r.archetypeIndex(next)
}

func (r *Registry) archetypeFor(comps map[reflect.Type]any) *archetype {
	types := make([]reflect.Type, 0, len(comps))
	for typ := range comps {
		types = append(types, typ)
	}
	sort.Slice(types, func(i, j int) bool {
		return elementSorterLess(types[i], types[j])
	})

	ids := make([]int, len(types))
	for i, typ := range types {
		ids[i] = r.typeID(typ)
	}
	sort.Ints(ids)

	key := archetypeKey(ids)
	if arc, ok := r.byTypeKey[key]; ok {
		return arc
	}

	arc := newArchetype(types)
	r.byTypeKey[key] = arc
	arc.index = len(r.archetypes)
	r.archetypes = append(r.archetypes, arc)
	return arc
}

func (r *Registry) archetypeIndex(arc *archetype) int {
	return arc.index
}

func (r *Registry) allocEntity() (EntityID, *entityRecord) {
	if n := len(r.free); n > 0 {
		index := r.free[n-1]
		r.free = r.free[:n-1]
		rec := &r.entities[index]
		rec.alive = true
		rec.archetype = -1
		rec.id = makeEntityID(index, rec.gen)
		return rec.id, rec
	}

	index := len(r.entities)
	rec := &entityRecord{
		alive:     true,
		archetype: -1,
		gen:       1,
	}
	rec.id = makeEntityID(index, rec.gen)
	r.entities = append(r.entities, *rec)
	return rec.id, &r.entities[index]
}

func (r *Registry) typeID(typ reflect.Type) int {
	if id, ok := r.typeIDs[typ]; ok {
		return id
	}
	id := len(r.typeIDs)
	r.typeIDs[typ] = id
	return id
}

func archetypeKey(ids []int) string {
	b := make([]byte, 0, len(ids)*4)
	for _, id := range ids {
		b = append(b, byte(id), byte(id>>8), byte(id>>16), byte(id>>24))
	}
	return string(b)
}

func (r *Registry) checkEntityIndex(id EntityID, op string) {
	if idx := id.index(); idx < 0 || idx >= len(r.entities) {
		panic("ecs: " + op + ": entity index " + strconv.Itoa(idx) + " out of range (len " + strconv.Itoa(len(r.entities)) + ")")
	}
}
