//go:build windows

package render

import (
	"fmt"
	"unsafe"

	"github.com/weqqr/rio/pkg/com"
	"github.com/weqqr/rio/pkg/sys/d3d12"
	"github.com/weqqr/rio/pkg/sys/dxgi"
	"github.com/weqqr/rio/pkg/sys/win32"
)

const frameCount = 2

var clearRed = [4]float32{1, 0, 0, 1}

func errWin32(fn string) error {
	return fmt.Errorf("render: %s failed with error %d", fn, win32.GetLastError())
}

type Renderer struct {
	hwnd uintptr

	factory   dxgi.IDXGIFactory2
	swapChain dxgi.IDXGISwapChain3

	device   d3d12.ID3D12Device
	queue    d3d12.ID3D12CommandQueue
	cmdAlloc d3d12.ID3D12CommandAllocator
	cmdList  d3d12.ID3D12GraphicsCommandList

	rtvHeap     d3d12.ID3D12DescriptorHeap
	rtvStride   uint32
	backBuffers [frameCount]com.Object // ID3D12Resource
	rtvs        [frameCount]d3d12.D3D12_CPU_DESCRIPTOR_HANDLE

	fence      d3d12.ID3D12Fence
	fenceEvent uintptr
	fenceValue uint64
}

func New(rawWindowHandle uintptr) (*Renderer, error) {
	var clientRect win32.RECT
	if win32.GetClientRect(rawWindowHandle, &clientRect) == 0 {
		return nil, errWin32("GetClientRect")
	}

	r := &Renderer{hwnd: rawWindowHandle}

	if hr := dxgi.CreateDXGIFactory2(0, &dxgi.IID_IDXGIFactory2, &r.factory.Object); hr.Failed() {
		return nil, hr.Err()
	}
	if hr := d3d12.D3D12CreateDevice(0, d3d12.D3D_FEATURE_LEVEL_11_0, &d3d12.IID_ID3D12Device, &r.device.Object); hr.Failed() {
		r.Close()
		return nil, hr.Err()
	}

	queueDesc := d3d12.D3D12_COMMAND_QUEUE_DESC{Type: d3d12.D3D12_COMMAND_LIST_TYPE_DIRECT}
	if hr := r.device.CreateCommandQueue(&queueDesc, &r.queue); hr.Failed() {
		r.Close()
		return nil, hr.Err()
	}

	swapChainDesc := dxgi.DXGI_SWAP_CHAIN_DESC1{
		Width:       uint32(clientRect.Right - clientRect.Left),
		Height:      uint32(clientRect.Bottom - clientRect.Top),
		Format:      dxgi.DXGI_FORMAT_R8G8B8A8_UNORM,
		SampleDesc:  dxgi.DXGI_SAMPLE_DESC{Count: 1},
		BufferUsage: dxgi.DXGI_USAGE_RENDER_TARGET_OUTPUT,
		BufferCount: frameCount,
		Scaling:     dxgi.DXGI_SCALING_NONE,
		SwapEffect:  dxgi.DXGI_SWAP_EFFECT_FLIP_DISCARD,
		AlphaMode:   dxgi.DXGI_ALPHA_MODE_UNSPECIFIED,
	}
	var swapChain1 dxgi.IDXGISwapChain1
	if hr := r.factory.CreateSwapChainForHwnd(r.queue.Object, r.hwnd, &swapChainDesc, nil, com.Object{}, &swapChain1); hr.Failed() {
		r.Close()
		return nil, hr.Err()
	}
	swapChain3, err := swapChain1.Object.QueryInterface(&dxgi.IID_IDXGISwapChain3)
	swapChain1.Object.Release()
	if err != nil {
		r.Close()
		return nil, err
	}
	r.swapChain = dxgi.IDXGISwapChain3{Object: swapChain3}

	if hr := r.device.CreateCommandAllocator(d3d12.D3D12_COMMAND_LIST_TYPE_DIRECT, &r.cmdAlloc); hr.Failed() {
		r.Close()
		return nil, hr.Err()
	}
	if hr := r.device.CreateCommandList(0, d3d12.D3D12_COMMAND_LIST_TYPE_DIRECT, r.cmdAlloc.Object, com.Object{}, &r.cmdList); hr.Failed() {
		r.Close()
		return nil, hr.Err()
	}

	if hr := r.cmdList.Close(); hr.Failed() {
		r.Close()
		return nil, hr.Err()
	}

	if hr := r.device.CreateFence(0, d3d12.D3D12_FENCE_FLAG_NONE, &r.fence); hr.Failed() {
		r.Close()
		return nil, hr.Err()
	}
	r.fenceEvent = win32.CreateEventW(0, 0, 0, 0)
	if r.fenceEvent == 0 {
		r.Close()
		return nil, errWin32("CreateEventW")
	}

	rtvHeapDesc := d3d12.D3D12_DESCRIPTOR_HEAP_DESC{
		Type:           d3d12.D3D12_DESCRIPTOR_HEAP_TYPE_RTV,
		NumDescriptors: frameCount,
	}
	if hr := r.device.CreateDescriptorHeap(&rtvHeapDesc, &r.rtvHeap); hr.Failed() {
		r.Close()
		return nil, hr.Err()
	}
	r.rtvStride = r.device.GetDescriptorHandleIncrementSize(d3d12.D3D12_DESCRIPTOR_HEAP_TYPE_RTV)

	if err := r.createRenderTargets(); err != nil {
		r.Close()
		return nil, err
	}
	return r, nil
}

func (r *Renderer) createRenderTargets() error {
	rtv := r.rtvHeap.GetCPUDescriptorHandleForHeapStart()
	for i := range r.backBuffers {
		if hr := r.swapChain.GetBuffer(uint32(i), &d3d12.IID_ID3D12Resource, &r.backBuffers[i]); hr.Failed() {
			return hr.Err()
		}
		r.device.CreateRenderTargetView(r.backBuffers[i], nil, rtv)
		r.rtvs[i] = rtv
		rtv.Ptr += uintptr(r.rtvStride)
	}
	return nil
}

func (r *Renderer) Render() error {
	index := r.swapChain.GetCurrentBackBufferIndex()

	if hr := r.cmdAlloc.Reset(); hr.Failed() {
		return hr.Err()
	}
	if hr := r.cmdList.Reset(r.cmdAlloc.Object, com.Object{}); hr.Failed() {
		return hr.Err()
	}

	transition := d3d12.D3D12_RESOURCE_BARRIER{
		Type: d3d12.D3D12_RESOURCE_BARRIER_TYPE_TRANSITION,
		Transition: d3d12.D3D12_RESOURCE_TRANSITION_BARRIER{
			Resource:    r.backBuffers[index].Ptr(),
			Subresource: d3d12.D3D12_RESOURCE_BARRIER_ALL_SUBRESOURCES,
			StateBefore: d3d12.D3D12_RESOURCE_STATE_PRESENT,
			StateAfter:  d3d12.D3D12_RESOURCE_STATE_RENDER_TARGET,
		},
	}
	r.cmdList.ResourceBarrier(1, &transition)

	r.cmdList.OMSetRenderTargets(1, &r.rtvs[index], 1, 0)
	r.cmdList.ClearRenderTargetView(r.rtvs[index], &clearRed, 0, 0)

	transition.Transition.StateBefore = d3d12.D3D12_RESOURCE_STATE_RENDER_TARGET
	transition.Transition.StateAfter = d3d12.D3D12_RESOURCE_STATE_PRESENT
	r.cmdList.ResourceBarrier(1, &transition)

	if hr := r.cmdList.Close(); hr.Failed() {
		return hr.Err()
	}

	lists := [1]unsafe.Pointer{r.cmdList.Ptr()}
	r.queue.ExecuteCommandLists(1, &lists[0])

	if hr := r.swapChain.Present(1, 0); hr.Failed() {
		return hr.Err()
	}
	return r.waitForGpu()
}

func (r *Renderer) Resize(width, height uint32) error {
	if width == 0 || height == 0 {
		return nil
	}
	for i := range r.backBuffers {
		r.backBuffers[i].Release()
		r.backBuffers[i] = com.Object{}
	}
	if hr := r.swapChain.ResizeBuffers(0, width, height, dxgi.DXGI_FORMAT_UNKNOWN, 0); hr.Failed() {
		return hr.Err()
	}
	return r.createRenderTargets()
}

func (r *Renderer) waitForGpu() error {
	r.fenceValue++
	if hr := r.queue.Signal(r.fence.Object, r.fenceValue); hr.Failed() {
		return hr.Err()
	}
	if hr := r.fence.SetEventOnCompletion(r.fenceValue, r.fenceEvent); hr.Failed() {
		return hr.Err()
	}
	if win32.WaitForSingleObject(r.fenceEvent, win32.INFINITE) == win32.WAIT_FAILED {
		return errWin32("WaitForSingleObject")
	}
	return nil
}

func (r *Renderer) Close() error {
	if r.fence.Valid() && r.queue.Valid() {
		r.waitForGpu()
	}
	if r.cmdList.Valid() {
		r.cmdList.Release()
		r.cmdList = d3d12.ID3D12GraphicsCommandList{}
	}
	if r.cmdAlloc.Valid() {
		r.cmdAlloc.Release()
		r.cmdAlloc = d3d12.ID3D12CommandAllocator{}
	}
	for i := range r.backBuffers {
		if r.backBuffers[i].Valid() {
			r.backBuffers[i].Release()
			r.backBuffers[i] = com.Object{}
		}
	}
	if r.rtvHeap.Valid() {
		r.rtvHeap.Release()
		r.rtvHeap = d3d12.ID3D12DescriptorHeap{}
	}
	if r.fence.Valid() {
		r.fence.Release()
		r.fence = d3d12.ID3D12Fence{}
	}
	if r.swapChain.Valid() {
		r.swapChain.Release()
		r.swapChain = dxgi.IDXGISwapChain3{}
	}
	if r.queue.Valid() {
		r.queue.Release()
		r.queue = d3d12.ID3D12CommandQueue{}
	}
	if r.device.Valid() {
		r.device.Release()
		r.device = d3d12.ID3D12Device{}
	}
	if r.factory.Valid() {
		r.factory.Release()
		r.factory = dxgi.IDXGIFactory2{}
	}
	if r.fenceEvent != 0 {
		win32.CloseHandle(r.fenceEvent)
		r.fenceEvent = 0
	}
	return nil
}
