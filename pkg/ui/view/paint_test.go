package view

import (
	"fmt"
	"testing"

	"github.com/weqqr/rio/pkg/ui/paint"
	"github.com/weqqr/rio/pkg/xmath"
)

var _ Canvas = (*paint.Canvas)(nil)

func meshPositions(mesh paint.Mesh) []xmath.Pos2 {
	out := make([]xmath.Pos2, 0, len(mesh.Vertices))
	for _, vertex := range mesh.Vertices {
		out = append(out, vertex.Pos)
	}
	return out
}

func TestMeasurerFollowsTheRenderer(t *testing.T) {
	renderer := paint.NewStubText()
	style := TextStyle{Size: 20}
	measured := Measurer(renderer).Measure("abc", style, Constraints{})
	if measured != (Size{Width: 30, Height: 20}) {
		t.Fatalf("measured = %v, want {30 20}", measured)
	}
	layout := renderer.Layout("abc", style)
	if measured != (Size{Width: layout.Size.X, Height: layout.Size.Y}) {
		t.Fatal("the measurer and the text renderer disagree")
	}
}

func TestCompositionTessellatesThroughTheCanvas(t *testing.T) {
	renderer := paint.NewStubText()
	comp := New(&Config{TextMeasurer: Measurer(renderer)})
	comp.SetContent(func(c *Scope) {
		c.Column(func(c *Scope) {
			c.Text("hi")
			c.Spacer(Sized(10, 5), Background(Color{G: 1, A: 1}))
		}, Padding(1))
	})

	canvas := paint.New(renderer)
	if got, want := comp.Measure(Unbounded), (Size{Width: 18, Height: 23}); got != want {
		t.Fatalf("size = %v, want %v", got, want)
	}
	comp.Draw(canvas)

	primitives := canvas.Primitives()
	if len(primitives) != 2 {
		t.Fatalf("primitives = %d, want text then background", len(primitives))
	}
	if primitives[0].Mesh.Texture != 1 || primitives[1].Mesh.Texture != paint.SolidTexture {
		t.Fatalf("textures = %d and %d, want the glyph texture then the solid one",
			primitives[0].Mesh.Texture, primitives[1].Mesh.Texture)
	}
	if got, want := fmt.Sprint(meshPositions(primitives[0].Mesh)), "[{1 1} {9 1} {9 17} {1 17} {9 1} {17 1} {17 17} {9 17}]"; got != want {
		t.Fatalf("glyph positions = %s, want %s", got, want)
	}
	if got, want := fmt.Sprint(meshPositions(primitives[1].Mesh)), "[{1 17} {11 17} {11 22} {1 22}]"; got != want {
		t.Fatalf("background positions = %s, want %s", got, want)
	}
	if primitives[1].Mesh.Vertices[0].Color != paint.RGBA(0, 255, 0, 255) {
		t.Fatalf("background colour = %v, want opaque green", primitives[1].Mesh.Vertices[0].Color)
	}
}

func TestRecorderAndCanvasAgreeOnGeometry(t *testing.T) {
	renderer := paint.NewStubText()
	comp := New(&Config{TextMeasurer: Measurer(renderer)})
	comp.SetContent(func(c *Scope) {
		c.Column(func(c *Scope) {
			c.Text("hi")
			c.Spacer(Sized(10, 5), Background(Color{B: 1, A: 1}))
		}, Padding(1))
	})

	recorder := &Recorder{}
	comp.Draw(recorder)
	canvas := paint.New(renderer)
	comp.Draw(canvas)

	var textAt []xmath.Pos2
	var fills []xmath.Rect
	for _, op := range recorder.Ops {
		switch op := op.(type) {
		case DrawTextOp:
			textAt = append(textAt, op.At)
		case FillRectOp:
			fills = append(fills, op.Rect)
		}
	}
	if len(textAt) != 1 || len(fills) != 1 {
		t.Fatalf("recorder saw %d texts and %d fills, want one each", len(textAt), len(fills))
	}

	primitives := canvas.Primitives()
	if len(primitives) != 2 {
		t.Fatalf("primitives = %d, want 2", len(primitives))
	}
	if got := meshPositions(primitives[0].Mesh)[0]; got != textAt[0] {
		t.Fatalf("first glyph origin = %v, want the recorded text position %v", got, textAt[0])
	}
	quad := meshPositions(primitives[1].Mesh)
	if quad[0] != fills[0].Min || quad[2] != fills[0].Max {
		t.Fatalf("fill quad %v-%v, want the recorded rect %v", quad[0], quad[2], fills[0])
	}
	if primitives[1].ClipRect != paint.Unclipped {
		t.Fatalf("clip = %v, want the unclipped rect", primitives[1].ClipRect)
	}
}

func TestCanvasClipFlowsIntoThePrimitives(t *testing.T) {
	comp := New(monospace())
	comp.SetContent(func(c *Scope) {
		c.Column(func(c *Scope) {
			c.Spacer(Sized(4, 4), Background(Color{A: 1}))
			c.Spacer(Sized(4, 4), Background(Color{B: 1, A: 1}))
		})
	})

	canvas := paint.New(nil)
	canvas.SetClip(xmath.Rect{Max: xmath.Pos2{X: 2, Y: 2}})
	comp.Draw(canvas)
	canvas.SetClip(xmath.Rect{Max: xmath.Pos2{X: 4, Y: 4}})
	comp.Draw(canvas)

	primitives := canvas.Primitives()
	if len(primitives) != 2 {
		t.Fatalf("primitives = %d, want one per clip rect", len(primitives))
	}
	if got, want := fmt.Sprint(primitives[0].ClipRect), "{{0 0} {2 2}}"; got != want {
		t.Fatalf("clip = %s, want %s", got, want)
	}
	if got, want := fmt.Sprint(primitives[1].ClipRect), "{{0 0} {4 4}}"; got != want {
		t.Fatalf("clip = %s, want %s", got, want)
	}
}
