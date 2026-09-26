//go:build windows

package dxgi

import (
	"structs"

	"github.com/weqqr/rio/pkg/com"
)

var (
	IID_IDXGIFactory2   = com.IID{Data1: 0x50C83A1C, Data2: 0xE072, Data3: 0x4C48, Data4: [8]byte{0x87, 0xB0, 0x36, 0x30, 0xFA, 0x36, 0xA6, 0xD0}}
	IID_IDXGISwapChain1 = com.IID{Data1: 0x790A45F7, Data2: 0x0D42, Data3: 0x4876, Data4: [8]byte{0x98, 0x3A, 0x0A, 0x55, 0xCF, 0xE6, 0xF4, 0xAA}}
	IID_IDXGISwapChain3 = com.IID{Data1: 0x94D99BDB, Data2: 0xF1F8, Data3: 0x4AB0, Data4: [8]byte{0xB2, 0x36, 0x7D, 0xA0, 0x17, 0x0E, 0xDA, 0xB1}}
)

type DXGI_FORMAT int32

const (
	DXGI_FORMAT_UNKNOWN        DXGI_FORMAT = 0
	DXGI_FORMAT_R8G8B8A8_UNORM DXGI_FORMAT = 28
)

type DXGI_USAGE uint32

const DXGI_USAGE_RENDER_TARGET_OUTPUT DXGI_USAGE = 0x20

type DXGI_SCALING int32

const DXGI_SCALING_NONE DXGI_SCALING = 1

type DXGI_SWAP_EFFECT int32

const DXGI_SWAP_EFFECT_FLIP_DISCARD DXGI_SWAP_EFFECT = 4

type DXGI_ALPHA_MODE int32

const DXGI_ALPHA_MODE_UNSPECIFIED DXGI_ALPHA_MODE = 0

type DXGI_SAMPLE_DESC struct {
	_       structs.HostLayout
	Count   uint32
	Quality uint32
}

type DXGI_SWAP_CHAIN_DESC1 struct {
	_           structs.HostLayout
	Width       uint32
	Height      uint32
	Format      DXGI_FORMAT
	Stereo      uint32 // BOOL
	SampleDesc  DXGI_SAMPLE_DESC
	BufferUsage DXGI_USAGE
	BufferCount uint32
	Scaling     DXGI_SCALING
	SwapEffect  DXGI_SWAP_EFFECT
	AlphaMode   DXGI_ALPHA_MODE
	Flags       uint32
}
