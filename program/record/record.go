package record

import (
	"iter"
	"sync/atomic"
)

// Record is a single data item. It is always matches one of these types:
//
//	untyped nil, bool, float64, string, []any, map[string]any
//
// json.Marshal/json.Unmarshal may be used on Record at will.
type Record = any

// Convenience aliases (for readability)
type (
	Array  = []any
	Object = map[string]any
)

type OldStream interface{ Next() (Record, error) }

type Stream = iter.Seq2[Record, error]

// CollectStream buffers a Stream of Record into an Array.
func CollectStream(s Stream) (arr Array, err error) {
	for r, err := range s {
		if err != nil {
			return nil, err
		}
		arr = append(arr, r)
	}
	return arr, nil
}

func SingletonStream(rec Record) Stream {
	return func(yield func(Record, error) bool) { yield(rec, nil) }
}

func CountStream(s Stream) (Stream, *Counter) {
	cnt := &Counter{stream: s}
	return cnt.seq, cnt
}

type Counter struct {
	stream Stream
	cnt    atomic.Int64
}

func (c *Counter) seq(yield func(Record, error) bool) {
	for r, err := range c.stream {
		c.cnt.Add(1)
		if !yield(r, err) {
			break
		}
	}
	c.cnt.Store(-c.cnt.Load() - 1) // -n-1 the count to indicate completion
}

func (c *Counter) Count() (n int, done bool) {
	if n = int(c.cnt.Load()); n < 0 {
		n, done = -(n + 1), true
	}
	return n, done
}

type RecordAndError struct {
	Record
	Err error
}

type ChannelStream struct {
	Ch chan RecordAndError
}

func (c ChannelStream) All() Stream { return c.iter }
func (c ChannelStream) iter(yield func(Record, error) bool) {
	for re := range c.Ch {
		if !yield(re.Record, re.Err) {
			break
		}
	}
}
