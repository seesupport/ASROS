package mm

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

// InitMemoryMap parses the UEFI memory map and stores it.
func InitMemoryMap(boot *lib.EFIBootServices) {
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
	// Allocate buffer from UEFI boot services (temporary)
	buffer, err := boot.AllocatePool(mapSize)
	if err != lib.EFI_SUCCESS {
		lib.PrintString("Failed to allocate memory map buffer\n")
		return
	}
	defer boot.FreePool(buffer)

	status = boot.GetMemoryMap(&mapSize, buffer, &mapKey, &descSize, &descVersion)
	if status != lib.EFI_SUCCESS {
		lib.PrintString("GetMemoryMap failed\n")
		return
	}

	numEntries := uintptr(mapSize) / uintptr(descSize)
	lib.PrintString("Memory map entries: ")
	lib.PrintUint64(uint64(numEntries))
	lib.PrintString("\n")

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

// GetMemoryRegions returns the parsed memory map.
func GetMemoryRegions() []MemoryRegion {
	return memoryRegions
}

// DumpMemoryMap prints all regions.
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
