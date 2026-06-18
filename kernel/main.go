package kernel

import (
	"rtos/cpu"
	"rtos/interrupts"
	"rtos/lib"
	"rtos/mm"
	"rtos/scheduler"
)

// KernelEntry is called from UEFI.
//export KernelEntry
func KernelEntry(imageHandle lib.EFIHandle, systemTable *lib.EFISystemTable) {
	lib.SetSystemTable(systemTable)

	lib.PrintString("\nRTOS for x86_64 (UEFI) - Phase 4\n")
	lib.PrintString("Kernel entry successful.\n")

	// Phase 1: memory map
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

	// Phase 4: Scheduler
	lib.PrintString("Setting up IDT...\n")
	cpu.InitIDT()

	lib.PrintString("Initializing APIC...\n")
	cpu.InitAPIC()

	lib.PrintString("Initializing timer...\n")
	interrupts.TimerInit()

	// Create a test task
	lib.PrintString("Creating test task...\n")
	testTask := scheduler.CreateTask(testTaskFunc, 1)
	if testTask == nil {
		lib.PrintString("Failed to create test task\n")
	} else {
		lib.PrintString("Test task created\n")
	}

	lib.PrintString("Initializing scheduler...\n")
	scheduler.Init()

	lib.PrintString("Starting scheduler...\n")
	scheduler.Start()

	// Should never reach here
	for {
		cpu.Halt()
	}
}

// testTaskFunc is a simple task that prints and yields.
func testTaskFunc() {
	for {
		lib.PrintString("Hello from test task\n")
		scheduler.Yield()
	}
}
