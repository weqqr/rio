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
const (
	D3D12_DESCRIPTOR_HEAP_TYPE_CBV_SRV_UAV D3D12_DESCRIPTOR_HEAP_TYPE = 0
	D3D12_DESCRIPTOR_HEAP_TYPE_SAMPLER     D3D12_DESCRIPTOR_HEAP_TYPE = 1
)

type D3D12_DESCRIPTOR_HEAP_FLAGS int32

const (
	D3D12_DESCRIPTOR_HEAP_FLAG_NONE           D3D12_DESCRIPTOR_HEAP_FLAGS = 0
	D3D12_DESCRIPTOR_HEAP_FLAG_SHADER_VISIBLE D3D12_DESCRIPTOR_HEAP_FLAGS = 0x1
)

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

type D3D12_GPU_DESCRIPTOR_HANDLE struct {
	_   structs.HostLayout
	Ptr uint64
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
	Resource    unsafe.Pointer
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

type D3D12_ROOT_PARAMETER_TYPE int32

const D3D12_ROOT_PARAMETER_TYPE_DESCRIPTOR_TABLE D3D12_ROOT_PARAMETER_TYPE = 0

type D3D12_SHADER_VISIBILITY int32

const (
	D3D12_SHADER_VISIBILITY_ALL    D3D12_SHADER_VISIBILITY = 0
	D3D12_SHADER_VISIBILITY_VERTEX D3D12_SHADER_VISIBILITY = 1
	D3D12_SHADER_VISIBILITY_PIXEL  D3D12_SHADER_VISIBILITY = 5
)

type D3D12_DESCRIPTOR_RANGE_TYPE int32

const (
	D3D12_DESCRIPTOR_RANGE_TYPE_SRV     D3D12_DESCRIPTOR_RANGE_TYPE = 0
	D3D12_DESCRIPTOR_RANGE_TYPE_SAMPLER D3D12_DESCRIPTOR_RANGE_TYPE = 3
)

type D3D12_DESCRIPTOR_RANGE struct {
	_                                 structs.HostLayout
	RangeType                         D3D12_DESCRIPTOR_RANGE_TYPE
	NumDescriptors                    uint32
	BaseShaderRegister                uint32
	RegisterSpace                     uint32
	OffsetInDescriptorsFromTableStart uint32
}

type D3D12_ROOT_DESCRIPTOR_TABLE struct {
	_                   structs.HostLayout
	NumDescriptorRanges uint32
	_                   uint32
	DescriptorRanges    *D3D12_DESCRIPTOR_RANGE
}

type D3D12_ROOT_PARAMETER struct {
	_                structs.HostLayout
	ParameterType    D3D12_ROOT_PARAMETER_TYPE
	_                uint32
	DescriptorTable  D3D12_ROOT_DESCRIPTOR_TABLE
	ShaderVisibility D3D12_SHADER_VISIBILITY
	_                uint32
}

type D3D12_TEXTURE_COPY_TYPE int32

const (
	D3D12_TEXTURE_COPY_TYPE_SUBRESOURCE_INDEX D3D12_TEXTURE_COPY_TYPE = 0
	D3D12_TEXTURE_COPY_TYPE_PLACED_FOOTPRINT  D3D12_TEXTURE_COPY_TYPE = 1
)

type D3D12_SUBRESOURCE_FOOTPRINT struct {
	_        structs.HostLayout
	Format   int32
	Width    uint32
	Height   uint32
	Depth    uint32
	RowPitch uint32
}

type D3D12_PLACED_SUBRESOURCE_FOOTPRINT struct {
	_         structs.HostLayout
	Offset    uint64
	Footprint D3D12_SUBRESOURCE_FOOTPRINT
}

type D3D12_TEXTURE_COPY_LOCATION struct {
	_               structs.HostLayout
	Resource        unsafe.Pointer
	Type            D3D12_TEXTURE_COPY_TYPE
	_               uint32
	PlacedFootprint D3D12_PLACED_SUBRESOURCE_FOOTPRINT
}

func (l *D3D12_TEXTURE_COPY_LOCATION) SetSubresourceIndex(index uint32) {
	*(*uint32)(unsafe.Pointer(&l.PlacedFootprint)) = index
}

type D3D12_BOX struct {
	_                   structs.HostLayout
	Left, Top, Front    uint32
	Right, Bottom, Back uint32
}

type D3D12_SRV_DIMENSION int32

const D3D12_SRV_DIMENSION_TEXTURE2D D3D12_SRV_DIMENSION = 4

type D3D12_TEX2D_SRV struct {
	_                   structs.HostLayout
	MostDetailedMip     uint32
	MipLevels           uint32
	PlaneSlice          uint32
	ResourceMinLODClamp float32
}

type D3D12_SHADER_RESOURCE_VIEW_DESC struct {
	_                       structs.HostLayout
	Format                  int32
	ViewDimension           D3D12_SRV_DIMENSION
	Shader4ComponentMapping uint32
	_                       uint32
	Texture2D               D3D12_TEX2D_SRV
	_                       [8]byte
}

type D3D12_TEXTURE_ADDRESS_MODE int32

const (
	D3D12_TEXTURE_ADDRESS_MODE_WRAP   D3D12_TEXTURE_ADDRESS_MODE = 1
	D3D12_TEXTURE_ADDRESS_MODE_MIRROR D3D12_TEXTURE_ADDRESS_MODE = 2
	D3D12_TEXTURE_ADDRESS_MODE_CLAMP  D3D12_TEXTURE_ADDRESS_MODE = 3
)

type D3D12_FILTER int32

const (
	D3D12_FILTER_MIN_MAG_MIP_POINT               D3D12_FILTER = 0x0
	D3D12_FILTER_MIN_MAG_POINT_MIP_LINEAR        D3D12_FILTER = 0x1
	D3D12_FILTER_MIN_POINT_MAG_LINEAR_MIP_POINT  D3D12_FILTER = 0x4
	D3D12_FILTER_MIN_POINT_MAG_MIP_LINEAR        D3D12_FILTER = 0x5
	D3D12_FILTER_MIN_LINEAR_MAG_MIP_POINT        D3D12_FILTER = 0x10
	D3D12_FILTER_MIN_LINEAR_MAG_POINT_MIP_LINEAR D3D12_FILTER = 0x11
	D3D12_FILTER_MIN_MAG_LINEAR_MIP_POINT        D3D12_FILTER = 0x14
	D3D12_FILTER_MIN_MAG_MIP_LINEAR              D3D12_FILTER = 0x15
	D3D12_FILTER_ANISOTROPIC                     D3D12_FILTER = 0x55
)

type D3D12_SAMPLER_DESC struct {
	_              structs.HostLayout
	Filter         D3D12_FILTER
	AddressU       D3D12_TEXTURE_ADDRESS_MODE
	AddressV       D3D12_TEXTURE_ADDRESS_MODE
	AddressW       D3D12_TEXTURE_ADDRESS_MODE
	MipLODBias     float32
	MaxAnisotropy  uint32
	ComparisonFunc D3D12_COMPARISON_FUNC
	BorderColor    [4]float32
	MinLOD         float32
	MaxLOD         float32
}
