package stream

import (
	"sync/atomic"
)

// Stream is a push stream of items with error propagation. When called, it should repeatedly call yield until all items
// are sent, or until yield returns an error, which should be returned upward.
// yield is safe for concurrent use by multiple goroutines.
// Stream's signature mismatch with iter.Seq is intentional to prevent misuse (iter.Seq does not allow concurrent yield).
// Whether Stream is single use (as per definition in iter.Seq) is implementation defined.
// If a type has a method that acts as a Stream, it is conventionally named Stream.
type Stream[T any] func(yield func(T) error) error

// Collect buffers a Stream into a slice.
func Collect[T any](in Stream[T]) (out []T, err error) {
	c := (*CollectSink[T])(&out)
	return out, in(c.Sink)
}

func Run[T any](in Stream[T], out Sink[T]) error { return in(out) }

func Slice[T any](vals []T) Stream[T] {
	return Stream[T](func(yield func(T) error) error {
		for _, v := range vals {
			if err := yield(v); err != nil {
				return err
			}
		}
		return nil
	})
}

func Singleton[T any](v T) Stream[T] {
	return Stream[T](func(yield func(T) error) error { return yield(v) })
}

func Counting[T any](in Stream[T]) (counter *Counter, counted Stream[T]) {
	var c Counter
	return &c, func(yield func(T) error) error {
		err := in(func(v T) error {
			if err := yield(v); err != nil {
				return err
			}
			c.count.Add(1)
			return nil
		})
		c.count.Store(-c.count.Load() - 1) // -n-1 the count to indicate completion
		return err
	}
}

type Counter struct {
	count atomic.Int64
}

func (c *Counter) Count() (n int, done bool) {
	if n = int(c.count.Load()); n < 0 {
		n, done = -(n + 1), true
	}
	return n, done
}
