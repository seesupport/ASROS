package scheduler

const PRIORITY_LEVELS = 32 // we can support 0..31, 0 highest

// List is a doubly linked list of tasks.
type List struct {
	head *TaskControlBlock
	tail *TaskControlBlock
}

// AddTail adds a task to the end of the list.
func (l *List) AddTail(task *TaskControlBlock) {
	if task == nil {
		return
	}
	task.Next = nil
	task.Prev = l.tail
	if l.tail != nil {
		l.tail.Next = task
	} else {
		l.head = task
	}
	l.tail = task
}

// AddHead adds a task to the front of the list.
func (l *List) AddHead(task *TaskControlBlock) {
	if task == nil {
		return
	}
	task.Prev = nil
	task.Next = l.head
	if l.head != nil {
		l.head.Prev = task
	} else {
		l.tail = task
	}
	l.head = task
}

// Remove removes a task from the list.
func (l *List) Remove(task *TaskControlBlock) {
	if task == nil {
		return
	}
	if task.Prev != nil {
		task.Prev.Next = task.Next
	} else {
		l.head = task.Next
	}
	if task.Next != nil {
		task.Next.Prev = task.Prev
	} else {
		l.tail = task.Prev
	}
	task.Next = nil
	task.Prev = nil
}

// PopHead removes and returns the first task.
func (l *List) PopHead() *TaskControlBlock {
	if l.head == nil {
		return nil
	}
	task := l.head
	l.Remove(task)
	return task
}

// IsEmpty returns true if the list has no tasks.
func (l *List) IsEmpty() bool {
	return l.head == nil
}

// PeekHead returns the first task without removing it.
func (l *List) PeekHead() *TaskControlBlock {
	return l.head
}

// RunQueues (adapted to use List)
var runQueues [PRIORITY_LEVELS]List
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
	runQueues[prio].AddTail(task)
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
	task := q.PopHead()
	if q.IsEmpty() {
		priorityBitmap &^= (1 << prio)
	}
	return task
}

// removeTask removes a specific task from its queue.
func removeTask(task *TaskControlBlock) {
	if task == nil {
		return
	}
	prio := task.Priority
	if prio < 0 || prio >= PRIORITY_LEVELS {
		return
	}
	runQueues[prio].Remove(task)
	if runQueues[prio].IsEmpty() {
		priorityBitmap &^= (1 << uint(prio))
	}
}
