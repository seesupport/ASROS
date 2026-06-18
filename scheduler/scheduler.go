package scheduler

import (
	"rtos/cpu"
	"rtos/interrupts"
	"rtos/lib"
	"unsafe"
)

const DEFAULT_TIME_SLICE = 10 // ticks

var idleTask *TaskControlBlock

// switchContext is implemented in assembly.
// It saves the current task's registers and loads the next task's registers.
//go:noescape
func switchContext(current, next *TaskControlBlock)

// Init initialises the scheduler and creates the idle task.
func Init() {
	// Initialise run queues
	for i := 0; i < PRIORITY_LEVELS; i++ {
		runQueues[i] = RunQueue{}
	}
	priorityBitmap = 0

	// Create idle task (priority = PRIORITY_LEVELS-1, lowest)
	idle := NewTask(idleTaskFunc, PRIORITY_LEVELS-1)
	if idle == nil {
		lib.PrintString("FATAL: Failed to create idle task\n")
		for {
		}
	}
	idle.State = StateReady
	enqueue(idle)
	idleTask = idle
	currentTask = idle

	// Register the timer tick handler with the interrupt system
	interrupts.RegisterTimerTick(Tick)

	lib.PrintString("Scheduler initialised\n")
}

// idleTaskFunc is the idle task: just HLT.
func idleTaskFunc() {
	for {
		cpu.Halt()
	}
}

// Start begins the scheduler. It enables interrupts and switches to the first task.
func Start() {
	// Ensure we have an idle task
	if idleTask == nil {
		idleTask = NewTask(idleTaskFunc, PRIORITY_LEVELS-1)
		idleTask.State = StateReady
		enqueue(idleTask)
	}
	// Set current task to the first ready task (or idle)
	next := dequeue()
	if next == nil {
		next = idleTask
		enqueue(idleTask) // re-enqueue
	}
	currentTask = next
	// Enable interrupts
	cpu.EnableInterrupts()
	// Switch to the first task
	switchContext(nil, next) // nil means no current task, just load new
	// Should never return
}

// schedule selects the highest-priority ready task and returns it.
func schedule() *TaskControlBlock {
	// Disable interrupts while manipulating queues
	cpu.DisableInterrupts()
	defer cpu.EnableInterrupts()

	// If current task is still running, re-enqueue it.
	if currentTask != nil && currentTask.State == StateRunning {
		currentTask.State = StateReady
		enqueue(currentTask)
	} else if currentTask != nil && currentTask.State == StateTerminated {
		// Free resources? We'll keep it simple.
	}

	// Pick next task
	next := dequeue()
	if next == nil {
		// If no ready task, use idle task (which should always be ready).
		if idleTask == nil {
			idleTask = NewTask(idleTaskFunc, PRIORITY_LEVELS-1)
			idleTask.State = StateReady
			enqueue(idleTask)
		}
		next = dequeue()
		if next == nil {
			// Still nil? Something wrong.
			lib.PrintString("FATAL: No tasks available!\n")
			for {
			}
		}
	}
	next.State = StateRunning
	return next
}

// Tick is called from the timer interrupt (every time slice).
func Tick() {
	// Disable interrupts (they are already disabled in ISR)
	// Check if current task's time slice expired.
	if currentTask != nil && currentTask.State == StateRunning {
		currentTask.TimeSlice--
		if currentTask.TimeSlice == 0 {
			// Preempt current task
			currentTask.TimeSlice = DEFAULT_TIME_SLICE
			// Re-schedule
			next := schedule()
			if next != currentTask {
				old := currentTask
				currentTask = next
				switchContext(old, next)
				// After switch, we return here only if the task yields or is preempted again.
				// The switchContext function does not return; it jumps to the next task.
			}
		}
	}
}

// Yield is called by a task to voluntarily give up the CPU.
func Yield() {
	cpu.DisableInterrupts()
	// Re-enqueue current task if it's still running
	if currentTask != nil && currentTask.State == StateRunning {
		currentTask.State = StateReady
		currentTask.TimeSlice = DEFAULT_TIME_SLICE // reset slice
		enqueue(currentTask)
	}
	next := schedule()
	if next != currentTask {
		old := currentTask
		currentTask = next
		switchContext(old, next)
	}
	// If no switch, re-enable interrupts
	cpu.EnableInterrupts()
}

// CreateTask creates a new user task and adds it to the ready queue.
func CreateTask(entry func(), priority int) *TaskControlBlock {
	task := NewTask(entry, priority)
	if task == nil {
		return nil
	}
	task.State = StateReady
	enqueue(task)
	return task
}
