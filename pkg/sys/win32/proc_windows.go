//go:build windows

package win32

import (
	"syscall"

	"github.com/ebitengine/purego"
)

var (
	GetModuleHandleW func(name *uint16) uintptr
	GetLastError     func() uint32

	RegisterClassExW func(wcx *WNDCLASSEXW) uint16
	CreateWindowExW  func(exStyle uintptr, className, windowName *uint16, style uintptr, x, y, width, height int32, parent, menu, instance uintptr, param uintptr) uintptr
	DefWindowProcW   func(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr
	DestroyWindow    func(hwnd uintptr) int32
	ShowWindow       func(hwnd uintptr, cmd int32) int32
	GetMessageW      func(msg *MSG, hwnd uintptr, min, max uint32) int32
	TranslateMessage func(msg *MSG) int32
	DispatchMessageW func(msg *MSG) uintptr
	PostQuitMessage  func(exitCode uint32)
	LoadCursorW      func(instance uintptr, name uintptr) uintptr

	GetClientRect                 func(hwnd uintptr, rect *RECT) int32
	GetDpiForWindow               func(hwnd uintptr) uint32
	SetProcessDpiAwarenessContext func(context uintptr) int32
)

func init() {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	user32 := syscall.NewLazyDLL("user32.dll")

	purego.RegisterFunc(&GetModuleHandleW, kernel32.NewProc("GetModuleHandleW").Addr())
	purego.RegisterFunc(&GetLastError, kernel32.NewProc("GetLastError").Addr())

	purego.RegisterFunc(&RegisterClassExW, user32.NewProc("RegisterClassExW").Addr())
	purego.RegisterFunc(&CreateWindowExW, user32.NewProc("CreateWindowExW").Addr())
	purego.RegisterFunc(&DefWindowProcW, user32.NewProc("DefWindowProcW").Addr())
	purego.RegisterFunc(&DestroyWindow, user32.NewProc("DestroyWindow").Addr())
	purego.RegisterFunc(&ShowWindow, user32.NewProc("ShowWindow").Addr())
	purego.RegisterFunc(&GetMessageW, user32.NewProc("GetMessageW").Addr())
	purego.RegisterFunc(&TranslateMessage, user32.NewProc("TranslateMessage").Addr())
	purego.RegisterFunc(&DispatchMessageW, user32.NewProc("DispatchMessageW").Addr())
	purego.RegisterFunc(&PostQuitMessage, user32.NewProc("PostQuitMessage").Addr())
	purego.RegisterFunc(&LoadCursorW, user32.NewProc("LoadCursorW").Addr())
	purego.RegisterFunc(&GetClientRect, user32.NewProc("GetClientRect").Addr())
	purego.RegisterFunc(&GetDpiForWindow, user32.NewProc("GetDpiForWindow").Addr())
	purego.RegisterFunc(&SetProcessDpiAwarenessContext, user32.NewProc("SetProcessDpiAwarenessContext").Addr())
}
