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
	Close() error
}

type queueBackend interface {
	Submit(cmds []*CommandBuffer) error
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
	Finish() (commandBufferBackend, error)
}

type renderPassBackend interface {
	End()
}

type commandBufferBackend interface{}
