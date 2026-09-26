//go:build windows

package d3d12

import (
	"structs"
	"unsafe"

	"github.com/weqqr/rio/pkg/com"
)

var (
	IID_ID3D12Device              = com.IID{Data1: 0x189819F1, Data2: 0x1DB6, Data3: 0x4B57, Data4: [8]byte{0xBE, 0x54, 0x18, 0x21, 0x33, 0x9B, 0x85, 0xF7}}
	IID_ID3D12CommandQueue        = com.IID{Data1: 0x0EC870A6, Data2: 0x5D7E, Data3: 0x4C22, Data4: [8]byte{0x8C, 0xFC, 0x5B, 0xAA, 0xE0, 0x76, 0x16, 0xED}}
	IID_ID3D12CommandAllocator    = com.IID{Data1: 0x6102DEE4, Data2: 0xAF59, Data3: 0x4B09, Data4: [8]byte{0xB9, 0x99, 0xB4, 0x4D, 0x73, 0xF0, 0x9B, 0x24}}
	IID_ID3D12GraphicsCommandList = com.IID{Data1: 0x5B160D0F, Data2: 0xAC1B, Data3: 0x4185, Data4: [8]byte{0x8B, 0xA8, 0xB3, 0xAE, 0x42, 0xA5, 0xA4, 0x55}}
	IID_ID3D12Fence               = com.IID{Data1: 0x0A753DCF, Data2: 0xC4D8, Data3: 0x4B91, Data4: [8]byte{0xAD, 0xF6, 0xBE, 0x5A, 0x60, 0xD9, 0x5A, 0x76}}
	IID_ID3D12DescriptorHeap      = com.IID{Data1: 0x8EFB471D, Data2: 0x616C, Data3: 0x4F49, Data4: [8]byte{0x90, 0xF7, 0x12, 0x7B, 0xB7, 0x63, 0xFA, 0x51}}
	IID_ID3D12Resource            = com.IID{Data1: 0x696442BE, Data2: 0xA72E, Data3: 0x4059, Data4: [8]byte{0xBC, 0x79, 0x5B, 0x5C, 0x98, 0x04, 0x0F, 0xAD}}
)

type D3D_FEATURE_LEVEL int32

const D3D_FEATURE_LEVEL_11_0 D3D_FEATURE_LEVEL = 0xB000

type D3D12_COMMAND_LIST_TYPE int32

const D3D12_COMMAND_LIST_TYPE_DIRECT D3D12_COMMAND_LIST_TYPE = 0

type D3D12_COMMAND_QUEUE_FLAGS int32

const D3D12_COMMAND_QUEUE_FLAG_NONE D3D12_COMMAND_QUEUE_FLAGS = 0

type D3D12_FENCE_FLAGS int32

const D3D12_FENCE_FLAG_NONE D3D12_FENCE_FLAGS = 0

type D3D12_DESCRIPTOR_HEAP_TYPE int32

const D3D12_DESCRIPTOR_HEAP_TYPE_RTV D3D12_DESCRIPTOR_HEAP_TYPE = 2

type D3D12_DESCRIPTOR_HEAP_FLAGS int32

const D3D12_DESCRIPTOR_HEAP_FLAG_NONE D3D12_DESCRIPTOR_HEAP_FLAGS = 0

type D3D12_RESOURCE_STATES int32

const (
	D3D12_RESOURCE_STATE_RENDER_TARGET D3D12_RESOURCE_STATES = 0x4
	D3D12_RESOURCE_STATE_PRESENT       D3D12_RESOURCE_STATES = 0x0
)

type D3D12_RESOURCE_BARRIER_TYPE int32

type D3D12_RESOURCE_BARRIER_FLAGS int32

const (
	D3D12_RESOURCE_BARRIER_TYPE_TRANSITION D3D12_RESOURCE_BARRIER_TYPE  = 0
	D3D12_RESOURCE_BARRIER_FLAG_NONE       D3D12_RESOURCE_BARRIER_FLAGS = 0

	D3D12_RESOURCE_BARRIER_ALL_SUBRESOURCES uint32 = 0xFFFFFFFF
)

type D3D12_CPU_DESCRIPTOR_HANDLE struct {
	_   structs.HostLayout
	Ptr uintptr
}

type D3D12_COMMAND_QUEUE_DESC struct {
	_        structs.HostLayout
	Type     D3D12_COMMAND_LIST_TYPE
	Priority int32
	Flags    D3D12_COMMAND_QUEUE_FLAGS
	NodeMask uint32
}

type D3D12_DESCRIPTOR_HEAP_DESC struct {
	_              structs.HostLayout
	Type           D3D12_DESCRIPTOR_HEAP_TYPE
	NumDescriptors uint32
	Flags          D3D12_DESCRIPTOR_HEAP_FLAGS
	NodeMask       uint32
}

type D3D12_RESOURCE_TRANSITION_BARRIER struct {
	_           structs.HostLayout
	Resource    unsafe.Pointer // ID3D12Resource*
	Subresource uint32
	StateBefore D3D12_RESOURCE_STATES
	StateAfter  D3D12_RESOURCE_STATES
}

type D3D12_RESOURCE_BARRIER struct {
	_          structs.HostLayout
	Type       D3D12_RESOURCE_BARRIER_TYPE
	Flags      D3D12_RESOURCE_BARRIER_FLAGS
	Transition D3D12_RESOURCE_TRANSITION_BARRIER
}
