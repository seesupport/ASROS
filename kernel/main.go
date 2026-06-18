package kernel

import (
	"rtos/cpu"
	"rtos/interrupts"
	"rtos/lib"
	"rtos/mm"
)

// KernelEntry is called from UEFI.
//export KernelEntry
func KernelEntry(imageHandle lib.EFIHandle, systemTable *lib.EFISystemTable) {
	lib.SetSystemTable(systemTable)

	lib.PrintString("\nRTOS for x86_64 (UEFI) - Phase 3\n")
	lib.PrintString("Kernel entry successful.\n")

	// Phase 1: get memory map
	boot := systemTable.Boot
	if boot == nil {
		lib.PrintString("ERROR: Boot services not available\n")
		return
	}
	mm.InitMemoryMap(boot)
	mm.DumpMemoryMap()

	// Phase 2: CPU and interrupt setup
	lib.PrintString("Initializing CPU...\n")
	cpu.InitCPU()

	lib.PrintString("Setting up GDT...\n")
	cpu.InitGDT()

	// Phase 3: Memory management
	lib.PrintString("Initializing physical memory manager...\n")
	mm.InitPhysMem()

	lib.PrintString("Enabling paging...\n")
	mm.InitPaging()

	lib.PrintString("Initializing VMM...\n")
	mm.InitVMM()

	lib.PrintString("Initializing heap allocator...\n")
	mm.InitHeap()

	// Test allocation
	lib.PrintString("Testing slab allocator...\n")
	p1 := mm.Alloc(64)
	if p1 != nil {
		lib.PrintString("Allocated 64 bytes at 0x")
		lib.PrintHex64(uint64(uintptr(p1)))
		lib.PrintString("\n")
		mm.Free(p1, 64)
		lib.PrintString("Freed\n")
	} else {
		lib.PrintString("Allocation failed\n")
	}

	// Now set up IDT (needs to be done after paging? It can be before, but we do it after for clarity)
	lib.PrintString("Setting up IDT...\n")
	cpu.InitIDT()

	lib.PrintString("Initializing APIC...\n")
	cpu.InitAPIC()

	lib.PrintString("Initializing timer...\n")
	interrupts.TimerInit()

	// Enable interrupts
	lib.PrintString("Enabling interrupts...\n")
	cpu.EnableInterrupts()

	lib.PrintString("Phase 3 complete. Idling with interrupts enabled.\n")
	for {
		// Wait for interrupts
		cpu.Halt()
	}
}
