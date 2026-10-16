package xmath

import (
	"fmt"
	"testing"
)

func TestVec2Arithmetic(t *testing.T) {
	a := Vec2{X: 3, Y: 4}
	b := Vec2{X: 1, Y: 2}

	if got, want := fmt.Sprint(a.Add(b)), "{4 6}"; got != want {
		t.Fatalf("Add = %s, want %s", got, want)
	}
	if got, want := fmt.Sprint(a.Sub(b)), "{2 2}"; got != want {
		t.Fatalf("Sub = %s, want %s", got, want)
	}
	if got, want := fmt.Sprint(a.Mul(2)), "{6 8}"; got != want {
		t.Fatalf("Mul = %s, want %s", got, want)
	}
	if got := a.Length(); got != 5 {
		t.Fatalf("Length = %v, want 5", got)
	}
	if got, want := fmt.Sprint(Vec2{}.Lerp(a, 0.5)), "{1.5 2}"; got != want {
		t.Fatalf("Lerp = %s, want %s", got, want)
	}
}

func TestPos2Arithmetic(t *testing.T) {
	p := Pos2{X: 10, Y: 20}

	if got, want := fmt.Sprint(p.Add(Vec2{X: 1, Y: -2})), "{11 18}"; got != want {
		t.Fatalf("Add = %s, want %s", got, want)
	}
	if got, want := fmt.Sprint(p.Sub(Vec2{X: 1, Y: -2})), "{9 22}"; got != want {
		t.Fatalf("Sub = %s, want %s", got, want)
	}
	if got, want := fmt.Sprint(p.Delta(Pos2{X: 4, Y: 5})), "{6 15}"; got != want {
		t.Fatalf("Delta = %s, want %s", got, want)
	}
	if got, want := fmt.Sprint(Pos2{}.Lerp(p, 0.25)), "{2.5 5}"; got != want {
		t.Fatalf("Lerp = %s, want %s", got, want)
	}
}

func TestRectGeometry(t *testing.T) {
	r := RectFromMinSize(Pos2{X: 10, Y: 20}, Vec2{X: 30, Y: 40})

	if got, want := fmt.Sprint(r), "{{10 20} {40 60}}"; got != want {
		t.Fatalf("Rect = %s, want %s", got, want)
	}
	if r.Width() != 30 || r.Height() != 40 {
		t.Fatalf("size = %vx%v, want 30x40", r.Width(), r.Height())
	}
	if got, want := fmt.Sprint(r.Size()), "{30 40}"; got != want {
		t.Fatalf("Size = %s, want %s", got, want)
	}
	if got, want := fmt.Sprint(r.Center()), "{25 40}"; got != want {
		t.Fatalf("Center = %s, want %s", got, want)
	}
	if !r.IsPositive() {
		t.Fatal("a sized rect must be positive")
	}
	if (Rect{Min: Pos2{X: 1, Y: 1}, Max: Pos2{X: 1, Y: 5}}).IsPositive() {
		t.Fatal("a rect without width must not be positive")
	}
}

func TestRectTransform(t *testing.T) {
	r := Rect{Min: Pos2{X: 0, Y: 0}, Max: Pos2{X: 10, Y: 10}}

	if got, want := fmt.Sprint(r.Translate(Vec2{X: 2, Y: -3})), "{{2 -3} {12 7}}"; got != want {
		t.Fatalf("Translate = %s, want %s", got, want)
	}
	if got, want := fmt.Sprint(r.Expand(2)), "{{-2 -2} {12 12}}"; got != want {
		t.Fatalf("Expand = %s, want %s", got, want)
	}
}

func TestRectIntersectAndContains(t *testing.T) {
	r := Rect{Min: Pos2{X: 0, Y: 0}, Max: Pos2{X: 10, Y: 10}}
	other := Rect{Min: Pos2{X: 5, Y: -5}, Max: Pos2{X: 20, Y: 5}}

	if got, want := fmt.Sprint(r.Intersect(other)), "{{5 0} {10 5}}"; got != want {
		t.Fatalf("Intersect = %s, want %s", got, want)
	}
	if got := r.Intersect(Rect{Min: Pos2{X: 20, Y: 20}, Max: Pos2{X: 30, Y: 30}}); got.IsPositive() {
		t.Fatalf("disjoint rects intersected to %v", got)
	}
	if !r.Contains(Pos2{X: 0, Y: 10}) || !r.Contains(Pos2{X: 5, Y: 5}) {
		t.Fatal("the rect must contain its own edges")
	}
	if r.Contains(Pos2{X: -1, Y: 5}) || r.Contains(Pos2{X: 5, Y: 11}) {
		t.Fatal("the rect must not contain outside points")
	}
}
