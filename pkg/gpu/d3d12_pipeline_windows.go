package gpu

import (
	"errors"
	"fmt"
	"math"
	"runtime"
	"sort"
	"syscall"
	"unsafe"

	"github.com/weqqr/rio/pkg/com"
	"github.com/weqqr/rio/pkg/sys/d3d12"
	"github.com/weqqr/rio/pkg/sys/d3dcompiler"
	"github.com/weqqr/rio/pkg/sys/dxgi"
)

const maxColorTargets = 8

type d3d12ShaderModule struct {
	code []byte
}

func (d *d3d12Device) CreateShaderModule(desc *ShaderModuleDescriptor) (shaderModuleBackend, error) {
	if len(desc.Code) == 0 {
		return nil, errors.New("gpu: shader module source is empty")
	}
	return &d3d12ShaderModule{code: append([]byte(nil), desc.Code...)}, nil
}

func (m *d3d12ShaderModule) Code() []byte { return m.code }

func (m *d3d12ShaderModule) Destroy() {}

type d3d12Pipeline struct {
	state         d3d12.ID3D12PipelineState
	strides       []uint64
	layout        *d3d12PipelineLayout
	rootSignature d3d12.ID3D12RootSignature
}

func (d *d3d12Device) CreateRenderPipeline(desc *RenderPipelineDescriptor) (renderPipelineBackend, error) {
	if desc.Vertex.Module == nil {
		return nil, errors.New("gpu: vertex shader module is missing")
	}
	vertexModule, ok := desc.Vertex.Module.b.(*d3d12ShaderModule)
	if !ok {
		return nil, errors.New("gpu: vertex shader module is not a d3d12 shader module")
	}
	vertexCode, err := compileShader(vertexModule.code, desc.Vertex.EntryPoint, "vs_5_0")
	if err != nil {
		return nil, err
	}

	var pixelCode []byte
	if desc.Fragment != nil {
		if desc.Fragment.Module == nil {
			return nil, errors.New("gpu: fragment shader module is missing")
		}
		pixelModule, ok := desc.Fragment.Module.b.(*d3d12ShaderModule)
		if !ok {
			return nil, errors.New("gpu: fragment shader module is not a d3d12 shader module")
		}
		pixelCode, err = compileShader(pixelModule.code, desc.Fragment.EntryPoint, "ps_5_0")
		if err != nil {
			return nil, err
		}
		if len(desc.Fragment.Targets) == 0 {
			return nil, errors.New("gpu: fragment state requires at least one color target")
		}
		if len(desc.Fragment.Targets) > maxColorTargets {
			return nil, fmt.Errorf("gpu: %d color targets exceed the d3d12 limit", len(desc.Fragment.Targets))
		}
	}

	var rootSignature d3d12.ID3D12RootSignature
	var pipelineLayout *d3d12PipelineLayout
	if desc.Layout != nil {
		var ok bool
		pipelineLayout, ok = desc.Layout.b.(*d3d12PipelineLayout)
		if !ok {
			return nil, errors.New("gpu: pipeline layout is not a d3d12 pipeline layout")
		}
		rootSignature = pipelineLayout.signature
	} else {
		var err error
		rootSignature, err = d.ensureRootSignature()
		if err != nil {
			return nil, err
		}
	}

	inputElements, semantic, strides, err := buildInputLayout(desc.Vertex.Buffers)
	if err != nil {
		return nil, err
	}

	cullMode, err := d3d12CullMode(desc.Primitive.CullMode)
	if err != nil {
		return nil, err
	}
	switch desc.Primitive.Topology {
	case PrimitiveTopologyTriangleList:
	default:
		return nil, fmt.Errorf("gpu: primitive topology %d is not supported by the d3d12 backend", int(desc.Primitive.Topology))
	}

	sampleCount := desc.Multisample.Count
	if sampleCount == 0 {
		sampleCount = 1
	}
	if sampleCount != 1 {
		return nil, fmt.Errorf("gpu: sample count %d is not supported by the d3d12 backend", sampleCount)
	}
	sampleMask := desc.Multisample.Mask
	if sampleMask == 0 {
		sampleMask = math.MaxUint32
	}

	pipelineDesc := d3d12.D3D12_GRAPHICS_PIPELINE_STATE_DESC{
		RootSignature: rootSignature.Ptr(),
		VS: d3d12.D3D12_SHADER_BYTECODE{
			Bytecode:   unsafe.Pointer(&vertexCode[0]),
			ByteLength: uintptr(len(vertexCode)),
		},
		BlendState: defaultBlendState(),
		SampleMask: sampleMask,
		RasterizerState: d3d12.D3D12_RASTERIZER_DESC{
			FillMode:              d3d12.D3D12_FILL_MODE_SOLID,
			CullMode:              cullMode,
			FrontCounterClockwise: boolToUint32(desc.Primitive.FrontFace == FrontFaceCCW),
			DepthClipEnable:       1,
		},
		DepthStencilState: d3d12.D3D12_DEPTH_STENCIL_DESC{
			DepthEnable:      0,
			DepthWriteMask:   d3d12.D3D12_DEPTH_WRITE_MASK_ZERO,
			DepthFunc:        d3d12.D3D12_COMPARISON_FUNC_LESS,
			StencilReadMask:  0xFF,
			StencilWriteMask: 0xFF,
			FrontFace: d3d12.D3D12_DEPTH_STENCILOP_DESC{
				StencilFailOp:      d3d12.D3D12_STENCIL_OP_KEEP,
				StencilDepthFailOp: d3d12.D3D12_STENCIL_OP_KEEP,
				StencilPassOp:      d3d12.D3D12_STENCIL_OP_KEEP,
				StencilFunc:        d3d12.D3D12_COMPARISON_FUNC_LESS,
			},
			BackFace: d3d12.D3D12_DEPTH_STENCILOP_DESC{
				StencilFailOp:      d3d12.D3D12_STENCIL_OP_KEEP,
				StencilDepthFailOp: d3d12.D3D12_STENCIL_OP_KEEP,
				StencilPassOp:      d3d12.D3D12_STENCIL_OP_KEEP,
				StencilFunc:        d3d12.D3D12_COMPARISON_FUNC_LESS,
			},
		},
		IBStripCutValue:       d3d12.D3D12_INDEX_BUFFER_STRIP_CUT_VALUE_DISABLED,
		PrimitiveTopologyType: d3d12.D3D12_PRIMITIVE_TOPOLOGY_TYPE_TRIANGLE,
		DSVFormat:             dxgi.DXGI_FORMAT_UNKNOWN,
		SampleDesc:            dxgi.DXGI_SAMPLE_DESC{Count: 1},
	}
	if len(pixelCode) > 0 {
		pipelineDesc.PS = d3d12.D3D12_SHADER_BYTECODE{
			Bytecode:   unsafe.Pointer(&pixelCode[0]),
			ByteLength: uintptr(len(pixelCode)),
		}
	}
	if len(inputElements) > 0 {
		pipelineDesc.InputLayout = d3d12.D3D12_INPUT_LAYOUT_DESC{
			InputElementDescs: &inputElements[0],
			NumElements:       uint32(len(inputElements)),
		}
	}
	if desc.Fragment != nil {
		pipelineDesc.NumRenderTargets = uint32(len(desc.Fragment.Targets))
		for i := range desc.Fragment.Targets {
			format, err := dxgiFormat(desc.Fragment.Targets[i].Format)
			if err != nil {
				return nil, err
			}
			pipelineDesc.RTVFormats[i] = format
			pipelineDesc.BlendState.RenderTarget[i].RenderTargetWriteMask = uint8(desc.Fragment.Targets[i].WriteMask)
		}
	}

	pipeline := &d3d12Pipeline{strides: strides, layout: pipelineLayout, rootSignature: rootSignature}
	if hr := d.device.CreateGraphicsPipelineState(&pipelineDesc, &pipeline.state); hr.Failed() {
		return nil, fmt.Errorf("gpu: create graphics pipeline state: %w", hr.Err())
	}
	runtime.KeepAlive(vertexCode)
	runtime.KeepAlive(pixelCode)
	runtime.KeepAlive(semantic)
	return pipeline, nil
}

func defaultBlendState() d3d12.D3D12_BLEND_DESC {
	target := d3d12.D3D12_RENDER_TARGET_BLEND_DESC{
		SrcBlend:              d3d12.D3D12_BLEND_ONE,
		DestBlend:             d3d12.D3D12_BLEND_ZERO,
		BlendOp:               d3d12.D3D12_BLEND_OP_ADD,
		SrcBlendAlpha:         d3d12.D3D12_BLEND_ONE,
		DestBlendAlpha:        d3d12.D3D12_BLEND_ZERO,
		BlendOpAlpha:          d3d12.D3D12_BLEND_OP_ADD,
		LogicOp:               d3d12.D3D12_LOGIC_OP_NOOP,
		RenderTargetWriteMask: 0xF,
	}
	state := d3d12.D3D12_BLEND_DESC{}
	for i := range state.RenderTarget {
		state.RenderTarget[i] = target
	}
	return state
}

func (p *d3d12Pipeline) Destroy() {
	if p.state.Valid() {
		p.state.Release()
		p.state = d3d12.ID3D12PipelineState{}
	}
}

func (d *d3d12Device) ensureRootSignature() (d3d12.ID3D12RootSignature, error) {
	if d.rootSignature.Valid() {
		return d.rootSignature, nil
	}
	signature, err := d.createRootSignature(nil, nil)
	if err != nil {
		return d3d12.ID3D12RootSignature{}, err
	}
	d.rootSignature = signature
	return signature, nil
}

func compileShader(source []byte, entryPoint, target string) ([]byte, error) {
	if entryPoint == "" {
		return nil, errors.New("gpu: shader entry point is empty")
	}
	sourceName, err := syscall.BytePtrFromString("shader.hlsl")
	if err != nil {
		return nil, err
	}
	entry, err := syscall.BytePtrFromString(entryPoint)
	if err != nil {
		return nil, err
	}
	targetName, err := syscall.BytePtrFromString(target)
	if err != nil {
		return nil, err
	}

	var code, errorsBlob com.Object
	hr := d3dcompiler.D3DCompile(
		unsafe.Pointer(&source[0]), uintptr(len(source)),
		sourceName, 0, 0,
		entry, targetName, 0, 0,
		&code, &errorsBlob,
	)
	if errorsBlob.Valid() {
		message := blobString(errorsBlob)
		errorsBlob.Release()
		if hr.Failed() {
			return nil, fmt.Errorf("gpu: compile %s: %s (%s)", entryPoint, message, hr)
		}
	}
	if hr.Failed() {
		return nil, hr.Err()
	}
	if !code.Valid() {
		return nil, errors.New("gpu: shader compiler returned no bytecode")
	}

	blob := d3d12.ID3DBlob{Object: code}
	defer blob.Release()
	size := blob.GetBufferSize()
	compiled := make([]byte, size)
	copy(compiled, unsafe.Slice((*byte)(blob.GetBufferPointer()), size))
	runtime.KeepAlive(source)
	if len(compiled) < 4 || string(compiled[:4]) != "DXBC" {
		return nil, fmt.Errorf("gpu: compile %s produced unexpected bytecode", entryPoint)
	}
	return compiled, nil
}

func blobString(blob com.Object) string {
	b := d3d12.ID3DBlob{Object: blob}
	size := b.GetBufferSize()
	if size == 0 {
		return ""
	}
	return string(unsafe.Slice((*byte)(b.GetBufferPointer()), size))
}

func buildInputLayout(buffers []VertexBufferLayout) ([]d3d12.D3D12_INPUT_ELEMENT_DESC, []byte, []uint64, error) {
	var elements []d3d12.D3D12_INPUT_ELEMENT_DESC
	semantic := []byte("TEXCOORD\x00")
	strides := make([]uint64, len(buffers))
	for slot, buffer := range buffers {
		if buffer.ArrayStride == 0 {
			return nil, nil, nil, fmt.Errorf("gpu: vertex buffer %d has a zero array stride", slot)
		}
		if buffer.ArrayStride > math.MaxUint32 {
			return nil, nil, nil, fmt.Errorf("gpu: vertex buffer %d array stride exceeds the d3d12 limit", slot)
		}
		strides[slot] = buffer.ArrayStride
		classification := d3d12.D3D12_INPUT_CLASSIFICATION_PER_VERTEX_DATA
		if buffer.StepMode == VertexStepModeInstance {
			classification = d3d12.D3D12_INPUT_CLASSIFICATION_PER_INSTANCE_DATA
		}
		for _, attribute := range buffer.Attributes {
			format, err := vertexFormat(attribute.Format)
			if err != nil {
				return nil, nil, nil, err
			}
			if attribute.Offset > math.MaxUint32 {
				return nil, nil, nil, fmt.Errorf("gpu: vertex attribute offset %d exceeds the d3d12 limit", attribute.Offset)
			}
			elements = append(elements, d3d12.D3D12_INPUT_ELEMENT_DESC{
				SemanticName:         &semantic[0],
				SemanticIndex:        attribute.ShaderLocation,
				Format:               format,
				InputSlot:            uint32(slot),
				AlignedByteOffset:    uint32(attribute.Offset),
				InputSlotClass:       classification,
				InstanceDataStepRate: boolToUint32(buffer.StepMode == VertexStepModeInstance),
			})
		}
	}
	return elements, semantic, strides, nil
}

func vertexFormat(format VertexFormat) (dxgi.DXGI_FORMAT, error) {
	switch format {
	case VertexFormatFloat32:
		return dxgi.DXGI_FORMAT_R32_FLOAT, nil
	case VertexFormatFloat32x2:
		return dxgi.DXGI_FORMAT_R32G32_FLOAT, nil
	case VertexFormatFloat32x3:
		return dxgi.DXGI_FORMAT_R32G32B32_FLOAT, nil
	case VertexFormatFloat32x4:
		return dxgi.DXGI_FORMAT_R32G32B32A32_FLOAT, nil
	}
	return dxgi.DXGI_FORMAT_UNKNOWN, fmt.Errorf("gpu: vertex format %d is not supported by the d3d12 backend", int(format))
}

func d3d12CullMode(mode CullMode) (d3d12.D3D12_CULL_MODE, error) {
	switch mode {
	case CullModeNone:
		return d3d12.D3D12_CULL_MODE_NONE, nil
	case CullModeFront:
		return d3d12.D3D12_CULL_MODE_FRONT, nil
	case CullModeBack:
		return d3d12.D3D12_CULL_MODE_BACK, nil
	}
	return d3d12.D3D12_CULL_MODE_NONE, fmt.Errorf("gpu: cull mode %d is not supported by the d3d12 backend", int(mode))
}

func boolToUint32(value bool) uint32 {
	if value {
		return 1
	}
	return 0
}

func (p *d3d12RenderPass) SetPipeline(pipeline *RenderPipeline) error {
	impl, ok := pipeline.b.(*d3d12Pipeline)
	if !ok {
		return errors.New("gpu: render pipeline is not a d3d12 pipeline")
	}
	p.pipeline = impl
	p.device.list.SetPipelineState(impl.state.Object)
	p.device.list.SetGraphicsRootSignature(impl.rootSignature.Object)
	return nil
}

func (p *d3d12RenderPass) SetVertexBuffer(slot uint32, buffer *Buffer) error {
	impl, ok := buffer.b.(*d3d12Buffer)
	if !ok {
		return errors.New("gpu: vertex buffer is not a d3d12 buffer")
	}
	if impl.usage&BufferUsageVertex == 0 {
		return errors.New("gpu: buffer was not created with vertex usage")
	}
	if slot >= 32 {
		return fmt.Errorf("gpu: vertex buffer slot %d is out of range", slot)
	}
	if p.vertexBuffers == nil {
		p.vertexBuffers = make(map[uint32]*d3d12Buffer)
	}
	p.vertexBuffers[slot] = impl
	return nil
}

func (p *d3d12RenderPass) SetBindGroup(index uint32, group *BindGroup) error {
	impl, ok := group.b.(*d3d12BindGroup)
	if !ok {
		return errors.New("gpu: bind group is not a d3d12 bind group")
	}
	if index >= maxBindGroups {
		return fmt.Errorf("gpu: bind group index %d is out of range", index)
	}
	if p.bindGroups == nil {
		p.bindGroups = make(map[uint32]*d3d12BindGroup)
	}
	p.bindGroups[index] = impl
	return nil
}

func (p *d3d12RenderPass) SetViewport(x, y, width, height, minDepth, maxDepth float32) error {
	if width <= 0 || height <= 0 {
		return errors.New("gpu: viewport width and height must be positive")
	}
	viewport := d3d12.D3D12_VIEWPORT{
		TopLeftX: x,
		TopLeftY: y,
		Width:    width,
		Height:   height,
		MinDepth: minDepth,
		MaxDepth: maxDepth,
	}
	p.device.list.RSSetViewports(1, &viewport)
	return nil
}

func (p *d3d12RenderPass) SetScissorRect(x, y, width, height uint32) error {
	if width == 0 || height == 0 {
		return errors.New("gpu: scissor width and height must be positive")
	}
	if uint64(x)+uint64(width) > math.MaxInt32 || uint64(y)+uint64(height) > math.MaxInt32 {
		return errors.New("gpu: scissor rect exceeds the d3d12 coordinate range")
	}
	rect := d3d12.D3D12_RECT{
		Left:   int32(x),
		Top:    int32(y),
		Right:  int32(x + width),
		Bottom: int32(y + height),
	}
	p.device.list.RSSetScissorRects(1, &rect)
	return nil
}

func (p *d3d12RenderPass) Draw(vertexCount, instanceCount, firstVertex, firstInstance uint32) error {
	if p.pipeline == nil {
		return errors.New("gpu: draw requires a render pipeline")
	}
	if vertexCount == 0 || instanceCount == 0 {
		return nil
	}
	if err := p.bindResourceGroups(); err != nil {
		return err
	}

	slots := make([]int, 0, len(p.vertexBuffers))
	for slot := range p.vertexBuffers {
		slots = append(slots, int(slot))
	}
	sort.Ints(slots)
	for _, value := range slots {
		slot := uint32(value)
		if int(slot) >= len(p.pipeline.strides) {
			return fmt.Errorf("gpu: vertex buffer slot %d has no layout in the pipeline", slot)
		}
		buffer := p.vertexBuffers[slot]
		view := d3d12.D3D12_VERTEX_BUFFER_VIEW{
			BufferLocation: buffer.resource.GetGPUVirtualAddress(),
			SizeInBytes:    uint32(min(buffer.size, math.MaxUint32)),
			StrideInBytes:  uint32(p.pipeline.strides[slot]),
		}
		p.device.list.IASetVertexBuffers(slot, 1, &view)
	}
	p.device.list.IASetPrimitiveTopology(d3d12.D3D_PRIMITIVE_TOPOLOGY_TRIANGLELIST)
	p.device.list.DrawInstanced(vertexCount, instanceCount, firstVertex, firstInstance)
	return nil
}

func (p *d3d12RenderPass) bindResourceGroups() error {
	if p.pipeline.layout == nil {
		return nil
	}
	type groupBinding struct {
		group    *d3d12BindGroup
		bindings []d3d12RootBinding
	}
	var groups []groupBinding
	var srvHeap, samplerHeap d3d12.ID3D12DescriptorHeap
	for i, layout := range p.pipeline.layout.groups {
		group, exists := p.bindGroups[uint32(i)]
		if !exists {
			if len(layout.entries) == 0 {
				continue
			}
			return fmt.Errorf("gpu: render pass is missing bind group %d", i)
		}
		if group.layout != layout {
			return fmt.Errorf("gpu: bind group %d uses a different layout than the pipeline", i)
		}
		if layout.hasTexture {
			if srvHeap.Valid() && srvHeap.Ptr() != group.srvHeap.heap.Ptr() {
				return errors.New("gpu: multiple bind groups with texture bindings are not supported by the d3d12 backend")
			}
			srvHeap = group.srvHeap.heap
		}
		if layout.hasSampler {
			if samplerHeap.Valid() && samplerHeap.Ptr() != group.samplerHeap.heap.Ptr() {
				return errors.New("gpu: multiple bind groups with sampler bindings are not supported by the d3d12 backend")
			}
			samplerHeap = group.samplerHeap.heap
		}
		groups = append(groups, groupBinding{group: group, bindings: p.pipeline.layout.bindings[i]})
	}

	heaps := make([]unsafe.Pointer, 0, 2)
	if srvHeap.Valid() {
		heaps = append(heaps, srvHeap.Ptr())
	}
	if samplerHeap.Valid() {
		heaps = append(heaps, samplerHeap.Ptr())
	}
	if len(heaps) > 0 {
		p.device.list.SetDescriptorHeaps(uint32(len(heaps)), &heaps[0])
	}
	for _, bindingGroup := range groups {
		for _, binding := range bindingGroup.bindings {
			var handle d3d12.D3D12_GPU_DESCRIPTOR_HANDLE
			if binding.entry.Type == BindingTypeSampler {
				handle = bindingGroup.group.samplerHeap.gpuStart
				register := bindingGroup.group.layout.samplerRegister[binding.entry.Binding]
				handle.Ptr += uint64(register) * uint64(bindingGroup.group.samplerHeap.stride)
			} else {
				handle = bindingGroup.group.srvHeap.gpuStart
				register := bindingGroup.group.layout.textureRegister[binding.entry.Binding]
				handle.Ptr += uint64(register) * uint64(bindingGroup.group.srvHeap.stride)
			}
			p.device.list.SetGraphicsRootDescriptorTable(binding.param, handle)
		}
	}
	return nil
}
