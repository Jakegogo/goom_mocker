//go:build !go1.16
// +build !go1.16

// Package hack 对 go 系统包的 hack, 包含一些系统结构体的 copy，需要和不同的 go 版本保持同步
package hack

import (
	"unsafe"
	// nolint
	_ "unsafe"
)

// InterceptCallerSkip 拦截器 callerskip
const InterceptCallerSkip = 7

// Firstmoduledata keep async with runtime.Firstmoduledata
//
//go:linkname Firstmoduledata runtime.firstmoduledata
var Firstmoduledata Moduledata

// Moduledata keep async with runtime.Moduledata
type Moduledata struct {
	Pclntable []byte
	Ftab      []Functab
	// nolint
	filetab []uint32
	// nolint
	findfunctab uintptr
	// nolint
	minpc, maxpc uintptr
	// nolint
	text, etext uintptr
	// nolint
	noptrdata, enoptrdata uintptr
	// nolint
	data, edata uintptr
	// nolint
	bss, ebss uintptr
	// nolint
	noptrbss, enoptrbss uintptr
	// nolint
	end, gcdata, gcbss uintptr
	// nolint
	types, etypes uintptr
	// nolint
	textsectmap []textsect
	// Original type was []*Type
	// nolint
	typelinks []int32
	// nolint
	itablinks []*uintptr
	// nolint
	ptab []interface{}
	// nolint
	pluginpath string
	// nolint
	pkghashes []interface{}
	// nolint
	modulename string
	// nolint
	// Original type was []modulehash
	modulehashes []interface{}
	// nolint
	hasmain uint8 // 1 if module contains the main function, 0 otherwise
	// nolint
	gcdatamask, gcbssmask Bitvector
	// nolint
	typemap map[typeOff]*interface{} // offset to *_rtype in previous module
	// nolint
	bad bool // module failed to load and should be ignored
	// Next links next Moduledata
	Next *Moduledata
}

// Functab Functab
type Functab struct {
	Entry   uintptr
	Funcoff uintptr
}

// Func Convenience struct for modifying the underlying code pointer of a function
// value. The actual struct has other values, but always starts with a code
// pointer.
// TODO 不同 go 版本兼容
type Func struct {
	CodePtr uintptr
}

// Bitvector Bitvector
type Bitvector struct {
	// nolint
	n int32 // # of bits
	// nolint
	bytedata *uint8
}

// nolint
type textsect struct {
	// nolint
	vaddr    uintptr // prelinked section vaddr
	length   uintptr // section length
	baseaddr uintptr // relocated section address
}

// nolint
type typeOff int32 // offset to an *rtype

// Value reflect.Value
// TODO 不同 go 版本兼容
type Value struct {
	Typ  *uintptr
	Ptr  unsafe.Pointer
	Flag uintptr
}

// CheckNameOffOverflow check nameOff overflow
func CheckNameOffOverflow(f *runtime.Func, md *Moduledata) bool {
	return false
}
