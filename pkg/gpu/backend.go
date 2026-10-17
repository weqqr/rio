package gpu

type instanceBackend interface {
	Backend() Backend
	CreateSurface(target SurfaceTarget) (surfaceBackend, error)
	RequestAdapter(opts *RequestAdapterOptions) (adapterBackend, error)
	Close()
}

type adapterBackend interface {
	RequestDevice() (deviceBackend, error)
}

type deviceBackend interface {
	Queue() queueBackend
	CreateCommandEncoder() (encoderBackend, error)
	CreateBuffer(desc *BufferDescriptor) (bufferBackend, error)
	CreateShaderModule(desc *ShaderModuleDescriptor) (shaderModuleBackend, error)
	CreateRenderPipeline(desc *RenderPipelineDescriptor) (renderPipelineBackend, error)
	Close() error
}

type queueBackend interface {
	Submit(cmds []*CommandBuffer) error
	WriteBuffer(buffer *Buffer, offset uint64, data []byte) error
}

type surfaceBackend interface {
	Configure(device deviceBackend, config SurfaceConfiguration) error
	AcquireFrame() (frameBackend, error)
	Close()
}

type frameBackend interface {
	Texture() textureBackend
	Present() error
}

type textureBackend interface {
	CreateView() viewBackend
}

type viewBackend interface{}

type encoderBackend interface {
	BeginRenderPass(desc *RenderPassDescriptor) (renderPassBackend, error)
	CopyBufferToBuffer(src *Buffer, srcOffset uint64, dst *Buffer, dstOffset uint64, size uint64) error
	Finish() (commandBufferBackend, error)
}

type renderPassBackend interface {
	SetPipeline(pipeline *RenderPipeline) error
	SetVertexBuffer(slot uint32, buffer *Buffer) error
	SetViewport(x, y, width, height, minDepth, maxDepth float32) error
	SetScissorRect(x, y, width, height uint32) error
	Draw(vertexCount, instanceCount, firstVertex, firstInstance uint32) error
	End()
}

type commandBufferBackend interface{}

type bufferBackend interface {
	Size() uint64
	Usage() BufferUsage
	MapAsync(mode MapMode, offset, size uint64) error
	GetMappedRange(offset, size uint64) ([]byte, error)
	Unmap()
	Destroy()
}

type shaderModuleBackend interface {
	Code() []byte
	Destroy()
}

type renderPipelineBackend interface {
	Destroy()
}
