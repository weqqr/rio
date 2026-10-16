package react

var (
	activeEffect *Effect
	tracking     = true
)

func isTracking() bool { return tracking && activeEffect != nil && activeEffect.active }

type Effect struct {
	fn        func()
	scheduler func()
	deps      []*dep
	cleanups  []func()
	scope     *EffectScope

	active       bool
	syncTrigger  bool
	post         bool
	manualClean  bool
	queued       bool
	pendingSched bool
	notified     uint64
}

type EffectConfig struct {
	Scheduler func()

	Lazy bool
}

func NewEffect(fn func(), config *EffectConfig) *Effect {
	e := &Effect{fn: fn, active: true}
	lazy := false
	if config != nil {
		e.scheduler = config.Scheduler
		lazy = config.Lazy
	}
	recordScope(e)
	if !lazy {
		e.run()
	}
	return e
}

func (e *Effect) Run() { e.run() }

func (e *Effect) Stop() {
	if !e.active {
		return
	}
	e.cleanupDeps()
	e.active = false
	e.runCleanups()
	e.fn = nil
	if e.scope != nil {
		e.scope.forget(e)
		e.scope = nil
	}
}

func (e *Effect) Active() bool { return e.active }

func (e *Effect) run() {
	if !e.active {
		return
	}
	e.cleanupDeps()
	if !e.manualClean {
		e.runCleanups()
	}
	previous := activeEffect
	activeEffect = e
	defer func() { activeEffect = previous }()
	e.fn()
}

func (e *Effect) trigger() {
	if e.syncTrigger {
		e.scheduler()
		return
	}
	if batchDepth == 0 && !e.post {
		if e.scheduler != nil {
			e.scheduler()
			return
		}
		e.run()
		return
	}
	e.pendingSched = e.scheduler != nil
	e.enqueue()
}

func (e *Effect) enqueue() {
	if e.queued {
		return
	}
	e.queued = true
	if e.post {
		postQueue = append(postQueue, e)
	} else {
		queue = append(queue, e)
	}
	if batchDepth == 0 {
		flush()
	}
}

func (e *Effect) perform() {
	if e.pendingSched && e.scheduler != nil {
		e.scheduler()
		return
	}
	e.run()
}

func (e *Effect) cleanupDeps() {
	for _, d := range e.deps {
		d.remove(e)
	}
	e.deps = e.deps[:0]
}

func (e *Effect) runCleanups() {
	if len(e.cleanups) == 0 {
		return
	}
	cleanups := e.cleanups
	e.cleanups = nil
	for _, fn := range cleanups {
		fn()
	}
}

func OnCleanup(fn func()) {
	if activeEffect != nil {
		activeEffect.cleanups = append(activeEffect.cleanups, fn)
	}
}

func Untrack[T any](fn func() T) T {
	previous := tracking
	tracking = false
	defer func() { tracking = previous }()
	return fn()
}
