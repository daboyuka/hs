package stream

// A Sink consumes items. It can function as the tail end of a Stream (i.e. the yield function).
// It is safe for concurrent use by multiple goroutines.
type Sink[T any] func(T) error

// DiscardSink is a Sink that discards all values without error.
func DiscardSink[T any](T) error { return nil }

// CollectSink is a Sink that collects all values into itself as a slice.
type CollectSink[T any] []T

func (c *CollectSink[T]) Sink(r T) error { *c = append(*c, r); return nil }
