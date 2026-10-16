package paint

import "github.com/weqqr/rio/pkg/xmath"

const DefaultTextSize float32 = 16

type TextStyle struct {
	Size float32
}

type Glyph struct {
	UV   xmath.Rect
	Pos  xmath.Pos2
	Size xmath.Vec2
}

type TextLayout struct {
	Size   xmath.Vec2
	Glyphs []Glyph
}

type TextRenderer interface {
	Atlas() (TextureID, xmath.Vec2)
	Layout(text string, style TextStyle) TextLayout
}

type StubText struct {
	AdvanceRatio     float32
	LineHeightRatio  float32
	GlyphTexture     TextureID
	GlyphTextureSize xmath.Vec2
}

func NewStubText() StubText {
	return StubText{
		AdvanceRatio:     0.5,
		LineHeightRatio:  1,
		GlyphTexture:     1,
		GlyphTextureSize: xmath.Vec2{X: 1, Y: 1},
	}
}

func (s StubText) Atlas() (TextureID, xmath.Vec2) {
	return s.GlyphTexture, s.GlyphTextureSize
}

func (s StubText) Layout(text string, style TextStyle) TextLayout {
	size := style.Size
	if size <= 0 {
		size = DefaultTextSize
	}
	advance := size * s.AdvanceRatio
	lineHeight := size * s.LineHeightRatio
	glyphSize := xmath.Vec2{X: advance, Y: lineHeight}
	uv := xmath.Rect{Max: xmath.Pos2{X: s.GlyphTextureSize.X, Y: s.GlyphTextureSize.Y}}

	var (
		glyphs          []Glyph
		width, y, lineX float32
	)
	for _, r := range text {
		if r == '\n' {
			width = max(width, lineX)
			lineX = 0
			y += lineHeight
			continue
		}
		glyphs = append(glyphs, Glyph{
			UV:   uv,
			Pos:  xmath.Pos2{X: lineX, Y: y},
			Size: glyphSize,
		})
		lineX += advance
	}
	return TextLayout{
		Size:   xmath.Vec2{X: max(width, lineX), Y: y + lineHeight},
		Glyphs: glyphs,
	}
}
