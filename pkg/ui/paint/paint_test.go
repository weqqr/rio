package paint

import (
	"fmt"
	"testing"

	"github.com/weqqr/rio/pkg/xmath"
)

func white() Rgba { return Rgba{R: 1, G: 1, B: 1, A: 1} }

func rect(min, max xmath.Pos2) xmath.Rect { return xmath.Rect{Min: min, Max: max} }

func positions(mesh Mesh) []xmath.Pos2 {
	out := make([]xmath.Pos2, 0, len(mesh.Vertices))
	for _, vertex := range mesh.Vertices {
		out = append(out, vertex.Pos)
	}
	return out
}

func uvs(mesh Mesh) []xmath.Pos2 {
	out := make([]xmath.Pos2, 0, len(mesh.Vertices))
	for _, vertex := range mesh.Vertices {
		out = append(out, vertex.UV)
	}
	return out
}

func textures(primitives []ClippedPrimitive) []TextureID {
	out := make([]TextureID, 0, len(primitives))
	for _, primitive := range primitives {
		out = append(out, primitive.Mesh.Texture)
	}
	return out
}

func fill(canvas *Canvas, width, height float32) {
	canvas.FillRect(rect(xmath.Pos2{}, xmath.Pos2{X: width, Y: height}), white())
}

func TestFillRectTessellatesAQuad(t *testing.T) {
	canvas := New(nil)
	canvas.FillRect(rect(xmath.Pos2{X: 2, Y: 3}, xmath.Pos2{X: 12, Y: 8}), Rgba{R: 1, A: 0.5})

	primitives := canvas.Primitives()
	if len(primitives) != 1 {
		t.Fatalf("primitives = %d, want 1", len(primitives))
	}
	primitive := primitives[0]
	if primitive.ClipRect != Unclipped {
		t.Fatalf("clip = %v, want the unclipped rect", primitive.ClipRect)
	}
	if primitive.Mesh.Texture != SolidTexture {
		t.Fatalf("texture = %d, want the solid texture", primitive.Mesh.Texture)
	}
	if got, want := fmt.Sprint(positions(primitive.Mesh)), "[{2 3} {12 3} {12 8} {2 8}]"; got != want {
		t.Fatalf("positions = %s, want %s", got, want)
	}
	if got, want := fmt.Sprint(uvs(primitive.Mesh)), "[{0 0} {1 0} {1 1} {0 1}]"; got != want {
		t.Fatalf("uvs = %s, want %s", got, want)
	}
	if got, want := fmt.Sprint(primitive.Mesh.Indices), "[0 1 2 0 2 3]"; got != want {
		t.Fatalf("indices = %s, want %s", got, want)
	}
	for _, vertex := range primitive.Mesh.Vertices {
		if r, g, b, a := vertex.Color.RGBA(); r != 128 || g != 0 || b != 0 || a != 128 {
			t.Fatalf("vertex colour = %d,%d,%d,%d, want premultiplied 128,0,0,128", r, g, b, a)
		}
	}
}

func TestConsecutiveFillsShareOneMesh(t *testing.T) {
	canvas := New(nil)
	fill(canvas, 4, 4)
	canvas.FillRect(rect(xmath.Pos2{X: 4}, xmath.Pos2{X: 8, Y: 4}), white())

	primitives := canvas.Primitives()
	if len(primitives) != 1 {
		t.Fatalf("primitives = %d, want the fills batched into one", len(primitives))
	}
	mesh := primitives[0].Mesh
	if len(mesh.Vertices) != 8 || len(mesh.Indices) != 12 {
		t.Fatalf("mesh = %d vertices, %d indices, want 8 and 12", len(mesh.Vertices), len(mesh.Indices))
	}
	if got, want := fmt.Sprint(mesh.Indices), "[0 1 2 0 2 3 4 5 6 4 6 7]"; got != want {
		t.Fatalf("indices = %s, want %s", got, want)
	}
}

func TestGeometrySplitsByTexture(t *testing.T) {
	renderer := NewStubText()
	canvas := New(renderer)
	fill(canvas, 4, 4)
	canvas.DrawText("ab", TextStyle{Size: 16}, xmath.Pos2{X: 4, Y: 5})
	fill(canvas, 2, 2)

	primitives := canvas.Primitives()
	if len(primitives) != 3 {
		t.Fatalf("primitives = %d, want one per texture run", len(primitives))
	}
	if got, want := fmt.Sprint(textures(primitives)), "[0 1 0]"; got != want {
		t.Fatalf("textures = %s, want %s", got, want)
	}
	glyphs := primitives[1].Mesh
	if len(glyphs.Vertices) != 8 {
		t.Fatalf("glyph vertices = %d, want two quads", len(glyphs.Vertices))
	}
	if got, want := fmt.Sprint(positions(glyphs)), "[{4 5} {12 5} {12 21} {4 21} {12 5} {20 5} {20 21} {12 21}]"; got != want {
		t.Fatalf("glyph positions = %s, want %s", got, want)
	}
	if got, want := fmt.Sprint(uvs(glyphs)), "[{0 0} {1 0} {1 1} {0 1} {0 0} {1 0} {1 1} {0 1}]"; got != want {
		t.Fatalf("glyph uvs = %s, want %s", got, want)
	}
	if r, g, b, a := glyphs.Vertices[0].Color.RGBA(); r != 255 || g != 255 || b != 255 || a != 255 {
		t.Fatalf("glyph colour = %d,%d,%d,%d, want opaque white", r, g, b, a)
	}
}

func TestClipStartsANewPrimitive(t *testing.T) {
	first := rect(xmath.Pos2{}, xmath.Pos2{X: 10, Y: 10})
	second := rect(xmath.Pos2{X: 10}, xmath.Pos2{X: 20, Y: 10})
	canvas := New(nil)
	canvas.SetClip(first)
	fill(canvas, 4, 4)
	fill(canvas, 2, 2)
	canvas.SetClip(second)
	fill(canvas, 3, 3)

	primitives := canvas.Primitives()
	if len(primitives) != 2 {
		t.Fatalf("primitives = %d, want one per clip rect", len(primitives))
	}
	if primitives[0].ClipRect != first || primitives[1].ClipRect != second {
		t.Fatalf("clip rects = %v and %v, want %v and %v", primitives[0].ClipRect, primitives[1].ClipRect, first, second)
	}
	if len(primitives[0].Mesh.Vertices) != 8 {
		t.Fatalf("first mesh = %d vertices, want the two fills merged", len(primitives[0].Mesh.Vertices))
	}
	if canvas.Clip() != second {
		t.Fatalf("clip = %v, want %v", canvas.Clip(), second)
	}
}

func TestInvisibleGeometryIsSkipped(t *testing.T) {
	canvas := New(nil)
	canvas.FillRect(rect(xmath.Pos2{}, xmath.Pos2{X: 4, Y: 4}), Rgba{})
	canvas.FillRect(rect(xmath.Pos2{X: 2}, xmath.Pos2{X: 2, Y: 4}), white())
	canvas.FillRect(rect(xmath.Pos2{}, xmath.Pos2{X: 4, Y: -1}), white())
	canvas.DrawText("", TextStyle{}, xmath.Pos2{})
	canvas.DrawText("no renderer", TextStyle{}, xmath.Pos2{})

	if got := len(canvas.Primitives()); got != 0 {
		t.Fatalf("primitives = %d, want none", got)
	}
}

func TestStubTextLayout(t *testing.T) {
	renderer := NewStubText()
	if id, size := renderer.Atlas(); id != 1 || size != (xmath.Vec2{X: 1, Y: 1}) {
		t.Fatalf("atlas = %d %v, want texture 1 of 1x1", id, size)
	}
	layout := renderer.Layout("ab\nc", TextStyle{Size: 16})
	if layout.Size != (xmath.Vec2{X: 16, Y: 32}) {
		t.Fatalf("layout size = %v, want {16 32}", layout.Size)
	}
	var got []string
	for _, glyph := range layout.Glyphs {
		got = append(got, fmt.Sprintf("%v%v", glyph.Pos, glyph.Size))
	}
	want := "[{0 0}{8 16} {8 0}{8 16} {0 16}{8 16}]"
	if joined := fmt.Sprint(got); joined != want {
		t.Fatalf("glyphs = %s, want %s", joined, want)
	}
}

type fakeText struct{}

func (fakeText) Atlas() (TextureID, xmath.Vec2) { return 7, xmath.Vec2{X: 64, Y: 64} }

func (fakeText) Layout(text string, style TextStyle) TextLayout {
	return TextLayout{
		Size: xmath.Vec2{X: 4, Y: 6},
		Glyphs: []Glyph{{
			UV:   rect(xmath.Pos2{X: 0.25, Y: 0.5}, xmath.Pos2{X: 0.5, Y: 0.75}),
			Pos:  xmath.Pos2{X: 2, Y: 3},
			Size: xmath.Vec2{X: 4, Y: 6},
		}},
	}
}

type emptyText struct{}

func (emptyText) Atlas() (TextureID, xmath.Vec2) { return 3, xmath.Vec2{X: 8, Y: 8} }

func (emptyText) Layout(text string, style TextStyle) TextLayout {
	return TextLayout{Size: xmath.Vec2{X: 4}, Glyphs: []Glyph{{Size: xmath.Vec2{X: 4}}}}
}

func TestGlyphsWithoutGeometryAreSkipped(t *testing.T) {
	canvas := New(emptyText{})
	canvas.DrawText("a", TextStyle{Size: 16}, xmath.Pos2{})
	if got := len(canvas.Primitives()); got != 0 {
		t.Fatalf("primitives = %d, want none for a text without geometry", got)
	}
	fill(canvas, 1, 1)
	primitives := canvas.Primitives()
	if len(primitives) != 1 || len(primitives[0].Mesh.Vertices) != 4 {
		t.Fatalf("primitives = %d, want the fill alone", len(primitives))
	}
}

func TestCanvasHonoursTheRenderer(t *testing.T) {
	canvas := New(fakeText{})
	canvas.SetTextColor(Rgba{G: 1, A: 1})
	canvas.DrawText("x", TextStyle{Size: 16}, xmath.Pos2{X: 10, Y: 20})

	primitives := canvas.Primitives()
	if len(primitives) != 1 {
		t.Fatalf("primitives = %d, want 1", len(primitives))
	}
	mesh := primitives[0].Mesh
	if mesh.Texture != 7 {
		t.Fatalf("texture = %d, want the renderer atlas 7", mesh.Texture)
	}
	if got, want := fmt.Sprint(positions(mesh)), "[{12 23} {16 23} {16 29} {12 29}]"; got != want {
		t.Fatalf("positions = %s, want %s", got, want)
	}
	if got, want := fmt.Sprint(uvs(mesh)), "[{0.25 0.5} {0.5 0.5} {0.5 0.75} {0.25 0.75}]"; got != want {
		t.Fatalf("uvs = %s, want %s", got, want)
	}
	if r, g, b, a := mesh.Vertices[0].Color.RGBA(); r != 0 || g != 255 || b != 0 || a != 255 {
		t.Fatalf("colour = %d,%d,%d,%d, want opaque green", r, g, b, a)
	}
}

func TestClearReusesBuffersAndResetsState(t *testing.T) {
	canvas := New(NewStubText())
	canvas.SetClip(rect(xmath.Pos2{}, xmath.Pos2{X: 5, Y: 5}))
	canvas.SetTextColor(Rgba{R: 1, A: 1})
	canvas.DrawText("a", TextStyle{Size: 16}, xmath.Pos2{})
	capacity := cap(canvas.Primitives()[0].Mesh.Vertices)

	canvas.Clear()
	if len(canvas.Primitives()) != 0 {
		t.Fatalf("primitives = %d after Clear, want none", len(canvas.Primitives()))
	}
	if canvas.Clip() != Unclipped {
		t.Fatalf("clip = %v after Clear, want the unclipped rect", canvas.Clip())
	}

	fill(canvas, 2, 2)
	primitives := canvas.Primitives()
	if len(primitives) != 1 || len(primitives[0].Mesh.Vertices) != 4 {
		t.Fatalf("geometry survived Clear: %d primitives, %d vertices", len(primitives), len(primitives[0].Mesh.Vertices))
	}
	if got := cap(primitives[0].Mesh.Vertices); got != capacity {
		t.Fatalf("vertex capacity = %d, want the reused %d", got, capacity)
	}
	if r, g, b, a := primitives[0].Mesh.Vertices[0].Color.RGBA(); r != 255 || g != 255 || b != 255 || a != 255 {
		t.Fatalf("colour = %d,%d,%d,%d, want the reset white", r, g, b, a)
	}
}

func TestColor32IsPremultiplied(t *testing.T) {
	cases := []struct {
		in   Rgba
		want string
	}{
		{Rgba{R: 1, G: 1, B: 1, A: 1}, "255,255,255,255"},
		{Rgba{R: 1, A: 0.5}, "128,0,0,128"},
		{Rgba{R: 1, G: 1, B: 1}, "0,0,0,0"},
		{Rgba{R: 2, G: -1, B: 0.5, A: 2}, "255,0,128,255"},
	}
	for _, test := range cases {
		r, g, b, a := FromColor(test.in).RGBA()
		got := fmt.Sprintf("%d,%d,%d,%d", r, g, b, a)
		if got != test.want {
			t.Fatalf("FromColor(%v) = %s, want %s", test.in, got, test.want)
		}
	}
	if got, want := fmt.Sprint(RGBA(1, 2, 3, 4).RGBA()), "1 2 3 4"; got != want {
		t.Fatalf("RGBA round trip = %s, want %s", got, want)
	}
}
