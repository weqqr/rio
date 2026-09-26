//go:build windows

package com

import (
	"fmt"
	"structs"
)

type GUID struct {
	_     structs.HostLayout
	Data1 uint32
	Data2 uint16
	Data3 uint16
	Data4 [8]byte
}

type IID = GUID

type CLSID = GUID

var IID_IUnknown = IID{Data1: 0x00000000, Data2: 0x0000, Data3: 0x0000, Data4: [8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

func (g GUID) String() string {
	return fmt.Sprintf("{%08X-%04X-%04X-%02X%02X-%02X%02X%02X%02X%02X%02X}",
		g.Data1, g.Data2, g.Data3,
		g.Data4[0], g.Data4[1], g.Data4[2], g.Data4[3],
		g.Data4[4], g.Data4[5], g.Data4[6], g.Data4[7])
}

type HRESULT int32

const (
	S_OK          HRESULT = 0
	S_FALSE       HRESULT = 1
	E_NOTIMPL     HRESULT = -2147467263 // 0x80004001
	E_NOINTERFACE HRESULT = -2147467262 // 0x80004002
	E_POINTER     HRESULT = -2147467261 // 0x80004003
	E_ABORT       HRESULT = -2147467260 // 0x80004004
	E_FAIL        HRESULT = -2147467259 // 0x80004005
	E_INVALIDARG  HRESULT = -2147024809 // 0x80070057
)

func (hr HRESULT) Succeeded() bool { return hr >= 0 }

func (hr HRESULT) Failed() bool { return hr < 0 }

func (hr HRESULT) Err() error {
	if hr >= 0 {
		return nil
	}
	return &Error{Code: hr}
}

func (hr HRESULT) String() string {
	name := ""
	switch hr {
	case S_OK:
		name = "S_OK"
	case S_FALSE:
		name = "S_FALSE"
	case E_NOTIMPL:
		name = "E_NOTIMPL"
	case E_NOINTERFACE:
		name = "E_NOINTERFACE"
	case E_POINTER:
		name = "E_POINTER"
	case E_ABORT:
		name = "E_ABORT"
	case E_FAIL:
		name = "E_FAIL"
	case E_INVALIDARG:
		name = "E_INVALIDARG"
	}
	if name == "" {
		return fmt.Sprintf("0x%08X", uint32(hr))
	}
	return fmt.Sprintf("%s (0x%08X)", name, uint32(hr))
}

type Error struct {
	Code HRESULT
}

func (e *Error) Error() string { return "com: " + e.Code.String() }
