//go:build windows

package dxgi

import (
	"unsafe"

	"github.com/weqqr/rio/pkg/com"
)

const slotFactoryCreateSwapChainForHwnd = 15

type IDXGIFactory2 struct{ com.Object }

func (f IDXGIFactory2) CreateSwapChainForHwnd(device com.Object, hwnd uintptr, desc *DXGI_SWAP_CHAIN_DESC1, fullscreenDesc unsafe.Pointer, restrictToOutput com.Object, swapChain *IDXGISwapChain1) com.HRESULT {
	return com.HRESULT(f.Object.Call(slotFactoryCreateSwapChainForHwnd,
		uintptr(device.Ptr()), hwnd,
		uintptr(unsafe.Pointer(desc)), uintptr(fullscreenDesc),
		uintptr(restrictToOutput.Ptr()),
		uintptr(unsafe.Pointer(&swapChain.Object))))
}

const (
	slotSwapChainPresent                   = 8
	slotSwapChainGetBuffer                 = 9
	slotSwapChainResizeBuffers             = 13
	slotSwapChainGetCurrentBackBufferIndex = 36
)

type IDXGISwapChain1 struct{ com.Object }

type IDXGISwapChain3 struct{ com.Object }

func (s IDXGISwapChain3) Present(syncInterval uint32, flags uint32) com.HRESULT {
	return com.HRESULT(s.Object.Call(slotSwapChainPresent, uintptr(syncInterval), uintptr(flags)))
}

func (s IDXGISwapChain3) GetBuffer(buffer uint32, iid *com.IID, surface *com.Object) com.HRESULT {
	return com.HRESULT(s.Object.Call(slotSwapChainGetBuffer,
		uintptr(buffer), uintptr(unsafe.Pointer(iid)), uintptr(unsafe.Pointer(surface))))
}

func (s IDXGISwapChain3) ResizeBuffers(bufferCount uint32, width uint32, height uint32, newFormat DXGI_FORMAT, flags uint32) com.HRESULT {
	return com.HRESULT(s.Object.Call(slotSwapChainResizeBuffers,
		uintptr(bufferCount), uintptr(width), uintptr(height), uintptr(newFormat), uintptr(flags)))
}

func (s IDXGISwapChain3) GetCurrentBackBufferIndex() uint32 {
	return uint32(s.Object.Call(slotSwapChainGetCurrentBackBufferIndex))
}
