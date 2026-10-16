package view

import "slices"

type partKind uint8

const (
	partPadding partKind = iota
	partSize
	partFill
	partBackground
	partAlign
	partWeight
	partGap
	partTextSize
)

type part struct {
	kind          partKind
	all, top      float32
	right, bottom float32
	left, width   float32
	height        float32
	fillWidth     bool
	fillHeight    bool
	align         Alignment
	color         Color
}

type Modifier struct {
	parts []part
}

func (m Modifier) Padding(all float32) Modifier {
	return m.append(part{kind: partPadding, all: all})
}

func (m Modifier) PaddingEach(top, right, bottom, left float32) Modifier {
	return m.append(part{kind: partPadding, top: top, right: right, bottom: bottom, left: left})
}

func (m Modifier) Sized(width, height float32) Modifier {
	return m.append(part{kind: partSize, width: width, height: height})
}

func (m Modifier) FillMaxWidth() Modifier {
	return m.append(part{kind: partFill, fillWidth: true})
}

func (m Modifier) FillMaxHeight() Modifier {
	return m.append(part{kind: partFill, fillHeight: true})
}

func (m Modifier) FillMaxSize() Modifier {
	return m.append(part{kind: partFill, fillWidth: true, fillHeight: true})
}

func (m Modifier) Background(color Color) Modifier {
	return m.append(part{kind: partBackground, color: color})
}

func (m Modifier) Align(align Alignment) Modifier {
	return m.append(part{kind: partAlign, align: align})
}

func (m Modifier) Weight(weight float32) Modifier {
	return m.append(part{kind: partWeight, all: weight})
}

func (m Modifier) Gap(gap float32) Modifier {
	return m.append(part{kind: partGap, all: gap})
}

func (m Modifier) TextSize(size float32) Modifier {
	return m.append(part{kind: partTextSize, all: size})
}

func Padding(all float32) Modifier { return Modifier{}.Padding(all) }

func PaddingEach(top, right, bottom, left float32) Modifier {
	return Modifier{}.PaddingEach(top, right, bottom, left)
}

func Sized(width, height float32) Modifier { return Modifier{}.Sized(width, height) }

func FillMaxWidth() Modifier { return Modifier{}.FillMaxWidth() }

func FillMaxHeight() Modifier { return Modifier{}.FillMaxHeight() }

func FillMaxSize() Modifier { return Modifier{}.FillMaxSize() }

func Background(color Color) Modifier { return Modifier{}.Background(color) }

func Align(align Alignment) Modifier { return Modifier{}.Align(align) }

func Weight(weight float32) Modifier { return Modifier{}.Weight(weight) }

func Gap(gap float32) Modifier { return Modifier{}.Gap(gap) }

func TextSize(size float32) Modifier { return Modifier{}.TextSize(size) }

func (m Modifier) append(p part) Modifier {
	parts := make([]part, len(m.parts), len(m.parts)+1)
	copy(parts, m.parts)
	return Modifier{parts: append(parts, p)}
}

func (m Modifier) same(other Modifier) bool {
	return slices.Equal(m.parts, other.parts)
}

func merge(mods []Modifier) Modifier {
	switch len(mods) {
	case 0:
		return Modifier{}
	case 1:
		return mods[0]
	}
	var parts []part
	for _, m := range mods {
		parts = append(parts, m.parts...)
	}
	return Modifier{parts: parts}
}

func (m Modifier) alignment() (Alignment, bool) {
	for _, p := range m.parts {
		if p.kind == partAlign {
			return p.align, true
		}
	}
	return AlignStart, false
}

func (m Modifier) flex() float32 {
	var weight float32
	for _, p := range m.parts {
		if p.kind == partWeight {
			weight = p.all
		}
	}
	return weight
}

func (m Modifier) spacing() float32 {
	var gap float32
	for _, p := range m.parts {
		if p.kind == partGap {
			gap = p.all
		}
	}
	return gap
}

func (m Modifier) font() float32 {
	var size float32
	for _, p := range m.parts {
		if p.kind == partTextSize {
			size = p.all
		}
	}
	return size
}

func (m Modifier) wrap(inner View) View {
	for i := len(m.parts) - 1; i >= 0; i-- {
		p := m.parts[i]
		switch p.kind {
		case partPadding:
			top, right, bottom, left := p.top, p.right, p.bottom, p.left
			if p.all != 0 {
				top, right, bottom, left = p.all, p.all, p.all, p.all
			}
			inner = &paddingView{inner: inner, top: top, right: right, bottom: bottom, left: left}
		case partSize:
			inner = &sizeView{inner: inner, width: p.width, height: p.height}
		case partFill:
			inner = &fillView{inner: inner, width: p.fillWidth, height: p.fillHeight}
		case partBackground:
			inner = &backgroundView{inner: inner, color: p.color}
		}
	}
	return inner
}
