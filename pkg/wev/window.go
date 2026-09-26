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
	w.loop.removeWindow(w)
	return w.loop.platform.DestroyWindow(w.handle)
}

func (w *Window) RawWindowHandle() uintptr {
	return w.handle.rawHandle()
}

func (w *Window) Size() (width, height int) {
	return w.handle.size()
}

func (w *Window) DPI() int {
	return w.handle.dpi()
}
