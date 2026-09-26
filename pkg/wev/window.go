package wev

type WindowConfig struct {
	Title  string
	Width  int
	Height int
}

type Window struct {
	loop   *EventLoop
	handle WindowHandle
}

func (w *Window) Destroy() error {
	delete(w.loop.windows, w.handle.id())
	return w.loop.platform.DestroyWindow(w.handle)
}
