package paint

import (
	"math"

	"github.com/weqqr/rio/pkg/ui/view"
)

type TextureID uint32

const SolidTexture TextureID = 0

const infinite = float32(math.MaxFloat32)

var (
	SolidUV          = view.Rect{Max: view.Point{X: 1, Y: 1}}
	SolidTextureSize = view.Size{Width: 1, Height: 1}
	Unclipped        = view.Rect{
		Min: view.Point{X: -infinite, Y: -infinite},
		Max: view.Point{X: infinite, Y: infinite},
	}
)

type Color32 uint32

func RGBA(r, g, b, a uint8) Color32 {
	return Color32(uint32(r) | uint32(g)<<8 | uint32(b)<<16 | uint32(a)<<24)
}

func FromColor(c view.Color) Color32 {
	a := clamp01(c.A)
	return RGBA(
		uint8(clamp01(c.R)*a*255+0.5),
		uint8(clamp01(c.G)*a*255+0.5),
		uint8(clamp01(c.B)*a*255+0.5),
		uint8(a*255+0.5),
	)
}

func (c Color32) RGBA() (r, g, b, a uint8) {
	return uint8(c), uint8(c >> 8), uint8(c >> 16), uint8(c >> 24)
}

func clamp01(v float32) float32 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

type Vertex struct {
	Pos   view.Point
	UV    view.Point
	Color Color32
}

type Mesh struct {
	Texture  TextureID
	Vertices []Vertex
	Indices  []uint32
}

type ClippedPrimitive struct {
	ClipRect view.Rect
	Mesh     Mesh
}
