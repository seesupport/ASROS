package sync

import (
	"rtos/cpu"
	"sync/atomic"
)

type Spinlock struct {
	locked uint32
}

// Lock spins until it acquires the lock.
func (s *Spinlock) Lock() {
	for !atomic.CompareAndSwapUint32(&s.locked, 0, 1) {
		cpu.Pause() // PAUSE instruction to reduce CPU usage
	}
}

// Unlock releases the lock.
func (s *Spinlock) Unlock() {
	atomic.StoreUint32(&s.locked, 0)
}
