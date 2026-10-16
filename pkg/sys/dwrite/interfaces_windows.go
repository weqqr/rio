package dwrite

import (
	"math"
	"unsafe"

	"github.com/weqqr/rio/pkg/com"
	"github.com/weqqr/rio/pkg/sys/win32"
)

const (
	slotFactoryCreateFontFileReference       = 7
	slotFactoryCreateCustomFontFileReference = 8
	slotFactoryCreateFontFace                = 9
	slotFactoryRegisterFontFileLoader        = 13
	slotFactoryCreateGlyphRunAnalysis        = 23
)

type IDWriteFactory struct{ com.Object }

func (f IDWriteFactory) CreateFontFileReference(filePath *uint16, lastWriteTime unsafe.Pointer, fontFile *IDWriteFontFile) com.HRESULT {
	return com.HRESULT(f.Object.Call(slotFactoryCreateFontFileReference,
		uintptr(unsafe.Pointer(filePath)), uintptr(lastWriteTime),
		uintptr(unsafe.Pointer(&fontFile.Object))))
}

func (f IDWriteFactory) CreateCustomFontFileReference(key unsafe.Pointer, keySize uint32, loader com.Object, fontFile *IDWriteFontFile) com.HRESULT {
	return com.HRESULT(f.Object.Call(slotFactoryCreateCustomFontFileReference,
		uintptr(key), uintptr(keySize), uintptr(loader.Ptr()),
		uintptr(unsafe.Pointer(&fontFile.Object))))
}

func (f IDWriteFactory) CreateFontFace(faceType DWRITE_FONT_FACE_TYPE, fileCount uint32, files *unsafe.Pointer, faceIndex uint32, simulations DWRITE_FONT_SIMULATIONS, fontFace *IDWriteFontFace) com.HRESULT {
	return com.HRESULT(f.Object.Call(slotFactoryCreateFontFace,
		uintptr(faceType), uintptr(fileCount), uintptr(unsafe.Pointer(files)),
		uintptr(faceIndex), uintptr(simulations),
		uintptr(unsafe.Pointer(&fontFace.Object))))
}

func (f IDWriteFactory) RegisterFontFileLoader(loader com.Object) com.HRESULT {
	return com.HRESULT(f.Object.Call(slotFactoryRegisterFontFileLoader, uintptr(loader.Ptr())))
}

func (f IDWriteFactory) CreateGlyphRunAnalysis(run *DWRITE_GLYPH_RUN, pixelsPerDip float32, transform unsafe.Pointer, renderingMode DWRITE_RENDERING_MODE, measuringMode DWRITE_MEASURING_MODE, baselineOriginX, baselineOriginY float32, analysis *IDWriteGlyphRunAnalysis) com.HRESULT {
	return com.HRESULT(f.Object.Call(slotFactoryCreateGlyphRunAnalysis,
		uintptr(unsafe.Pointer(run)), uintptr(math.Float32bits(pixelsPerDip)),
		uintptr(transform), uintptr(renderingMode), uintptr(measuringMode),
		uintptr(math.Float32bits(baselineOriginX)), uintptr(math.Float32bits(baselineOriginY)),
		uintptr(unsafe.Pointer(&analysis.Object))))
}

type IDWriteFontFile struct{ com.Object }

const (
	slotFontFaceGetMetrics            = 8
	slotFontFaceGetDesignGlyphMetrics = 10
	slotFontFaceGetGlyphIndices       = 11
)

type IDWriteFontFace struct{ com.Object }

func (f IDWriteFontFace) GetMetrics() DWRITE_FONT_METRICS {
	var metrics DWRITE_FONT_METRICS
	f.Object.Call(slotFontFaceGetMetrics, uintptr(unsafe.Pointer(&metrics)))
	return metrics
}

func (f IDWriteFontFace) GetDesignGlyphMetrics(glyphIndices *uint16, glyphCount uint32, glyphMetrics *DWRITE_GLYPH_METRICS, isSideways int32) com.HRESULT {
	return com.HRESULT(f.Object.Call(slotFontFaceGetDesignGlyphMetrics,
		uintptr(unsafe.Pointer(glyphIndices)), uintptr(glyphCount),
		uintptr(unsafe.Pointer(glyphMetrics)), uintptr(isSideways)))
}

func (f IDWriteFontFace) GetGlyphIndices(codePoints *uint32, codePointCount uint32, glyphIndices *uint16) com.HRESULT {
	return com.HRESULT(f.Object.Call(slotFontFaceGetGlyphIndices,
		uintptr(unsafe.Pointer(codePoints)), uintptr(codePointCount),
		uintptr(unsafe.Pointer(glyphIndices))))
}

const (
	slotFontFace1GetKerningPairAdjustments = 25
	slotFontFace1HasKerningPairs           = 26
)

type IDWriteFontFace1 struct{ com.Object }

func (f IDWriteFontFace1) GetKerningPairAdjustments(glyphCount uint32, glyphIndices *uint16, kerningPairAdjustments *int32) com.HRESULT {
	return com.HRESULT(f.Object.Call(slotFontFace1GetKerningPairAdjustments,
		uintptr(glyphCount), uintptr(unsafe.Pointer(glyphIndices)),
		uintptr(unsafe.Pointer(kerningPairAdjustments))))
}

func (f IDWriteFontFace1) HasKerningPairs() bool {
	return f.Object.Call(slotFontFace1HasKerningPairs) != 0
}

type IDWriteGlyphRunAnalysis struct{ com.Object }

const (
	slotGlyphRunAnalysisGetAlphaTextureBounds = 3
	slotGlyphRunAnalysisCreateAlphaTexture    = 4
)

func (a IDWriteGlyphRunAnalysis) GetAlphaTextureBounds(textureType DWRITE_TEXTURE_TYPE, bounds *win32.RECT) com.HRESULT {
	return com.HRESULT(a.Object.Call(slotGlyphRunAnalysisGetAlphaTextureBounds,
		uintptr(textureType), uintptr(unsafe.Pointer(bounds))))
}

func (a IDWriteGlyphRunAnalysis) CreateAlphaTexture(textureType DWRITE_TEXTURE_TYPE, bounds *win32.RECT, alphaValues *byte, bufferSize uint32) com.HRESULT {
	return com.HRESULT(a.Object.Call(slotGlyphRunAnalysisCreateAlphaTexture,
		uintptr(textureType), uintptr(unsafe.Pointer(bounds)),
		uintptr(unsafe.Pointer(alphaValues)), uintptr(bufferSize)))
}
