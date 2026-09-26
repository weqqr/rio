//go:build windows

package win32

import (
	"fmt"
	"unicode/utf16"
	"unsafe"

	"github.com/ebitengine/purego"

	"github.com/weqqr/rio/pkg/sys/win32"
)

const className = "rio.wev.Window"

type Config struct {
	Title  string
	Width  int
	Height int
}

type Window struct {
	hwnd uintptr
}

func (w *Window) ID() uintptr { return w.hwnd }

type Dispatcher interface {
	CloseRequested(window *Window)
}

type Backend struct {
	instance   uintptr
	className  []uint16
	wndProcPtr uintptr
	dispatcher Dispatcher
}

func NewBackend() (*Backend, error) {
	b := &Backend{}
	b.instance = win32.GetModuleHandleW(nil)
	if b.instance == 0 {
		return nil, fmt.Errorf("wev: GetModuleHandleW failed: %d", win32.GetLastError())
	}
	b.className = utf16.Encode([]rune(className))
	b.className = append(b.className, 0)
	b.wndProcPtr = purego.NewCallback(b.wndProc)

	wcx := win32.WNDCLASSEXW{
		Size:       uint32(unsafe.Sizeof(win32.WNDCLASSEXW{})),
		WndProc:    b.wndProcPtr,
		Instance:   b.instance,
		Cursor:     win32.LoadCursorW(0, win32.IDC_ARROW),
		Background: win32.COLOR_WINDOW + 1,
		ClassName:  &b.className[0],
	}
	if win32.RegisterClassExW(&wcx) == 0 {
		return nil, fmt.Errorf("wev: RegisterClassExW failed: %d", win32.GetLastError())
	}
	return b, nil
}

func (b *Backend) CreateWindow(config Config) (*Window, error) {
	title := utf16.Encode([]rune(config.Title))
	title = append(title, 0)
	width, height := int32(config.Width), int32(config.Height)
	if width == 0 {
		width = 800
	}
	if height == 0 {
		height = 600
	}

	hwnd := win32.CreateWindowExW(
		0,
		&b.className[0],
		&title[0],
		win32.WS_OVERLAPPEDWINDOW,
		win32.CW_USEDEFAULT, 0,
		width, height,
		0, 0, b.instance, 0,
	)
	if hwnd == 0 {
		return nil, fmt.Errorf("wev: CreateWindowExW failed: %d", win32.GetLastError())
	}
	win32.ShowWindow(hwnd, win32.SW_SHOW)
	return &Window{hwnd: hwnd}, nil
}

func (b *Backend) DestroyWindow(window *Window) error {
	if win32.DestroyWindow(window.hwnd) == 0 {
		return fmt.Errorf("wev: DestroyWindow failed: %d", win32.GetLastError())
	}
	return nil
}

func (b *Backend) Run(dispatcher Dispatcher) error {
	b.dispatcher = dispatcher
	for {
		var msg win32.MSG
		r := win32.GetMessageW(&msg, 0, 0, 0)
		if r == 0 {
			return nil
		}
		if r == -1 {
			return fmt.Errorf("wev: GetMessageW failed: %d", win32.GetLastError())
		}
		win32.TranslateMessage(&msg)
		win32.DispatchMessageW(&msg)
	}
}

func (b *Backend) PostQuit() {
	win32.PostQuitMessage(0)
}

func (b *Backend) wndProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case win32.WM_CLOSE:
		if b.dispatcher != nil {
			b.dispatcher.CloseRequested(&Window{hwnd: hwnd})
		}
		return 0
	}
	return win32.DefWindowProcW(hwnd, msg, wParam, lParam)
}
