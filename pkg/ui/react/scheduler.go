package react

var (
	queue     []*Effect
	queueNext int

	postQueue []*Effect
	postNext  int

	tickFns []func()

	flushing   bool
	batchDepth int
)

func Batch(fn func()) {
	batchDepth++
	defer func() {
		batchDepth--
		if batchDepth == 0 {
			flush()
		}
	}()
	fn()
}

func NextTick(fn func()) {
	if fn == nil {
		return
	}
	if flushing || batchDepth > 0 {
		tickFns = append(tickFns, fn)
		return
	}
	flush()
	fn()
}

func flush() {
	if flushing {
		return
	}
	flushing = true
	for queueNext < len(queue) || postNext < len(postQueue) {
		if queueNext < len(queue) {
			e := queue[queueNext]
			queueNext++
			e.queued = false
			e.perform()
			continue
		}
		e := postQueue[postNext]
		postNext++
		e.queued = false
		e.perform()
	}
	queue = queue[:0]
	queueNext = 0
	postQueue = postQueue[:0]
	postNext = 0
	flushing = false

	for len(tickFns) > 0 {
		fns := tickFns
		tickFns = nil
		for _, fn := range fns {
			fn()
		}
	}
}
