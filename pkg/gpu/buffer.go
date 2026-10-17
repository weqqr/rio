package gpu

import "errors"

type BufferUsage uint32

const (
	BufferUsageNone     BufferUsage = 0
	BufferUsageMapRead  BufferUsage = 1 << 0
	BufferUsageMapWrite BufferUsage = 1 << 1
	BufferUsageCopySrc  BufferUsage = 1 << 2
	BufferUsageCopyDst  BufferUsage = 1 << 3
	BufferUsageIndex    BufferUsage = 1 << 4
	BufferUsageVertex   BufferUsage = 1 << 5
	BufferUsageUniform  BufferUsage = 1 << 6
	BufferUsageStorage  BufferUsage = 1 << 7
	BufferUsageIndirect BufferUsage = 1 << 8
)

type MapMode uint32

const (
	MapModeRead  MapMode = 1 << 0
	MapModeWrite MapMode = 1 << 1
)

type BufferDescriptor struct {
	Label            string
	Size             uint64
	Usage            BufferUsage
	MappedAtCreation bool
}

type Buffer struct {
	b bufferBackend
}

func (d *Device) CreateBuffer(desc *BufferDescriptor) (*Buffer, error) {
	if desc == nil {
		return nil, errors.New("gpu: buffer descriptor is nil")
	}
	impl, err := d.b.CreateBuffer(desc)
	if err != nil {
		return nil, err
	}
	return &Buffer{b: impl}, nil
}

func (b *Buffer) Size() uint64 { return b.b.Size() }

func (b *Buffer) Usage() BufferUsage { return b.b.Usage() }

func (b *Buffer) MapAsync(mode MapMode, offset, size uint64) error {
	return b.b.MapAsync(mode, offset, size)
}

func (b *Buffer) GetMappedRange(offset, size uint64) ([]byte, error) {
	return b.b.GetMappedRange(offset, size)
}

func (b *Buffer) Unmap() { b.b.Unmap() }

func (b *Buffer) Destroy() { b.b.Destroy() }

func (q *Queue) WriteBuffer(buffer *Buffer, offset uint64, data []byte) error {
	if buffer == nil {
		return errors.New("gpu: write buffer target is nil")
	}
	return q.b.WriteBuffer(buffer, offset, data)
}

func (e *CommandEncoder) CopyBufferToBuffer(src *Buffer, srcOffset uint64, dst *Buffer, dstOffset uint64, size uint64) error {
	if src == nil || dst == nil {
		return errors.New("gpu: copy buffer to buffer needs both a source and a destination")
	}
	return e.b.CopyBufferToBuffer(src, srcOffset, dst, dstOffset, size)
}
