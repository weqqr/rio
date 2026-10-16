package plugin

import (
	"fmt"
	"reflect"
	"slices"

	"github.com/weqqr/rio/pkg/ecs"
	"github.com/weqqr/rio/pkg/sets"
)

type Mounter interface {
	Mount(*Context)
}

type Unmounter interface {
	Unmount(*Context) error
}

type Context struct {
	Registry *ecs.Registry
	Schedule *ecs.Schedule

	added   []Mounter
	present sets.Set[reflect.Type]
	closers []func() error
}

func New() *Context {
	schedule := ecs.NewSchedule()
	return &Context{
		Registry: ecs.NewRegistry(),
		Schedule: &schedule,
		present:  sets.New[reflect.Type](),
	}
}

func (e *Context) AddPlugins(plugins ...Mounter) {
	for _, plugin := range plugins {
		typ := reflect.TypeOf(plugin)
		if e.present.Contains(typ) {
			continue
		}
		e.present.Insert(typ)
		e.added = append(e.added, plugin)
		plugin.Mount(e)
	}
}

func (e *Context) OnClose(close func() error) {
	e.closers = append(e.closers, close)
}

func (e *Context) Finish() error {
	for _, plugin := range e.added {
		if unmounter, ok := plugin.(Unmounter); ok {
			if err := unmounter.Unmount(e); err != nil {
				return fmt.Errorf("unmount %T: %w", plugin, err)
			}
		}
	}
	return nil
}

func (e *Context) RunClosers() {
	for _, v := range slices.Backward(e.closers) {
		_ = v()
	}
}
