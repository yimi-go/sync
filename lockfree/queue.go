package lockfree

import (
	"math"
	stdSync "sync"
	"sync/atomic"
	"unsafe"

	"github.com/yimi-go/sync"
)

// QueueOption is option function to creating new Queue.
type QueueOption[E any] func(q *Queue[E])

// Cap returns a QueueOption that set capacity of the Queue.
// The default capacity of a new Queue is infinity.
func Cap[E any](cap uint64) QueueOption[E] {
	return func(q *Queue[E]) {
		q.cap = cap
	}
}

// Queue is a lock-free Queue.
type Queue[E any] struct {
	cap   uint64
	count uint64
	dummy *node
	_pad  [8]uint64 // avoid false sharing
	pool  stdSync.Pool
}

// NewQueue creates a new Queue by options.
// By default, it creates an infinity capacity Queue.
func NewQueue[E any](opts ...QueueOption[E]) *Queue[E] {
	dummy := node{}
	dummy.prv, dummy.nxt = unsafe.Pointer(&dummy), unsafe.Pointer(&dummy)
	q := &Queue[E]{
		cap:   math.MaxUint64, // unlimited by default.
		dummy: &dummy,
		pool:  stdSync.Pool{New: func() any { return &node{} }},
	}
	for _, opt := range opts {
		opt(q)
	}
	return q
}

func (q *Queue[E]) Len() uint64 {
	return atomic.LoadUint64(&q.count)
}

func (q *Queue[E]) Enqueue(v E) error {
	n := q.pool.Get().(*node)
	atomic.StorePointer(&n.val, unsafe.Pointer(&v))
	atomic.StorePointer(&n.nxt, unsafe.Pointer(q.dummy))
	for {
		// Load last first, then check capacity.
		last := q.dummy.prev()
		atomic.StorePointer(&n.prv, unsafe.Pointer(last))
		if q.Len() == q.cap {
			return sync.ErrQueueFull()
		}
		// If last is changed after capacity check, the if check fails.
		if last.casNext(q.dummy, n) {
			atomic.StorePointer(&q.dummy.prv, unsafe.Pointer(n))
			atomic.AddUint64(&q.count, 1)
			return nil
		}
	}
}

func (q *Queue[E]) Dequeue() (v E, err error) {
	for {
		first := q.dummy.next()
		if first == q.dummy {
			err = sync.ErrQueueEmpty()
			return
		}
		afterFirst := first.next()
		if q.dummy.casNext(first, afterFirst) {
			atomic.StorePointer(&afterFirst.prv, unsafe.Pointer(q.dummy))
			atomic.AddUint64(&q.count, ^uint64(0))
			v = *(*E)(first.value())
			atomic.StorePointer(&first.val, nil)
			atomic.StorePointer(&first.prv, nil)
			atomic.StorePointer(&first.nxt, nil)
			q.pool.Put(first)
			return v, nil
		}
	}
}

// Iterate creates an Iterator that iterates over the Queue, without dequeue any element.
func (q *Queue[E]) Iterate() sync.Iterator[E] {
	return &queueIter[E]{
		dummy: q.dummy,
		next:  q.dummy.next(),
	}
}

type queueIter[E any] struct {
	dummy *node
	next  *node
}

func (it *queueIter[E]) Next() (v E, ok bool) {
	if it.next == it.dummy {
		return
	}
	v = *(*E)(it.next.value())
	it.next = it.next.next()
	return v, true
}
