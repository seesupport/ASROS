package sync

import (
	"rtos/scheduler"
)

type EventFlags struct {
	flags   uint32
	waiters scheduler.List
	mu      Spinlock
}

// Set sets one or more bits.
func (e *EventFlags) Set(bits uint32) {
	e.mu.Lock()
	e.flags |= bits
	// Unblock all waiters whose conditions are met.
	// We'll unblock all and let them recheck.
	for !e.waiters.IsEmpty() {
		task := e.waiters.PopHead()
		if task != nil {
			scheduler.UnblockTask(task)
		}
	}
	e.mu.Unlock()
}

// Clear clears bits.
func (e *EventFlags) Clear(bits uint32) {
	e.mu.Lock()
	e.flags &^= bits
	e.mu.Unlock()
}

// WaitAny blocks until any of the specified bits are set.
func (e *EventFlags) WaitAny(bits uint32) {
	e.mu.Lock()
	if (e.flags & bits) != 0 {
		e.mu.Unlock()
		return
	}
	current := scheduler.GetCurrentTask()
	if current == nil {
		e.mu.Unlock()
		return
	}
	e.waiters.AddTail(current)
	e.mu.Unlock()
	scheduler.BlockTask(current, &e.waiters)
	// After unblock, check again; if condition still not met, wait again (loop).
}

// WaitAll blocks until all specified bits are set.
func (e *EventFlags) WaitAll(bits uint32) {
	for {
		e.mu.Lock()
		if (e.flags & bits) == bits {
			e.mu.Unlock()
			return
		}
		current := scheduler.GetCurrentTask()
		if current == nil {
			e.mu.Unlock()
			return
		}
		e.waiters.AddTail(current)
		e.mu.Unlock()
		scheduler.BlockTask(current, &e.waiters)
	}
}
