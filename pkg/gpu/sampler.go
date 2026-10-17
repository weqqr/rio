package gpu

import "errors"

type AddressMode int

const (
	AddressModeClampToEdge AddressMode = iota
	AddressModeRepeat
	AddressModeMirrorRepeat
)

type FilterMode int

const (
	FilterModeNearest FilterMode = iota
	FilterModeLinear
)

type MipmapFilterMode int

const (
	MipmapFilterModeNearest MipmapFilterMode = iota
	MipmapFilterModeLinear
)

type SamplerDescriptor struct {
	Label         string
	AddressModeU  AddressMode
	AddressModeV  AddressMode
	AddressModeW  AddressMode
	MagFilter     FilterMode
	MinFilter     FilterMode
	MipmapFilter  MipmapFilterMode
	LodMinClamp   float32
	LodMaxClamp   float32
	MaxAnisotropy uint16
}

type Sampler struct {
	b samplerBackend
}

func (d *Device) CreateSampler(desc *SamplerDescriptor) (*Sampler, error) {
	if desc == nil {
		return nil, errors.New("gpu: sampler descriptor is nil")
	}
	impl, err := d.b.CreateSampler(desc)
	if err != nil {
		return nil, err
	}
	return &Sampler{b: impl}, nil
}

func (s *Sampler) Destroy() { s.b.Destroy() }
