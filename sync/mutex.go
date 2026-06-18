package sync

import (
	"rtos/scheduler"
)

type Mutex struct {
	owner   *scheduler.TaskControlBlock
	waiters scheduler.List
	locked  bool
}

// Lock attempts to acquire the mutex. If locked, it blocks the current task.
func (m *Mutex) Lock() {
	current := scheduler.GetCurrentTask()
	if current == nil {
		return
	}
	// Fast path: if not locked, acquire.
	if !m.locked {
		m.locked = true
		m.owner = current
		return
	}
	// If already locked by this task, deadlock (we'll ignore for now).
	if m.owner == current {
		return // or panic
	}

	// Block: add current task to waiters list.
	scheduler.BlockTask(current, &m.waiters)

	// Priority inheritance: if the owner has lower priority than current, boost it.
	if m.owner != nil && m.owner.Priority > current.Priority {
		// Boost owner to current's priority.
		scheduler.SetTaskPriority(m.owner, current.Priority)
	}
	// After being unblocked, we'll own the mutex.
	// The unlocking task will set the owner and unblock us.
}

// Unlock releases the mutex and unblocks the highest priority waiter.
func (m *Mutex) Unlock() {
	current := scheduler.GetCurrentTask()
	if current == nil || m.owner != current {
		return // not owner
	}
	// If there are waiters, unblock the highest priority one.
	next := m.waiters.PopHead()
	if next != nil {
		// Pass ownership to next waiter.
		m.owner = next
		// Unblock the waiter.
		scheduler.UnblockTask(next)
		// Restore priority of original owner? We'll handle when task is scheduled again.
		// But we need to revert priority of the current task if it was boosted.
		// We'll keep a base priority and current priority in TCB.
		// We'll implement: when a task releases a mutex, it should lower its priority to the max of its base priority and the priority of any waiters it still holds.
		// For simplicity, we'll just set priority to base (we'll add basePriority field).
		// We'll add BasePriority to TCB.
		// When we boost, we set Priority to the higher value, but we store the original in BasePriority.
		// When we release, we check if we still own other mutexes with waiters; we'll simplify by setting to BasePriority.
		// We'll assume tasks don't hold multiple mutexes.
		// For now, we'll set to BasePriority.
		if current.BasePriority != current.Priority {
			scheduler.SetTaskPriority(current, current.BasePriority)
		}
	} else {
		// No waiters, release mutex.
		m.locked = false
		m.owner = nil
		// Revert priority to base if boosted.
		if current.BasePriority != current.Priority {
			scheduler.SetTaskPriority(current, current.BasePriority)
		}
	}
}
