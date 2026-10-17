package gpu

import (
	"errors"
	"fmt"
	"runtime"
	"sort"
	"unsafe"

	"github.com/weqqr/rio/pkg/com"
	"github.com/weqqr/rio/pkg/sys/d3d12"
	"github.com/weqqr/rio/pkg/sys/dxgi"
)

const (
	textureDataPlacementAlignment = 512
	textureDataPitchAlignment     = 256
	maxBindGroups                 = 4
)

type d3d12Texture struct {
	device        *d3d12Device
	resource      d3d12.ID3D12Resource
	format        dxgi.DXGI_FORMAT
	usage         TextureUsage
	state         d3d12.D3D12_RESOURCE_STATES
	mipLevelCount uint32
	width         uint32
	height        uint32
}

func (d *d3d12Device) CreateTexture(desc *TextureDescriptor) (textureBackend, error) {
	if desc.Size.DepthOrArrayLayers != 1 {
		return nil, errors.New("gpu: array textures are not implemented by the d3d12 backend")
	}
	if desc.MipLevelCount > 0xFFFF {
		return nil, errors.New("gpu: texture mip level count exceeds the d3d12 limit")
	}
	if desc.Usage&TextureUsageStorageBinding != 0 {
		return nil, errors.New("gpu: storage texture bindings are not implemented by the d3d12 backend")
	}
	format, err := dxgiFormat(desc.Format)
	if err != nil {
		return nil, err
	}
	flags := d3d12.D3D12_RESOURCE_FLAG_NONE
	if desc.Usage&TextureUsageRenderAttachment != 0 {
		flags |= d3d12.D3D12_RESOURCE_FLAG_ALLOW_RENDER_TARGET
	}
	initialState := d3d12.D3D12_RESOURCE_STATE_COMMON
	switch {
	case desc.Usage&TextureUsageCopyDst != 0:
		initialState = d3d12.D3D12_RESOURCE_STATE_COPY_DEST
	case desc.Usage&TextureUsageTextureBinding != 0:
		initialState = d3d12.D3D12_RESOURCE_STATE_PIXEL_SHADER_RESOURCE
	}

	heap := d3d12.D3D12_HEAP_PROPERTIES{Type: d3d12.D3D12_HEAP_TYPE_DEFAULT}
	resourceDesc := d3d12.D3D12_RESOURCE_DESC{
		Dimension:        d3d12.D3D12_RESOURCE_DIMENSION_TEXTURE2D,
		Width:            uint64(desc.Size.Width),
		Height:           desc.Size.Height,
		DepthOrArraySize: 1,
		MipLevels:        uint16(desc.MipLevelCount),
		Format:           format,
		SampleDesc:       dxgi.DXGI_SAMPLE_DESC{Count: 1},
		Layout:           d3d12.D3D12_TEXTURE_LAYOUT_UNKNOWN,
		Flags:            flags,
	}
	texture := &d3d12Texture{
		device:        d,
		format:        format,
		usage:         desc.Usage,
		state:         initialState,
		mipLevelCount: desc.MipLevelCount,
		width:         desc.Size.Width,
		height:        desc.Size.Height,
	}
	if hr := d.device.CreateCommittedResource(&heap, d3d12.D3D12_HEAP_FLAG_NONE, &resourceDesc, initialState, nil, &texture.resource); hr.Failed() {
		return nil, fmt.Errorf("gpu: create texture: %w", hr.Err())
	}
	return texture, nil
}

func (t *d3d12Texture) CreateView(desc *TextureViewDescriptor) (viewBackend, error) {
	baseMipLevel := uint32(0)
	mipLevelCount := t.mipLevelCount
	format := t.format
	if desc != nil {
		if desc.Dimension == TextureViewDimension2DArray {
			return nil, errors.New("gpu: array texture views are not implemented by the d3d12 backend")
		}
		if desc.Format != TextureFormatUndefined {
			viewFormat, err := dxgiFormat(desc.Format)
			if err != nil {
				return nil, err
			}
			if viewFormat != t.format {
				return nil, errors.New("gpu: texture view format must match the texture format")
			}
		}
		baseMipLevel = desc.BaseMipLevel
		if desc.MipLevelCount != 0 {
			mipLevelCount = desc.MipLevelCount
		}
	}
	if baseMipLevel >= t.mipLevelCount || mipLevelCount == 0 || uint64(baseMipLevel)+uint64(mipLevelCount) > uint64(t.mipLevelCount) {
		return nil, errors.New("gpu: texture view mip range is out of bounds")
	}

	view := &d3d12View{
		resource:      t.resource.Object,
		format:        format,
		baseMipLevel:  baseMipLevel,
		mipLevelCount: mipLevelCount,
		texture:       t,
	}
	if t.usage&TextureUsageRenderAttachment != 0 {
		heapDesc := d3d12.D3D12_DESCRIPTOR_HEAP_DESC{
			Type:           d3d12.D3D12_DESCRIPTOR_HEAP_TYPE_RTV,
			NumDescriptors: 1,
		}
		if hr := t.device.device.CreateDescriptorHeap(&heapDesc, &view.rtvHeap); hr.Failed() {
			return nil, hr.Err()
		}
		view.handle = view.rtvHeap.GetCPUDescriptorHandleForHeapStart()
		t.device.device.CreateRenderTargetView(t.resource.Object, nil, view.handle)
		view.ownsHeap = true
	}
	return view, nil
}

func (t *d3d12Texture) Destroy() {
	if t.resource.Valid() {
		t.resource.Release()
		t.resource = d3d12.ID3D12Resource{}
	}
}

type d3d12Sampler struct {
	desc d3d12.D3D12_SAMPLER_DESC
}

func (d *d3d12Device) CreateSampler(desc *SamplerDescriptor) (samplerBackend, error) {
	samplerDesc, err := samplerDescriptor(desc)
	if err != nil {
		return nil, err
	}
	return &d3d12Sampler{desc: samplerDesc}, nil
}

func (s *d3d12Sampler) Destroy() {}

type d3d12BindGroupLayout struct {
	entries         []BindGroupLayoutEntry
	entryByBinding  map[uint32]BindGroupLayoutEntry
	textureRegister map[uint32]uint32
	samplerRegister map[uint32]uint32
	hasTexture      bool
	hasSampler      bool
	textureCount    uint32
	samplerCount    uint32
}

func (d *d3d12Device) CreateBindGroupLayout(desc *BindGroupLayoutDescriptor) (bindGroupLayoutBackend, error) {
	layout := &d3d12BindGroupLayout{
		entries:         append([]BindGroupLayoutEntry(nil), desc.Entries...),
		entryByBinding:  make(map[uint32]BindGroupLayoutEntry, len(desc.Entries)),
		textureRegister: make(map[uint32]uint32),
		samplerRegister: make(map[uint32]uint32),
	}
	sort.Slice(layout.entries, func(i, j int) bool {
		return layout.entries[i].Binding < layout.entries[j].Binding
	})
	for _, entry := range layout.entries {
		if _, exists := layout.entryByBinding[entry.Binding]; exists {
			return nil, errors.New("gpu: bind group layout entry binding is duplicated")
		}
		layout.entryByBinding[entry.Binding] = entry
		switch entry.Type {
		case BindingTypeTexture:
			layout.hasTexture = true
			layout.textureRegister[entry.Binding] = layout.textureCount
			layout.textureCount++
		case BindingTypeSampler:
			layout.hasSampler = true
			layout.samplerRegister[entry.Binding] = layout.samplerCount
			layout.samplerCount++
		default:
			return nil, errors.New("gpu: bind group layout entry has an unsupported binding type")
		}
		if _, err := shaderVisibility(entry.Visibility); err != nil {
			return nil, err
		}
	}
	return layout, nil
}

func (l *d3d12BindGroupLayout) Destroy() {}

type d3d12DescriptorHeapInfo struct {
	heap     d3d12.ID3D12DescriptorHeap
	cpuStart d3d12.D3D12_CPU_DESCRIPTOR_HANDLE
	gpuStart d3d12.D3D12_GPU_DESCRIPTOR_HANDLE
	stride   uint32
}

func (d *d3d12Device) createShaderVisibleHeap(heapType d3d12.D3D12_DESCRIPTOR_HEAP_TYPE, count uint32) (d3d12DescriptorHeapInfo, error) {
	heapDesc := d3d12.D3D12_DESCRIPTOR_HEAP_DESC{
		Type:           heapType,
		NumDescriptors: count,
		Flags:          d3d12.D3D12_DESCRIPTOR_HEAP_FLAG_SHADER_VISIBLE,
	}
	var heap d3d12.ID3D12DescriptorHeap
	if hr := d.device.CreateDescriptorHeap(&heapDesc, &heap); hr.Failed() {
		return d3d12DescriptorHeapInfo{}, fmt.Errorf("gpu: create descriptor heap: %w", hr.Err())
	}
	return d3d12DescriptorHeapInfo{
		heap:     heap,
		cpuStart: heap.GetCPUDescriptorHandleForHeapStart(),
		gpuStart: heap.GetGPUDescriptorHandleForHeapStart(),
		stride:   d.device.GetDescriptorHandleIncrementSize(heapType),
	}, nil
}

type d3d12BindGroup struct {
	device      *d3d12Device
	layout      *d3d12BindGroupLayout
	srvHeap     d3d12DescriptorHeapInfo
	samplerHeap d3d12DescriptorHeapInfo
}

func (d *d3d12Device) CreateBindGroup(desc *BindGroupDescriptor) (bindGroupBackend, error) {
	layout, ok := desc.Layout.b.(*d3d12BindGroupLayout)
	if !ok {
		return nil, errors.New("gpu: bind group layout is not a d3d12 layout")
	}

	seen := make(map[uint32]struct{}, len(desc.Entries))
	for i := range desc.Entries {
		entry := &desc.Entries[i]
		layoutEntry, exists := layout.entryByBinding[entry.Binding]
		if !exists {
			return nil, fmt.Errorf("gpu: bind group entry %d is not declared by the layout", entry.Binding)
		}
		if _, exists := seen[entry.Binding]; exists {
			return nil, fmt.Errorf("gpu: bind group entry %d is duplicated", entry.Binding)
		}
		seen[entry.Binding] = struct{}{}
		switch layoutEntry.Type {
		case BindingTypeTexture:
			if entry.Resource.TextureView == nil {
				return nil, fmt.Errorf("gpu: bind group entry %d requires a texture view", entry.Binding)
			}
		case BindingTypeSampler:
			if entry.Resource.Sampler == nil {
				return nil, fmt.Errorf("gpu: bind group entry %d requires a sampler", entry.Binding)
			}
		}
	}
	for _, entry := range layout.entries {
		if _, exists := seen[entry.Binding]; !exists {
			return nil, fmt.Errorf("gpu: bind group is missing an entry for binding %d", entry.Binding)
		}
	}

	group := &d3d12BindGroup{device: d, layout: layout}
	if layout.hasTexture {
		heap, err := d.createShaderVisibleHeap(d3d12.D3D12_DESCRIPTOR_HEAP_TYPE_CBV_SRV_UAV, layout.textureCount)
		if err != nil {
			return nil, err
		}
		group.srvHeap = heap
	}
	if layout.hasSampler {
		heap, err := d.createShaderVisibleHeap(d3d12.D3D12_DESCRIPTOR_HEAP_TYPE_SAMPLER, layout.samplerCount)
		if err != nil {
			group.Destroy()
			return nil, err
		}
		group.samplerHeap = heap
	}

	for i := range desc.Entries {
		entry := &desc.Entries[i]
		layoutEntry := layout.entryByBinding[entry.Binding]
		switch layoutEntry.Type {
		case BindingTypeTexture:
			view, ok := entry.Resource.TextureView.b.(*d3d12View)
			if !ok {
				group.Destroy()
				return nil, errors.New("gpu: texture view is not a d3d12 texture view")
			}
			if view.texture != nil && view.texture.usage&TextureUsageTextureBinding == 0 {
				group.Destroy()
				return nil, errors.New("gpu: texture was not created with texture-binding usage")
			}
			handle := group.srvHeap.cpuStart
			handle.Ptr += uintptr(layout.textureRegister[entry.Binding]) * uintptr(group.srvHeap.stride)
			if err := d.createSRV(view, handle); err != nil {
				group.Destroy()
				return nil, err
			}
		case BindingTypeSampler:
			sampler, ok := entry.Resource.Sampler.b.(*d3d12Sampler)
			if !ok {
				group.Destroy()
				return nil, errors.New("gpu: sampler is not a d3d12 sampler")
			}
			handle := group.samplerHeap.cpuStart
			handle.Ptr += uintptr(layout.samplerRegister[entry.Binding]) * uintptr(group.samplerHeap.stride)
			d.device.CreateSampler(&sampler.desc, handle)
		}
	}
	return group, nil
}

func (g *d3d12BindGroup) Destroy() {
	if g.srvHeap.heap.Valid() {
		g.srvHeap.heap.Release()
		g.srvHeap = d3d12DescriptorHeapInfo{}
	}
	if g.samplerHeap.heap.Valid() {
		g.samplerHeap.heap.Release()
		g.samplerHeap = d3d12DescriptorHeapInfo{}
	}
}

func (d *d3d12Device) createSRV(view *d3d12View, handle d3d12.D3D12_CPU_DESCRIPTOR_HANDLE) error {
	if !view.resource.Valid() {
		return errors.New("gpu: texture view has no resource")
	}
	desc := d3d12.D3D12_SHADER_RESOURCE_VIEW_DESC{
		Format:                  int32(view.format),
		ViewDimension:           d3d12.D3D12_SRV_DIMENSION_TEXTURE2D,
		Shader4ComponentMapping: d3d12.D3D12_DEFAULT_SHADER_4_COMPONENT_MAPPING,
		Texture2D: d3d12.D3D12_TEX2D_SRV{
			MostDetailedMip: view.baseMipLevel,
			MipLevels:       view.mipLevelCount,
		},
	}
	d.device.CreateShaderResourceView(view.resource, &desc, handle)
	return nil
}

type d3d12RootBinding struct {
	param uint32
	entry BindGroupLayoutEntry
}

type d3d12PipelineLayout struct {
	signature d3d12.ID3D12RootSignature
	groups    []*d3d12BindGroupLayout
	bindings  [][]d3d12RootBinding
}

func (d *d3d12Device) CreatePipelineLayout(desc *PipelineLayoutDescriptor) (pipelineLayoutBackend, error) {
	layout := &d3d12PipelineLayout{
		groups:   make([]*d3d12BindGroupLayout, len(desc.BindGroupLayouts)),
		bindings: make([][]d3d12RootBinding, len(desc.BindGroupLayouts)),
	}
	var ranges []d3d12.D3D12_DESCRIPTOR_RANGE
	rangeStarts := make([]int, len(desc.BindGroupLayouts))
	for i, bindGroupLayout := range desc.BindGroupLayouts {
		if bindGroupLayout == nil {
			return nil, errors.New("gpu: pipeline layout has a nil bind group layout")
		}
		group, ok := bindGroupLayout.b.(*d3d12BindGroupLayout)
		if !ok {
			return nil, errors.New("gpu: bind group layout is not a d3d12 layout")
		}
		layout.groups[i] = group
		rangeStarts[i] = len(ranges)
		for _, entry := range group.entries {
			rangeType := d3d12.D3D12_DESCRIPTOR_RANGE_TYPE_SRV
			baseRegister := group.textureRegister[entry.Binding]
			if entry.Type == BindingTypeSampler {
				rangeType = d3d12.D3D12_DESCRIPTOR_RANGE_TYPE_SAMPLER
				baseRegister = group.samplerRegister[entry.Binding]
			}
			ranges = append(ranges, d3d12.D3D12_DESCRIPTOR_RANGE{
				RangeType:          rangeType,
				NumDescriptors:     1,
				BaseShaderRegister: baseRegister,
				RegisterSpace:      uint32(i),
			})
		}
	}

	var params []d3d12.D3D12_ROOT_PARAMETER
	paramIndex := uint32(0)
	for i, group := range layout.groups {
		rangeIndex := rangeStarts[i]
		for _, entry := range group.entries {
			visibility, err := shaderVisibility(entry.Visibility)
			if err != nil {
				return nil, err
			}
			params = append(params, d3d12.D3D12_ROOT_PARAMETER{
				ParameterType: d3d12.D3D12_ROOT_PARAMETER_TYPE_DESCRIPTOR_TABLE,
				DescriptorTable: d3d12.D3D12_ROOT_DESCRIPTOR_TABLE{
					NumDescriptorRanges: 1,
					DescriptorRanges:    &ranges[rangeIndex],
				},
				ShaderVisibility: visibility,
			})
			layout.bindings[i] = append(layout.bindings[i], d3d12RootBinding{param: paramIndex, entry: entry})
			rangeIndex++
			paramIndex++
		}
	}

	signature, err := d.createRootSignature(params, ranges)
	if err != nil {
		return nil, err
	}
	layout.signature = signature
	return layout, nil
}

func (l *d3d12PipelineLayout) Destroy() {
	if l.signature.Valid() {
		l.signature.Release()
		l.signature = d3d12.ID3D12RootSignature{}
	}
}

func (d *d3d12Device) createRootSignature(params []d3d12.D3D12_ROOT_PARAMETER, ranges []d3d12.D3D12_DESCRIPTOR_RANGE) (d3d12.ID3D12RootSignature, error) {
	desc := d3d12.D3D12_ROOT_SIGNATURE_DESC{
		Flags: d3d12.D3D12_ROOT_SIGNATURE_FLAG_ALLOW_INPUT_ASSEMBLER_INPUT_LAYOUT,
	}
	if len(params) > 0 {
		desc.NumParameters = uint32(len(params))
		desc.Parameters = unsafe.Pointer(&params[0])
	}
	var blob, errorBlob com.Object
	hr := d3d12.D3D12SerializeRootSignature(&desc, d3d12.D3D_ROOT_SIGNATURE_VERSION_1, &blob, &errorBlob)
	if errorBlob.Valid() {
		message := blobString(errorBlob)
		errorBlob.Release()
		if hr.Failed() {
			return d3d12.ID3D12RootSignature{}, fmt.Errorf("gpu: serialize root signature: %s (%s)", message, hr)
		}
	}
	if hr.Failed() {
		return d3d12.ID3D12RootSignature{}, fmt.Errorf("gpu: serialize root signature: %w", hr.Err())
	}
	if !blob.Valid() {
		return d3d12.ID3D12RootSignature{}, errors.New("gpu: root signature serializer returned no blob")
	}

	serialized := d3d12.ID3DBlob{Object: blob}
	defer serialized.Release()

	var signature d3d12.ID3D12RootSignature
	if hr := d.device.CreateRootSignature(0, serialized.GetBufferPointer(), serialized.GetBufferSize(), &signature); hr.Failed() {
		return d3d12.ID3D12RootSignature{}, fmt.Errorf("gpu: create root signature: %w", hr.Err())
	}
	runtime.KeepAlive(params)
	runtime.KeepAlive(ranges)
	return signature, nil
}

func (q *d3d12Queue) WriteTexture(dst *ImageCopyTexture, data []byte, layout *ImageDataLayout, size Extent3D) error {
	texture, ok := dst.Texture.b.(*d3d12Texture)
	if !ok {
		return errors.New("gpu: write texture target is not a d3d12 texture")
	}
	if texture.usage&TextureUsageCopyDst == 0 {
		return errors.New("gpu: texture was not created with copy-destination usage")
	}
	if size.DepthOrArrayLayers != 1 || dst.Origin.Z != 0 {
		return errors.New("gpu: d3d12 write texture supports a single layer")
	}
	if dst.MipLevel >= texture.mipLevelCount {
		return errors.New("gpu: write texture mip level is out of bounds")
	}
	mipWidth := max(uint32(1), texture.width>>dst.MipLevel)
	mipHeight := max(uint32(1), texture.height>>dst.MipLevel)
	if dst.Origin.X+size.Width > mipWidth || dst.Origin.Y+size.Height > mipHeight {
		return errors.New("gpu: write texture region is out of bounds")
	}
	bytesPerPixel, err := textureBytesPerPixel(texture.format)
	if err != nil {
		return err
	}
	rowBytes := uint64(size.Width) * uint64(bytesPerPixel)
	userRowPitch := uint64(layout.BytesPerRow)
	if userRowPitch == 0 {
		userRowPitch = rowBytes
	}
	if userRowPitch < rowBytes {
		return fmt.Errorf("gpu: write texture row pitch %d is smaller than %d", userRowPitch, rowBytes)
	}
	rowsPerImage := uint64(layout.RowsPerImage)
	if rowsPerImage == 0 {
		rowsPerImage = uint64(size.Height)
	}
	if rowsPerImage < uint64(size.Height) {
		return errors.New("gpu: write texture rows per image is smaller than the image height")
	}
	dataEnd := layout.Offset + (uint64(size.Height)-1)*userRowPitch + rowBytes
	if dataEnd > uint64(len(data)) {
		return errors.New("gpu: write texture data is too small for the given layout")
	}
	rowPitch := alignUp(userRowPitch, textureDataPitchAlignment)
	uploadSize := alignUp(rowPitch*uint64(size.Height), textureDataPlacementAlignment)

	upload, err := q.device.CreateBuffer(&BufferDescriptor{
		Label:            "texture upload",
		Size:             uploadSize,
		Usage:            BufferUsageMapWrite | BufferUsageCopySrc,
		MappedAtCreation: true,
	})
	if err != nil {
		return fmt.Errorf("gpu: create texture upload buffer: %w", err)
	}
	uploadImpl, ok := upload.(*d3d12Buffer)
	if !ok {
		return errors.New("gpu: texture upload buffer is not a d3d12 buffer")
	}
	defer uploadImpl.Destroy()
	mapped, err := uploadImpl.GetMappedRange(0, uploadSize)
	if err != nil {
		return err
	}
	for row := uint64(0); row < uint64(size.Height); row++ {
		src := layout.Offset + row*userRowPitch
		dst := row * rowPitch
		copy(mapped[dst:dst+rowBytes], data[src:src+rowBytes])
	}
	uploadImpl.Unmap()

	encoderImpl, err := q.device.CreateCommandEncoder()
	if err != nil {
		return fmt.Errorf("gpu: write texture encoder: %w", err)
	}
	encoder := encoderImpl.(*d3d12Encoder)
	if texture.state != d3d12.D3D12_RESOURCE_STATE_COPY_DEST {
		encoder.transition(texture.resource.Object, texture.state, d3d12.D3D12_RESOURCE_STATE_COPY_DEST)
	}
	destination := d3d12.D3D12_TEXTURE_COPY_LOCATION{
		Resource: texture.resource.Ptr(),
		Type:     d3d12.D3D12_TEXTURE_COPY_TYPE_SUBRESOURCE_INDEX,
	}
	destination.SetSubresourceIndex(dst.MipLevel)
	source := d3d12.D3D12_TEXTURE_COPY_LOCATION{
		Resource: uploadImpl.resource.Ptr(),
		Type:     d3d12.D3D12_TEXTURE_COPY_TYPE_PLACED_FOOTPRINT,
		PlacedFootprint: d3d12.D3D12_PLACED_SUBRESOURCE_FOOTPRINT{
			Footprint: d3d12.D3D12_SUBRESOURCE_FOOTPRINT{
				Format:   int32(texture.format),
				Width:    size.Width,
				Height:   size.Height,
				Depth:    1,
				RowPitch: uint32(rowPitch),
			},
		},
	}
	encoder.device.list.CopyTextureRegion(&destination, dst.Origin.X, dst.Origin.Y, 0, &source, nil)
	finalState := d3d12.D3D12_RESOURCE_STATE_COMMON
	if texture.usage&TextureUsageTextureBinding != 0 {
		finalState = d3d12.D3D12_RESOURCE_STATE_PIXEL_SHADER_RESOURCE
	}
	if finalState != d3d12.D3D12_RESOURCE_STATE_COPY_DEST {
		encoder.transition(texture.resource.Object, d3d12.D3D12_RESOURCE_STATE_COPY_DEST, finalState)
	}
	texture.state = finalState
	commands, err := encoder.Finish()
	if err != nil {
		return fmt.Errorf("gpu: write texture commands: %w", err)
	}
	if err := q.Submit([]*CommandBuffer{{b: commands}}); err != nil {
		return fmt.Errorf("gpu: submit texture upload: %w", err)
	}
	return nil
}

func alignUp(value, alignment uint64) uint64 {
	return (value + alignment - 1) &^ (alignment - 1)
}

func textureStateAfterUse(usage TextureUsage) d3d12.D3D12_RESOURCE_STATES {
	if usage&TextureUsageTextureBinding != 0 {
		return d3d12.D3D12_RESOURCE_STATE_PIXEL_SHADER_RESOURCE
	}
	return d3d12.D3D12_RESOURCE_STATE_COMMON
}

func textureBytesPerPixel(format dxgi.DXGI_FORMAT) (uint32, error) {
	switch format {
	case dxgi.DXGI_FORMAT_R8G8B8A8_UNORM, dxgi.DXGI_FORMAT_B8G8R8A8_UNORM:
		return 4, nil
	}
	return 0, fmt.Errorf("gpu: texture format %d is not supported by the d3d12 backend", int(format))
}

func shaderVisibility(stage ShaderStage) (d3d12.D3D12_SHADER_VISIBILITY, error) {
	switch stage {
	case ShaderStageVertex:
		return d3d12.D3D12_SHADER_VISIBILITY_VERTEX, nil
	case ShaderStageFragment:
		return d3d12.D3D12_SHADER_VISIBILITY_PIXEL, nil
	case ShaderStageVertex | ShaderStageFragment:
		return d3d12.D3D12_SHADER_VISIBILITY_ALL, nil
	}
	return d3d12.D3D12_SHADER_VISIBILITY_ALL, errors.New("gpu: shader stage combination is not supported by the d3d12 backend")
}

func samplerDescriptor(desc *SamplerDescriptor) (d3d12.D3D12_SAMPLER_DESC, error) {
	addressU, err := addressMode(desc.AddressModeU)
	if err != nil {
		return d3d12.D3D12_SAMPLER_DESC{}, err
	}
	addressV, err := addressMode(desc.AddressModeV)
	if err != nil {
		return d3d12.D3D12_SAMPLER_DESC{}, err
	}
	addressW, err := addressMode(desc.AddressModeW)
	if err != nil {
		return d3d12.D3D12_SAMPLER_DESC{}, err
	}
	filter, err := samplerFilter(desc)
	if err != nil {
		return d3d12.D3D12_SAMPLER_DESC{}, err
	}
	maxAnisotropy := uint32(desc.MaxAnisotropy)
	if maxAnisotropy == 0 {
		maxAnisotropy = 1
	}
	maxLod := desc.LodMaxClamp
	if maxLod == 0 {
		maxLod = 32
	}
	return d3d12.D3D12_SAMPLER_DESC{
		Filter:         filter,
		AddressU:       addressU,
		AddressV:       addressV,
		AddressW:       addressW,
		MaxAnisotropy:  maxAnisotropy,
		ComparisonFunc: d3d12.D3D12_COMPARISON_FUNC_LESS,
		MinLOD:         desc.LodMinClamp,
		MaxLOD:         maxLod,
	}, nil
}

func addressMode(mode AddressMode) (d3d12.D3D12_TEXTURE_ADDRESS_MODE, error) {
	switch mode {
	case AddressModeClampToEdge:
		return d3d12.D3D12_TEXTURE_ADDRESS_MODE_CLAMP, nil
	case AddressModeRepeat:
		return d3d12.D3D12_TEXTURE_ADDRESS_MODE_WRAP, nil
	case AddressModeMirrorRepeat:
		return d3d12.D3D12_TEXTURE_ADDRESS_MODE_MIRROR, nil
	}
	return d3d12.D3D12_TEXTURE_ADDRESS_MODE_CLAMP, fmt.Errorf("gpu: address mode %d is not supported", int(mode))
}

func samplerFilter(desc *SamplerDescriptor) (d3d12.D3D12_FILTER, error) {
	if desc.MagFilter != FilterModeNearest && desc.MagFilter != FilterModeLinear {
		return d3d12.D3D12_FILTER_MIN_MAG_MIP_POINT, fmt.Errorf("gpu: filter mode %d is not supported", int(desc.MagFilter))
	}
	if desc.MinFilter != FilterModeNearest && desc.MinFilter != FilterModeLinear {
		return d3d12.D3D12_FILTER_MIN_MAG_MIP_POINT, fmt.Errorf("gpu: filter mode %d is not supported", int(desc.MinFilter))
	}
	if desc.MipmapFilter != MipmapFilterModeNearest && desc.MipmapFilter != MipmapFilterModeLinear {
		return d3d12.D3D12_FILTER_MIN_MAG_MIP_POINT, fmt.Errorf("gpu: mipmap filter mode %d is not supported", int(desc.MipmapFilter))
	}
	if desc.MaxAnisotropy > 1 {
		return d3d12.D3D12_FILTER_ANISOTROPIC, nil
	}
	minLinear := desc.MinFilter == FilterModeLinear
	magLinear := desc.MagFilter == FilterModeLinear
	mipLinear := desc.MipmapFilter == MipmapFilterModeLinear
	switch {
	case !minLinear && !magLinear && !mipLinear:
		return d3d12.D3D12_FILTER_MIN_MAG_MIP_POINT, nil
	case !minLinear && !magLinear && mipLinear:
		return d3d12.D3D12_FILTER_MIN_MAG_POINT_MIP_LINEAR, nil
	case !minLinear && magLinear && !mipLinear:
		return d3d12.D3D12_FILTER_MIN_POINT_MAG_LINEAR_MIP_POINT, nil
	case !minLinear && magLinear && mipLinear:
		return d3d12.D3D12_FILTER_MIN_POINT_MAG_MIP_LINEAR, nil
	case minLinear && !magLinear && !mipLinear:
		return d3d12.D3D12_FILTER_MIN_LINEAR_MAG_MIP_POINT, nil
	case minLinear && !magLinear && mipLinear:
		return d3d12.D3D12_FILTER_MIN_LINEAR_MAG_POINT_MIP_LINEAR, nil
	case minLinear && magLinear && !mipLinear:
		return d3d12.D3D12_FILTER_MIN_MAG_LINEAR_MIP_POINT, nil
	}
	return d3d12.D3D12_FILTER_MIN_MAG_MIP_LINEAR, nil
}
