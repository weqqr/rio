package gpu

import (
	"errors"
	"fmt"
)

type Backend int

const (
	BackendD3D12 Backend = iota
	BackendVulkan
)

func (b Backend) String() string {
	switch b {
	case BackendD3D12:
		return "d3d12"
	case BackendVulkan:
		return "vulkan"
	}
	return fmt.Sprintf("Backend(%d)", int(b))
}

var ErrNoBackend = errors.New("gpu: no graphics backend available")

type InstanceDescriptor struct {
	Backends []Backend
}

type Instance struct {
	backends []instanceBackend
}

func New(desc *InstanceDescriptor) (*Instance, error) {
	requested := availableBackends()
	if desc != nil && len(desc.Backends) > 0 {
		requested = desc.Backends
	}

	instance := &Instance{}
	var errs []error
	for _, backend := range requested {
		impl, err := newBackend(backend)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		instance.backends = append(instance.backends, impl)
	}
	if len(instance.backends) == 0 {
		if len(errs) == 0 {
			return nil, ErrNoBackend
		}
		return nil, errors.Join(errs...)
	}
	return instance, nil
}

type SurfaceTarget struct {
	Handle uintptr
	Width  uint32
	Height uint32
}

func (i *Instance) CreateSurface(target SurfaceTarget) (*Surface, error) {
	var firstErr error
	for _, backend := range i.backends {
		impl, err := backend.CreateSurface(target)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		return &Surface{b: impl}, nil
	}
	if firstErr == nil {
		firstErr = ErrNoBackend
	}
	return nil, firstErr
}

type RequestAdapterOptions struct {
	Surface *Surface
}

type Adapter struct {
	b adapterBackend
}

func (i *Instance) RequestAdapter(opts *RequestAdapterOptions) (*Adapter, error) {
	var firstErr error
	for _, backend := range i.backends {
		impl, err := backend.RequestAdapter(opts)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		return &Adapter{b: impl}, nil
	}
	if firstErr == nil {
		firstErr = ErrNoBackend
	}
	return nil, firstErr
}

func (i *Instance) Close() {
	for _, backend := range i.backends {
		backend.Close()
	}
	i.backends = nil
}

func (a *Adapter) RequestDevice() (*Device, error) {
	impl, err := a.b.RequestDevice()
	if err != nil {
		return nil, err
	}
	return &Device{b: impl, queue: &Queue{b: impl.Queue()}}, nil
}

type Device struct {
	b     deviceBackend
	queue *Queue
}

func (d *Device) Queue() *Queue { return d.queue }

func (d *Device) CreateCommandEncoder() (*CommandEncoder, error) {
	impl, err := d.b.CreateCommandEncoder()
	if err != nil {
		return nil, err
	}
	return &CommandEncoder{b: impl}, nil
}

func (d *Device) Close() error { return d.b.Close() }

type Queue struct {
	b queueBackend
}

func (q *Queue) Submit(cmds ...*CommandBuffer) error { return q.b.Submit(cmds) }

type TextureFormat int

const (
	TextureFormatRGBA8Unorm TextureFormat = iota
	TextureFormatBGRA8Unorm
)

type PresentMode int

const (
	PresentModeFifo PresentMode = iota
	PresentModeImmediate
)

type SurfaceConfiguration struct {
	Width       uint32
	Height      uint32
	Format      TextureFormat
	PresentMode PresentMode
}

type Surface struct {
	b surfaceBackend
}

func (s *Surface) Configure(device *Device, config SurfaceConfiguration) error {
	return s.b.Configure(device.b, config)
}

func (s *Surface) AcquireFrame() (*Frame, error) {
	impl, err := s.b.AcquireFrame()
	if err != nil {
		return nil, err
	}
	return &Frame{b: impl}, nil
}

func (s *Surface) Close() { s.b.Close() }

type Frame struct {
	b frameBackend
}

func (f *Frame) Texture() *Texture { return &Texture{b: f.b.Texture()} }

func (f *Frame) Present() error { return f.b.Present() }

type Texture struct {
	b textureBackend
}

func (t *Texture) CreateView() *TextureView { return &TextureView{b: t.b.CreateView()} }

type TextureView struct {
	b viewBackend
}

type Color struct {
	R, G, B, A float64
}

type LoadOp int

const (
	LoadOpLoad LoadOp = iota
	LoadOpClear
)

type StoreOp int

const (
	StoreOpStore StoreOp = iota
	StoreOpDiscard
)

type ColorAttachment struct {
	View       *TextureView
	LoadOp     LoadOp
	ClearValue Color
	StoreOp    StoreOp
}

type RenderPassDescriptor struct {
	ColorAttachments []ColorAttachment
}

type CommandEncoder struct {
	b encoderBackend
}

func (e *CommandEncoder) BeginRenderPass(desc *RenderPassDescriptor) (*RenderPass, error) {
	if desc == nil || len(desc.ColorAttachments) == 0 {
		return nil, errors.New("gpu: render pass requires a color attachment")
	}
	for i := range desc.ColorAttachments {
		if desc.ColorAttachments[i].View == nil {
			return nil, fmt.Errorf("gpu: color attachment %d has no view", i)
		}
	}
	impl, err := e.b.BeginRenderPass(desc)
	if err != nil {
		return nil, err
	}
	return &RenderPass{b: impl}, nil
}

func (e *CommandEncoder) Finish() (*CommandBuffer, error) {
	impl, err := e.b.Finish()
	if err != nil {
		return nil, err
	}
	return &CommandBuffer{b: impl}, nil
}

type CommandBuffer struct {
	b commandBufferBackend
}

type RenderPass struct {
	b renderPassBackend
}

func (p *RenderPass) SetPipeline(pipeline *RenderPipeline) error {
	if pipeline == nil {
		return errors.New("gpu: cannot set a nil render pipeline")
	}
	return p.b.SetPipeline(pipeline)
}

func (p *RenderPass) SetVertexBuffer(slot uint32, buffer *Buffer) error {
	if buffer == nil {
		return errors.New("gpu: cannot set a nil vertex buffer")
	}
	return p.b.SetVertexBuffer(slot, buffer)
}

func (p *RenderPass) SetViewport(x, y, width, height, minDepth, maxDepth float32) error {
	return p.b.SetViewport(x, y, width, height, minDepth, maxDepth)
}

func (p *RenderPass) SetScissorRect(x, y, width, height uint32) error {
	return p.b.SetScissorRect(x, y, width, height)
}

func (p *RenderPass) Draw(vertexCount, instanceCount, firstVertex, firstInstance uint32) error {
	return p.b.Draw(vertexCount, instanceCount, firstVertex, firstInstance)
}

func (p *RenderPass) End() { p.b.End() }
