package lockfree

import (
	"sync/atomic"
	"unsafe"
)

type node struct {
	val      unsafe.Pointer
	prv, nxt unsafe.Pointer
}

func (n *node) value() unsafe.Pointer {
	return atomic.LoadPointer(&n.val)
}

func (n *node) prev() *node {
	return (*node)(atomic.LoadPointer(&n.prv))
}

func (n *node) next() *node {
	return (*node)(atomic.LoadPointer(&n.nxt))
}

func (n *node) casNext(expected, target *node) bool {
	return atomic.CompareAndSwapPointer(&n.nxt, unsafe.Pointer(expected), unsafe.Pointer(target))
}
