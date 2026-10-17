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

type Instance struct {
	backends []InstanceBackend
}

func New(backends ...InstanceBackend) (*Instance, error) {
	if len(backends) == 0 {
		return nil, ErrNoBackend
	}
	for i, backend := range backends {
		if backend == nil {
			return nil, fmt.Errorf("gpu: backend %d is nil", i)
		}
	}
	return &Instance{backends: backends}, nil
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
	b AdapterBackend
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
	b     DeviceBackend
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
	b QueueBackend
}

func (q *Queue) Submit(cmds ...*CommandBuffer) error {
	impls := make([]CommandBufferBackend, len(cmds))
	for i, cmd := range cmds {
		impls[i] = cmd.b
	}
	return q.b.Submit(impls)
}

type TextureFormat int

const (
	TextureFormatUndefined TextureFormat = iota
	TextureFormatRGBA8Unorm
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
	b SurfaceBackend
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

func (s *Surface) Backend() SurfaceBackend { return s.b }

type Frame struct {
	b FrameBackend
}

func (f *Frame) Texture() *Texture { return &Texture{b: f.b.Texture()} }

func (f *Frame) Present() error { return f.b.Present() }

type TextureView struct {
	b ViewBackend
}

func (v *TextureView) Destroy() { v.b.Destroy() }

func (v *TextureView) Backend() ViewBackend { return v.b }

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
	b EncoderBackend
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
	b CommandBufferBackend
}

type RenderPass struct {
	b RenderPassBackend
}

func (p *RenderPass) SetPipeline(pipeline *RenderPipeline) error {
	if pipeline == nil {
		return errors.New("gpu: cannot set a nil render pipeline")
	}
	return p.b.SetPipeline(pipeline.b)
}

func (p *RenderPass) SetVertexBuffer(slot uint32, buffer *Buffer) error {
	if buffer == nil {
		return errors.New("gpu: cannot set a nil vertex buffer")
	}
	return p.b.SetVertexBuffer(slot, buffer.b)
}

func (p *RenderPass) SetBindGroup(index uint32, group *BindGroup) error {
	if group == nil {
		return errors.New("gpu: cannot set a nil bind group")
	}
	return p.b.SetBindGroup(index, group.b)
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
