package d3d12backend

import (
	"errors"
	"fmt"

	"github.com/weqqr/rio/pkg/com"
	"github.com/weqqr/rio/pkg/gpu"
	"github.com/weqqr/rio/pkg/sys/d3d12"
	"github.com/weqqr/rio/pkg/sys/dxgi"
)

type d3d12Surface struct {
	factory dxgi.IDXGIFactory2
	target  gpu.SurfaceTarget

	swapChain dxgi.IDXGISwapChain3
	config    gpu.SurfaceConfiguration
	format    dxgi.DXGI_FORMAT

	rtvHeap     d3d12.ID3D12DescriptorHeap
	rtvStride   uint32
	backBuffers [frameCount]com.Object
	views       [frameCount]*d3d12View
}

func (s *d3d12Surface) Configure(device gpu.DeviceBackend, config gpu.SurfaceConfiguration) error {
	d3d12Device, ok := device.(*d3d12Device)
	if !ok {
		return errors.New("gpu: d3d12 surface cannot be configured with a device from another backend")
	}
	if config.Width == 0 || config.Height == 0 {
		return errors.New("gpu: surface configuration needs a non-zero size")
	}
	format, err := dxgiFormat(config.Format)
	if err != nil {
		return err
	}

	if s.swapChain.Valid() {
		if config.Width == s.config.Width && config.Height == s.config.Height && format == s.format {
			s.config = config
			return nil
		}
		s.releaseBackBuffers()
		if hr := s.swapChain.ResizeBuffers(0, config.Width, config.Height, dxgi.DXGI_FORMAT_UNKNOWN, 0); hr.Failed() {
			return hr.Err()
		}
	} else {
		desc := dxgi.DXGI_SWAP_CHAIN_DESC1{
			Width:       config.Width,
			Height:      config.Height,
			Format:      format,
			SampleDesc:  dxgi.DXGI_SAMPLE_DESC{Count: 1},
			BufferUsage: dxgi.DXGI_USAGE_RENDER_TARGET_OUTPUT,
			BufferCount: frameCount,
			Scaling:     dxgi.DXGI_SCALING_NONE,
			SwapEffect:  dxgi.DXGI_SWAP_EFFECT_FLIP_DISCARD,
			AlphaMode:   dxgi.DXGI_ALPHA_MODE_UNSPECIFIED,
		}
		var swapChain1 dxgi.IDXGISwapChain1
		if hr := s.factory.CreateSwapChainForHwnd(d3d12Device.queue.Object, s.target.Handle, &desc, nil, com.Object{}, &swapChain1); hr.Failed() {
			return hr.Err()
		}
		object, err := swapChain1.QueryInterface(&dxgi.IID_IDXGISwapChain3)
		swapChain1.Release()
		if err != nil {
			return err
		}
		s.swapChain = dxgi.IDXGISwapChain3{Object: object}

		heapDesc := d3d12.D3D12_DESCRIPTOR_HEAP_DESC{
			Type:           d3d12.D3D12_DESCRIPTOR_HEAP_TYPE_RTV,
			NumDescriptors: frameCount,
		}
		if hr := d3d12Device.device.CreateDescriptorHeap(&heapDesc, &s.rtvHeap); hr.Failed() {
			return hr.Err()
		}
		s.rtvStride = d3d12Device.device.GetDescriptorHandleIncrementSize(d3d12.D3D12_DESCRIPTOR_HEAP_TYPE_RTV)
	}

	s.config = config
	s.format = format
	return s.createRenderTargets(d3d12Device)
}

func (s *d3d12Surface) createRenderTargets(device *d3d12Device) error {
	handle := s.rtvHeap.GetCPUDescriptorHandleForHeapStart()
	for i := range s.backBuffers {
		if s.backBuffers[i].Valid() {
			s.backBuffers[i].Release()
		}
		if hr := s.swapChain.GetBuffer(uint32(i), &d3d12.IID_ID3D12Resource, &s.backBuffers[i]); hr.Failed() {
			return hr.Err()
		}
		device.device.CreateRenderTargetView(s.backBuffers[i], nil, handle)
		s.views[i] = &d3d12View{
			handle:        handle,
			resource:      s.backBuffers[i],
			format:        s.format,
			mipLevelCount: 1,
		}
		handle.Ptr += uintptr(s.rtvStride)
	}
	return nil
}

func (s *d3d12Surface) releaseBackBuffers() {
	for i := range s.backBuffers {
		if s.backBuffers[i].Valid() {
			s.backBuffers[i].Release()
			s.backBuffers[i] = com.Object{}
		}
		s.views[i] = nil
	}
}

func (s *d3d12Surface) AcquireFrame() (gpu.FrameBackend, error) {
	if !s.swapChain.Valid() {
		return nil, errors.New("gpu: surface is not configured")
	}
	index := s.swapChain.GetCurrentBackBufferIndex()
	if index >= frameCount {
		return nil, fmt.Errorf("gpu: swap chain back buffer index %d is out of range", index)
	}
	return &d3d12Frame{surface: s, index: index}, nil
}

func (s *d3d12Surface) Close() {
	s.releaseBackBuffers()
	if s.rtvHeap.Valid() {
		s.rtvHeap.Release()
		s.rtvHeap = d3d12.ID3D12DescriptorHeap{}
	}
	if s.swapChain.Valid() {
		s.swapChain.Release()
		s.swapChain = dxgi.IDXGISwapChain3{}
	}
}

type d3d12Frame struct {
	surface *d3d12Surface
	index   uint32
}

func (f *d3d12Frame) Texture() gpu.TextureBackend {
	return &d3d12SurfaceTexture{surface: f.surface, index: f.index}
}

func (f *d3d12Frame) Present() error {
	interval := uint32(1)
	if f.surface.config.PresentMode == gpu.PresentModeImmediate {
		interval = 0
	}
	if hr := f.surface.swapChain.Present(interval, 0); hr.Failed() {
		return hr.Err()
	}
	return nil
}

type d3d12SurfaceTexture struct {
	surface *d3d12Surface
	index   uint32
}

func (t *d3d12SurfaceTexture) CreateView(desc *gpu.TextureViewDescriptor) (gpu.ViewBackend, error) {
	if desc != nil {
		if desc.Dimension == gpu.TextureViewDimension2DArray {
			return nil, errors.New("gpu: array texture views are not implemented by the d3d12 backend")
		}
		if desc.Format != gpu.TextureFormatUndefined {
			format, err := dxgiFormat(desc.Format)
			if err != nil {
				return nil, err
			}
			if format != t.surface.format {
				return nil, errors.New("gpu: texture view format must match the surface format")
			}
		}
		if desc.BaseMipLevel != 0 || desc.MipLevelCount > 1 {
			return nil, errors.New("gpu: surface textures only have one mip level")
		}
	}
	return t.surface.views[t.index], nil
}

func (t *d3d12SurfaceTexture) Destroy() {}

type d3d12View struct {
	handle        d3d12.D3D12_CPU_DESCRIPTOR_HANDLE
	resource      com.Object
	format        dxgi.DXGI_FORMAT
	baseMipLevel  uint32
	mipLevelCount uint32
	texture       *d3d12Texture
	rtvHeap       d3d12.ID3D12DescriptorHeap
	ownsHeap      bool
}

func (v *d3d12View) Destroy() {
	if v.ownsHeap && v.rtvHeap.Valid() {
		v.rtvHeap.Release()
		v.rtvHeap = d3d12.ID3D12DescriptorHeap{}
		v.ownsHeap = false
	}
}

func dxgiFormat(format gpu.TextureFormat) (dxgi.DXGI_FORMAT, error) {
	switch format {
	case gpu.TextureFormatRGBA8Unorm:
		return dxgi.DXGI_FORMAT_R8G8B8A8_UNORM, nil
	case gpu.TextureFormatBGRA8Unorm:
		return dxgi.DXGI_FORMAT_B8G8R8A8_UNORM, nil
	}
	return dxgi.DXGI_FORMAT_UNKNOWN, fmt.Errorf("gpu: texture format %d is not supported by the d3d12 backend", int(format))
}

func rgba(c gpu.Color) [4]float32 {
	return [4]float32{float32(c.R), float32(c.G), float32(c.B), float32(c.A)}
}
