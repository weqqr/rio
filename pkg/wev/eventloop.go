package wev

type EventHandler any

type CloseRequestedHandler interface {
	OnCloseRequested(e *EventLoop, window *Window) error
}

type WindowCreatedHandler interface {
	OnWindowCreated(e *EventLoop, window *Window) error
}

type WindowResizeHandler interface {
	OnWindowResize(e *EventLoop, window *Window, width, height int) error
}

type WindowDPIChangedHandler interface {
	OnWindowDPIChanged(e *EventLoop, window *Window, dpi int) error
}

type Platform interface {
	CreateWindow(config WindowConfig) (WindowHandle, error)
	DestroyWindow(handle WindowHandle) error
	Run() error
	PostQuit()
}

type Dispatcher interface {
	CloseRequested(handle WindowHandle)
	WindowResized(handle WindowHandle, width, height int)
	WindowDPIChanged(handle WindowHandle, dpi int)
}

type WindowHandle interface {
	id() uintptr
	rawHandle() uintptr
	size() (int, int)
	dpi() int
}

type EventLoop struct {
	handler  EventHandler
	platform Platform
	windows  map[uintptr]*Window
	pending  []*Window // windows created before Run; OnWindowCreated is deferred until then
	err      error
}

func (e *EventLoop) Run(handler EventHandler) error {
	e.handler = handler
	pending := e.pending
	e.pending = nil
	for _, w := range pending {
		e.fireWindowCreated(w)
	}
	err := e.platform.Run()
	if e.err != nil {
		return e.err
	}
	return err
}

func (e *EventLoop) NewWindow(config WindowConfig) (*Window, error) {
	handle, err := e.platform.CreateWindow(config)
	if err != nil {
		return nil, err
	}
	w := &Window{loop: e, handle: handle}
	e.windows[handle.id()] = w
	if e.handler != nil {
		e.fireWindowCreated(w)
	} else {
		e.pending = append(e.pending, w)
	}
	return w, nil
}

func (e *EventLoop) fireWindowCreated(w *Window) {
	if h, ok := e.handler.(WindowCreatedHandler); ok {
		if err := h.OnWindowCreated(e, w); err != nil && e.err == nil {
			e.err = err
			e.Quit()
		}
	}
}

func (e *EventLoop) removeWindow(w *Window) {
	delete(e.windows, w.handle.id())
	for i, p := range e.pending {
		if p == w {
			e.pending = append(e.pending[:i], e.pending[i+1:]...)
			break
		}
	}
}

func (e *EventLoop) Quit() {
	e.platform.PostQuit()
}

func (e *EventLoop) Close() error {
	var firstErr error
	for _, w := range e.windows {
		if err := e.platform.DestroyWindow(w.handle); err != nil && firstErr == nil {
			firstErr = err
		}
		e.removeWindow(w)
	}
	return firstErr
}

func (e *EventLoop) CloseRequested(handle WindowHandle) {
	w, ok := e.windows[handle.id()]
	if !ok {
		return
	}
	if h, ok := e.handler.(CloseRequestedHandler); ok {
		if err := h.OnCloseRequested(e, w); err != nil && e.err == nil {
			e.err = err
			e.Quit()
		}
	}
}

func (e *EventLoop) WindowResized(handle WindowHandle, width, height int) {
	w, ok := e.windows[handle.id()]
	if !ok {
		return
	}
	if h, ok := e.handler.(WindowResizeHandler); ok {
		if err := h.OnWindowResize(e, w, width, height); err != nil && e.err == nil {
			e.err = err
			e.Quit()
		}
	}
}

func (e *EventLoop) WindowDPIChanged(handle WindowHandle, dpi int) {
	w, ok := e.windows[handle.id()]
	if !ok {
		return
	}
	if h, ok := e.handler.(WindowDPIChangedHandler); ok {
		if err := h.OnWindowDPIChanged(e, w, dpi); err != nil && e.err == nil {
			e.err = err
			e.Quit()
		}
	}
}
