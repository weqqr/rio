package renderer

import (
	"unsafe"

	"github.com/weqqr/rio/pkg/gpu"
)

var clearColor = gpu.Color{R: 0.1, G: 0.1, B: 0.1, A: 1}

type Renderer struct {
	instance     *gpu.Instance
	surface      *gpu.Surface
	device       *gpu.Device
	queue        *gpu.Queue
	config       gpu.SurfaceConfiguration
	pipeline     *gpu.RenderPipeline
	vertexBuffer *gpu.Buffer
}

func New(target gpu.SurfaceTarget) (*Renderer, error) {
	instance, err := gpu.New(nil)
	if err != nil {
		return nil, err
	}
	surface, err := instance.CreateSurface(target)
	if err != nil {
		instance.Close()
		return nil, err
	}
	adapter, err := instance.RequestAdapter(&gpu.RequestAdapterOptions{Surface: surface})
	if err != nil {
		surface.Close()
		instance.Close()
		return nil, err
	}
	device, err := adapter.RequestDevice()
	if err != nil {
		surface.Close()
		instance.Close()
		return nil, err
	}

	r := &Renderer{
		instance: instance,
		surface:  surface,
		device:   device,
		queue:    device.Queue(),
		config: gpu.SurfaceConfiguration{
			Width:       target.Width,
			Height:      target.Height,
			Format:      gpu.TextureFormatRGBA8Unorm,
			PresentMode: gpu.PresentModeFifo,
		},
	}
	if err := r.surface.Configure(r.device, r.config); err != nil {
		_ = r.Close()
		return nil, err
	}
	if err := r.createTriangle(); err != nil {
		_ = r.Close()
		return nil, err
	}
	return r, nil
}

func (r *Renderer) createTriangle() error {
	shader, err := r.device.CreateShaderModule(&gpu.ShaderModuleDescriptor{
		Label: "triangle",
		Code:  []byte(triangleShader),
	})
	if err != nil {
		return err
	}
	defer shader.Destroy()

	pipeline, err := r.device.CreateRenderPipeline(&gpu.RenderPipelineDescriptor{
		Label: "triangle",
		Vertex: gpu.VertexState{
			Module:     shader,
			EntryPoint: "vs_main",
			Buffers: []gpu.VertexBufferLayout{{
				ArrayStride: uint64(unsafe.Sizeof(Vertex{})),
				StepMode:    gpu.VertexStepModeVertex,
				Attributes: []gpu.VertexAttribute{
					{Format: gpu.VertexFormatFloat32x2, Offset: 0, ShaderLocation: 0},
					{Format: gpu.VertexFormatFloat32x4, Offset: 8, ShaderLocation: 1},
				},
			}},
		},
		Primitive: gpu.PrimitiveState{
			Topology:  gpu.PrimitiveTopologyTriangleList,
			FrontFace: gpu.FrontFaceCCW,
			CullMode:  gpu.CullModeNone,
		},
		Fragment: &gpu.FragmentState{
			Module:     shader,
			EntryPoint: "ps_main",
			Targets: []gpu.ColorTargetState{{
				Format:    gpu.TextureFormatRGBA8Unorm,
				WriteMask: gpu.ColorWriteMaskAll,
			}},
		},
		Multisample: gpu.MultisampleState{Count: 1, Mask: 0xFFFFFFFF},
	})
	if err != nil {
		return err
	}
	r.pipeline = pipeline

	vertexSize := uint64(unsafe.Sizeof(triangleVertices[0]))
	vertexBytes := uint64(len(triangleVertices)) * vertexSize
	vertexBuffer, err := r.device.CreateBuffer(&gpu.BufferDescriptor{
		Label: "triangle vertices",
		Size:  vertexBytes,
		Usage: gpu.BufferUsageVertex | gpu.BufferUsageCopyDst,
	})
	if err != nil {
		return err
	}
	r.vertexBuffer = vertexBuffer

	data := unsafe.Slice((*byte)(unsafe.Pointer(&triangleVertices[0])), int(vertexBytes))
	if err := r.queue.WriteBuffer(r.vertexBuffer, 0, data); err != nil {
		return err
	}
	return nil
}

func (r *Renderer) Render() error {
	frame, err := r.surface.AcquireFrame()
	if err != nil {
		return err
	}

	encoder, err := r.device.CreateCommandEncoder()
	if err != nil {
		return err
	}
	pass, err := encoder.BeginRenderPass(&gpu.RenderPassDescriptor{
		ColorAttachments: []gpu.ColorAttachment{{
			View:       frame.Texture().CreateView(),
			LoadOp:     gpu.LoadOpClear,
			ClearValue: clearColor,
			StoreOp:    gpu.StoreOpStore,
		}},
	})
	if err != nil {
		return err
	}
	if err := r.draw(pass); err != nil {
		pass.End()
		return err
	}
	pass.End()

	commands, err := encoder.Finish()
	if err != nil {
		return err
	}
	if err := r.queue.Submit(commands); err != nil {
		return err
	}
	return frame.Present()
}

func (r *Renderer) draw(pass *gpu.RenderPass) error {
	width := float32(r.config.Width)
	height := float32(r.config.Height)
	if err := pass.SetViewport(0, 0, width, height, 0, 1); err != nil {
		return err
	}
	if err := pass.SetScissorRect(0, 0, r.config.Width, r.config.Height); err != nil {
		return err
	}
	if err := pass.SetPipeline(r.pipeline); err != nil {
		return err
	}
	if err := pass.SetVertexBuffer(0, r.vertexBuffer); err != nil {
		return err
	}
	return pass.Draw(uint32(len(triangleVertices)), 1, 0, 0)
}

func (r *Renderer) Resize(width, height uint32) error {
	if width == 0 || height == 0 {
		return nil
	}
	r.config.Width = width
	r.config.Height = height
	return r.surface.Configure(r.device, r.config)
}

func (r *Renderer) Close() error {
	if r.vertexBuffer != nil {
		r.vertexBuffer.Destroy()
		r.vertexBuffer = nil
	}
	if r.pipeline != nil {
		r.pipeline.Destroy()
		r.pipeline = nil
	}
	if r.surface != nil {
		r.surface.Close()
		r.surface = nil
	}
	var err error
	if r.device != nil {
		err = r.device.Close()
		r.device = nil
	}
	if r.instance != nil {
		r.instance.Close()
		r.instance = nil
	}
	return err
}
