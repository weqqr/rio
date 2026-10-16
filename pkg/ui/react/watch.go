package react

type Flush int

const (
	FlushSync Flush = iota

	FlushPost
)

type WatchConfig struct {
	Immediate bool

	Flush Flush
}

func WatchEffect(fn func(), flush Flush) *Effect {
	e := &Effect{fn: fn, active: true, post: flush == FlushPost}
	recordScope(e)
	e.run()
	return e
}

func Watch[T any](get func() T, cb func(newValue, oldValue T), config *WatchConfig) *Effect {
	unchanged := same[T]()
	immediate := config != nil && config.Immediate
	e := &Effect{active: true, manualClean: true, post: config != nil && config.Flush == FlushPost}
	var (
		oldValue T
		started  bool
	)
	e.fn = func() {
		newValue := get()
		if !started {
			started = true
			oldValue = newValue
			if immediate {
				e.runCleanups()
				cb(newValue, newValue)
			}
			return
		}
		if unchanged(oldValue, newValue) {
			return
		}
		previous := oldValue
		oldValue = newValue
		e.runCleanups()
		cb(newValue, previous)
	}
	recordScope(e)
	e.run()
	return e
}
