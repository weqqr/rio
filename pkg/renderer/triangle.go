package renderer

import "structs"

type Vertex struct {
	_        structs.HostLayout
	Position [2]float32
	Color    [4]float32
	UV       [2]float32
}

var triangleVertices = [...]Vertex{
	{Position: [2]float32{0.0, -0.5}, Color: [4]float32{1.0, 0.0, 0.0, 1.0}, UV: [2]float32{0.5, 0.0}},
	{Position: [2]float32{0.5, 0.5}, Color: [4]float32{0.0, 1.0, 0.0, 1.0}, UV: [2]float32{1.0, 1.0}},
	{Position: [2]float32{-0.5, 0.5}, Color: [4]float32{0.0, 0.0, 1.0, 1.0}, UV: [2]float32{0.0, 1.0}},
}

const triangleShader = `
struct VertexInput {
    float2 position : TEXCOORD0;
    float4 color    : TEXCOORD1;
    float2 uv       : TEXCOORD2;
};

struct VertexOutput {
    float4 position : SV_POSITION;
    float4 color    : COLOR0;
    float2 uv       : TEXCOORD0;
};

VertexOutput vs_main(VertexInput input) {
    VertexOutput output;
    output.position = float4(input.position, 0.0, 1.0);
    output.color = input.color;
    output.uv = input.uv;
    return output;
}

Texture2D xor_texture : register(t0);
SamplerState xor_sampler : register(s0);

float4 ps_main(VertexOutput input) : SV_TARGET {
    return xor_texture.Sample(xor_sampler, input.uv) * input.color;
}
`
