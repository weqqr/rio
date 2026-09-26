package wev

type EventHandler any

type CloseRequestedHandler interface {
	OnCloseRequested(e *EventLoop, window *Window) error
}

type Platform interface {
	CreateWindow(config WindowConfig) (WindowHandle, error)
	DestroyWindow(handle WindowHandle) error
	Run(dispatch Dispatcher) error
	PostQuit()
}

type Dispatcher interface {
	CloseRequested(handle WindowHandle)
}

type WindowHandle interface {
	id() uintptr
}

type EventLoop struct {
	handler  EventHandler
	platform Platform
	windows  map[uintptr]*Window
	err      error
}

func (e *EventLoop) Run(handler EventHandler) error {
	e.handler = handler
	err := e.platform.Run(e)
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
	return w, nil
}

func (e *EventLoop) Quit() {
	e.platform.PostQuit()
}

func (e *EventLoop) Close() error {
	var firstErr error
	for id, w := range e.windows {
		if err := e.platform.DestroyWindow(w.handle); err != nil && firstErr == nil {
			firstErr = err
		}
		delete(e.windows, id)
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
