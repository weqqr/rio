package ecs

import (
	"reflect"
	"runtime"
	"strings"
	"sync"
)

type Phase int

const (
	PhaseStartup Phase = iota
	PhasePreUpdate
	PhaseUpdate
	PhasePostUpdate
	PhaseFixedUpdate
	PhaseRender
)

const phaseHook Phase = -1

type Schedule struct {
	systems map[Phase][]boundSystem
	onEnter map[reflect.Type][]boundSystem
	onExit  map[reflect.Type][]boundSystem
}

func NewSchedule() Schedule {
	return Schedule{
		systems: make(map[Phase][]boundSystem),
		onEnter: make(map[reflect.Type][]boundSystem),
		onExit:  make(map[reflect.Type][]boundSystem),
	}
}

func (s *Schedule) Add(phase Phase, system any) {
	s.systems[phase] = append(s.systems[phase], bindSystem(system, phase))
}

func (s *Schedule) AddOnEnter[T interface{ *TBase }, TBase any](system any) {
	ty := reflect.TypeFor[T]()
	s.onEnter[ty] = append(s.onEnter[ty], bindSystem(system, phaseHook))
}

func (s *Schedule) AddOnExit[T interface{ *TBase }, TBase any](system any) {
	ty := reflect.TypeFor[T]()
	s.onExit[ty] = append(s.onExit[ty], bindSystem(system, phaseHook))
}

type Executor struct{}

func NewExecutor() *Executor {
	return &Executor{}
}

func (e *Executor) Execute(registry *Registry, schedule Schedule, phase Phase) {
	dispatchResourceEvents(registry, schedule)
	for _, sys := range schedule.systems[phase] {
		sys.call(registry)
	}
}

const maxResourceEventRounds = 64

func dispatchResourceEvents(r *Registry, schedule Schedule) {
	for round := 0; len(r.resourceEvents) > 0; round++ {
		if round == maxResourceEventRounds {
			panic("ecs: resource transitions did not settle after " + itoa(maxResourceEventRounds) + " rounds")
		}
		pending := r.resourceEvents
		r.resourceEvents = nil
		for _, ev := range pending {
			if ev.added {
				for _, sys := range schedule.onEnter[ev.typ] {
					sys.callWith(r, ev.typ, ev.value)
				}
				continue
			}
			for _, sys := range schedule.onExit[ev.typ] {
				sys.callWith(r, ev.typ, ev.value)
			}
		}
	}
}

type paramKind uint8

const (
	paramRegistry paramKind = iota
	paramQuery
	paramResource
)

type paramBinding struct {
	kind      paramKind
	typ       reflect.Type
	construct reflect.Value
}

type boundSystem struct {
	fn     reflect.Value
	params []paramBinding
	phase  Phase
	label  string
}

func bindSystem(system any, phase Phase) boundSystem {
	fn := reflect.ValueOf(system)
	if fn.Kind() != reflect.Func {
		panic("ecs: system must be a function")
	}
	return boundSystem{fn: fn, params: paramBindings(fn), phase: phase, label: zoneLabel(fn)}
}

var paramBindingsCache sync.Map

func paramBindings(fn reflect.Value) []paramBinding {
	ft := fn.Type()
	if v, ok := paramBindingsCache.Load(ft); ok {
		return v.([]paramBinding)
	}

	params := make([]paramBinding, 0, ft.NumIn())
	for i := 0; i < ft.NumIn(); i++ {
		pt := ft.In(i)
		switch {
		case pt == reflect.TypeOf((*Registry)(nil)):
			params = append(params, paramBinding{kind: paramRegistry, typ: pt})

		case pt.Kind() == reflect.Struct && pt.Implements(queryMarkerType):
			params = append(params, paramBinding{
				kind:      paramQuery,
				typ:       pt,
				construct: reflect.Zero(pt).MethodByName("WithRegistry"),
			})

		case pt.Kind() == reflect.Pointer && pt.Elem().Implements(queryMarkerType):
			panic("ecs: system " + systemName(fn) + ": pass Entities query by value, not *Entities")

		case pt.Kind() == reflect.Pointer:
			params = append(params, paramBinding{kind: paramResource, typ: pt})

		default:
			panic("ecs: system " + systemName(fn) + ": unsupported parameter type " + pt.String() + " at index " + itoa(i))
		}
	}

	actual, _ := paramBindingsCache.LoadOrStore(ft, params)
	return actual.([]paramBinding)
}

type queryMarker interface {
	isQuery()
}

var queryMarkerType = reflect.TypeFor[queryMarker]()

func (b *boundSystem) call(r *Registry) {
	b.callWith(r, nil, nil)
}

func (b *boundSystem) callWith(r *Registry, lastTyp reflect.Type, lastVal any) {
	args := make([]reflect.Value, len(b.params))
	for i, p := range b.params {
		switch p.kind {
		case paramRegistry:
			args[i] = reflect.ValueOf(r)

		case paramQuery:
			args[i] = p.construct.Call([]reflect.Value{reflect.ValueOf(r)})[0]

		case paramResource:
			if p.typ == lastTyp {
				args[i] = reflect.ValueOf(lastVal)
				continue
			}
			res, ok := r.resources[p.typ]
			if !ok {
				if b.phase == PhaseStartup {
					panic("ecs: system " + systemName(b.fn) + ": resource " + p.typ.String() + " is not registered")
				}
				return
			}
			args[i] = reflect.ValueOf(res)
		}
	}
	b.fn.Call(args)
}

func systemName(fn reflect.Value) string {
	return runtime.FuncForPC(fn.Pointer()).Name()
}

func zoneLabel(fn reflect.Value) string {
	name := strings.TrimSuffix(systemName(fn), "-fm")
	return name[strings.LastIndex(name, "/")+1:]
}

func itoa(v int) string {
	if v < 10 {
		return string(rune('0' + v))
	}
	return string(rune('0'+v/10)) + string(rune('0'+v%10))
}
