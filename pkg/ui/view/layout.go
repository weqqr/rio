package view

type paddingView struct {
	inner                    View
	top, right, bottom, left float32
}

func (v *paddingView) Measure(c Constraints) Size {
	inner := Constraints{
		MinWidth:  max(0, c.MinWidth-v.left-v.right),
		MaxWidth:  max(0, c.MaxWidth-v.left-v.right),
		MinHeight: max(0, c.MinHeight-v.top-v.bottom),
		MaxHeight: max(0, c.MaxHeight-v.top-v.bottom),
	}
	child := v.inner.Measure(inner)
	return Size{Width: child.Width + v.left + v.right, Height: child.Height + v.top + v.bottom}
}

func (v *paddingView) Place(offset Point) {
	v.inner.Place(Point{X: offset.X + v.left, Y: offset.Y + v.top})
}

func (v *paddingView) Draw(canvas Canvas) { v.inner.Draw(canvas) }

type sizeView struct {
	inner         View
	width, height float32
}

func (v *sizeView) Measure(c Constraints) Size {
	if v.width > 0 {
		c.MinWidth = v.width
		c.MaxWidth = v.width
	}
	if v.height > 0 {
		c.MinHeight = v.height
		c.MaxHeight = v.height
	}
	return v.inner.Measure(c)
}

func (v *sizeView) Place(offset Point) { v.inner.Place(offset) }

func (v *sizeView) Draw(canvas Canvas) { v.inner.Draw(canvas) }

type fillView struct {
	inner         View
	width, height bool
}

func (v *fillView) Measure(c Constraints) Size {
	if v.width && c.MaxWidth < infinite {
		c.MinWidth = c.MaxWidth
	}
	if v.height && c.MaxHeight < infinite {
		c.MinHeight = c.MaxHeight
	}
	return v.inner.Measure(c)
}

func (v *fillView) Place(offset Point) { v.inner.Place(offset) }

func (v *fillView) Draw(canvas Canvas) { v.inner.Draw(canvas) }

type backgroundView struct {
	inner  View
	color  Color
	size   Size
	offset Point
}

func (v *backgroundView) Measure(c Constraints) Size {
	v.size = v.inner.Measure(c)
	return v.size
}

func (v *backgroundView) Place(offset Point) {
	v.offset = offset
	v.inner.Place(offset)
}

func (v *backgroundView) Draw(canvas Canvas) {
	canvas.FillRect(Rect{Min: v.offset, Max: Point{X: v.offset.X + v.size.Width, Y: v.offset.Y + v.size.Height}}, v.color)
	v.inner.Draw(canvas)
}

type textView struct {
	measurer            TextMeasurer
	value               string
	style               TextStyle
	size                Size
	offset              Point
	measured            bool
	measuredConstraints Constraints
}

func (v *textView) Measure(c Constraints) Size {
	if v.measured && v.measuredConstraints == c {
		return v.size
	}
	v.size = c.Constrain(v.measurer.Measure(v.value, v.style, c))
	v.measuredConstraints = c
	v.measured = true
	return v.size
}

func (v *textView) Place(offset Point) { v.offset = offset }

func (v *textView) Draw(canvas Canvas) { canvas.DrawText(v.value, v.style, v.offset) }

type spacerView struct {
	size   Size
	offset Point
}

func (v *spacerView) Measure(c Constraints) Size {
	v.size = c.Constrain(Size{})
	return v.size
}

func (v *spacerView) Place(offset Point) { v.offset = offset }

func (v *spacerView) Draw(canvas Canvas) {}

type columnView struct {
	owner *node
	gap   float32
	align Alignment
	size  Size
}

func (v *columnView) Measure(c Constraints) Size {
	kids := v.owner.layoutChildren()
	var weightTotal float32
	for _, kid := range kids {
		weightTotal += kid.weight
	}
	var used float32
	for i, kid := range kids {
		if i > 0 {
			used += v.gap
		}
		if kid.weight > 0 {
			continue
		}
		kid.measured = kid.view().Measure(Constraints{MaxWidth: c.MaxWidth, MaxHeight: c.MaxHeight})
		used += kid.measured.Height
	}
	var share float32
	if weightTotal > 0 && c.MaxHeight < infinite && c.MaxHeight > used {
		share = c.MaxHeight - used
	}
	for _, kid := range kids {
		if kid.weight <= 0 {
			continue
		}
		main := share * kid.weight / weightTotal
		kid.measured = kid.view().Measure(Constraints{MinWidth: 0, MaxWidth: c.MaxWidth, MinHeight: main, MaxHeight: main})
	}
	var cross, total float32
	for i, kid := range kids {
		if i > 0 {
			total += v.gap
		}
		total += kid.measured.Height
		cross = max(cross, kid.measured.Width)
	}
	v.size = c.Constrain(Size{Width: cross, Height: total})
	return v.size
}

func (v *columnView) Place(offset Point) {
	y := offset.Y
	for i, kid := range v.owner.layoutChildren() {
		if i > 0 {
			y += v.gap
		}
		align := v.align
		if kid.alignSet {
			align = kid.align
		}
		kid.view().Place(Point{X: offset.X + alignOffset(align, v.size.Width, kid.measured.Width), Y: y})
		y += kid.measured.Height
	}
}

func (v *columnView) Draw(canvas Canvas) {
	for _, kid := range v.owner.layoutChildren() {
		kid.view().Draw(canvas)
	}
}

type rowView struct {
	owner *node
	gap   float32
	align Alignment
	size  Size
}

func (v *rowView) Measure(c Constraints) Size {
	kids := v.owner.layoutChildren()
	var weightTotal float32
	for _, kid := range kids {
		weightTotal += kid.weight
	}
	var used float32
	for i, kid := range kids {
		if i > 0 {
			used += v.gap
		}
		if kid.weight > 0 {
			continue
		}
		kid.measured = kid.view().Measure(Constraints{MaxWidth: c.MaxWidth, MaxHeight: c.MaxHeight})
		used += kid.measured.Width
	}
	var share float32
	if weightTotal > 0 && c.MaxWidth < infinite && c.MaxWidth > used {
		share = c.MaxWidth - used
	}
	for _, kid := range kids {
		if kid.weight <= 0 {
			continue
		}
		main := share * kid.weight / weightTotal
		kid.measured = kid.view().Measure(Constraints{MinWidth: main, MaxWidth: main, MinHeight: 0, MaxHeight: c.MaxHeight})
	}
	var cross, total float32
	for i, kid := range kids {
		if i > 0 {
			total += v.gap
		}
		total += kid.measured.Width
		cross = max(cross, kid.measured.Height)
	}
	v.size = c.Constrain(Size{Width: total, Height: cross})
	return v.size
}

func (v *rowView) Place(offset Point) {
	x := offset.X
	for i, kid := range v.owner.layoutChildren() {
		if i > 0 {
			x += v.gap
		}
		align := v.align
		if kid.alignSet {
			align = kid.align
		}
		kid.view().Place(Point{X: x, Y: offset.Y + alignOffset(align, v.size.Height, kid.measured.Height)})
		x += kid.measured.Width
	}
}

func (v *rowView) Draw(canvas Canvas) {
	for _, kid := range v.owner.layoutChildren() {
		kid.view().Draw(canvas)
	}
}

type boxView struct {
	owner *node
	align Alignment
	size  Size
}

func (v *boxView) Measure(c Constraints) Size {
	var cross, main float32
	for _, kid := range v.owner.layoutChildren() {
		kid.measured = kid.view().Measure(c)
		cross = max(cross, kid.measured.Width)
		main = max(main, kid.measured.Height)
	}
	v.size = c.Constrain(Size{Width: cross, Height: main})
	return v.size
}

func (v *boxView) Place(offset Point) {
	for _, kid := range v.owner.layoutChildren() {
		align := v.align
		if kid.alignSet {
			align = kid.align
		}
		kid.view().Place(Point{
			X: offset.X + alignOffset(align, v.size.Width, kid.measured.Width),
			Y: offset.Y + alignOffset(align, v.size.Height, kid.measured.Height),
		})
	}
}

func (v *boxView) Draw(canvas Canvas) {
	for _, kid := range v.owner.layoutChildren() {
		kid.view().Draw(canvas)
	}
}
