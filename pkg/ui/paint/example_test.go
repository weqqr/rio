package paint

import (
	"fmt"

	"github.com/weqqr/rio/pkg/xmath"
)

func Example() {
	text := NewStubText()
	glyphTexture, glyphSize := text.Atlas()
	fmt.Printf("textures: %d %v, %d %v\n", SolidTexture, SolidTextureSize, glyphTexture, glyphSize)

	canvas := New(text)
	canvas.SetClip(xmath.Rect{Max: xmath.Pos2{X: 40, Y: 40}})
	canvas.FillRect(xmath.Rect{Max: xmath.Pos2{X: 3, Y: 20}}, Rgba{G: 0.5, A: 1})
	canvas.DrawText("hi", TextStyle{Size: 16}, xmath.Pos2{X: 4, Y: 2})

	for _, primitive := range canvas.Primitives() {
		fmt.Printf("mesh texture %d: %d vertices, %d triangles, clip %v\n",
			primitive.Mesh.Texture, len(primitive.Mesh.Vertices), len(primitive.Mesh.Indices)/3, primitive.ClipRect)
		for _, vertex := range primitive.Mesh.Vertices {
			fmt.Printf("  %v uv %v colour %d\n", vertex.Pos, vertex.UV, vertex.Color)
		}
	}

	// Output:
	// textures: 0 {1 1}, 1 {1 1}
	// mesh texture 0: 4 vertices, 2 triangles, clip {{0 0} {40 40}}
	//   {0 0} uv {0 0} colour 4278222848
	//   {3 0} uv {1 0} colour 4278222848
	//   {3 20} uv {1 1} colour 4278222848
	//   {0 20} uv {0 1} colour 4278222848
	// mesh texture 1: 8 vertices, 4 triangles, clip {{0 0} {40 40}}
	//   {4 2} uv {0 0} colour 4294967295
	//   {12 2} uv {1 0} colour 4294967295
	//   {12 18} uv {1 1} colour 4294967295
	//   {4 18} uv {0 1} colour 4294967295
	//   {12 2} uv {0 0} colour 4294967295
	//   {20 2} uv {1 0} colour 4294967295
	//   {20 18} uv {1 1} colour 4294967295
	//   {12 18} uv {0 1} colour 4294967295
}
