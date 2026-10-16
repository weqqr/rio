package paint

import "github.com/weqqr/rio/pkg/ui/view"

var _ view.Canvas = (*Canvas)(nil)

type Canvas struct {
	text       TextRenderer
	clip       view.Rect
	textColor  Color32
	primitives []ClippedPrimitive
	used       int
}

func New(text TextRenderer) *Canvas {
	c := &Canvas{text: text}
	c.Clear()
	return c
}

func (c *Canvas) FillRect(rect view.Rect, color view.Color) {
	if rect.Max.X <= rect.Min.X || rect.Max.Y <= rect.Min.Y {
		return
	}
	vertexColor := FromColor(color)
	if vertexColor>>24 == 0 {
		return
	}
	addQuad(c.target(SolidTexture), rect.Min, rect.Max, SolidUV.Min, SolidUV.Max, vertexColor)
}

func (c *Canvas) DrawText(text string, style view.TextStyle, at view.Point) {
	if c.text == nil || text == "" {
		return
	}
	layout := c.text.Layout(text, style)
	if len(layout.Glyphs) == 0 {
		return
	}
	texture, _ := c.text.Atlas()
	var mesh *Mesh
	for _, glyph := range layout.Glyphs {
		if glyph.Size.Width <= 0 || glyph.Size.Height <= 0 {
			continue
		}
		if mesh == nil {
			mesh = c.target(texture)
		}
		min := view.Point{X: at.X + glyph.Pos.X, Y: at.Y + glyph.Pos.Y}
		max := view.Point{X: min.X + glyph.Size.Width, Y: min.Y + glyph.Size.Height}
		addQuad(mesh, min, max, glyph.UV.Min, glyph.UV.Max, c.textColor)
	}
}

func (c *Canvas) SetClip(rect view.Rect) {
	c.clip = rect
}

func (c *Canvas) Clip() view.Rect {
	return c.clip
}

func (c *Canvas) SetTextColor(color view.Color) {
	c.textColor = FromColor(color)
}

func (c *Canvas) Primitives() []ClippedPrimitive {
	return c.primitives[:c.used]
}

func (c *Canvas) Clear() {
	for i := range c.primitives {
		c.primitives[i].Mesh.Vertices = c.primitives[i].Mesh.Vertices[:0]
		c.primitives[i].Mesh.Indices = c.primitives[i].Mesh.Indices[:0]
	}
	c.used = 0
	c.clip = Unclipped
	c.textColor = FromColor(view.Color{R: 1, G: 1, B: 1, A: 1})
}

func (c *Canvas) target(texture TextureID) *Mesh {
	if c.used > 0 {
		last := &c.primitives[c.used-1]
		if last.Mesh.Texture == texture && last.ClipRect == c.clip {
			return &last.Mesh
		}
	}
	if c.used == len(c.primitives) {
		c.primitives = append(c.primitives, ClippedPrimitive{})
	}
	primitive := &c.primitives[c.used]
	primitive.ClipRect = c.clip
	primitive.Mesh.Texture = texture
	c.used++
	return &primitive.Mesh
}

func addQuad(mesh *Mesh, min, max, uvMin, uvMax view.Point, color Color32) {
	first := uint32(len(mesh.Vertices))
	mesh.Vertices = append(mesh.Vertices,
		Vertex{Pos: min, UV: uvMin, Color: color},
		Vertex{Pos: view.Point{X: max.X, Y: min.Y}, UV: view.Point{X: uvMax.X, Y: uvMin.Y}, Color: color},
		Vertex{Pos: max, UV: uvMax, Color: color},
		Vertex{Pos: view.Point{X: min.X, Y: max.Y}, UV: view.Point{X: uvMin.X, Y: uvMax.Y}, Color: color},
	)
	mesh.Indices = append(mesh.Indices, first, first+1, first+2, first, first+2, first+3)
}
