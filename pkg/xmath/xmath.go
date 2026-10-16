package xmath

import "math"

type Vec2 struct {
	X, Y float32
}

func (v Vec2) Add(o Vec2) Vec2 { return Vec2{X: v.X + o.X, Y: v.Y + o.Y} }

func (v Vec2) Sub(o Vec2) Vec2 { return Vec2{X: v.X - o.X, Y: v.Y - o.Y} }

func (v Vec2) Mul(scale float32) Vec2 { return Vec2{X: v.X * scale, Y: v.Y * scale} }

func (v Vec2) Length() float32 { return float32(math.Sqrt(float64(v.X*v.X + v.Y*v.Y))) }

func (v Vec2) Lerp(o Vec2, t float32) Vec2 {
	return Vec2{X: lerp(v.X, o.X, t), Y: lerp(v.Y, o.Y, t)}
}

type Pos2 struct {
	X, Y float32
}

func (p Pos2) Add(v Vec2) Pos2 { return Pos2{X: p.X + v.X, Y: p.Y + v.Y} }

func (p Pos2) Sub(v Vec2) Pos2 { return Pos2{X: p.X - v.X, Y: p.Y - v.Y} }

func (p Pos2) Delta(o Pos2) Vec2 { return Vec2{X: p.X - o.X, Y: p.Y - o.Y} }

func (p Pos2) ToVec2() Vec2 { return Vec2{X: p.X, Y: p.Y} }

func (p Pos2) Lerp(o Pos2, t float32) Pos2 {
	return Pos2{X: lerp(p.X, o.X, t), Y: lerp(p.Y, o.Y, t)}
}

type Rect struct {
	Min, Max Pos2
}

func RectFromMinSize(min Pos2, size Vec2) Rect {
	return Rect{Min: min, Max: min.Add(size)}
}

func (r Rect) Width() float32 { return r.Max.X - r.Min.X }

func (r Rect) Height() float32 { return r.Max.Y - r.Min.Y }

func (r Rect) Size() Vec2 { return Vec2{X: r.Width(), Y: r.Height()} }

func (r Rect) Center() Pos2 {
	return Pos2{X: r.Min.X + r.Width()/2, Y: r.Min.Y + r.Height()/2}
}

func (r Rect) IsPositive() bool { return r.Max.X > r.Min.X && r.Max.Y > r.Min.Y }

func (r Rect) Translate(v Vec2) Rect {
	return Rect{Min: r.Min.Add(v), Max: r.Max.Add(v)}
}

func (r Rect) Expand(amount float32) Rect {
	return Rect{
		Min: Pos2{X: r.Min.X - amount, Y: r.Min.Y - amount},
		Max: Pos2{X: r.Max.X + amount, Y: r.Max.Y + amount},
	}
}

func (r Rect) Intersect(o Rect) Rect {
	return Rect{
		Min: Pos2{X: max(r.Min.X, o.Min.X), Y: max(r.Min.Y, o.Min.Y)},
		Max: Pos2{X: min(r.Max.X, o.Max.X), Y: min(r.Max.Y, o.Max.Y)},
	}
}

func (r Rect) Contains(p Pos2) bool {
	return p.X >= r.Min.X && p.X <= r.Max.X && p.Y >= r.Min.Y && p.Y <= r.Max.Y
}

func lerp(a, b, t float32) float32 { return a + (b-a)*t }
