package view

import "math"

type Point struct {
	X, Y float32
}

type Size struct {
	Width, Height float32
}

type Rect struct {
	Min, Max Point
}

type Color struct {
	R, G, B, A float32
}

type Alignment int

const (
	AlignStart Alignment = iota
	AlignCenter
	AlignEnd
)

type Constraints struct {
	MinWidth  float32
	MaxWidth  float32
	MinHeight float32
	MaxHeight float32
}

var Unbounded = Constraints{MaxWidth: infinite, MaxHeight: infinite}

const infinite = float32(math.MaxFloat32)

func Tight(size Size) Constraints {
	return Constraints{
		MinWidth:  size.Width,
		MaxWidth:  size.Width,
		MinHeight: size.Height,
		MaxHeight: size.Height,
	}
}

func Loose(max Size) Constraints {
	return Constraints{MaxWidth: max.Width, MaxHeight: max.Height}
}

func (c Constraints) Constrain(size Size) Size {
	return Size{
		Width:  clamp(size.Width, c.MinWidth, c.MaxWidth),
		Height: clamp(size.Height, c.MinHeight, c.MaxHeight),
	}
}

func clamp(value, low, high float32) float32 {
	if value < low {
		return low
	}
	if value > high {
		return high
	}
	return value
}

func alignOffset(align Alignment, container, child float32) float32 {
	switch align {
	case AlignCenter:
		return (container - child) / 2
	case AlignEnd:
		return container - child
	}
	return 0
}

type View interface {
	Measure(constraints Constraints) Size
	Place(offset Point)
	Draw(canvas Canvas)
}
