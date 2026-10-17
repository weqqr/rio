package gpu

import "errors"

type ShaderModuleDescriptor struct {
	Label string
	Code  []byte
}

type ShaderModule struct {
	b shaderModuleBackend
}

func (d *Device) CreateShaderModule(desc *ShaderModuleDescriptor) (*ShaderModule, error) {
	if desc == nil {
		return nil, errors.New("gpu: shader module descriptor is nil")
	}
	impl, err := d.b.CreateShaderModule(desc)
	if err != nil {
		return nil, err
	}
	return &ShaderModule{b: impl}, nil
}

func (m *ShaderModule) Destroy() { m.b.Destroy() }

type VertexFormat int

const (
	VertexFormatFloat32 VertexFormat = iota
	VertexFormatFloat32x2
	VertexFormatFloat32x3
	VertexFormatFloat32x4
)

type VertexStepMode int

const (
	VertexStepModeVertex VertexStepMode = iota
	VertexStepModeInstance
)

type VertexAttribute struct {
	Format         VertexFormat
	Offset         uint64
	ShaderLocation uint32
}

type VertexBufferLayout struct {
	ArrayStride uint64
	StepMode    VertexStepMode
	Attributes  []VertexAttribute
}

type VertexState struct {
	Module     *ShaderModule
	EntryPoint string
	Buffers    []VertexBufferLayout
}

type PrimitiveTopology int

const (
	PrimitiveTopologyTriangleList PrimitiveTopology = iota
)

type FrontFace int

const (
	FrontFaceCCW FrontFace = iota
	FrontFaceCW
)

type CullMode int

const (
	CullModeNone CullMode = iota
	CullModeFront
	CullModeBack
)

type PrimitiveState struct {
	Topology  PrimitiveTopology
	FrontFace FrontFace
	CullMode  CullMode
}

type ColorWriteMask uint32

const (
	ColorWriteMaskRed   ColorWriteMask = 1 << 0
	ColorWriteMaskGreen ColorWriteMask = 1 << 1
	ColorWriteMaskBlue  ColorWriteMask = 1 << 2
	ColorWriteMaskAlpha ColorWriteMask = 1 << 3
	ColorWriteMaskAll                  = ColorWriteMaskRed | ColorWriteMaskGreen | ColorWriteMaskBlue | ColorWriteMaskAlpha
)

type ColorTargetState struct {
	Format    TextureFormat
	WriteMask ColorWriteMask
}

type FragmentState struct {
	Module     *ShaderModule
	EntryPoint string
	Targets    []ColorTargetState
}

type MultisampleState struct {
	Count uint32
	Mask  uint32
}

type RenderPipelineDescriptor struct {
	Label       string
	Vertex      VertexState
	Primitive   PrimitiveState
	Fragment    *FragmentState
	Multisample MultisampleState
}

type RenderPipeline struct {
	b renderPipelineBackend
}

func (d *Device) CreateRenderPipeline(desc *RenderPipelineDescriptor) (*RenderPipeline, error) {
	if desc == nil {
		return nil, errors.New("gpu: render pipeline descriptor is nil")
	}
	impl, err := d.b.CreateRenderPipeline(desc)
	if err != nil {
		return nil, err
	}
	return &RenderPipeline{b: impl}, nil
}

func (p *RenderPipeline) Destroy() { p.b.Destroy() }
