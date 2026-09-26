//go:build windows

package com

import (
	"syscall"

	"github.com/ebitengine/purego"
)

const (
	COINIT_APARTMENTTHREADED = 0x2
	COINIT_MULTITHREADED     = 0x0
)

var (
	CoInitializeEx func(reserved uintptr, initModel uint32) HRESULT
	CoUninitialize func()
)

func init() {
	ole32 := syscall.NewLazyDLL("ole32.dll")
	purego.RegisterFunc(&CoInitializeEx, ole32.NewProc("CoInitializeEx").Addr())
	purego.RegisterFunc(&CoUninitialize, ole32.NewProc("CoUninitialize").Addr())
}
