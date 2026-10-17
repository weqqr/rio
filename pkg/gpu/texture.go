package gpu

import "errors"

type TextureDimension int

const (
	TextureDimension1D TextureDimension = iota
	TextureDimension2D
	TextureDimension3D
)

type TextureUsage uint32

const (
	TextureUsageNone             TextureUsage = 0
	TextureUsageCopySrc          TextureUsage = 1 << 0
	TextureUsageCopyDst          TextureUsage = 1 << 1
	TextureUsageTextureBinding   TextureUsage = 1 << 2
	TextureUsageStorageBinding   TextureUsage = 1 << 3
	TextureUsageRenderAttachment TextureUsage = 1 << 4
)

type Extent3D struct {
	Width              uint32
	Height             uint32
	DepthOrArrayLayers uint32
}

type TextureDescriptor struct {
	Label         string
	Size          Extent3D
	MipLevelCount uint32
	SampleCount   uint32
	Dimension     TextureDimension
	Format        TextureFormat
	Usage         TextureUsage
}

type Texture struct {
	b TextureBackend
}

func (d *Device) CreateTexture(desc *TextureDescriptor) (*Texture, error) {
	if desc == nil {
		return nil, errors.New("gpu: texture descriptor is nil")
	}
	if desc.Size.Width == 0 || desc.Size.Height == 0 {
		return nil, errors.New("gpu: texture width and height must be greater than zero")
	}
	if desc.Size.DepthOrArrayLayers == 0 {
		return nil, errors.New("gpu: texture depth or array layer count must be greater than zero")
	}
	if desc.MipLevelCount == 0 {
		return nil, errors.New("gpu: texture requires at least one mip level")
	}
	if desc.SampleCount > 1 {
		return nil, errors.New("gpu: multisampled textures are not implemented")
	}
	if desc.Dimension != TextureDimension2D {
		return nil, errors.New("gpu: only two-dimensional textures are implemented")
	}
	if desc.Format == TextureFormatUndefined {
		return nil, errors.New("gpu: texture format is undefined")
	}
	if desc.Usage == TextureUsageNone {
		return nil, errors.New("gpu: texture requires at least one usage")
	}
	impl, err := d.b.CreateTexture(desc)
	if err != nil {
		return nil, err
	}
	return &Texture{b: impl}, nil
}

type TextureViewDimension int

const (
	TextureViewDimension2D TextureViewDimension = iota
	TextureViewDimension2DArray
)

type TextureViewDescriptor struct {
	Label         string
	Format        TextureFormat
	Dimension     TextureViewDimension
	BaseMipLevel  uint32
	MipLevelCount uint32
}

func (t *Texture) CreateView(desc *TextureViewDescriptor) (*TextureView, error) {
	impl, err := t.b.CreateView(desc)
	if err != nil {
		return nil, err
	}
	return &TextureView{b: impl}, nil
}

func (t *Texture) Destroy() { t.b.Destroy() }

func (t *Texture) Backend() TextureBackend { return t.b }

type Origin3D struct {
	X uint32
	Y uint32
	Z uint32
}

type ImageCopyTexture struct {
	Texture  *Texture
	MipLevel uint32
	Origin   Origin3D
}

type ImageDataLayout struct {
	Offset       uint64
	BytesPerRow  uint32
	RowsPerImage uint32
}

func (q *Queue) WriteTexture(dst *ImageCopyTexture, data []byte, layout *ImageDataLayout, size Extent3D) error {
	if dst == nil || dst.Texture == nil {
		return errors.New("gpu: write texture target is nil")
	}
	if layout == nil {
		return errors.New("gpu: write texture layout is nil")
	}
	if size.Width == 0 || size.Height == 0 || size.DepthOrArrayLayers == 0 {
		return errors.New("gpu: write texture size must be non-zero")
	}
	return q.b.WriteTexture(dst, data, layout, size)
}
