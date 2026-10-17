package renderer

import (
	"fmt"
	"image"
	"unsafe"

	"github.com/weqqr/rio/pkg/gpu"
)

var clearColor = gpu.Color{R: 0.1, G: 0.1, B: 0.1, A: 1}

type Renderer struct {
	instance        *gpu.Instance
	surface         *gpu.Surface
	device          *gpu.Device
	queue           *gpu.Queue
	config          gpu.SurfaceConfiguration
	pipeline        *gpu.RenderPipeline
	vertexBuffer    *gpu.Buffer
	texture         *gpu.Texture
	textureView     *gpu.TextureView
	sampler         *gpu.Sampler
	bindGroupLayout *gpu.BindGroupLayout
	pipelineLayout  *gpu.PipelineLayout
	bindGroup       *gpu.BindGroup
}

func New(backend gpu.InstanceBackend, target gpu.SurfaceTarget) (*Renderer, error) {
	instance, err := gpu.New(backend)
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
	if err := r.createTexture(); err != nil {
		_ = r.Close()
		return nil, err
	}
	if err := r.createTriangle(); err != nil {
		_ = r.Close()
		return nil, err
	}
	return r, nil
}

const textureSize = 256

func xorTexture(width, height int) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			i := img.PixOffset(x, y)
			r := byte(x ^ y)
			g := byte((x * 3) ^ (y * 5))
			b := byte((x * 7) ^ (y * 11))
			img.Pix[i+0] = r
			img.Pix[i+1] = g
			img.Pix[i+2] = b
			img.Pix[i+3] = 0xFF
		}
	}
	return img
}

func (r *Renderer) createTexture() error {
	img := xorTexture(textureSize, textureSize)
	texture, err := r.device.CreateTexture(&gpu.TextureDescriptor{
		Label:         "xor texture",
		Size:          gpu.Extent3D{Width: uint32(img.Rect.Dx()), Height: uint32(img.Rect.Dy()), DepthOrArrayLayers: 1},
		MipLevelCount: 1,
		SampleCount:   1,
		Dimension:     gpu.TextureDimension2D,
		Format:        gpu.TextureFormatRGBA8Unorm,
		Usage:         gpu.TextureUsageTextureBinding | gpu.TextureUsageCopyDst,
	})
	if err != nil {
		return err
	}
	r.texture = texture

	view, err := texture.CreateView(nil)
	if err != nil {
		return err
	}
	r.textureView = view

	sampler, err := r.device.CreateSampler(&gpu.SamplerDescriptor{
		Label:        "xor sampler",
		AddressModeU: gpu.AddressModeRepeat,
		AddressModeV: gpu.AddressModeRepeat,
		AddressModeW: gpu.AddressModeRepeat,
		MagFilter:    gpu.FilterModeLinear,
		MinFilter:    gpu.FilterModeLinear,
		MipmapFilter: gpu.MipmapFilterModeLinear,
		LodMaxClamp:  32,
	})
	if err != nil {
		return err
	}
	r.sampler = sampler

	return r.queue.WriteTexture(
		&gpu.ImageCopyTexture{Texture: texture},
		img.Pix,
		&gpu.ImageDataLayout{Offset: 0, BytesPerRow: uint32(img.Stride), RowsPerImage: uint32(img.Rect.Dy())},
		gpu.Extent3D{Width: uint32(img.Rect.Dx()), Height: uint32(img.Rect.Dy()), DepthOrArrayLayers: 1},
	)
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

	bindGroupLayout, err := r.device.CreateBindGroupLayout(&gpu.BindGroupLayoutDescriptor{
		Label: "triangle",
		Entries: []gpu.BindGroupLayoutEntry{
			{Binding: 0, Visibility: gpu.ShaderStageFragment, Type: gpu.BindingTypeTexture},
			{Binding: 1, Visibility: gpu.ShaderStageFragment, Type: gpu.BindingTypeSampler},
		},
	})
	if err != nil {
		return err
	}
	r.bindGroupLayout = bindGroupLayout

	pipelineLayout, err := r.device.CreatePipelineLayout(&gpu.PipelineLayoutDescriptor{
		Label:            "triangle",
		BindGroupLayouts: []*gpu.BindGroupLayout{bindGroupLayout},
	})
	if err != nil {
		return err
	}
	r.pipelineLayout = pipelineLayout

	pipeline, err := r.device.CreateRenderPipeline(&gpu.RenderPipelineDescriptor{
		Label:  "triangle",
		Layout: pipelineLayout,
		Vertex: gpu.VertexState{
			Module:     shader,
			EntryPoint: "vs_main",
			Buffers: []gpu.VertexBufferLayout{{
				ArrayStride: uint64(unsafe.Sizeof(Vertex{})),
				StepMode:    gpu.VertexStepModeVertex,
				Attributes: []gpu.VertexAttribute{
					{Format: gpu.VertexFormatFloat32x2, Offset: 0, ShaderLocation: 0},
					{Format: gpu.VertexFormatFloat32x4, Offset: uint64(unsafe.Offsetof(Vertex{}.Color)), ShaderLocation: 1},
					{Format: gpu.VertexFormatFloat32x2, Offset: uint64(unsafe.Offsetof(Vertex{}.UV)), ShaderLocation: 2},
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

	bindGroup, err := r.device.CreateBindGroup(&gpu.BindGroupDescriptor{
		Label:  "triangle",
		Layout: bindGroupLayout,
		Entries: []gpu.BindGroupEntry{
			{Binding: 0, Resource: gpu.BindingResource{TextureView: r.textureView}},
			{Binding: 1, Resource: gpu.BindingResource{Sampler: r.sampler}},
		},
	})
	if err != nil {
		return err
	}
	r.bindGroup = bindGroup

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
		return fmt.Errorf("renderer: acquire frame: %w", err)
	}
	view, err := frame.Texture().CreateView(nil)
	if err != nil {
		return err
	}
	defer view.Destroy()

	encoder, err := r.device.CreateCommandEncoder()
	if err != nil {
		return fmt.Errorf("renderer: create encoder: %w", err)
	}
	pass, err := encoder.BeginRenderPass(&gpu.RenderPassDescriptor{
		ColorAttachments: []gpu.ColorAttachment{{
			View:       view,
			LoadOp:     gpu.LoadOpClear,
			ClearValue: clearColor,
			StoreOp:    gpu.StoreOpStore,
		}},
	})
	if err != nil {
		return fmt.Errorf("renderer: begin pass: %w", err)
	}
	if err := r.draw(pass); err != nil {
		pass.End()
		return err
	}
	pass.End()

	commands, err := encoder.Finish()
	if err != nil {
		return fmt.Errorf("renderer: finish commands: %w", err)
	}
	if err := r.queue.Submit(commands); err != nil {
		return fmt.Errorf("renderer: submit: %w", err)
	}
	if err := frame.Present(); err != nil {
		return fmt.Errorf("renderer: present: %w", err)
	}
	return nil
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
	if err := pass.SetBindGroup(0, r.bindGroup); err != nil {
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
	if r.bindGroup != nil {
		r.bindGroup.Destroy()
		r.bindGroup = nil
	}
	if r.vertexBuffer != nil {
		r.vertexBuffer.Destroy()
		r.vertexBuffer = nil
	}
	if r.pipeline != nil {
		r.pipeline.Destroy()
		r.pipeline = nil
	}
	if r.pipelineLayout != nil {
		r.pipelineLayout.Destroy()
		r.pipelineLayout = nil
	}
	if r.bindGroupLayout != nil {
		r.bindGroupLayout.Destroy()
		r.bindGroupLayout = nil
	}
	if r.sampler != nil {
		r.sampler.Destroy()
		r.sampler = nil
	}
	if r.textureView != nil {
		r.textureView.Destroy()
		r.textureView = nil
	}
	if r.texture != nil {
		r.texture.Destroy()
		r.texture = nil
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
