package gpu

import "errors"

type ShaderStage uint32

const (
	ShaderStageNone     ShaderStage = 0
	ShaderStageVertex   ShaderStage = 1 << 0
	ShaderStageFragment ShaderStage = 1 << 1
	ShaderStageCompute  ShaderStage = 1 << 2
)

type BindingType int

const (
	BindingTypeUndefined BindingType = iota
	BindingTypeTexture
	BindingTypeSampler
)

type BindGroupLayoutEntry struct {
	Binding    uint32
	Visibility ShaderStage
	Type       BindingType
}

type BindGroupLayoutDescriptor struct {
	Label   string
	Entries []BindGroupLayoutEntry
}

type BindGroupLayout struct {
	b BindGroupLayoutBackend
}

func (d *Device) CreateBindGroupLayout(desc *BindGroupLayoutDescriptor) (*BindGroupLayout, error) {
	if desc == nil {
		return nil, errors.New("gpu: bind group layout descriptor is nil")
	}
	seen := make(map[uint32]struct{}, len(desc.Entries))
	for i := range desc.Entries {
		entry := &desc.Entries[i]
		if entry.Type != BindingTypeTexture && entry.Type != BindingTypeSampler {
			return nil, errors.New("gpu: bind group layout entry has an unsupported binding type")
		}
		if entry.Visibility == ShaderStageNone {
			return nil, errors.New("gpu: bind group layout entry has no shader visibility")
		}
		if _, exists := seen[entry.Binding]; exists {
			return nil, errors.New("gpu: bind group layout entry binding is duplicated")
		}
		seen[entry.Binding] = struct{}{}
	}
	impl, err := d.b.CreateBindGroupLayout(desc)
	if err != nil {
		return nil, err
	}
	return &BindGroupLayout{b: impl}, nil
}

func (l *BindGroupLayout) Destroy() { l.b.Destroy() }

func (l *BindGroupLayout) Backend() BindGroupLayoutBackend { return l.b }

type BindingResource struct {
	Buffer      *Buffer
	Sampler     *Sampler
	TextureView *TextureView
}

type BindGroupEntry struct {
	Binding  uint32
	Resource BindingResource
}

type BindGroupDescriptor struct {
	Label   string
	Layout  *BindGroupLayout
	Entries []BindGroupEntry
}

type BindGroup struct {
	b BindGroupBackend
}

func (d *Device) CreateBindGroup(desc *BindGroupDescriptor) (*BindGroup, error) {
	if desc == nil {
		return nil, errors.New("gpu: bind group descriptor is nil")
	}
	if desc.Layout == nil {
		return nil, errors.New("gpu: bind group requires a layout")
	}
	impl, err := d.b.CreateBindGroup(desc)
	if err != nil {
		return nil, err
	}
	return &BindGroup{b: impl}, nil
}

func (g *BindGroup) Destroy() { g.b.Destroy() }
