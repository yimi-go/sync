package lockfree

import (
	stdSync "sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yimi-go/sync"
)

func TestQueue(t *testing.T) {
	fq := NewQueue[float64]()
	assert.Equal(t, uint64(0), fq.count)
	_, err := fq.Dequeue()
	assert.True(t, sync.IsErrQueueEmpty(err))
	_ = fq.Enqueue(1.0)
	assert.Equal(t, uint64(1), fq.Len())
	_ = fq.Enqueue(1.0)
	assert.Equal(t, uint64(2), fq.Len())
	v, err := fq.Dequeue()
	assert.Nil(t, err)
	assert.Equal(t, 1.0, v)
	assert.Equal(t, uint64(1), fq.Len())
	v, err = fq.Dequeue()
	assert.Nil(t, err)
	assert.Equal(t, 1.0, v)
	assert.Equal(t, uint64(0), fq.Len())

	wg := &stdSync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			for i := 0; i < 100; i++ {
				_ = fq.Enqueue(1.0)
			}
			wg.Done()
		}()
	}
	wg.Wait()
	assert.Equal(t, uint64(10000), fq.Len())
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			for i := 0; i < 99; i++ {
				_ = fq.Enqueue(1.0)
				_, _ = fq.Dequeue()
				var v float64
				v, err := fq.Dequeue()
				assert.Nil(t, err)
				assert.Equal(t, 1.0, v)
			}
			wg.Done()
		}()
	}
	wg.Wait()
	assert.Equal(t, uint64(100), fq.Len())
	for i := 0; i < 100; i++ {
		_, _ = fq.Dequeue()
	}
	_, err = fq.Dequeue()
	assert.True(t, sync.IsErrQueueEmpty(err))

	cq := NewQueue[int](Cap[int](1))
	err = cq.Enqueue(1)
	assert.Nil(t, err)
	err = cq.Enqueue(1)
	assert.True(t, sync.IsErrQueueFull(err))
}

func TestQueueIterate(t *testing.T) {
	q := NewQueue[int]()
	for i := 0; i < 100; i++ {
		_ = q.Enqueue(i)
	}
	it := q.Iterate()
	for i := 0; i < 100; i++ {
		v, ok := it.Next()
		assert.True(t, ok)
		assert.Equal(t, i, v)
	}
	_, ok := it.Next()
	assert.False(t, ok)
}

func doBench(q *Queue[float64]) {
	wg := &stdSync.WaitGroup{}
	// 100并发分别100次写
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			for i := 0; i < 100; i++ {
				if i%4 == 3 {
					_, _ = q.Dequeue()
				} else {
					_ = q.Enqueue(1.0)
				}
			}
			wg.Done()
		}()
	}
	// 同时100并发100次读，读写数相当。因为用于窗口计数桶时，是并发读写相当的场景
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			for i := 0; i < 100; i++ {
				q.Len()
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func BenchmarkQueue(b *testing.B) {
	q := NewQueue[float64]()
	for i := 0; i < b.N; i++ {
		doBench(q)
	}
}
