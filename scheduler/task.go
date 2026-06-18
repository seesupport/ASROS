package scheduler

import (
	"rtos/mm"
	"rtos/lib"
	"unsafe"
)

// TaskState represents the state of a task.
type TaskState uint8

const (
	StateReady TaskState = iota
	StateRunning
	StateWaiting
	StateTerminated
)

// TaskControlBlock holds the state of a task.
type TaskControlBlock struct {
	// Registers saved during context switch
	R15 uint64
	R14 uint64
	R13 uint64
	R12 uint64
	RBP uint64
	RBX uint64
	// RSP and RIP must be saved last (or we can save them separately)
	RSP uint64
	RIP uint64

	// Task metadata
	PID         uint32
	Priority    int
	BasePriority int // original priority before boosting
	State       TaskState
	Stack       unsafe.Pointer // allocated stack memory
	StackSize   uint32

	// Scheduling fields
	TimeSlice   uint32 // remaining ticks in current time slice
	CPUAffinity uint64 // bitmask of allowed CPUs (future SMP)
	Next        *TaskControlBlock // for linked lists
	Prev        *TaskControlBlock
}

var nextPID uint32 = 1

// NewTask creates a new task with the given entry function and priority.
// Returns a TCB or nil on failure.
func NewTask(entry func(), priority int) *TaskControlBlock {
	// Validate priority
	if priority < 0 || priority >= PRIORITY_LEVELS {
		return nil
	}

	// Allocate stack (8KB)
	stackSize := uint32(8192)
	stackPtr := mm.Alloc(uint64(stackSize))
	if stackPtr == nil {
		lib.PrintString("Failed to allocate stack for task\n")
		return nil
	}

	// Allocate TCB from heap
	tcbPtr := mm.Alloc(uint64(unsafe.Sizeof(TaskControlBlock{})))
	if tcbPtr == nil {
		mm.Free(stackPtr, uint64(stackSize))
		lib.PrintString("Failed to allocate TCB\n")
		return nil
	}
	tcb := (*TaskControlBlock)(tcbPtr)

	// Initialize TCB
	tcb.PID = nextPID
	nextPID++
	tcb.Priority = priority
	tcb.BasePriority = priority
	tcb.State = StateReady
	tcb.Stack = stackPtr
	tcb.StackSize = stackSize
	tcb.TimeSlice = DEFAULT_TIME_SLICE
	tcb.CPUAffinity = ^uint64(0) // all CPUs

	// Set up initial stack frame.
	// We need to simulate a call to `entry` with a return address that calls `taskExit`.
	// We'll push a return address (taskExit) onto the stack, then arrange RSP to point to the top.
	// We'll also set RIP to entry.
	// Stack grows down, so we set RSP to stackPtr + stackSize.
	stackTop := uintptr(stackPtr) + uintptr(stackSize)

	// Push a return address (taskExit) onto the stack.
	// We'll use a small assembly stub `taskExit` that will be called when the task returns.
	// We'll define `taskExit` in assembly that simply calls `scheduler.TaskExit()`.
	// We'll put the address of taskExit at the top of the stack.
	// Then we set RIP to entry.
	// When the task returns (ret), it will pop the return address and jump to taskExit.
	// This works because we set up the stack like a normal call.
	// We'll store the return address at stackTop-8.
	*(*uint64)(unsafe.Pointer(stackTop - 8)) = uint64(uintptr(unsafe.Pointer(&taskExitStub)))
	// Now set RSP to point to stackTop - 8 (so that ret will pop that address).
	tcb.RSP = uint64(stackTop - 8)
	// Set RIP to entry point.
	tcb.RIP = uint64(uintptr(unsafe.Pointer(entry)))

	// Save the current task's registers (we'll initialize others to zero).
	// They will be loaded on first switch.

	return tcb
}

// taskExitStub is a small assembly function that calls TaskExit.
// It is placed at the bottom of the stack.
//go:noescape
func taskExitStub()

// TaskExit is called when a task returns from its entry function.
// It marks the task as terminated and yields the CPU.
//export TaskExit
func TaskExit() {
	// Get current task
	current := getCurrentTask()
	if current != nil {
		current.State = StateTerminated
		// We could free resources here, but we'll keep it simple.
	}
	// Yield to scheduler (never returns)
	Yield()
}

// getCurrentTask returns the currently running task (set by scheduler).
// This is a global variable managed by the scheduler.
var currentTask *TaskControlBlock

func getCurrentTask() *TaskControlBlock {
	return currentTask
}
