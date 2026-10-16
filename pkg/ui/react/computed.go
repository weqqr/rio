package react

type Computed[T any] struct {
	getter    func() T
	unchanged sameFunc[T]
	dep       dep
	effect    *Effect
	value     T
	dirty     bool
}

func NewComputed[T any](get func() T) *Computed[T] {
	return NewComputedEqual(get, nil)
}

func NewComputedEqual[T any](get func() T, equal func(newValue, oldValue T) bool) *Computed[T] {
	c := &Computed[T]{getter: get, dirty: true, unchanged: same[T]()}
	if equal != nil {
		c.unchanged = equal
	}
	c.effect = &Effect{active: true, syncTrigger: true, scheduler: c.invalidate}
	c.effect.fn = func() {

		c.dirty = false
		c.value = c.getter()
	}
	recordScope(c.effect)
	return c
}

func (c *Computed[T]) Get() T {
	if !c.effect.active {
		return c.value
	}
	c.dep.track()
	if c.dirty {
		c.effect.run()
	}
	return c.value
}

func (c *Computed[T]) Peek() T {
	if c.effect.active && c.dirty {
		c.effect.run()
	}
	return c.value
}

func (c *Computed[T]) Stop() {
	c.effect.Stop()
	c.dep.subs = nil
}

func (c *Computed[T]) invalidate() {
	if c.dirty {
		return
	}
	c.dirty = true
	if !c.dep.hasSubs() {
		return
	}
	previous := c.value
	c.effect.run()
	if c.unchanged(previous, c.value) {
		return
	}
	c.dep.trigger()
}
