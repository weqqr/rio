package renderer

import "structs"

type Vertex struct {
	_        structs.HostLayout
	Position [2]float32
	Color    [4]float32
}

var triangleVertices = [...]Vertex{
	{Position: [2]float32{0.0, -0.5}, Color: [4]float32{1.0, 0.0, 0.0, 1.0}},
	{Position: [2]float32{0.5, 0.5}, Color: [4]float32{0.0, 1.0, 0.0, 1.0}},
	{Position: [2]float32{-0.5, 0.5}, Color: [4]float32{0.0, 0.0, 1.0, 1.0}},
}

const triangleShader = `
struct VertexInput {
    float2 position : TEXCOORD0;
    float4 color    : TEXCOORD1;
};

struct VertexOutput {
    float4 position : SV_POSITION;
    float4 color    : COLOR0;
};

VertexOutput vs_main(VertexInput input) {
    VertexOutput output;
    output.position = float4(input.position, 0.0, 1.0);
    output.color = input.color;
    return output;
}

float4 ps_main(VertexOutput input) : SV_TARGET {
    return input.color;
}
`
