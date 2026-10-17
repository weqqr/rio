package d3d12

import (
	"syscall"
	"unsafe"

	"github.com/ebitengine/purego"

	"github.com/weqqr/rio/pkg/com"
)

var (
	IID_ID3D12RootSignature = com.IID{Data1: 0xC54A6B66, Data2: 0x72DF, Data3: 0x4EE8, Data4: [8]byte{0x8B, 0xE5, 0xA9, 0x46, 0xA1, 0x42, 0x92, 0x14}}
	IID_ID3D12PipelineState = com.IID{Data1: 0x765A30F3, Data2: 0xF624, Data3: 0x4C6F, Data4: [8]byte{0xA8, 0x28, 0xAC, 0xE9, 0x48, 0x62, 0x24, 0x45}}
)

const (
	slotDeviceCreateGraphicsPipelineState = 10
	slotDeviceCreateRootSignature         = 16
	slotDeviceCreateCommittedResource     = 27
)

type ID3D12RootSignature struct{ com.Object }

type ID3D12PipelineState struct{ com.Object }

type ID3D12Resource struct{ com.Object }

func (d ID3D12Device) CreateGraphicsPipelineState(desc *D3D12_GRAPHICS_PIPELINE_STATE_DESC, state *ID3D12PipelineState) com.HRESULT {
	return com.HRESULT(d.Call(slotDeviceCreateGraphicsPipelineState,
		uintptr(unsafe.Pointer(desc)),
		uintptr(unsafe.Pointer(&IID_ID3D12PipelineState)),
		uintptr(unsafe.Pointer(&state.Object))))
}

func (d ID3D12Device) CreateRootSignature(nodeMask uint32, blob unsafe.Pointer, blobLength uintptr, signature *ID3D12RootSignature) com.HRESULT {
	return com.HRESULT(d.Call(slotDeviceCreateRootSignature,
		uintptr(nodeMask), uintptr(blob), blobLength,
		uintptr(unsafe.Pointer(&IID_ID3D12RootSignature)),
		uintptr(unsafe.Pointer(&signature.Object))))
}

func (d ID3D12Device) CreateCommittedResource(heapProperties *D3D12_HEAP_PROPERTIES, heapFlags D3D12_HEAP_FLAGS, desc *D3D12_RESOURCE_DESC, initialState D3D12_RESOURCE_STATES, optimizedClearValue unsafe.Pointer, resource *ID3D12Resource) com.HRESULT {
	return com.HRESULT(d.Call(slotDeviceCreateCommittedResource,
		uintptr(unsafe.Pointer(heapProperties)),
		uintptr(heapFlags),
		uintptr(unsafe.Pointer(desc)),
		uintptr(initialState),
		uintptr(optimizedClearValue),
		uintptr(unsafe.Pointer(&IID_ID3D12Resource)),
		uintptr(unsafe.Pointer(&resource.Object))))
}

const (
	slotResourceMap                  = 8
	slotResourceUnmap                = 9
	slotResourceGetGPUVirtualAddress = 11
)

func (r ID3D12Resource) Map(subresource uint32, readRange *D3D12_RANGE) (unsafe.Pointer, com.HRESULT) {
	var data unsafe.Pointer
	hr := com.HRESULT(r.Call(slotResourceMap,
		uintptr(subresource),
		uintptr(unsafe.Pointer(readRange)),
		uintptr(unsafe.Pointer(&data))))
	return data, hr
}

func (r ID3D12Resource) Unmap(subresource uint32, writtenRange *D3D12_RANGE) {
	r.Call(slotResourceUnmap, uintptr(subresource), uintptr(unsafe.Pointer(writtenRange)))
}

func (r ID3D12Resource) GetGPUVirtualAddress() D3D12_GPU_VIRTUAL_ADDRESS {
	return D3D12_GPU_VIRTUAL_ADDRESS(r.Call(slotResourceGetGPUVirtualAddress))
}

const (
	slotListDrawInstanced            = 12
	slotListCopyBufferRegion         = 15
	slotListIASetPrimitiveTopology   = 20
	slotListRSSetViewports           = 21
	slotListRSSetScissorRects        = 22
	slotListSetPipelineState         = 25
	slotListSetGraphicsRootSignature = 30
	slotListIASetVertexBuffers       = 44
)

func (l ID3D12GraphicsCommandList) DrawInstanced(vertexCount, instanceCount, firstVertex, firstInstance uint32) {
	l.Call(slotListDrawInstanced,
		uintptr(vertexCount), uintptr(instanceCount),
		uintptr(firstVertex), uintptr(firstInstance))
}

func (l ID3D12GraphicsCommandList) CopyBufferRegion(dst com.Object, dstOffset uint64, src com.Object, srcOffset uint64, size uint64) {
	l.Call(slotListCopyBufferRegion,
		uintptr(dst.Ptr()), uintptr(dstOffset),
		uintptr(src.Ptr()), uintptr(srcOffset), uintptr(size))
}

func (l ID3D12GraphicsCommandList) IASetPrimitiveTopology(topology D3D_PRIMITIVE_TOPOLOGY) {
	l.Call(slotListIASetPrimitiveTopology, uintptr(topology))
}

func (l ID3D12GraphicsCommandList) RSSetViewports(count uint32, viewports *D3D12_VIEWPORT) {
	l.Call(slotListRSSetViewports, uintptr(count), uintptr(unsafe.Pointer(viewports)))
}

func (l ID3D12GraphicsCommandList) RSSetScissorRects(count uint32, rects *D3D12_RECT) {
	l.Call(slotListRSSetScissorRects, uintptr(count), uintptr(unsafe.Pointer(rects)))
}

func (l ID3D12GraphicsCommandList) SetPipelineState(state com.Object) {
	l.Call(slotListSetPipelineState, uintptr(state.Ptr()))
}

func (l ID3D12GraphicsCommandList) SetGraphicsRootSignature(signature com.Object) {
	l.Call(slotListSetGraphicsRootSignature, uintptr(signature.Ptr()))
}

func (l ID3D12GraphicsCommandList) IASetVertexBuffers(startSlot, count uint32, views *D3D12_VERTEX_BUFFER_VIEW) {
	l.Call(slotListIASetVertexBuffers,
		uintptr(startSlot), uintptr(count), uintptr(unsafe.Pointer(views)))
}

type ID3DBlob struct{ com.Object }

const (
	slotBlobGetBufferPointer = 3
	slotBlobGetBufferSize    = 4
)

func (b ID3DBlob) GetBufferPointer() unsafe.Pointer {
	pointer := b.Call(slotBlobGetBufferPointer)
	return *(*unsafe.Pointer)(unsafe.Pointer(&pointer))
}

func (b ID3DBlob) GetBufferSize() uintptr {
	return b.Call(slotBlobGetBufferSize)
}

var D3D12SerializeRootSignature func(desc *D3D12_ROOT_SIGNATURE_DESC, version D3D_ROOT_SIGNATURE_VERSION, blob *com.Object, errorBlob *com.Object) com.HRESULT

func init() {
	d3d12DLL := syscall.NewLazyDLL("d3d12.dll")
	purego.RegisterFunc(&D3D12SerializeRootSignature, d3d12DLL.NewProc("D3D12SerializeRootSignature").Addr())
}
