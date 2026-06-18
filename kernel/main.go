package kernel

import (
	"rtos/cpu"
	"rtos/interrupts"
	"rtos/lib"
	"rtos/mm"
	"rtos/scheduler"
	"rtos/sync"
)

var (
	mutex sync.Mutex
	shared int
)

func producer() {
	for {
		mutex.Lock()
		shared++
		lib.PrintString("Producer: ")
		lib.PrintUint64(uint64(shared))
		lib.PrintString("\n")
		mutex.Unlock()
		scheduler.Yield()
	}
}

func consumer() {
	for {
		mutex.Lock()
		if shared > 0 {
			shared--
			lib.PrintString("Consumer: ")
			lib.PrintUint64(uint64(shared))
			lib.PrintString("\n")
		}
		mutex.Unlock()
		scheduler.Yield()
	}
}

// KernelEntry is called from UEFI.
//export KernelEntry
func KernelEntry(imageHandle lib.EFIHandle, systemTable *lib.EFISystemTable) {
	lib.SetSystemTable(systemTable)

	lib.PrintString("\nRTOS for x86_64 (UEFI) - Phase 5\n")
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

	// Phase 5: Synchronization
	// Create producer and consumer tasks
	lib.PrintString("Creating producer task...\n")
	scheduler.CreateTask(producer, 2)
	lib.PrintString("Creating consumer task...\n")
	scheduler.CreateTask(consumer, 2)

	lib.PrintString("Initializing scheduler...\n")
	scheduler.Init()

	lib.PrintString("Starting scheduler...\n")
	scheduler.Start()

	// Should never reach here
	for {
		cpu.Halt()
	}
}
