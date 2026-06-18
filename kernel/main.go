package kernel

import (
	"rtos/lib"
)

// KernelEntry is called from the UEFI bootloader.
//export KernelEntry
func KernelEntry(imageHandle lib.EFIHandle, systemTable *lib.EFISystemTable) {
	// Ensure system table is set
	lib.SetSystemTable(systemTable)

	// Print initial banner
	lib.PrintString("\nRTOS for x86_64 (UEFI) - Phase 1\n")
	lib.PrintString("Kernel entry successful.\n")

	// Retrieve memory map
	boot := systemTable.Boot
	if boot == nil {
		lib.PrintString("ERROR: Boot services not available\n")
		return
	}
	getMemoryMap(boot)
	DumpMemoryMap()

	lib.PrintString("Phase 1 complete. Halting.\n")
	// Halt (we'll later jump to Phase 2)
	for {
		// Spin
	}
}
