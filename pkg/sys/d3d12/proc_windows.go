//go:build windows

package d3d12

import (
	"syscall"

	"github.com/ebitengine/purego"

	"github.com/weqqr/rio/pkg/com"
)

var D3D12CreateDevice func(adapter uintptr, minimumFeatureLevel D3D_FEATURE_LEVEL, riid *com.IID, ppDevice *com.Object) com.HRESULT

func init() {
	d3d12 := syscall.NewLazyDLL("d3d12.dll")
	purego.RegisterFunc(&D3D12CreateDevice, d3d12.NewProc("D3D12CreateDevice").Addr())
}
