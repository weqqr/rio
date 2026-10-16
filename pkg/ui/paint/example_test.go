package paint

import (
	"fmt"

	"github.com/weqqr/rio/pkg/ui/view"
)

func Example() {
	renderer := NewStubText()
	comp := view.New(&view.Config{TextMeasurer: Measurer(renderer)})
	comp.SetContent(func(c *view.Scope) {
		c.Column(func(c *view.Scope) {
			c.Text("hi", view.Padding(1))
			c.Spacer(view.Sized(6, 3), view.Background(view.Color{B: 1, A: 1}))
		}, view.Gap(2))
	})

	glyphTexture, glyphSize := renderer.Atlas()
	fmt.Printf("textures: %d %v, %d %v\n", SolidTexture, SolidTextureSize, glyphTexture, glyphSize)

	canvas := New(renderer)
	canvas.SetClip(view.Rect{Max: view.Point{X: 40, Y: 40}})
	fmt.Println("size", comp.Measure(view.Unbounded))
	comp.Draw(canvas)
	for _, primitive := range canvas.Primitives() {
		fmt.Printf("mesh texture %d: %d vertices, %d triangles, clip %v\n",
			primitive.Mesh.Texture, len(primitive.Mesh.Vertices), len(primitive.Mesh.Indices)/3, primitive.ClipRect)
		for _, vertex := range primitive.Mesh.Vertices {
			fmt.Printf("  %v uv %v\n", vertex.Pos, vertex.UV)
		}
	}

	// Output:
	// textures: 0 {1 1}, 1 {1 1}
	// size {18 23}
	// mesh texture 1: 8 vertices, 4 triangles, clip {{0 0} {40 40}}
	//   {1 1} uv {0 0}
	//   {9 1} uv {1 0}
	//   {9 17} uv {1 1}
	//   {1 17} uv {0 1}
	//   {9 1} uv {0 0}
	//   {17 1} uv {1 0}
	//   {17 17} uv {1 1}
	//   {9 17} uv {0 1}
	// mesh texture 0: 4 vertices, 2 triangles, clip {{0 0} {40 40}}
	//   {0 20} uv {0 0}
	//   {6 20} uv {1 0}
	//   {6 23} uv {1 1}
	//   {0 23} uv {0 1}
}
