package sync

import (
	"rtos/scheduler"
	"unsafe"
)

// MessageQueue holds fixed-size messages.
type MessageQueue struct {
	buffer   []byte
	msgSize  uint32
	capacity uint32
	head     uint32
	tail     uint32
	count    uint32
	sendWait scheduler.List
	recvWait scheduler.List
	mu       Spinlock // protect internal state
}

// NewMessageQueue creates a queue with capacity messages of given size.
func NewMessageQueue(capacity uint32, msgSize uint32) *MessageQueue {
	return &MessageQueue{
		buffer:   make([]byte, capacity*msgSize),
		msgSize:  msgSize,
		capacity: capacity,
	}
}

// Send blocks until a message can be sent.
func (mq *MessageQueue) Send(msg unsafe.Pointer) {
	mq.mu.Lock()
	if mq.count < mq.capacity {
		// Copy message into buffer
		offset := mq.tail * mq.msgSize
		// Manual copy
		for i := uint32(0); i < mq.msgSize; i++ {
			mq.buffer[offset+i] = *(*byte)(unsafe.Pointer(uintptr(msg) + uintptr(i)))
		}
		mq.tail = (mq.tail + 1) % mq.capacity
		mq.count++
		// If there are receivers waiting, unblock one.
		if !mq.recvWait.IsEmpty() {
			task := mq.recvWait.PopHead()
			if task != nil {
				scheduler.UnblockTask(task)
			}
		}
		mq.mu.Unlock()
		return
	}
	// Need to block sending task.
	current := scheduler.GetCurrentTask()
	if current == nil {
		mq.mu.Unlock()
		return
	}
	// Add current to send wait list.
	mq.sendWait.AddTail(current)
	// Unlock and block (BlockTask will disable interrupts, but we are already locked? We need to release lock before blocking.)
	mq.mu.Unlock()
	scheduler.BlockTask(current, &mq.sendWait)
	// After being unblocked, we need to try sending again.
	// We'll loop back.
}

// Receive blocks until a message is available.
func (mq *MessageQueue) Receive(msg unsafe.Pointer) {
	mq.mu.Lock()
	if mq.count > 0 {
		offset := mq.head * mq.msgSize
		// Manual copy
		for i := uint32(0); i < mq.msgSize; i++ {
			*(*byte)(unsafe.Pointer(uintptr(msg) + uintptr(i))) = mq.buffer[offset+i]
		}
		mq.head = (mq.head + 1) % mq.capacity
		mq.count--
		// If there are senders waiting, unblock one.
		if !mq.sendWait.IsEmpty() {
			task := mq.sendWait.PopHead()
			if task != nil {
				scheduler.UnblockTask(task)
			}
		}
		mq.mu.Unlock()
		return
	}
	// Need to block receiving task.
	current := scheduler.GetCurrentTask()
	if current == nil {
		mq.mu.Unlock()
		return
	}
	mq.recvWait.AddTail(current)
	mq.mu.Unlock()
	scheduler.BlockTask(current, &mq.recvWait)
	// After being unblocked, try again.
}
