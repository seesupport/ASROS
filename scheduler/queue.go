package scheduler

const PRIORITY_LEVELS = 32 // we can support 0..31, 0 highest

// RunQueue is a doubly linked list of tasks for a given priority.
type RunQueue struct {
	head *TaskControlBlock
	tail *TaskControlBlock
}

var runQueues [PRIORITY_LEVELS]RunQueue
var priorityBitmap uint32 // bit i set if runQueues[i] non-empty

// enqueue adds a task to the appropriate run queue.
func enqueue(task *TaskControlBlock) {
	if task == nil {
		return
	}
	prio := task.Priority
	if prio < 0 || prio >= PRIORITY_LEVELS {
		return
	}
	q := &runQueues[prio]
	task.Next = nil
	task.Prev = q.tail
	if q.tail != nil {
		q.tail.Next = task
	} else {
		q.head = task
	}
	q.tail = task
	// set bitmap
	priorityBitmap |= (1 << uint(prio))
}

// dequeue removes and returns the first task from the highest priority non-empty queue.
func dequeue() *TaskControlBlock {
	if priorityBitmap == 0 {
		return nil
	}
	// Find highest priority (lowest index) with bit set.
	prio := uint(0)
	// We can use clz or simple loop. We'll use a loop for simplicity.
	for prio = 0; prio < PRIORITY_LEVELS; prio++ {
		if priorityBitmap&(1<<prio) != 0 {
			break
		}
	}
	if prio >= PRIORITY_LEVELS {
		return nil
	}
	q := &runQueues[prio]
	task := q.head
	if task == nil {
		// Clear bitmap if queue empty
		priorityBitmap &^= (1 << prio)
		return nil
	}
	q.head = task.Next
	if q.head != nil {
		q.head.Prev = nil
	} else {
		q.tail = nil
	}
	task.Next = nil
	task.Prev = nil
	// If queue becomes empty, clear bitmap
	if q.head == nil {
		priorityBitmap &^= (1 << prio)
	}
	return task
}

// removeTask removes a specific task from its queue (used for yielding or waiting).
func removeTask(task *TaskControlBlock) {
	if task == nil {
		return
	}
	prio := task.Priority
	if prio < 0 || prio >= PRIORITY_LEVELS {
		return
	}
	q := &runQueues[prio]
	if task.Prev != nil {
		task.Prev.Next = task.Next
	} else {
		q.head = task.Next
	}
	if task.Next != nil {
		task.Next.Prev = task.Prev
	} else {
		q.tail = task.Prev
	}
	task.Next = nil
	task.Prev = nil
	if q.head == nil {
		priorityBitmap &^= (1 << uint(prio))
	}
}
