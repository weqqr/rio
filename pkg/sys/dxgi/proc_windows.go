//go:build windows

package dxgi

import (
	"syscall"

	"github.com/ebitengine/purego"

	"github.com/weqqr/rio/pkg/com"
)

var CreateDXGIFactory2 func(flags uint32, riid *com.IID, ppFactory *com.Object) com.HRESULT

func init() {
	dxgi := syscall.NewLazyDLL("dxgi.dll")
	purego.RegisterFunc(&CreateDXGIFactory2, dxgi.NewProc("CreateDXGIFactory2").Addr())
}
