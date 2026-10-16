package paint

import (
	"math"

	"github.com/weqqr/rio/pkg/xmath"
)

type TextureID uint32

const SolidTexture TextureID = 0

const infinite = float32(math.MaxFloat32)

var (
	SolidUV          = xmath.Rect{Max: xmath.Pos2{X: 1, Y: 1}}
	SolidTextureSize = xmath.Vec2{X: 1, Y: 1}
	Unclipped        = xmath.Rect{
		Min: xmath.Pos2{X: -infinite, Y: -infinite},
		Max: xmath.Pos2{X: infinite, Y: infinite},
	}
)

type Rgba struct {
	R, G, B, A float32
}

type Color32 uint32

func RGBA(r, g, b, a uint8) Color32 {
	return Color32(uint32(r) | uint32(g)<<8 | uint32(b)<<16 | uint32(a)<<24)
}

func FromColor(c Rgba) Color32 {
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
	Pos   xmath.Pos2
	UV    xmath.Pos2
	Color Color32
}

type Mesh struct {
	Texture  TextureID
	Vertices []Vertex
	Indices  []uint32
}

type ClippedPrimitive struct {
	ClipRect xmath.Rect
	Mesh     Mesh
}
