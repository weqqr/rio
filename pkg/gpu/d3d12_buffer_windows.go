package gpu

import (
	"errors"
	"fmt"
	"unsafe"

	"github.com/weqqr/rio/pkg/com"
	"github.com/weqqr/rio/pkg/sys/d3d12"
	"github.com/weqqr/rio/pkg/sys/dxgi"
)

const bufferCopyAlignment = 4

type d3d12Buffer struct {
	device   *d3d12Device
	resource d3d12.ID3D12Resource
	size     uint64
	usage    BufferUsage
	heapType d3d12.D3D12_HEAP_TYPE
	state    d3d12.D3D12_RESOURCE_STATES

	mapped       unsafe.Pointer
	mappedOffset uint64
	mappedSize   uint64
}

func (d *d3d12Device) CreateBuffer(desc *BufferDescriptor) (bufferBackend, error) {
	if desc.Size == 0 {
		return nil, errors.New("gpu: buffer size must be greater than zero")
	}
	if desc.Size%bufferCopyAlignment != 0 {
		return nil, fmt.Errorf("gpu: buffer size %d is not a multiple of %d", desc.Size, bufferCopyAlignment)
	}
	if desc.Usage == BufferUsageNone {
		return nil, errors.New("gpu: buffer requires at least one usage")
	}
	if desc.Usage&BufferUsageMapRead != 0 {
		return nil, errors.New("gpu: d3d12 map-read buffers are not implemented")
	}
	if desc.MappedAtCreation && desc.Usage&BufferUsageMapWrite == 0 {
		return nil, errors.New("gpu: a mapped-at-creation buffer requires map-write usage")
	}
	if desc.MappedAtCreation && desc.Usage&BufferUsageCopyDst != 0 {
		return nil, errors.New("gpu: a mapped-at-creation buffer cannot also be a copy destination")
	}

	upload := desc.MappedAtCreation || desc.Usage&BufferUsageMapWrite != 0
	heapType := d3d12.D3D12_HEAP_TYPE_DEFAULT
	initialState := resourceStateForUsage(desc.Usage)
	if upload {
		heapType = d3d12.D3D12_HEAP_TYPE_UPLOAD
		initialState = d3d12.D3D12_RESOURCE_STATE_GENERIC_READ
	} else if desc.Usage&BufferUsageCopyDst != 0 {
		initialState = d3d12.D3D12_RESOURCE_STATE_COPY_DEST
	}

	heap := d3d12.D3D12_HEAP_PROPERTIES{Type: heapType}
	resourceDesc := d3d12.D3D12_RESOURCE_DESC{
		Dimension:        d3d12.D3D12_RESOURCE_DIMENSION_BUFFER,
		Width:            desc.Size,
		Height:           1,
		DepthOrArraySize: 1,
		MipLevels:        1,
		Format:           dxgi.DXGI_FORMAT_UNKNOWN,
		SampleDesc:       dxgi.DXGI_SAMPLE_DESC{Count: 1},
		Layout:           d3d12.D3D12_TEXTURE_LAYOUT_ROW_MAJOR,
		Flags:            d3d12.D3D12_RESOURCE_FLAG_NONE,
	}

	buffer := &d3d12Buffer{
		device:   d,
		size:     desc.Size,
		usage:    desc.Usage,
		heapType: heapType,
		state:    initialState,
	}
	if hr := d.device.CreateCommittedResource(&heap, d3d12.D3D12_HEAP_FLAG_NONE, &resourceDesc, initialState, nil, &buffer.resource); hr.Failed() {
		return nil, fmt.Errorf("gpu: create buffer: %w", hr.Err())
	}
	if desc.MappedAtCreation {
		if err := buffer.MapAsync(MapModeWrite, 0, desc.Size); err != nil {
			buffer.Destroy()
			return nil, err
		}
	}
	return buffer, nil
}

func resourceStateForUsage(usage BufferUsage) d3d12.D3D12_RESOURCE_STATES {
	switch {
	case usage&(BufferUsageVertex|BufferUsageUniform) != 0:
		return d3d12.D3D12_RESOURCE_STATE_VERTEX_AND_CONSTANT_BUFFER
	case usage&BufferUsageIndex != 0:
		return d3d12.D3D12_RESOURCE_STATE_INDEX_BUFFER
	}
	return d3d12.D3D12_RESOURCE_STATE_COMMON
}

func (b *d3d12Buffer) Size() uint64 { return b.size }

func (b *d3d12Buffer) Usage() BufferUsage { return b.usage }

func (b *d3d12Buffer) MapAsync(mode MapMode, offset, size uint64) error {
	if mode != MapModeWrite {
		return errors.New("gpu: d3d12 only supports write mapping")
	}
	if b.heapType != d3d12.D3D12_HEAP_TYPE_UPLOAD {
		return errors.New("gpu: buffer was not created with map-write usage")
	}
	if b.mapped != nil {
		return errors.New("gpu: buffer is already mapped")
	}
	if offset%8 != 0 {
		return fmt.Errorf("gpu: map offset %d is not 8-byte aligned", offset)
	}
	if size%bufferCopyAlignment != 0 {
		return fmt.Errorf("gpu: map size %d is not a multiple of %d", size, bufferCopyAlignment)
	}
	if offset > b.size || size > b.size-offset {
		return fmt.Errorf("gpu: map range [%d, %d) is out of bounds for a %d-byte buffer", offset, offset+size, b.size)
	}

	data, hr := b.resource.Map(0, &d3d12.D3D12_RANGE{})
	if hr.Failed() {
		return hr.Err()
	}
	b.mapped = data
	b.mappedOffset = offset
	b.mappedSize = size
	return nil
}

func (b *d3d12Buffer) GetMappedRange(offset, size uint64) ([]byte, error) {
	if b.mapped == nil {
		return nil, errors.New("gpu: buffer is not mapped")
	}
	if size > b.mappedSize || offset < b.mappedOffset || offset-b.mappedOffset > b.mappedSize-size {
		return nil, fmt.Errorf("gpu: mapped range [%d, %d) is outside the mapped range [%d, %d)", offset, offset+size, b.mappedOffset, b.mappedOffset+b.mappedSize)
	}
	bytes := unsafe.Slice((*byte)(b.mapped), int(b.size))
	return bytes[offset : offset+size], nil
}

func (b *d3d12Buffer) Unmap() {
	if b.mapped == nil {
		return
	}
	b.resource.Unmap(0, nil)
	b.mapped = nil
	b.mappedOffset = 0
	b.mappedSize = 0
}

func (b *d3d12Buffer) Destroy() {
	b.Unmap()
	if b.resource.Valid() {
		b.resource.Release()
		b.resource = d3d12.ID3D12Resource{}
	}
}

func (q *d3d12Queue) WriteBuffer(buffer *Buffer, offset uint64, data []byte) error {
	dst, ok := buffer.b.(*d3d12Buffer)
	if !ok {
		return errors.New("gpu: write buffer target is not a d3d12 buffer")
	}
	if dst.usage&BufferUsageCopyDst == 0 {
		return errors.New("gpu: buffer was not created with copy-destination usage")
	}
	if len(data) == 0 {
		return nil
	}
	if offset%bufferCopyAlignment != 0 || len(data)%bufferCopyAlignment != 0 {
		return fmt.Errorf("gpu: write buffer offset and size must be multiples of %d", bufferCopyAlignment)
	}
	size := uint64(len(data))
	if offset > dst.size || size > dst.size-offset {
		return fmt.Errorf("gpu: write buffer range [%d, %d) is out of bounds for a %d-byte buffer", offset, offset+size, dst.size)
	}

	stagingImpl, err := q.device.CreateBuffer(&BufferDescriptor{
		Label: "staging",
		Size:  size,
		Usage: BufferUsageMapWrite | BufferUsageCopySrc,
	})
	if err != nil {
		return err
	}
	staging := &Buffer{b: stagingImpl}
	defer staging.Destroy()

	if err := staging.MapAsync(MapModeWrite, 0, size); err != nil {
		return err
	}
	mapped, err := staging.GetMappedRange(0, size)
	if err != nil {
		return err
	}
	copy(mapped, data)
	staging.Unmap()

	encoderImpl, err := q.device.CreateCommandEncoder()
	if err != nil {
		return err
	}
	encoder := &CommandEncoder{b: encoderImpl}
	if err := encoder.CopyBufferToBuffer(staging, 0, buffer, offset, size); err != nil {
		return err
	}
	commands, err := encoder.Finish()
	if err != nil {
		return err
	}
	return q.Submit([]*CommandBuffer{commands})
}

func (e *d3d12Encoder) CopyBufferToBuffer(src *Buffer, srcOffset uint64, dst *Buffer, dstOffset uint64, size uint64) error {
	srcBuffer, ok := src.b.(*d3d12Buffer)
	if !ok {
		return errors.New("gpu: copy source is not a d3d12 buffer")
	}
	dstBuffer, ok := dst.b.(*d3d12Buffer)
	if !ok {
		return errors.New("gpu: copy destination is not a d3d12 buffer")
	}
	if srcBuffer.usage&BufferUsageCopySrc == 0 {
		return errors.New("gpu: source buffer was not created with copy-source usage")
	}
	if dstBuffer.usage&BufferUsageCopyDst == 0 {
		return errors.New("gpu: destination buffer was not created with copy-destination usage")
	}
	if size == 0 {
		return nil
	}
	if srcOffset%bufferCopyAlignment != 0 || dstOffset%bufferCopyAlignment != 0 || size%bufferCopyAlignment != 0 {
		return fmt.Errorf("gpu: copy offsets and size must be multiples of %d", bufferCopyAlignment)
	}
	if srcOffset > srcBuffer.size || size > srcBuffer.size-srcOffset {
		return fmt.Errorf("gpu: copy source range [%d, %d) is out of bounds", srcOffset, srcOffset+size)
	}
	if dstOffset > dstBuffer.size || size > dstBuffer.size-dstOffset {
		return fmt.Errorf("gpu: copy destination range [%d, %d) is out of bounds", dstOffset, dstOffset+size)
	}
	if dstBuffer.heapType == d3d12.D3D12_HEAP_TYPE_UPLOAD {
		return errors.New("gpu: an upload-heap buffer cannot be a copy destination")
	}

	sourceState := srcBuffer.state
	if srcBuffer.heapType != d3d12.D3D12_HEAP_TYPE_UPLOAD && sourceState&d3d12.D3D12_RESOURCE_STATE_COPY_SOURCE == 0 {
		e.transition(srcBuffer.resource.Object, sourceState, d3d12.D3D12_RESOURCE_STATE_COPY_SOURCE)
	}
	if dstBuffer.state != d3d12.D3D12_RESOURCE_STATE_COPY_DEST {
		e.transition(dstBuffer.resource.Object, dstBuffer.state, d3d12.D3D12_RESOURCE_STATE_COPY_DEST)
	}

	e.device.list.CopyBufferRegion(dstBuffer.resource.Object, dstOffset, srcBuffer.resource.Object, srcOffset, size)

	if srcBuffer.heapType != d3d12.D3D12_HEAP_TYPE_UPLOAD && sourceState&d3d12.D3D12_RESOURCE_STATE_COPY_SOURCE == 0 {
		e.transition(srcBuffer.resource.Object, d3d12.D3D12_RESOURCE_STATE_COPY_SOURCE, sourceState)
	}
	finalState := resourceStateForUsage(dstBuffer.usage)
	if finalState != d3d12.D3D12_RESOURCE_STATE_COPY_DEST {
		e.transition(dstBuffer.resource.Object, d3d12.D3D12_RESOURCE_STATE_COPY_DEST, finalState)
		dstBuffer.state = finalState
	}
	return nil
}

func (e *d3d12Encoder) transition(resource com.Object, before, after d3d12.D3D12_RESOURCE_STATES) {
	if before == after {
		return
	}
	e.device.list.ResourceBarrier(1, &d3d12.D3D12_RESOURCE_BARRIER{
		Type: d3d12.D3D12_RESOURCE_BARRIER_TYPE_TRANSITION,
		Transition: d3d12.D3D12_RESOURCE_TRANSITION_BARRIER{
			Resource:    resource.Ptr(),
			Subresource: d3d12.D3D12_RESOURCE_BARRIER_ALL_SUBRESOURCES,
			StateBefore: before,
			StateAfter:  after,
		},
	})
}
