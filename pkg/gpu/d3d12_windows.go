package gpu

import (
	"errors"
	"fmt"
	"unsafe"

	"github.com/weqqr/rio/pkg/com"
	"github.com/weqqr/rio/pkg/sys/d3d12"
	"github.com/weqqr/rio/pkg/sys/dxgi"
	"github.com/weqqr/rio/pkg/sys/win32"
)

const frameCount = 2

func errWin32(fn string) error {
	return fmt.Errorf("gpu: %s failed with error %d", fn, win32.GetLastError())
}

type d3d12Instance struct {
	factory dxgi.IDXGIFactory2
}

func newD3D12Instance() (*d3d12Instance, error) {
	instance := &d3d12Instance{}
	if hr := dxgi.CreateDXGIFactory2(0, &dxgi.IID_IDXGIFactory2, &instance.factory.Object); hr.Failed() {
		return nil, hr.Err()
	}
	return instance, nil
}

func (i *d3d12Instance) Backend() Backend { return BackendD3D12 }

func (i *d3d12Instance) CreateSurface(target SurfaceTarget) (surfaceBackend, error) {
	if target.Handle == 0 {
		return nil, errors.New("gpu: surface target has no window handle")
	}
	return &d3d12Surface{factory: i.factory, target: target}, nil
}

func (i *d3d12Instance) RequestAdapter(opts *RequestAdapterOptions) (adapterBackend, error) {
	if opts != nil && opts.Surface != nil {
		if _, ok := opts.Surface.b.(*d3d12Surface); !ok {
			return nil, errors.New("gpu: d3d12 cannot present to a surface from another backend")
		}
	}
	return &d3d12Adapter{}, nil
}

func (i *d3d12Instance) Close() {
	if i.factory.Valid() {
		i.factory.Release()
		i.factory = dxgi.IDXGIFactory2{}
	}
}

type d3d12Adapter struct{}

func (a *d3d12Adapter) RequestDevice() (deviceBackend, error) {
	device := &d3d12Device{}
	if hr := d3d12.D3D12CreateDevice(0, d3d12.D3D_FEATURE_LEVEL_11_0, &d3d12.IID_ID3D12Device, &device.device.Object); hr.Failed() {
		return nil, hr.Err()
	}

	queueDesc := d3d12.D3D12_COMMAND_QUEUE_DESC{Type: d3d12.D3D12_COMMAND_LIST_TYPE_DIRECT}
	if hr := device.device.CreateCommandQueue(&queueDesc, &device.queue); hr.Failed() {
		_ = device.Close()
		return nil, hr.Err()
	}
	if hr := device.device.CreateCommandAllocator(d3d12.D3D12_COMMAND_LIST_TYPE_DIRECT, &device.allocator); hr.Failed() {
		_ = device.Close()
		return nil, hr.Err()
	}
	if hr := device.device.CreateCommandList(0, d3d12.D3D12_COMMAND_LIST_TYPE_DIRECT, device.allocator.Object, com.Object{}, &device.list); hr.Failed() {
		_ = device.Close()
		return nil, hr.Err()
	}
	if hr := device.list.Close(); hr.Failed() {
		_ = device.Close()
		return nil, hr.Err()
	}
	if hr := device.device.CreateFence(0, d3d12.D3D12_FENCE_FLAG_NONE, &device.fence); hr.Failed() {
		_ = device.Close()
		return nil, hr.Err()
	}
	device.fenceEvent = win32.CreateEventW(0, 0, 0, 0)
	if device.fenceEvent == 0 {
		_ = device.Close()
		return nil, errWin32("CreateEventW")
	}
	return device, nil
}

type d3d12Device struct {
	device    d3d12.ID3D12Device
	queue     d3d12.ID3D12CommandQueue
	allocator d3d12.ID3D12CommandAllocator
	list      d3d12.ID3D12GraphicsCommandList

	rootSignature d3d12.ID3D12RootSignature

	fence      d3d12.ID3D12Fence
	fenceEvent uintptr
	fenceValue uint64
}

func (d *d3d12Device) Queue() queueBackend { return &d3d12Queue{device: d} }

func (d *d3d12Device) CreateCommandEncoder() (encoderBackend, error) {
	if hr := d.allocator.Reset(); hr.Failed() {
		return nil, hr.Err()
	}
	if hr := d.list.Reset(d.allocator.Object, com.Object{}); hr.Failed() {
		return nil, hr.Err()
	}
	return &d3d12Encoder{device: d}, nil
}

func (d *d3d12Device) waitForGpu() error {
	d.fenceValue++
	if hr := d.queue.Signal(d.fence.Object, d.fenceValue); hr.Failed() {
		return hr.Err()
	}
	if hr := d.fence.SetEventOnCompletion(d.fenceValue, d.fenceEvent); hr.Failed() {
		return hr.Err()
	}
	if win32.WaitForSingleObject(d.fenceEvent, win32.INFINITE) == win32.WAIT_FAILED {
		return errWin32("WaitForSingleObject")
	}
	return nil
}

func (d *d3d12Device) Close() error {
	var firstErr error
	if d.fence.Valid() && d.queue.Valid() && d.fenceEvent != 0 {
		firstErr = d.waitForGpu()
	}
	if d.list.Valid() {
		d.list.Release()
		d.list = d3d12.ID3D12GraphicsCommandList{}
	}
	if d.allocator.Valid() {
		d.allocator.Release()
		d.allocator = d3d12.ID3D12CommandAllocator{}
	}
	if d.rootSignature.Valid() {
		d.rootSignature.Release()
		d.rootSignature = d3d12.ID3D12RootSignature{}
	}
	if d.fence.Valid() {
		d.fence.Release()
		d.fence = d3d12.ID3D12Fence{}
	}
	if d.queue.Valid() {
		d.queue.Release()
		d.queue = d3d12.ID3D12CommandQueue{}
	}
	if d.device.Valid() {
		d.device.Release()
		d.device = d3d12.ID3D12Device{}
	}
	if d.fenceEvent != 0 {
		win32.CloseHandle(d.fenceEvent)
		d.fenceEvent = 0
	}
	return firstErr
}

type d3d12Queue struct {
	device *d3d12Device
}

func (q *d3d12Queue) Submit(cmds []*CommandBuffer) error {
	if len(cmds) == 0 {
		return nil
	}
	lists := make([]unsafe.Pointer, len(cmds))
	for i, cmd := range cmds {
		buffer, ok := cmd.b.(*d3d12CommandBuffer)
		if !ok {
			return errors.New("gpu: command buffer is not a d3d12 command buffer")
		}
		lists[i] = buffer.list.Ptr()
	}
	q.device.queue.ExecuteCommandLists(uint32(len(lists)), &lists[0])
	return q.device.waitForGpu()
}

type d3d12Encoder struct {
	device *d3d12Device
}

func (e *d3d12Encoder) BeginRenderPass(desc *RenderPassDescriptor) (renderPassBackend, error) {
	views := make([]*d3d12View, len(desc.ColorAttachments))
	handles := make([]d3d12.D3D12_CPU_DESCRIPTOR_HANDLE, len(desc.ColorAttachments))
	afterStates := make([]d3d12.D3D12_RESOURCE_STATES, len(desc.ColorAttachments))
	for i := range desc.ColorAttachments {
		attachment := &desc.ColorAttachments[i]
		if attachment.StoreOp != StoreOpStore {
			return nil, fmt.Errorf("gpu: store op %d is not implemented by the d3d12 backend", int(attachment.StoreOp))
		}
		view, ok := attachment.View.b.(*d3d12View)
		if !ok {
			return nil, errors.New("gpu: color attachment view is not a d3d12 texture view")
		}
		views[i] = view
		handles[i] = view.handle
		before := d3d12.D3D12_RESOURCE_STATE_PRESENT
		after := d3d12.D3D12_RESOURCE_STATE_PRESENT
		if view.texture != nil {
			before = view.texture.state
			after = textureStateAfterUse(view.texture.usage)
			view.texture.state = d3d12.D3D12_RESOURCE_STATE_RENDER_TARGET
		}
		afterStates[i] = after
		e.device.list.ResourceBarrier(1, &d3d12.D3D12_RESOURCE_BARRIER{
			Type: d3d12.D3D12_RESOURCE_BARRIER_TYPE_TRANSITION,
			Transition: d3d12.D3D12_RESOURCE_TRANSITION_BARRIER{
				Resource:    view.resource.Ptr(),
				Subresource: d3d12.D3D12_RESOURCE_BARRIER_ALL_SUBRESOURCES,
				StateBefore: before,
				StateAfter:  d3d12.D3D12_RESOURCE_STATE_RENDER_TARGET,
			},
		})
	}

	e.device.list.OMSetRenderTargets(uint32(len(handles)), &handles[0], 0, 0)

	for i := range desc.ColorAttachments {
		if desc.ColorAttachments[i].LoadOp != LoadOpClear {
			continue
		}
		color := desc.ColorAttachments[i].ClearValue.rgba()
		e.device.list.ClearRenderTargetView(handles[i], &color, 0, 0)
	}
	return &d3d12RenderPass{device: e.device, views: views, afterStates: afterStates}, nil
}

func (e *d3d12Encoder) Finish() (commandBufferBackend, error) {
	if hr := e.device.list.Close(); hr.Failed() {
		return nil, hr.Err()
	}
	return &d3d12CommandBuffer{list: e.device.list}, nil
}

type d3d12CommandBuffer struct {
	list d3d12.ID3D12GraphicsCommandList
}

type d3d12RenderPass struct {
	device      *d3d12Device
	views       []*d3d12View
	afterStates []d3d12.D3D12_RESOURCE_STATES

	pipeline      *d3d12Pipeline
	vertexBuffers map[uint32]*d3d12Buffer
	bindGroups    map[uint32]*d3d12BindGroup
}

func (p *d3d12RenderPass) End() {
	for i, view := range p.views {
		after := p.afterStates[i]
		if view.texture != nil {
			view.texture.state = after
		}
		p.device.list.ResourceBarrier(1, &d3d12.D3D12_RESOURCE_BARRIER{
			Type: d3d12.D3D12_RESOURCE_BARRIER_TYPE_TRANSITION,
			Transition: d3d12.D3D12_RESOURCE_TRANSITION_BARRIER{
				Resource:    view.resource.Ptr(),
				Subresource: d3d12.D3D12_RESOURCE_BARRIER_ALL_SUBRESOURCES,
				StateBefore: d3d12.D3D12_RESOURCE_STATE_RENDER_TARGET,
				StateAfter:  after,
			},
		})
	}
}
