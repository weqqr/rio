package d3dcompiler

import (
	"syscall"
	"unsafe"

	"github.com/ebitengine/purego"

	"github.com/weqqr/rio/pkg/com"
)

var D3DCompile func(srcData unsafe.Pointer, srcDataSize uintptr, sourceName *byte, defines uintptr, include uintptr, entryPoint *byte, target *byte, flags1 uint32, flags2 uint32, code *com.Object, errorMsgs *com.Object) com.HRESULT

func init() {
	d3dcompiler := syscall.NewLazyDLL("d3dcompiler_47.dll")
	purego.RegisterFunc(&D3DCompile, d3dcompiler.NewProc("D3DCompile").Addr())
}
