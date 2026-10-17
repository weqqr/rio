package d3d12

import (
	"structs"
	"unsafe"

	"github.com/weqqr/rio/pkg/sys/dxgi"
)

type D3D12_HEAP_TYPE int32

const (
	D3D12_HEAP_TYPE_DEFAULT  D3D12_HEAP_TYPE = 1
	D3D12_HEAP_TYPE_UPLOAD   D3D12_HEAP_TYPE = 2
	D3D12_HEAP_TYPE_READBACK D3D12_HEAP_TYPE = 3
)

type D3D12_CPU_PAGE_PROPERTY int32

const D3D12_CPU_PAGE_PROPERTY_UNKNOWN D3D12_CPU_PAGE_PROPERTY = 0

type D3D12_MEMORY_POOL int32

const D3D12_MEMORY_POOL_UNKNOWN D3D12_MEMORY_POOL = 0

type D3D12_HEAP_FLAGS int32

const D3D12_HEAP_FLAG_NONE D3D12_HEAP_FLAGS = 0

type D3D12_RESOURCE_DIMENSION int32

const D3D12_RESOURCE_DIMENSION_BUFFER D3D12_RESOURCE_DIMENSION = 1

type D3D12_TEXTURE_LAYOUT int32

const (
	D3D12_TEXTURE_LAYOUT_UNKNOWN   D3D12_TEXTURE_LAYOUT = 0
	D3D12_TEXTURE_LAYOUT_ROW_MAJOR D3D12_TEXTURE_LAYOUT = 1
)

type D3D12_RESOURCE_FLAGS uint32

const D3D12_RESOURCE_FLAG_NONE D3D12_RESOURCE_FLAGS = 0

const (
	D3D12_RESOURCE_STATE_COMMON                     D3D12_RESOURCE_STATES = 0x0
	D3D12_RESOURCE_STATE_VERTEX_AND_CONSTANT_BUFFER D3D12_RESOURCE_STATES = 0x1
	D3D12_RESOURCE_STATE_INDEX_BUFFER               D3D12_RESOURCE_STATES = 0x2
	D3D12_RESOURCE_STATE_COPY_DEST                  D3D12_RESOURCE_STATES = 0x400
	D3D12_RESOURCE_STATE_COPY_SOURCE                D3D12_RESOURCE_STATES = 0x800
	D3D12_RESOURCE_STATE_GENERIC_READ               D3D12_RESOURCE_STATES = 0x1 | 0x2 | 0x40 | 0x80 | 0x200 | 0x800
)

type D3D12_HEAP_PROPERTIES struct {
	_                    structs.HostLayout
	Type                 D3D12_HEAP_TYPE
	CPUPageProperty      D3D12_CPU_PAGE_PROPERTY
	MemoryPoolPreference D3D12_MEMORY_POOL
	CreationNodeMask     uint32
	VisibleNodeMask      uint32
}

type D3D12_RESOURCE_DESC struct {
	_                structs.HostLayout
	Dimension        D3D12_RESOURCE_DIMENSION
	Alignment        uint64
	Width            uint64
	Height           uint32
	DepthOrArraySize uint16
	MipLevels        uint16
	Format           dxgi.DXGI_FORMAT
	SampleDesc       dxgi.DXGI_SAMPLE_DESC
	Layout           D3D12_TEXTURE_LAYOUT
	Flags            D3D12_RESOURCE_FLAGS
}

type D3D12_RANGE struct {
	_     structs.HostLayout
	Begin uintptr
	End   uintptr
}

type D3D12_GPU_VIRTUAL_ADDRESS uint64

type D3D12_VERTEX_BUFFER_VIEW struct {
	_              structs.HostLayout
	BufferLocation D3D12_GPU_VIRTUAL_ADDRESS
	SizeInBytes    uint32
	StrideInBytes  uint32
}

type D3D12_VIEWPORT struct {
	_        structs.HostLayout
	TopLeftX float32
	TopLeftY float32
	Width    float32
	Height   float32
	MinDepth float32
	MaxDepth float32
}

type D3D12_RECT struct {
	_                        structs.HostLayout
	Left, Top, Right, Bottom int32
}

type D3D12_ROOT_SIGNATURE_FLAGS uint32

const D3D12_ROOT_SIGNATURE_FLAG_ALLOW_INPUT_ASSEMBLER_INPUT_LAYOUT D3D12_ROOT_SIGNATURE_FLAGS = 0x1

type D3D_ROOT_SIGNATURE_VERSION int32

const D3D_ROOT_SIGNATURE_VERSION_1 D3D_ROOT_SIGNATURE_VERSION = 0x1

type D3D12_ROOT_SIGNATURE_DESC struct {
	_                 structs.HostLayout
	NumParameters     uint32
	Parameters        unsafe.Pointer
	NumStaticSamplers uint32
	StaticSamplers    unsafe.Pointer
	Flags             D3D12_ROOT_SIGNATURE_FLAGS
}

type D3D12_INPUT_CLASSIFICATION int32

const (
	D3D12_INPUT_CLASSIFICATION_PER_VERTEX_DATA   D3D12_INPUT_CLASSIFICATION = 0
	D3D12_INPUT_CLASSIFICATION_PER_INSTANCE_DATA D3D12_INPUT_CLASSIFICATION = 1
)

type D3D12_INPUT_ELEMENT_DESC struct {
	_                    structs.HostLayout
	SemanticName         *byte
	SemanticIndex        uint32
	Format               dxgi.DXGI_FORMAT
	InputSlot            uint32
	AlignedByteOffset    uint32
	InputSlotClass       D3D12_INPUT_CLASSIFICATION
	InstanceDataStepRate uint32
}

type D3D12_INPUT_LAYOUT_DESC struct {
	_                 structs.HostLayout
	InputElementDescs *D3D12_INPUT_ELEMENT_DESC
	NumElements       uint32
}

type D3D12_SHADER_BYTECODE struct {
	_          structs.HostLayout
	Bytecode   unsafe.Pointer
	ByteLength uintptr
}

type D3D12_FILL_MODE int32

const D3D12_FILL_MODE_SOLID D3D12_FILL_MODE = 3

type D3D12_CULL_MODE int32

const (
	D3D12_CULL_MODE_NONE  D3D12_CULL_MODE = 1
	D3D12_CULL_MODE_FRONT D3D12_CULL_MODE = 2
	D3D12_CULL_MODE_BACK  D3D12_CULL_MODE = 3
)

type D3D12_CONSERVATIVE_RASTERIZATION_MODE int32

const D3D12_CONSERVATIVE_RASTERIZATION_MODE_OFF D3D12_CONSERVATIVE_RASTERIZATION_MODE = 0

type D3D12_DEPTH_WRITE_MASK int32

const D3D12_DEPTH_WRITE_MASK_ZERO D3D12_DEPTH_WRITE_MASK = 0

type D3D12_COMPARISON_FUNC int32

const D3D12_COMPARISON_FUNC_LESS D3D12_COMPARISON_FUNC = 2

type D3D12_STENCIL_OP int32

const D3D12_STENCIL_OP_KEEP D3D12_STENCIL_OP = 1

type D3D12_LOGIC_OP int32

const (
	D3D12_LOGIC_OP_CLEAR D3D12_LOGIC_OP = 0
	D3D12_LOGIC_OP_NOOP  D3D12_LOGIC_OP = 4
)

type D3D12_BLEND int32

const (
	D3D12_BLEND_ZERO D3D12_BLEND = 1
	D3D12_BLEND_ONE  D3D12_BLEND = 2
)

type D3D12_BLEND_OP int32

const D3D12_BLEND_OP_ADD D3D12_BLEND_OP = 1

type D3D12_DEPTH_STENCILOP_DESC struct {
	_                  structs.HostLayout
	StencilFailOp      D3D12_STENCIL_OP
	StencilDepthFailOp D3D12_STENCIL_OP
	StencilPassOp      D3D12_STENCIL_OP
	StencilFunc        D3D12_COMPARISON_FUNC
}

type D3D12_DEPTH_STENCIL_DESC struct {
	_                structs.HostLayout
	DepthEnable      uint32
	DepthWriteMask   D3D12_DEPTH_WRITE_MASK
	DepthFunc        D3D12_COMPARISON_FUNC
	StencilEnable    uint32
	StencilReadMask  uint8
	StencilWriteMask uint8
	FrontFace        D3D12_DEPTH_STENCILOP_DESC
	BackFace         D3D12_DEPTH_STENCILOP_DESC
}

type D3D12_RENDER_TARGET_BLEND_DESC struct {
	_                     structs.HostLayout
	BlendEnable           uint32
	LogicOpEnable         uint32
	SrcBlend              D3D12_BLEND
	DestBlend             D3D12_BLEND
	BlendOp               D3D12_BLEND_OP
	SrcBlendAlpha         D3D12_BLEND
	DestBlendAlpha        D3D12_BLEND
	BlendOpAlpha          D3D12_BLEND_OP
	LogicOp               D3D12_LOGIC_OP
	RenderTargetWriteMask uint8
}

type D3D12_BLEND_DESC struct {
	_                      structs.HostLayout
	AlphaToCoverageEnable  uint32
	IndependentBlendEnable uint32
	RenderTarget           [8]D3D12_RENDER_TARGET_BLEND_DESC
}

type D3D12_RASTERIZER_DESC struct {
	_                     structs.HostLayout
	FillMode              D3D12_FILL_MODE
	CullMode              D3D12_CULL_MODE
	FrontCounterClockwise uint32
	DepthBias             int32
	DepthBiasClamp        float32
	SlopeScaledDepthBias  float32
	DepthClipEnable       uint32
	MultisampleEnable     uint32
	AntialiasedLineEnable uint32
	ForcedSampleCount     uint32
	ConservativeRaster    D3D12_CONSERVATIVE_RASTERIZATION_MODE
}

type D3D12_SO_DECLARATION_ENTRY struct {
	_              structs.HostLayout
	Stream         uint32
	SemanticName   *byte
	SemanticIndex  uint32
	StartComponent uint8
	ComponentCount uint8
	OutputSlot     uint8
}

type D3D12_STREAM_OUTPUT_DESC struct {
	_                structs.HostLayout
	SODeclaration    *D3D12_SO_DECLARATION_ENTRY
	NumEntries       uint32
	BufferStrides    *uint32
	NumStrides       uint32
	RasterizedStream uint32
}

type D3D12_INDEX_BUFFER_STRIP_CUT_VALUE int32

const D3D12_INDEX_BUFFER_STRIP_CUT_VALUE_DISABLED D3D12_INDEX_BUFFER_STRIP_CUT_VALUE = 0

type D3D12_PRIMITIVE_TOPOLOGY_TYPE int32

const D3D12_PRIMITIVE_TOPOLOGY_TYPE_TRIANGLE D3D12_PRIMITIVE_TOPOLOGY_TYPE = 3

type D3D_PRIMITIVE_TOPOLOGY int32

const D3D_PRIMITIVE_TOPOLOGY_TRIANGLELIST D3D_PRIMITIVE_TOPOLOGY = 4

type D3D12_CACHED_PIPELINE_STATE struct {
	_                     structs.HostLayout
	CachedBlob            unsafe.Pointer
	CachedBlobSizeInBytes uintptr
}

type D3D12_PIPELINE_STATE_FLAGS uint32

const D3D12_PIPELINE_STATE_FLAG_NONE D3D12_PIPELINE_STATE_FLAGS = 0

type D3D12_GRAPHICS_PIPELINE_STATE_DESC struct {
	_                     structs.HostLayout
	RootSignature         unsafe.Pointer
	VS                    D3D12_SHADER_BYTECODE
	PS                    D3D12_SHADER_BYTECODE
	DS                    D3D12_SHADER_BYTECODE
	HS                    D3D12_SHADER_BYTECODE
	GS                    D3D12_SHADER_BYTECODE
	StreamOutput          D3D12_STREAM_OUTPUT_DESC
	BlendState            D3D12_BLEND_DESC
	SampleMask            uint32
	RasterizerState       D3D12_RASTERIZER_DESC
	DepthStencilState     D3D12_DEPTH_STENCIL_DESC
	InputLayout           D3D12_INPUT_LAYOUT_DESC
	IBStripCutValue       D3D12_INDEX_BUFFER_STRIP_CUT_VALUE
	PrimitiveTopologyType D3D12_PRIMITIVE_TOPOLOGY_TYPE
	NumRenderTargets      uint32
	RTVFormats            [8]dxgi.DXGI_FORMAT
	DSVFormat             dxgi.DXGI_FORMAT
	SampleDesc            dxgi.DXGI_SAMPLE_DESC
	NodeMask              uint32
	CachedPSO             D3D12_CACHED_PIPELINE_STATE
	Flags                 D3D12_PIPELINE_STATE_FLAGS
}
