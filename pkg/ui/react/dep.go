package react

type dep struct {
	subs  []*Effect
	depth int
}

func (d *dep) track() {
	e := activeEffect
	if e == nil || !e.active || !tracking || d.has(e) {
		return
	}
	d.subs = append(d.subs, e)
	e.deps = append(e.deps, d)
}

func (d *dep) trigger() { d.triggerPass(newPass()) }

func (d *dep) triggerPass(pass uint64) {
	d.depth++
	defer func() {
		d.depth--
		if d.depth == 0 {
			d.compact()
		}
	}()
	n := len(d.subs)
	for i := range n {
		e := d.subs[i]
		if e == nil || e == activeEffect || e.notified == pass {

			continue
		}
		e.notified = pass
		e.trigger()
	}
}

var passes uint64

func newPass() uint64 {
	passes++
	return passes
}

func (d *dep) has(e *Effect) bool {
	for _, s := range d.subs {
		if s == e {
			return true
		}
	}
	return false
}

func (d *dep) hasSubs() bool {
	for _, s := range d.subs {
		if s != nil {
			return true
		}
	}
	return false
}

func (d *dep) remove(e *Effect) {
	for i, s := range d.subs {
		if s != e {
			continue
		}
		if d.depth > 0 {
			d.subs[i] = nil
			return
		}
		d.subs = append(d.subs[:i], d.subs[i+1:]...)
		return
	}
}

func (d *dep) compact() {
	subs := d.subs[:0]
	for _, s := range d.subs {
		if s != nil {
			subs = append(subs, s)
		}
	}
	if len(subs) == 0 {
		d.subs = nil
		return
	}
	d.subs = subs
}
