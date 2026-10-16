package dwrite

import (
	"encoding/binary"
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/ebitengine/purego"

	"github.com/weqqr/rio/pkg/com"
)

const memFontKeySize = 8

type memFontLoader struct {
	vtbl uintptr
	refs int32
}

type memFontStream struct {
	vtbl uintptr
	refs int32
	data []byte
}

var (
	memFontVtablesOnce sync.Once
	loaderVtable       [4]uintptr
	streamVtable       [7]uintptr

	memFontLoaderInstance memFontLoader
	memFontLoaderObject   = com.WrapObject(unsafe.Pointer(&memFontLoaderInstance))

	memFonts = struct {
		sync.Mutex
		nextID    uint64
		data      map[uint64][]byte
		streams   map[uintptr]*memFontStream
		factories map[uintptr]struct{}
	}{
		data:      make(map[uint64][]byte),
		streams:   make(map[uintptr]*memFontStream),
		factories: make(map[uintptr]struct{}),
	}
)

func hrResult(hr com.HRESULT) uintptr { return uintptr(uint32(hr)) }

func initMemFontVtables() {
	memFontVtablesOnce.Do(func() {
		memFontLoaderInstance.vtbl = uintptr(unsafe.Pointer(&loaderVtable[0]))
		memFontLoaderInstance.refs = 1

		loaderVtable[0] = purego.NewCallback(memLoaderQueryInterface)
		loaderVtable[1] = purego.NewCallback(memLoaderAddRef)
		loaderVtable[2] = purego.NewCallback(memLoaderRelease)
		loaderVtable[3] = purego.NewCallback(memLoaderCreateStreamFromKey)

		streamVtable[0] = purego.NewCallback(memStreamQueryInterface)
		streamVtable[1] = purego.NewCallback(memStreamAddRef)
		streamVtable[2] = purego.NewCallback(memStreamRelease)
		streamVtable[3] = purego.NewCallback(memStreamReadFileFragment)
		streamVtable[4] = purego.NewCallback(memStreamReleaseFileFragment)
		streamVtable[5] = purego.NewCallback(memStreamGetFileSize)
		streamVtable[6] = purego.NewCallback(memStreamGetLastWriteTime)
	})
}

func memLoaderPtr() uintptr { return uintptr(unsafe.Pointer(&memFontLoaderInstance)) }

func (f IDWriteFactory) CreateFontFileReferenceFromMemory(data []byte, fontFile *IDWriteFontFile) com.HRESULT {
	if len(data) == 0 {
		return com.E_INVALIDARG
	}
	initMemFontVtables()
	if hr := registerMemFontLoader(f); hr.Failed() {
		return hr
	}

	key := make([]byte, memFontKeySize)
	memFonts.Lock()
	memFonts.nextID++
	id := memFonts.nextID
	binary.LittleEndian.PutUint64(key, id)
	memFonts.data[id] = data
	memFonts.Unlock()

	hr := f.CreateCustomFontFileReference(unsafe.Pointer(&key[0]), memFontKeySize, memFontLoaderObject, fontFile)
	if hr.Failed() {
		memFonts.Lock()
		delete(memFonts.data, id)
		memFonts.Unlock()
	}
	return hr
}

func registerMemFontLoader(f IDWriteFactory) com.HRESULT {
	ptr := uintptr(f.Ptr())
	memFonts.Lock()
	_, registered := memFonts.factories[ptr]
	memFonts.Unlock()
	if registered {
		return com.S_OK
	}
	hr := f.RegisterFontFileLoader(memFontLoaderObject)
	if hr.Failed() {
		return hr
	}
	memFonts.Lock()
	memFonts.factories[ptr] = struct{}{}
	memFonts.Unlock()
	return com.S_OK
}

func memLoaderQueryInterface(this uintptr, iid *com.IID, out *unsafe.Pointer) uintptr {
	if out == nil || this != memLoaderPtr() {
		return hrResult(com.E_POINTER)
	}
	if iid != nil && (*iid == com.IID_IUnknown || *iid == IID_IDWriteFontFileLoader) {
		atomic.AddInt32(&memFontLoaderInstance.refs, 1)
		*out = unsafe.Pointer(&memFontLoaderInstance)
		return hrResult(com.S_OK)
	}
	*out = nil
	return hrResult(com.E_NOINTERFACE)
}

func memLoaderAddRef(this uintptr) uintptr {
	if this != memLoaderPtr() {
		return 0
	}
	return uintptr(atomic.AddInt32(&memFontLoaderInstance.refs, 1))
}

func memLoaderRelease(this uintptr) uintptr {
	if this != memLoaderPtr() {
		return 0
	}
	return uintptr(atomic.AddInt32(&memFontLoaderInstance.refs, -1))
}

func memLoaderCreateStreamFromKey(this uintptr, key unsafe.Pointer, keySize uint32, out *unsafe.Pointer) uintptr {
	if this != memLoaderPtr() || out == nil || key == nil || keySize != memFontKeySize {
		return hrResult(com.E_INVALIDARG)
	}
	id := binary.LittleEndian.Uint64(unsafe.Slice((*byte)(key), memFontKeySize))

	memFonts.Lock()
	defer memFonts.Unlock()
	data, ok := memFonts.data[id]
	if !ok {
		*out = nil
		return hrResult(com.E_INVALIDARG)
	}
	stream := &memFontStream{
		vtbl: uintptr(unsafe.Pointer(&streamVtable[0])),
		refs: 1,
		data: data,
	}
	memFonts.streams[uintptr(unsafe.Pointer(stream))] = stream
	*out = unsafe.Pointer(stream)
	return hrResult(com.S_OK)
}

func memStreamQueryInterface(this uintptr, iid *com.IID, out *unsafe.Pointer) uintptr {
	if out == nil {
		return hrResult(com.E_POINTER)
	}
	wanted := iid != nil && (*iid == com.IID_IUnknown || *iid == IID_IDWriteFontFileStream)

	memFonts.Lock()
	stream := memFonts.streams[this]
	if stream != nil && wanted {
		atomic.AddInt32(&stream.refs, 1)
	}
	memFonts.Unlock()

	if stream == nil {
		return hrResult(com.E_POINTER)
	}
	if wanted {
		*out = unsafe.Pointer(stream)
		return hrResult(com.S_OK)
	}
	*out = nil
	return hrResult(com.E_NOINTERFACE)
}

func memStreamAddRef(this uintptr) uintptr {
	memFonts.Lock()
	defer memFonts.Unlock()
	if stream := memFonts.streams[this]; stream != nil {
		return uintptr(atomic.AddInt32(&stream.refs, 1))
	}
	return 0
}

func memStreamRelease(this uintptr) uintptr {
	memFonts.Lock()
	defer memFonts.Unlock()
	stream := memFonts.streams[this]
	if stream == nil {
		return 0
	}
	refs := atomic.AddInt32(&stream.refs, -1)
	if refs <= 0 {
		delete(memFonts.streams, this)
		return 0
	}
	return uintptr(refs)
}

func memStreamReadFileFragment(this uintptr, fragmentStart *unsafe.Pointer, fileOffset uint64, fragmentSize uint64, fragmentContext *unsafe.Pointer) uintptr {
	if fragmentStart == nil || fragmentContext == nil {
		return hrResult(com.E_POINTER)
	}
	memFonts.Lock()
	stream := memFonts.streams[this]
	memFonts.Unlock()
	if stream == nil {
		return hrResult(com.E_POINTER)
	}
	size := uint64(len(stream.data))
	if fileOffset > size || fragmentSize > size-fileOffset {
		return hrResult(com.E_INVALIDARG)
	}
	*fragmentContext = nil
	if fragmentSize == 0 {
		*fragmentStart = nil
		return hrResult(com.S_OK)
	}
	*fragmentStart = unsafe.Pointer(&stream.data[fileOffset])
	return hrResult(com.S_OK)
}

func memStreamReleaseFileFragment(_ uintptr, _ unsafe.Pointer) uintptr {
	return hrResult(com.S_OK)
}

func memStreamGetFileSize(this uintptr, fileSize *uint64) uintptr {
	if fileSize == nil {
		return hrResult(com.E_POINTER)
	}
	memFonts.Lock()
	stream := memFonts.streams[this]
	memFonts.Unlock()
	if stream == nil {
		return hrResult(com.E_POINTER)
	}
	*fileSize = uint64(len(stream.data))
	return hrResult(com.S_OK)
}

func memStreamGetLastWriteTime(_ uintptr, _ *uint64) uintptr {
	return hrResult(com.E_NOTIMPL)
}
