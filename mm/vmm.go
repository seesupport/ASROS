package mm

import (
	"rtos/lib"
)

// KERNEL_BASE is the virtual base of the kernel (we keep it identity mapped for simplicity).
// Later we can move to higher half.
const KERNEL_BASE = 0xFFFFFFFF80000000 // not used yet; we keep identity.

// VMM provides functions to map virtual to physical pages.
type VMM struct {
	// We'll maintain a list of mappings; for now, just a stub.
}

var vmm *VMM

// InitVMM initialises the virtual memory manager.
func InitVMM() {
	vmm = &VMM{}
	lib.PrintString("VMM initialised\n")
}

// MapPage maps a single virtual page to a physical page with given flags.
func (vmm *VMM) MapPage(virt, phys uint64, flags uint64) error {
	// Walk page tables and set entry.
	// For now, we'll just print a message.
	lib.PrintString("MapPage: virt=0x")
	lib.PrintHex64(virt)
	lib.PrintString(" phys=0x")
	lib.PrintHex64(phys)
	lib.PrintString("\n")
	return nil
}

// UnmapPage unmaps a virtual page.
func (vmm *VMM) UnmapPage(virt uint64) {
	// ...
}
