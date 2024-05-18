package common

import "sync"

type GenericChan[T any] chan T

type SafeChannel[T any] struct {
	C      GenericChan[T]
	closed bool
	mutex  sync.Mutex
}

func NewSafeChannel[T any]() *SafeChannel[T] {
	return &SafeChannel[T]{C: make(chan T)}
}

func (mc *SafeChannel[T]) SafeClose() {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	if !mc.closed {
		close(mc.C)
		mc.closed = true
	}
}

func (mc *SafeChannel[T]) IsClosed() bool {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	return mc.closed
}
