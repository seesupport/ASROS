package kernel

import (
	"rtos/lib"
	"unsafe"
)

// MemoryRegion represents a contiguous physical memory region
type MemoryRegion struct {
	Base    uint64
	Length  uint64
	Type    uint32 // EFI memory type
	Attribs uint64
}

var memoryRegions []MemoryRegion

// GetMemoryMap calls UEFI Boot Services GetMemoryMap and parses into our structure.
func getMemoryMap(boot *lib.EFIBootServices) {
	// First call to get buffer size
	var mapSize uintptr
	var mapKey uintptr
	var descSize uintptr
	var descVersion uint32
	status := boot.GetMemoryMap(&mapSize, nil, &mapKey, &descSize, &descVersion)
	if status != lib.EFI_SUCCESS && status != lib.EFI_ERROR {
		lib.PrintString("GetMemoryMap size failed\n")
		return
	}
	// Allocate buffer (we'll use a static array for simplicity, or allocate from pool later)
	// For now, we assume we have a fixed buffer; but we can use the UEFI boot services to allocate.
	// We'll allocate with AllocatePool.
	buffer, err := boot.AllocatePool(mapSize)
	if err != lib.EFI_SUCCESS {
		lib.PrintString("Failed to allocate memory map buffer\n")
		return
	}
	defer boot.FreePool(buffer)

	// Second call to get actual map
	status = boot.GetMemoryMap(&mapSize, buffer, &mapKey, &descSize, &descVersion)
	if status != lib.EFI_SUCCESS {
		lib.PrintString("GetMemoryMap failed\n")
		return
	}

	// Parse descriptors
	numEntries := uintptr(mapSize) / uintptr(descSize)
	lib.PrintString("Memory map entries: ")
	lib.PrintUint64(uint64(numEntries))
	lib.PrintString("\n")

	// Create slice for regions
	memoryRegions = make([]MemoryRegion, 0, numEntries)

	ptr := uintptr(buffer)
	for i := uintptr(0); i < numEntries; i++ {
		desc := (*lib.EFIMemoryDescriptor)(unsafe.Pointer(ptr))
		region := MemoryRegion{
			Base:    desc.PhysicalStart,
			Length:  desc.NumPages * 4096,
			Type:    desc.Type,
			Attribs: desc.Attribute,
		}
		memoryRegions = append(memoryRegions, region)
		ptr += uintptr(descSize)
	}
}

// DumpMemoryMap prints all regions for debugging.
func DumpMemoryMap() {
	lib.PrintString("Memory Map:\n")
	for _, r := range memoryRegions {
		lib.PrintString("  Base=0x")
		lib.PrintHex64(r.Base)
		lib.PrintString(" Length=0x")
		lib.PrintHex64(r.Length)
		lib.PrintString(" Type=")
		lib.PrintUint64(uint64(r.Type))
		lib.PrintString("\n")
	}
}
