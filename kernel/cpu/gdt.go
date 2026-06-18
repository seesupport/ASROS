package cpu

import (
	"unsafe"
)

// GDT entry structure (8 bytes)
type GDTEntry struct {
	LimitLow   uint16
	BaseLow    uint16
	BaseMid    uint8
	Access     uint8
	Granularity uint8
	BaseHigh   uint8
}

// GDTR structure (used with LGDT)
type GDTR struct {
	Limit uint16
	Base  uint64
}

// GDT selectors
const (
	GDTNull      = 0
	GDTKernelCode = 1 << 3
	GDTKernelData = 2 << 3
	GDTUserCode   = 3 << 3
	GDTUserData   = 4 << 3
)

var gdtEntries [5]GDTEntry
var gdtr GDTR

// initGDT sets up the GDT for long mode.
func InitGDT() {
	// Null descriptor
	gdtEntries[0] = GDTEntry{}

	// Kernel code: 64-bit, executable, readable, present, ring 0
	gdtEntries[1] = GDTEntry{
		LimitLow:   0xFFFF,
		BaseLow:    0,
		BaseMid:    0,
		Access:     0x9A, // Present, ring 0, code, executable, readable
		Granularity: 0xAF, // Page granular, 64-bit, limit high bits
		BaseHigh:   0,
	}

	// Kernel data: 64-bit, writable, present, ring 0
	gdtEntries[2] = GDTEntry{
		LimitLow:   0xFFFF,
		BaseLow:    0,
		BaseMid:    0,
		Access:     0x92, // Present, ring 0, data, writable
		Granularity: 0xAF,
		BaseHigh:   0,
	}

	// User code: 64-bit, executable, readable, present, ring 3
	gdtEntries[3] = GDTEntry{
		LimitLow:   0xFFFF,
		BaseLow:    0,
		BaseMid:    0,
		Access:     0xFA, // Present, ring 3, code, executable, readable
		Granularity: 0xAF,
		BaseHigh:   0,
	}

	// User data: 64-bit, writable, present, ring 3
	gdtEntries[4] = GDTEntry{
		LimitLow:   0xFFFF,
		BaseLow:    0,
		BaseMid:    0,
		Access:     0xF2, // Present, ring 3, data, writable
		Granularity: 0xAF,
		BaseHigh:   0,
	}

	// Set GDTR
	gdtr.Limit = uint16(len(gdtEntries)*8 - 1)
	gdtr.Base = uint64(uintptr(unsafe.Pointer(&gdtEntries[0])))

	// Load GDT
	loadGDT(&gdtr)

	// Reload segment registers (in assembly)
	reloadSegments()
}

// loadGDT is implemented in assembly
//go:linkname loadGDT
func loadGDT(gdtr *GDTR)

// reloadSegments reloads CS, DS, ES, FS, GS, SS (via far jump for CS)
//go:linkname reloadSegments
func reloadSegments()
