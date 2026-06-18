package kernel

import (
	"rtos/cpu"
	"rtos/interrupts"
	"rtos/lib"
)

// KernelEntry is called from UEFI.
//export KernelEntry
func KernelEntry(imageHandle lib.EFIHandle, systemTable *lib.EFISystemTable) {
	lib.SetSystemTable(systemTable)

	lib.PrintString("\nRTOS for x86_64 (UEFI) - Phase 2\n")
	lib.PrintString("Kernel entry successful.\n")

	// Phase 1: get memory map
	boot := systemTable.Boot
	if boot == nil {
		lib.PrintString("ERROR: Boot services not available\n")
		return
	}
	getMemoryMap(boot)
	DumpMemoryMap()

	// Phase 2: CPU and interrupt setup
	lib.PrintString("Initializing CPU...\n")
	cpu.InitCPU()

	lib.PrintString("Setting up GDT...\n")
	cpu.InitGDT()

	lib.PrintString("Setting up IDT...\n")
	cpu.InitIDT()

	lib.PrintString("Initializing APIC...\n")
	cpu.InitAPIC()

	lib.PrintString("Initializing timer...\n")
	interrupts.TimerInit()

	// Enable interrupts
	lib.PrintString("Enabling interrupts...\n")
	cpu.EnableInterrupts()

	lib.PrintString("Phase 2 complete. Idling with interrupts enabled.\n")
	// Idle loop
	for {
		// Wait for interrupts
		// We'll use HLT instruction
		asm("hlt")
	}
}
