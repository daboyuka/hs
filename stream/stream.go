package stream

import (
	"io"
	"sync"
	"sync/atomic"
)

// Stream is an iterator over a stream of values, which may produce errors during iteration.
type Stream[V any] interface {
	// Next returns the next Record in the stream, or io.EOF if end-of-stream, or other error if one occurred.
	// Next never returns both a Record and a non-nil error (as nil is indistinguishable from a real null Record).
	// Once Next returns a non-nil error, every subsequent call returns the same error.
	// Next is *not* safe for concurrent use by multiple goroutines.
	Next() (V, error)
}

//

// Singleton wraps a value to make a Stream with that value as its single value.
func Singleton[V any](v V) Stream[V] { return &singletonStream[V]{Val: v} }

// Slice wraps a slice to make a Stream over those values.
func Slice[V any](s []V) Stream[V] { return (*slice[V])(&s) }

// Error wraps an error to make a Stream that always returns that error.
func Error[V any](err error) Stream[V] { return &errStream[V]{err: err} }

// Sync wraps a Stream to make it safe for concurrent access by multiple goroutines.
func Sync[V any](s Stream[V]) Stream[V] { return &syncStream[V]{s: s} }

type slice[V any] []V

func (ss *slice[V]) Next() (out V, err error) {
	recs := *ss
	if len(recs) == 0 {
		var zero V
		return zero, io.EOF
	}
	out, *ss = recs[0], recs[1:]
	return out, nil
}

type singletonStream[V any] struct {
	Val  V
	once atomic.Uintptr
}

func (t *singletonStream[V]) Next() (V, error) {
	if t.once.Swap(1) != 0 {
		var zero V
		return zero, io.EOF
	}
	return t.Val, nil
}

type errStream[V any] struct{ err error }

func (e *errStream[V]) Next() (V, error) {
	var zero V
	return zero, e.err
}

type syncStream[V any] struct {
	mtx sync.Mutex
	s   Stream[V]
}

func (ss *syncStream[V]) Next() (V, error) {
	ss.mtx.Lock()
	defer ss.mtx.Unlock()
	return ss.Next()
}

type CountingStream[V any] struct {
	Stream[V]
	count atomic.Int64
}

func (c *CountingStream[V]) Next() (V, error) {
	r, err := c.Stream.Next()
	if err != nil {
		c.count.Store(-c.count.Load() - 1) // -n-1 the count to indicate completion
		return r, err
	}
	c.count.Add(1)
	return r, nil
}
func (c *CountingStream[V]) Count() (n int, done bool) {
	if n = int(c.count.Load()); n < 0 {
		n, done = -(n + 1), true
	}
	return n, done
}

type ValWithErr[V any] struct {
	Val V
	Err error
}

type ChannelStream[V any] struct {
	Ch chan ValWithErr[V]
}

func (c ChannelStream[V]) Next() (out V, err error) {
	r, ok := <-c.Ch
	if !ok {
		return out, io.EOF
	} else {
		return r.Val, r.Err
	}
}
