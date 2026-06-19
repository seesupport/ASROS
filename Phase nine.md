We now implement **Phase 9 – RTOS Runtime Layer**, providing a deterministic, GC‑free runtime environment for the kernel and tasks. This layer integrates our custom memory allocator and scheduler, ensuring that all kernel‑critical paths avoid unpredictable pauses.

---

## Phase 9: RTOS Runtime Layer

### 1. Design Overview

The standard Go runtime includes a garbage collector and a scheduler that are unsuitable for hard real‑time systems. We address this by:

- **Disabling garbage collection** via the TinyGo compiler flag `‑gc=none`.
- **Replacing all dynamic memory allocation** in the kernel with our own deterministic slab allocator (`mm` package).
- **Using our own task scheduler** (`scheduler` package) instead of the Go scheduler.
- **Providing a thin wrapper** (`runtime` package) that exposes safe allocation and task creation functions to the rest of the kernel.

This ensures that:
- Memory allocation is O(1) and free of fragmentation.
- No stop‑the‑world pauses occur.
- Task scheduling is fully deterministic and preemptive.

### 2. Architecture Diagram (ASCII)

```
+-----------------------+
|   User Tasks (Go fn)  |
+-----------------------+
|   RTOS Runtime Layer  |   <-- runtime.Alloc, runtime.CreateTask, runtime.Yield
| (GC‑free, deterministic) |
+-----------------------+
|  Scheduler & Task Mgmt |
+-----------------------+
|  Slab Allocator (mm)  |
+-----------------------+
|  Memory Manager (phys/virt) |
+-----------------------+
|      Hardware          |
+-----------------------+
```

The runtime layer acts as the interface between application‑level Go code and the kernel primitives.

### 3. Source Code

New files:

```
/runtime/runtime.go
/runtime/noop_gc.go   (optional, for compatibility)
```

Modified files:

```
/build.sh   (add `-gc=none`)
/kernel/main.go  (add `runtime.Init()`)
```

---

#### File: `/runtime/runtime.go`

```go
// Package runtime provides a deterministic, GC‑free runtime layer for the RTOS.
// It wraps kernel memory management and scheduling primitives.
package runtime

import (
	"rtos/mm"
	"rtos/scheduler"
	"unsafe"
)

// Init initializes the runtime environment.
// It ensures the slab allocator is ready and sets up any needed runtime hooks.
// This must be called before any other runtime functions.
func Init() {
	// The slab allocator is already initialized in kernel main,
	// but we ensure it's ready here.
	mm.InitHeap()
	// No GC to disable – the compiler flag `-gc=none` takes care of that.
}

// Alloc allocates a block of memory of at least `size` bytes.
// Returns a pointer to the block, or nil if allocation fails.
// This is the preferred method for all dynamic memory in the kernel.
func Alloc(size uint64) unsafe.Pointer {
	return mm.Alloc(size)
}

// Free releases a block previously allocated by Alloc.
// The `size` must match the original allocation size.
func Free(ptr unsafe.Pointer, size uint64) {
	mm.Free(ptr, size)
}

// CreateTask creates a new task with the given entry function and priority.
// The new task is added to the ready queue and will run when scheduled.
func CreateTask(fn func(), priority int) *scheduler.TaskControlBlock {
	return scheduler.CreateTask(fn, priority)
}

// Yield voluntarily relinquishes the CPU to the scheduler.
// The current task is moved to the back of its priority queue.
func Yield() {
	scheduler.Yield()
}

// GetCurrentTask returns the TaskControlBlock of the currently running task.
func GetCurrentTask() *scheduler.TaskControlBlock {
	return scheduler.GetCurrentTask()
}

// TaskExit is called when a task completes its entry function.
// It marks the task as terminated and yields.
//go:noinline
func TaskExit() {
	scheduler.TaskExit()
}
```

#### File: `/runtime/noop_gc.go`

This file is optional; the compiler flag `-gc=none` already disables the GC. We keep it as a placeholder to satisfy any explicit calls to `runtime.GC()` in third‑party code.

```go
//go:build !gc
package runtime

// GC is a no‑op because garbage collection is disabled.
func GC() {
	// nothing
}
```

---

#### Modified File: `/build.sh`

Add the `-gc=none` flag to disable garbage collection.

```bash
#!/bin/bash
TINYGO=${TINYGO:-tinygo}
# Generate ISR stubs
python3 tools/gen_isr.py
# Build with GC disabled and no debugging
$TINYGO build -o rtos.efi -target=x86_64-unknown-uefi -gc=none -no-debug ./boot/uefi
```

---

#### Modified File: `/kernel/main.go`

Insert `runtime.Init()` early in the boot sequence, before any dynamic allocations.

```go
package kernel

import (
	"rtos/cpu"
	"rtos/drivers/hpet"
	"rtos/drivers/keyboard"
	"rtos/drivers/pci"
	"rtos/drivers/serial"
	"rtos/drivers/storage"
	"rtos/fs"
	"rtos/interrupts"
	"rtos/lib"
	"rtos/mm"
	"rtos/net"
	"rtos/runtime"
	"rtos/scheduler"
)

// KernelEntry is the main entry point.
//export KernelEntry
func KernelEntry(imageHandle lib.EFIHandle, systemTable *lib.EFISystemTable) {
	lib.SetSystemTable(systemTable)

	lib.PrintString("\nRTOS for x86_64 (UEFI) - Phase 9\n")
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

	// Phase 9: Runtime initialization – must happen before any allocation.
	lib.PrintString("Initializing runtime (GC‑free)...\n")
	runtime.Init()

	// Phase 4: Scheduler (needs heap)
	lib.PrintString("Setting up IDT...\n")
	cpu.InitIDT()
	lib.PrintString("Initializing APIC...\n")
	cpu.InitAPIC()
	lib.PrintString("Initializing timer...\n")
	interrupts.TimerInit()

	// Phase 6: Drivers
	initDrivers()

	// Phase 7: Filesystem
	testFilesystem()

	// Phase 8: Networking
	testNetworking()

	// Create test tasks using the runtime wrapper
	lib.PrintString("Creating test tasks...\n")
	runtime.CreateTask(testTask, 2)

	// Phase 4: Scheduler start
	lib.PrintString("Initializing scheduler...\n")
	scheduler.Init()
	lib.PrintString("Starting scheduler...\n")
	scheduler.Start()

	// Should never reach here
	for {
		cpu.Halt()
	}
}
```

---

### 4. Interfaces (Public APIs)

The runtime package exposes the following functions:

- `runtime.Init()` – must be called once at boot.
- `runtime.Alloc(size uint64) unsafe.Pointer` – allocate memory.
- `runtime.Free(ptr unsafe.Pointer, size uint64)` – free memory.
- `runtime.CreateTask(fn func(), priority int) *scheduler.TaskControlBlock` – create a new task.
- `runtime.Yield()` – yield the CPU.
- `runtime.GetCurrentTask() *scheduler.TaskControlBlock` – get the running task.
- `runtime.TaskExit()` – terminate the current task (used internally).

All allocations are served by the slab allocator; no garbage collection is active.

---

### 5. Initialization Sequence

1. `KernelEntry` starts.
2. Memory map and physical memory manager are initialised.
3. Paging and virtual memory are enabled.
4. Slab allocator (`mm.InitHeap`) is initialised.
5. `runtime.Init()` is called – this ensures the allocator is ready and GC is off.
6. The scheduler, interrupts, drivers, filesystem, and network stack are initialised.
7. Tasks are created using `runtime.CreateTask`.
8. `scheduler.Start()` is called and never returns.

---

### 6. Testing Procedures

**Build**:
```bash
./build.sh
```

**Run**:
```bash
./run_qemu.sh
```

**Expected Output**:
- The system should boot and print all earlier phase messages.
- The test tasks should run and print their outputs.
- There should be **no GC‑related output** (no “GC” messages).
- Memory allocation tests (e.g., creating many tasks or allocating large objects) should succeed without unpredictable pauses.

**Performance Verification**:
- Interrupt latency and context‑switch times can be measured using HPET (already implemented) and should remain within the target bounds, as GC is disabled.

---

### 7. How Deterministic Behavior Is Achieved

| Concern                     | Solution |
|-----------------------------|----------|
| Garbage collection pauses   | Disabled via `-gc=none` compiler flag. |
| Non‑deterministic allocation | All kernel objects are allocated from fixed‑size slabs (O(1) allocate/free). |
| Unpredictable scheduling    | Our own priority‑based preemptive scheduler is used; no Go scheduler interference. |
| Heap fragmentation          | Slab allocator eliminates fragmentation within each size class. |
| Memory pooling              | Tasks and common structures are allocated from dedicated slabs. |
| Interrupt latency           | Minimal top‑half handlers; all work is deferred to tasks or bottom halves. |

All kernel‑critical paths avoid `new`, `make`, and `append` that could allocate from the Go heap. Where necessary, we allocate using `runtime.Alloc` and manage memory explicitly.

---

### 8. Unit and Integration Tests

We will include a test task that:

- Allocates and frees memory in a loop.
- Creates multiple tasks and measures the time to switch between them.
- Verifies that no GC‑related stalls occur.

Example test task (to be added to the test suite):

```go
func testAllocTask() {
	for {
		// Allocate 128 bytes
		ptr := runtime.Alloc(128)
		if ptr != nil {
			// Write pattern
			for i := 0; i < 128; i++ {
				*(*byte)(unsafe.Pointer(uintptr(ptr) + uintptr(i))) = byte(i)
			}
			// Verify
			ok := true
			for i := 0; i < 128; i++ {
				if *(*byte)(unsafe.Pointer(uintptr(ptr) + uintptr(i))) != byte(i) {
					ok = false
					break
				}
			}
			if ok {
				lib.PrintString("Memory test passed\n")
			} else {
				lib.PrintString("Memory test failed\n")
			}
			runtime.Free(ptr, 128)
		}
		runtime.Yield()
	}
}
```

---

### 9. What Was Implemented

- A runtime package that wraps slab allocation, task creation, and scheduling.
- Build script updated to disable garbage collection.
- Integration of runtime initialization into the boot sequence.
- Documentation of the deterministic memory and scheduling strategy.

---

### 10. What Is Pending

- **Phase 10:** Testing and benchmarks – measure interrupt latency, context‑switch time, and overall system performance to verify the RTOS meets the target metrics.

---

### 11. Next Step

Proceed to **Phase 10 – Testing + Benchmarks**. We will implement:
- Unit tests for critical kernel modules.
- Integration tests for scheduling, IPC, and drivers.
- Performance benchmarks for interrupt latency and context‑switch time.
- Stress tests to validate system stability under heavy load.

---

Proceed to Phase 10? If so, say "Continue to Phase 10".
