package view

const DefaultTextSize float32 = 16

type TextStyle struct {
	Size float32
}

type FillRectOp struct {
	Rect  Rect
	Color Color
}

type DrawTextOp struct {
	Text  string
	Style TextStyle
	At    Point
}

type Canvas interface {
	FillRect(rect Rect, color Color)
	DrawText(text string, style TextStyle, at Point)
}

type Recorder struct {
	Ops []any
}

func (r *Recorder) FillRect(rect Rect, color Color) {
	r.Ops = append(r.Ops, FillRectOp{Rect: rect, Color: color})
}

func (r *Recorder) DrawText(text string, style TextStyle, at Point) {
	r.Ops = append(r.Ops, DrawTextOp{Text: text, Style: style, At: at})
}

type TextMeasurer interface {
	Measure(text string, style TextStyle, constraints Constraints) Size
}

type Monospace struct {
	AdvanceRatio    float32
	LineHeightRatio float32
}

func (m Monospace) Measure(text string, style TextStyle, constraints Constraints) Size {
	size := style.Size
	if size <= 0 {
		size = DefaultTextSize
	}
	return Size{
		Width:  float32(len([]rune(text))) * size * m.AdvanceRatio,
		Height: size * m.LineHeightRatio,
	}
}
