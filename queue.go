package sync

import (
	"errors"
	"fmt"
)

var errQueueFull = fmt.Errorf("sync: queue is full")
var errQueueEmpty = fmt.Errorf("sync: queue is empty")

// ErrQueueFull returns an error that indicates the Queue is full.
func ErrQueueFull() error {
	return errQueueFull
}

// IsErrQueueFull checks whether an error is indicating the Queue is full.
// This function supports wrapped errors.
func IsErrQueueFull(err error) bool {
	return errors.Is(err, errQueueFull)
}

// ErrQueueEmpty returns an error that indicates the Queue is empty.
func ErrQueueEmpty() error {
	return errQueueEmpty
}

// IsErrQueueEmpty checks whether an error is indicating the Queue is empty.
// This function supports wrapped errors.
func IsErrQueueEmpty(err error) bool {
	return errors.Is(err, errQueueEmpty)
}

// Queue is a data structure that collect and offer elements in a FIFO (first-in-first-out) manner.
type Queue[E any] interface {
	// Len returns the elements count in the Queue.
	Len() uint64
	// Enqueue receives an element and put it at tail of the Queue.
	// If the Queue is full, an error is returned.
	Enqueue(v E) error
	// Dequeue removes the first element of the Queue and returns the element.
	// If the Queue is empty, an error is returned.
	Dequeue() (v E, err error)
}
