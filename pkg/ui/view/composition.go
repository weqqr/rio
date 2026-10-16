package view

import "github.com/weqqr/rio/pkg/ui/react"

type Config struct {
	TextMeasurer TextMeasurer
	OnInvalidate func()
}

type Composition struct {
	config      Config
	measurer    TextMeasurer
	root        *node
	queue       []*node
	constraints Constraints
	measured    bool
	size        Size
	notified    bool
	disposed    bool
}

func New(config *Config) *Composition {
	c := &Composition{constraints: Unbounded}
	if config != nil {
		c.config = *config
	}
	c.measurer = c.config.TextMeasurer
	if c.measurer == nil {
		c.measurer = Monospace{AdvanceRatio: 0.5, LineHeightRatio: 1}
	}
	return c
}

func (c *Composition) SetContent(content func(*Scope)) {
	if c.disposed || content == nil {
		return
	}
	if c.root == nil {
		root := &node{owner: c, kind: kindColumn}
		root.base = &columnView{owner: root}
		root.render = root.mods.wrap(root.base)
		root.body = func(s *Scope) { content(s) }
		c.root = root
		root.mount()
		return
	}
	c.root.body = func(s *Scope) { content(s) }
	c.root.run()
}

func (c *Composition) Recompose() {
	if c.disposed {
		return
	}
	c.notified = false
	if len(c.queue) == 0 {
		return
	}
	react.Batch(func() {
		for len(c.queue) > 0 {
			n := c.queue[0]
			c.queue = c.queue[1:]
			n.pending = false
			if n.mounted && n.eff != nil {
				n.eff.Run()
			}
		}
	})
}

func (c *Composition) Measure(constraints Constraints) Size {
	c.Recompose()
	c.layout(constraints)
	return c.size
}

func (c *Composition) Draw(canvas Canvas) {
	c.Recompose()
	c.layout(c.constraints)
	if c.root == nil {
		return
	}
	root := c.root.view()
	root.Place(Point{})
	root.Draw(canvas)
}

func (c *Composition) Dispose() {
	if c.disposed {
		return
	}
	c.disposed = true
	c.queue = nil
	if c.root != nil {
		c.root.unmount()
		c.root = nil
	}
}

func (c *Composition) layout(constraints Constraints) {
	if c.root == nil {
		c.size = Size{}
		c.constraints = constraints
		c.measured = true
		return
	}
	if c.measured && c.constraints == constraints {
		return
	}
	c.size = c.root.view().Measure(constraints)
	c.constraints = constraints
	c.measured = true
}

func (c *Composition) request(n *node) {
	if c.disposed || !n.mounted || n.pending {
		return
	}
	n.pending = true
	c.queue = append(c.queue, n)
	if c.notified {
		return
	}
	c.notified = true
	if c.config.OnInvalidate != nil {
		c.config.OnInvalidate()
	}
}
