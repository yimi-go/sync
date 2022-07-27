package mu

import (
	"math"
	stdSync "sync"

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

// Queue is a sync Queue protected by an RWMutex.
type Queue[E any] struct {
	mu    stdSync.RWMutex
	cap   uint64
	count uint64
	dummy *node[E]
	pool  stdSync.Pool
}

// NewQueue creates a new Queue by options.
// By default, it creates an infinity capacity Queue.
func NewQueue[E any](opts ...QueueOption[E]) *Queue[E] {
	dummy := node[E]{}
	dummy.prev, dummy.next = &dummy, &dummy
	q := &Queue[E]{
		cap:   math.MaxUint64, // unlimited by default.
		dummy: &dummy,
		pool:  stdSync.Pool{New: func() any { return &node[E]{} }},
	}
	for _, opt := range opts {
		opt(q)
	}
	return q
}

func (q *Queue[E]) Len() uint64 {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.count
}

func (q *Queue[E]) Enqueue(e E) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.count == q.cap {
		return sync.ErrQueueFull()
	}
	n := q.pool.Get().(*node[E])
	n.val = &e
	last := q.dummy.prev
	n.next, n.prev, last.next, q.dummy.prev = q.dummy, last, n, n
	q.count++
	return nil
}

func (q *Queue[E]) Dequeue() (v E, err error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.count == 0 {
		err = sync.ErrQueueEmpty()
		return
	}
	first := q.dummy.next
	second := first.next
	q.dummy.next, second.prev = second, q.dummy
	q.count--
	v = *first.val
	return
}

// Iterate creates an Iterator that iterates over the Queue, without dequeue any element.
func (q *Queue[E]) Iterate() sync.Iterator[E] {
	return &queueIter[E]{
		dummy: q.dummy,
		next:  q.dummy.next,
	}
}

type queueIter[E any] struct {
	dummy *node[E]
	next  *node[E]
}

func (it *queueIter[E]) Next() (v E, ok bool) {
	if it.next == it.dummy {
		return
	}
	v = *it.next.val
	it.next = it.next.next
	return v, true
}
