package gpu

type InstanceBackend interface {
	Backend() Backend
	CreateSurface(target SurfaceTarget) (SurfaceBackend, error)
	RequestAdapter(opts *RequestAdapterOptions) (AdapterBackend, error)
	Close()
}

type AdapterBackend interface {
	RequestDevice() (DeviceBackend, error)
}

type DeviceBackend interface {
	Queue() QueueBackend
	CreateCommandEncoder() (EncoderBackend, error)
	CreateBuffer(desc *BufferDescriptor) (BufferBackend, error)
	CreateTexture(desc *TextureDescriptor) (TextureBackend, error)
	CreateSampler(desc *SamplerDescriptor) (SamplerBackend, error)
	CreateShaderModule(desc *ShaderModuleDescriptor) (ShaderModuleBackend, error)
	CreateBindGroupLayout(desc *BindGroupLayoutDescriptor) (BindGroupLayoutBackend, error)
	CreateBindGroup(desc *BindGroupDescriptor) (BindGroupBackend, error)
	CreatePipelineLayout(desc *PipelineLayoutDescriptor) (PipelineLayoutBackend, error)
	CreateRenderPipeline(desc *RenderPipelineDescriptor) (RenderPipelineBackend, error)
	Close() error
}

type QueueBackend interface {
	Submit(cmds []CommandBufferBackend) error
	WriteBuffer(buffer BufferBackend, offset uint64, data []byte) error
	WriteTexture(dst *ImageCopyTexture, data []byte, layout *ImageDataLayout, size Extent3D) error
}

type SurfaceBackend interface {
	Configure(device DeviceBackend, config SurfaceConfiguration) error
	AcquireFrame() (FrameBackend, error)
	Close()
}

type FrameBackend interface {
	Texture() TextureBackend
	Present() error
}

type TextureBackend interface {
	CreateView(desc *TextureViewDescriptor) (ViewBackend, error)
	Destroy()
}

type ViewBackend interface {
	Destroy()
}

type SamplerBackend interface {
	Destroy()
}

type BindGroupLayoutBackend interface {
	Destroy()
}

type BindGroupBackend interface {
	Destroy()
}

type PipelineLayoutBackend interface {
	Destroy()
}

type EncoderBackend interface {
	BeginRenderPass(desc *RenderPassDescriptor) (RenderPassBackend, error)
	CopyBufferToBuffer(src BufferBackend, srcOffset uint64, dst BufferBackend, dstOffset uint64, size uint64) error
	Finish() (CommandBufferBackend, error)
}

type RenderPassBackend interface {
	SetPipeline(pipeline RenderPipelineBackend) error
	SetVertexBuffer(slot uint32, buffer BufferBackend) error
	SetBindGroup(index uint32, group BindGroupBackend) error
	SetViewport(x, y, width, height, minDepth, maxDepth float32) error
	SetScissorRect(x, y, width, height uint32) error
	Draw(vertexCount, instanceCount, firstVertex, firstInstance uint32) error
	End()
}

type CommandBufferBackend interface{}

type BufferBackend interface {
	Size() uint64
	Usage() BufferUsage
	MapAsync(mode MapMode, offset, size uint64) error
	GetMappedRange(offset, size uint64) ([]byte, error)
	Unmap()
	Destroy()
}

type ShaderModuleBackend interface {
	Code() []byte
	Destroy()
}

type RenderPipelineBackend interface {
	Destroy()
}
