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
	return &EventLoop{
		platform: win32Platform{backend: backend},
		windows:  make(map[uintptr]*Window),
	}
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

func (p win32Platform) Run(dispatch Dispatcher) error {
	return p.backend.Run(win32Dispatcher{dispatch})
}

func (p win32Platform) PostQuit() {
	p.backend.PostQuit()
}

type win32Handle struct {
	window *win32.Window
}

func (h win32Handle) id() uintptr { return h.window.ID() }

type win32Dispatcher struct {
	dispatcher Dispatcher
}

func (d win32Dispatcher) CloseRequested(window *win32.Window) {
	d.dispatcher.CloseRequested(win32Handle{window})
}
