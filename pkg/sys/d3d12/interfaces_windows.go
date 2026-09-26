//go:build windows

package d3d12

import (
	"unsafe"

	"github.com/weqqr/rio/pkg/com"
)

const (
	slotDeviceCreateCommandQueue               = 8
	slotDeviceCreateCommandAllocator           = 9
	slotDeviceCreateCommandList                = 12
	slotDeviceCreateDescriptorHeap             = 14
	slotDeviceGetDescriptorHandleIncrementSize = 15
	slotDeviceCreateRenderTargetView           = 20
	slotDeviceCreateFence                      = 36
)

type ID3D12Device struct{ com.Object }

func (d ID3D12Device) CreateCommandQueue(desc *D3D12_COMMAND_QUEUE_DESC, queue *ID3D12CommandQueue) com.HRESULT {
	return com.HRESULT(d.Object.Call(slotDeviceCreateCommandQueue,
		uintptr(unsafe.Pointer(desc)),
		uintptr(unsafe.Pointer(&IID_ID3D12CommandQueue)),
		uintptr(unsafe.Pointer(&queue.Object))))
}

func (d ID3D12Device) CreateCommandAllocator(listType D3D12_COMMAND_LIST_TYPE, allocator *ID3D12CommandAllocator) com.HRESULT {
	return com.HRESULT(d.Object.Call(slotDeviceCreateCommandAllocator,
		uintptr(listType),
		uintptr(unsafe.Pointer(&IID_ID3D12CommandAllocator)),
		uintptr(unsafe.Pointer(&allocator.Object))))
}

func (d ID3D12Device) CreateCommandList(nodeMask uint32, listType D3D12_COMMAND_LIST_TYPE, allocator com.Object, initialState com.Object, list *ID3D12GraphicsCommandList) com.HRESULT {
	return com.HRESULT(d.Object.Call(slotDeviceCreateCommandList,
		uintptr(nodeMask), uintptr(listType),
		uintptr(allocator.Ptr()), uintptr(initialState.Ptr()),
		uintptr(unsafe.Pointer(&IID_ID3D12GraphicsCommandList)),
		uintptr(unsafe.Pointer(&list.Object))))
}

func (d ID3D12Device) CreateDescriptorHeap(desc *D3D12_DESCRIPTOR_HEAP_DESC, heap *ID3D12DescriptorHeap) com.HRESULT {
	return com.HRESULT(d.Object.Call(slotDeviceCreateDescriptorHeap,
		uintptr(unsafe.Pointer(desc)),
		uintptr(unsafe.Pointer(&IID_ID3D12DescriptorHeap)),
		uintptr(unsafe.Pointer(&heap.Object))))
}

func (d ID3D12Device) GetDescriptorHandleIncrementSize(heapType D3D12_DESCRIPTOR_HEAP_TYPE) uint32 {
	return uint32(d.Object.Call(slotDeviceGetDescriptorHandleIncrementSize, uintptr(heapType)))
}

func (d ID3D12Device) CreateRenderTargetView(resource com.Object, desc unsafe.Pointer, dest D3D12_CPU_DESCRIPTOR_HANDLE) {
	d.Object.Call(slotDeviceCreateRenderTargetView,
		uintptr(resource.Ptr()), uintptr(desc), dest.Ptr)
}

func (d ID3D12Device) CreateFence(initialValue uint64, flags D3D12_FENCE_FLAGS, fence *ID3D12Fence) com.HRESULT {
	return com.HRESULT(d.Object.Call(slotDeviceCreateFence,
		uintptr(initialValue), uintptr(flags),
		uintptr(unsafe.Pointer(&IID_ID3D12Fence)),
		uintptr(unsafe.Pointer(&fence.Object))))
}

const (
	slotQueueExecuteCommandLists = 10
	slotQueueSignal              = 14
)

type ID3D12CommandQueue struct{ com.Object }

func (q ID3D12CommandQueue) ExecuteCommandLists(numLists uint32, lists *unsafe.Pointer) {
	q.Object.Call(slotQueueExecuteCommandLists, uintptr(numLists), uintptr(unsafe.Pointer(lists)))
}

func (q ID3D12CommandQueue) Signal(fence com.Object, value uint64) com.HRESULT {
	return com.HRESULT(q.Object.Call(slotQueueSignal, uintptr(fence.Ptr()), uintptr(value)))
}

const slotAllocatorReset = 8

type ID3D12CommandAllocator struct{ com.Object }

func (a ID3D12CommandAllocator) Reset() com.HRESULT {
	return com.HRESULT(a.Object.Call(slotAllocatorReset))
}

const (
	slotListReset                 = 10
	slotListClose                 = 9
	slotListResourceBarrier       = 26
	slotListOMSetRenderTargets    = 46
	slotListClearRenderTargetView = 48
)

type ID3D12GraphicsCommandList struct{ com.Object }

func (l ID3D12GraphicsCommandList) Reset(allocator com.Object, initialState com.Object) com.HRESULT {
	return com.HRESULT(l.Object.Call(slotListReset, uintptr(allocator.Ptr()), uintptr(initialState.Ptr())))
}

func (l ID3D12GraphicsCommandList) Close() com.HRESULT {
	return com.HRESULT(l.Object.Call(slotListClose))
}

func (l ID3D12GraphicsCommandList) ResourceBarrier(numBarriers uint32, barriers *D3D12_RESOURCE_BARRIER) {
	l.Object.Call(slotListResourceBarrier, uintptr(numBarriers), uintptr(unsafe.Pointer(barriers)))
}

func (l ID3D12GraphicsCommandList) OMSetRenderTargets(numRenderTargetViews uint32, rtvs *D3D12_CPU_DESCRIPTOR_HANDLE, singleHandleToDescriptorRange uint32, dsv uintptr) {
	l.Object.Call(slotListOMSetRenderTargets,
		uintptr(numRenderTargetViews), uintptr(unsafe.Pointer(rtvs)),
		uintptr(singleHandleToDescriptorRange), dsv)
}

func (l ID3D12GraphicsCommandList) ClearRenderTargetView(rtv D3D12_CPU_DESCRIPTOR_HANDLE, color *[4]float32, numRects uint32, rects uintptr) {
	l.Object.Call(slotListClearRenderTargetView,
		rtv.Ptr, uintptr(unsafe.Pointer(color)), uintptr(numRects), rects)
}

const (
	slotFenceGetCompletedValue    = 8
	slotFenceSetEventOnCompletion = 9
)

type ID3D12Fence struct{ com.Object }

func (f ID3D12Fence) GetCompletedValue() uint64 {
	return uint64(f.Object.Call(slotFenceGetCompletedValue))
}

func (f ID3D12Fence) SetEventOnCompletion(value uint64, event uintptr) com.HRESULT {
	return com.HRESULT(f.Object.Call(slotFenceSetEventOnCompletion, uintptr(value), event))
}

type ID3D12DescriptorHeap struct{ com.Object }

const slotHeapGetCPUDescriptorHandleForHeapStart = 9

func (h ID3D12DescriptorHeap) GetCPUDescriptorHandleForHeapStart() D3D12_CPU_DESCRIPTOR_HANDLE {
	var handle D3D12_CPU_DESCRIPTOR_HANDLE
	h.Object.Call(slotHeapGetCPUDescriptorHandleForHeapStart, uintptr(unsafe.Pointer(&handle)))
	return handle
}
