//go:build windows

package com

import (
	"unsafe"

	"github.com/ebitengine/purego"
)

const (
	SlotQueryInterface = 0
	SlotAddRef         = 1
	SlotRelease        = 2
)

type Object struct {
	self unsafe.Pointer
}

func WrapObject(ptr unsafe.Pointer) Object { return Object{self: ptr} }

func (o Object) Ptr() unsafe.Pointer { return o.self }

func (o Object) Valid() bool { return o.self != nil }

func (o Object) Call(slot int, args ...uintptr) uintptr {
	if o.self == nil {
		panic("com: method call on null Object")
	}
	vtable := *(*unsafe.Pointer)(o.self)
	fn := *(*uintptr)(unsafe.Add(vtable, uintptr(slot)*unsafe.Sizeof(uintptr(0))))
	call := make([]uintptr, len(args)+1)
	call[0] = uintptr(o.self)
	copy(call[1:], args)
	r1, _, _ := purego.SyscallN(fn, call...)
	return r1
}

func (o Object) QueryInterface(iid *IID) (Object, error) {
	var obj Object
	hr := HRESULT(o.Call(SlotQueryInterface, uintptr(unsafe.Pointer(iid)), uintptr(unsafe.Pointer(&obj))))
	if hr.Failed() {
		return Object{}, hr.Err()
	}
	return obj, nil
}

func (o Object) AddRef() uint32 {
	return uint32(o.Call(SlotAddRef))
}

func (o Object) Release() uint32 {
	return uint32(o.Call(SlotRelease))
}
