//go:build windows

package wev

import (
	"runtime"

	"github.com/weqqr/rio/pkg/wev/win32"
)

func NewEventLoop() *EventLoop {
	runtime.LockOSThread()
	backend, err := win32.NewBackend()
	if err != nil {
		panic("wev: " + err.Error())
	}
	e := &EventLoop{
		platform: win32Platform{backend: backend},
		windows:  make(map[uintptr]*Window),
	}
	backend.SetDispatcher(win32Dispatcher{dispatcher: e})
	return e
}

type win32Platform struct {
	backend *win32.Backend
}

func (p win32Platform) CreateWindow(config WindowConfig) (WindowHandle, error) {
	window, err := p.backend.CreateWindow(win32.Config{
		Title:  config.Title,
		Width:  config.Width,
		Height: config.Height,
	})
	if err != nil {
		return nil, err
	}
	return win32Handle{window}, nil
}

func (p win32Platform) DestroyWindow(handle WindowHandle) error {
	return p.backend.DestroyWindow(handle.(win32Handle).window)
}

func (p win32Platform) Run() error {
	return p.backend.Run()
}

func (p win32Platform) PostQuit() {
	p.backend.PostQuit()
}

type win32Handle struct {
	window *win32.Window
}

func (h win32Handle) rawHandle() uintptr { return h.window.HWND() }
func (h win32Handle) id() uintptr        { return h.window.ID() }

func (h win32Handle) size() (int, int) { return h.window.Size() }

func (h win32Handle) dpi() int { return h.window.DPI() }

type win32Dispatcher struct {
	dispatcher Dispatcher
}

func (d win32Dispatcher) CloseRequested(window *win32.Window) {
	d.dispatcher.CloseRequested(win32Handle{window})
}

func (d win32Dispatcher) Resized(window *win32.Window, width, height int) {
	d.dispatcher.WindowResized(win32Handle{window}, width, height)
}

func (d win32Dispatcher) DPIChanged(window *win32.Window, dpi int) {
	d.dispatcher.WindowDPIChanged(win32Handle{window}, dpi)
}
