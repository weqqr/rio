package ecs

import (
	"iter"
	"reflect"
	"sort"
	"sync"
	"unsafe"
)

type EntityID int64

func (id EntityID) index() int { return int(uint64(id) >> 32) }

func (id EntityID) generation() uint32 { return uint32(uint64(id)) }

func makeEntityID(index int, gen uint32) EntityID {
	return EntityID(uint64(uint32(index))<<32 | uint64(gen))
}

type Query[T any] struct {
	reg *Registry
}

func (e *Query[T]) Types() []reflect.Type {
	return queryPlanFor[T]().types
}

func (e Query[T]) WithRegistry(r *Registry) Query[T] {
	e.reg = r
	return e
}

func (Query[T]) isQuery() {}

func (e Query[T]) Iter() iter.Seq2[EntityID, T] {
	if e.reg == nil {
		panic("ecs: entity query is not bound to a registry")
	}
	r := e.reg
	plan := queryPlanFor[T]()
	return func(yield func(EntityID, T) bool) {
		sepoch := r.epoch
		for _, arc := range r.archetypes {
			if !plan.matches(arc) {
				continue
			}

			columns := arc.columns
			count := arc.count
			colOf := arc.columnByType

			ne := len(plan.entries)
			var basesBuf [8]unsafe.Pointer
			var sizesBuf [8]uintptr
			var bases []unsafe.Pointer
			var sizes []uintptr
			if ne <= len(basesBuf) {
				bases = basesBuf[:ne]
				sizes = sizesBuf[:ne]
			} else {
				bases = make([]unsafe.Pointer, ne)
				sizes = make([]uintptr, ne)
			}
			for i := range plan.entries {
				col := columns[colOf[plan.entries[i].typ]]
				bases[i] = col.basePtr()
				sizes[i] = col.size
			}

			for row := 0; row < count; row++ {
				var t T
				base := unsafe.Pointer(&t)
				for i := range plan.entries {
					fieldPtr := unsafe.Add(bases[i], uintptr(row)*sizes[i])
					*(*unsafe.Pointer)(unsafe.Add(base, plan.entries[i].offset)) = fieldPtr
				}
				keep := yield(arc.entityIDs[row], t)
				if r.epoch != sepoch {
					panic("ecs: structural change during query iteration; buffer Spawn/Despawn/Add/Remove outside Iter")
				}
				if !keep {
					return
				}
			}
		}
	}
}

type queryPlan struct {
	types   []reflect.Type
	entries []planEntry
}

type planEntry struct {
	typ    reflect.Type
	offset uintptr
}

func (p *queryPlan) matches(a *archetype) bool {
	for _, typ := range p.types {
		if !a.typeSet[typ] {
			return false
		}
	}
	return true
}

var queryPlans sync.Map

func queryPlanFor[T any]() *queryPlan {
	key := reflect.TypeFor[T]()
	if v, ok := queryPlans.Load(key); ok {
		return v.(*queryPlan)
	}

	plan := buildQueryPlan(key)
	actual, _ := queryPlans.LoadOrStore(key, plan)
	return actual.(*queryPlan)
}

func buildQueryPlan(ty reflect.Type) *queryPlan {
	if ty.Kind() != reflect.Struct {
		panic("ecs: query type must be a struct of pointer-to-component fields")
	}

	plan := &queryPlan{
		types:   make([]reflect.Type, 0, ty.NumField()),
		entries: make([]planEntry, 0, ty.NumField()),
	}

	seen := make(map[reflect.Type]struct{}, ty.NumField())
	for i := 0; i < ty.NumField(); i++ {
		field := ty.Field(i)
		if field.Type.Kind() != reflect.Pointer {
			panic("ecs: query field " + field.Name + " must be a pointer to a component type")
		}
		elem := field.Type.Elem()
		if _, ok := seen[elem]; !ok {
			seen[elem] = struct{}{}
			plan.types = append(plan.types, elem)
		}
		plan.entries = append(plan.entries, planEntry{
			typ:    elem,
			offset: field.Offset,
		})
	}

	sort.Slice(plan.types, func(i, j int) bool {
		return elementSorterLess(plan.types[i], plan.types[j])
	})

	return plan
}

type column struct {
	elemType reflect.Type
	size     uintptr
	slice    reflect.Value
}

func newColumn(elemType reflect.Type) *column {
	return &column{
		elemType: elemType,
		size:     elemType.Size(),
		slice:    reflect.MakeSlice(reflect.SliceOf(elemType), 0, 8),
	}
}

func (c *column) append(val any) {
	c.slice = reflect.Append(c.slice, reflect.ValueOf(val))
}

func (c *column) truncate(n int) {
	if c.slice.Len() <= n {
		return
	}
	c.slice = c.slice.Slice(0, n)
}

func (c *column) swapLastInto(index, last int) {
	c.slice.Index(index).Set(c.slice.Index(last))
}

func (c *column) copyRow(row int) any {
	return c.slice.Index(row).Interface()
}

func (c *column) setRow(row int, val any) {
	c.slice.Index(row).Set(reflect.ValueOf(val))
}

func (c *column) basePtr() unsafe.Pointer {
	return c.slice.UnsafePointer()
}

type archetype struct {
	types   []reflect.Type
	typeSet map[reflect.Type]bool

	columns      []*column
	entityIDs    []EntityID
	count        int
	columnByType map[reflect.Type]int
	index        int
}

func newArchetype(types []reflect.Type) *archetype {
	sorted := make([]reflect.Type, len(types))
	copy(sorted, types)
	sort.Slice(sorted, func(i, j int) bool {
		return elementSorterLess(sorted[i], sorted[j])
	})

	a := &archetype{
		types:        sorted,
		typeSet:      make(map[reflect.Type]bool, len(sorted)),
		columns:      make([]*column, 0, len(sorted)),
		entityIDs:    make([]EntityID, 0, 8),
		columnByType: make(map[reflect.Type]int, len(sorted)),
	}
	for _, typ := range sorted {
		a.typeSet[typ] = true
		a.columns = append(a.columns, newColumn(typ))
	}
	for idx, typ := range sorted {
		a.columnByType[typ] = idx
	}
	return a
}

func (a *archetype) appendRow(id EntityID, comps map[reflect.Type]any) {
	for _, col := range a.columns {
		col.append(comps[col.elemType])
	}
	a.entityIDs = append(a.entityIDs, id)
	a.count++
}

func (a *archetype) removeRow(index int) (EntityID, bool) {
	last := a.count - 1
	if index != last {
		for _, col := range a.columns {
			col.swapLastInto(index, last)
		}
		a.entityIDs[index] = a.entityIDs[last]
	}
	for _, col := range a.columns {
		col.truncate(last)
	}
	a.entityIDs = a.entityIDs[:last]
	a.count = last

	if index != last {
		return a.entityIDs[index], true
	}
	return 0, false
}

func elementSorterLess(a, b reflect.Type) bool {
	if a.PkgPath() != b.PkgPath() {
		return a.PkgPath() < b.PkgPath()
	}
	return a.Name() < b.Name()
}
