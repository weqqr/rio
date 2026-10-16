package paint

import "github.com/weqqr/rio/pkg/ui/view"

type Glyph struct {
	UV   view.Rect
	Pos  view.Point
	Size view.Size
}

type TextLayout struct {
	Size   view.Size
	Glyphs []Glyph
}

type TextRenderer interface {
	Atlas() (TextureID, view.Size)
	Layout(text string, style view.TextStyle) TextLayout
}

type StubText struct {
	AdvanceRatio     float32
	LineHeightRatio  float32
	GlyphTexture     TextureID
	GlyphTextureSize view.Size
}

func NewStubText() StubText {
	return StubText{
		AdvanceRatio:     0.5,
		LineHeightRatio:  1,
		GlyphTexture:     1,
		GlyphTextureSize: view.Size{Width: 1, Height: 1},
	}
}

func (s StubText) Atlas() (TextureID, view.Size) {
	return s.GlyphTexture, s.GlyphTextureSize
}

func (s StubText) Layout(text string, style view.TextStyle) TextLayout {
	size := style.Size
	if size <= 0 {
		size = view.DefaultTextSize
	}
	advance := size * s.AdvanceRatio
	lineHeight := size * s.LineHeightRatio
	uv := view.Rect{Max: view.Point{X: s.GlyphTextureSize.Width, Y: s.GlyphTextureSize.Height}}

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
			Pos:  view.Point{X: lineX, Y: y},
			Size: view.Size{Width: advance, Height: lineHeight},
		})
		lineX += advance
	}
	return TextLayout{
		Size:   view.Size{Width: max(width, lineX), Height: y + lineHeight},
		Glyphs: glyphs,
	}
}

type measurer struct {
	renderer TextRenderer
}

func Measurer(renderer TextRenderer) view.TextMeasurer {
	return measurer{renderer: renderer}
}

func (m measurer) Measure(text string, style view.TextStyle, constraints view.Constraints) view.Size {
	return m.renderer.Layout(text, style).Size
}
