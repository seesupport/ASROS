package sync

import (
	"rtos/scheduler"
)

type Semaphore struct {
	count   int
	waiters scheduler.List
}

// NewSemaphore creates a counting semaphore with initial count.
func NewSemaphore(initial int) *Semaphore {
	return &Semaphore{count: initial}
}

// Wait decrements the semaphore; blocks if count <= 0.
func (s *Semaphore) Wait() {
	current := scheduler.GetCurrentTask()
	if current == nil {
		return
	}
	if s.count > 0 {
		s.count--
		return
	}
	// Block
	scheduler.BlockTask(current, &s.waiters)
}

// Signal increments the semaphore and unblocks a waiter if any.
func (s *Semaphore) Signal() {
	if s.waiters.IsEmpty() {
		s.count++
	} else {
		// Unblock one waiter.
		task := s.waiters.PopHead()
		if task != nil {
			scheduler.UnblockTask(task)
		}
	}
}

// BinarySemaphore is a convenience for semaphore with initial 1.
type BinarySemaphore struct {
	Semaphore
}

func NewBinarySemaphore() *BinarySemaphore {
	return &BinarySemaphore{Semaphore: *NewSemaphore(1)}
}
