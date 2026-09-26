//go:build windows

package win32

import "structs"

type POINT struct {
	_    structs.HostLayout
	X, Y int32
}

type MSG struct {
	_       structs.HostLayout
	Hwnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      POINT
}

type WNDCLASSEXW struct {
	_          structs.HostLayout
	Size       uint32
	Style      uint32
	WndProc    uintptr
	ClsExtra   int32
	WndExtra   int32
	Instance   uintptr
	Icon       uintptr
	Cursor     uintptr
	Background uintptr
	MenuName   *uint16
	ClassName  *uint16
	IconSm     uintptr
}

const (
	WM_CLOSE = 0x0010

	WS_OVERLAPPEDWINDOW uintptr = 0x00CF0000

	SW_SHOW int32 = 5

	CW_USEDEFAULT int32 = -2147483648

	COLOR_WINDOW uintptr = 5

	IDC_ARROW uintptr = 32512
)
