package renderer

import "github.com/weqqr/rio/pkg/gpu"

var clearColor = gpu.Color{R: 1, A: 1}

type Renderer struct {
	instance *gpu.Instance
	surface  *gpu.Surface
	device   *gpu.Device
	queue    *gpu.Queue
	config   gpu.SurfaceConfiguration
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
	return r, nil
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

func (r *Renderer) Resize(width, height uint32) error {
	if width == 0 || height == 0 {
		return nil
	}
	r.config.Width = width
	r.config.Height = height
	return r.surface.Configure(r.device, r.config)
}

func (r *Renderer) Close() error {
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
