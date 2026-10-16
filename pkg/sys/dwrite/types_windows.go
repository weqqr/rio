package dwrite

import (
	"structs"
	"unsafe"

	"github.com/weqqr/rio/pkg/com"
)

var (
	IID_IDWriteFactory        = com.IID{Data1: 0xB859EE5A, Data2: 0xD838, Data3: 0x4B5B, Data4: [8]byte{0xA2, 0xE8, 0x1A, 0xDC, 0x7D, 0x93, 0xDB, 0x48}}
	IID_IDWriteFontFace1      = com.IID{Data1: 0xA71EFDB4, Data2: 0x9FDB, Data3: 0x4838, Data4: [8]byte{0xAD, 0x90, 0xCF, 0xC3, 0xBE, 0x8C, 0x3D, 0xAF}}
	IID_IDWriteFontFileLoader = com.IID{Data1: 0x727CAD4E, Data2: 0xD6AF, Data3: 0x4C9E, Data4: [8]byte{0x8A, 0x08, 0xD6, 0x95, 0xB1, 0x1C, 0xAA, 0x49}}
	IID_IDWriteFontFileStream = com.IID{Data1: 0x6D4865FE, Data2: 0x0AB8, Data3: 0x4D91, Data4: [8]byte{0x8F, 0x62, 0x5D, 0xD6, 0xBE, 0x34, 0xA3, 0xE0}}
)

type DWRITE_FACTORY_TYPE int32

const DWRITE_FACTORY_TYPE_SHARED DWRITE_FACTORY_TYPE = 0

type DWRITE_FONT_FACE_TYPE int32

const DWRITE_FONT_FACE_TYPE_TRUETYPE DWRITE_FONT_FACE_TYPE = 1

type DWRITE_FONT_SIMULATIONS int32

const DWRITE_FONT_SIMULATIONS_NONE DWRITE_FONT_SIMULATIONS = 0

type DWRITE_RENDERING_MODE int32

const DWRITE_RENDERING_MODE_NATURAL_SYMMETRIC DWRITE_RENDERING_MODE = 5

type DWRITE_MEASURING_MODE int32

const DWRITE_MEASURING_MODE_NATURAL DWRITE_MEASURING_MODE = 0

type DWRITE_TEXTURE_TYPE int32

const (
	DWRITE_TEXTURE_ALIASED_1x1   DWRITE_TEXTURE_TYPE = 0
	DWRITE_TEXTURE_CLEARTYPE_3x1 DWRITE_TEXTURE_TYPE = 1
)

type DWRITE_FONT_METRICS struct {
	_                      structs.HostLayout
	DesignUnitsPerEm       uint16
	Ascent                 uint16
	Descent                uint16
	LineGap                int16
	CapHeight              uint16
	XHeight                uint16
	UnderlinePosition      int16
	UnderlineThickness     uint16
	StrikethroughPosition  int16
	StrikethroughThickness uint16
}

type DWRITE_GLYPH_METRICS struct {
	_                 structs.HostLayout
	LeftSideBearing   int32
	AdvanceWidth      uint32
	RightSideBearing  int32
	TopSideBearing    int32
	AdvanceHeight     uint32
	BottomSideBearing int32
	VerticalOriginY   int32
}

type DWRITE_GLYPH_OFFSET struct {
	_              structs.HostLayout
	AdvanceOffset  float32
	AscenderOffset float32
}

type DWRITE_GLYPH_RUN struct {
	_             structs.HostLayout
	FontFace      unsafe.Pointer
	FontEmSize    float32
	GlyphCount    uint32
	GlyphIndices  *uint16
	GlyphAdvances *float32
	GlyphOffsets  *DWRITE_GLYPH_OFFSET
	IsSideways    int32
	BidiLevel     uint32
}
