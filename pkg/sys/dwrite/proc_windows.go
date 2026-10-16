package dwrite

import (
	"syscall"

	"github.com/ebitengine/purego"

	"github.com/weqqr/rio/pkg/com"
)

var DWriteCreateFactory func(factoryType DWRITE_FACTORY_TYPE, iid *com.IID, factory *com.Object) com.HRESULT

func init() {
	dwrite := syscall.NewLazyDLL("dwrite.dll")
	purego.RegisterFunc(&DWriteCreateFactory, dwrite.NewProc("DWriteCreateFactory").Addr())
}
