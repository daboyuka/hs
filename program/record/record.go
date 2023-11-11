package record

import (
	"io"
	"sync"
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

type Stream interface {
	// Next returns the next Record in the stream, or io.EOF if end-of-stream, or other error if one occurred.
	// Next never returns both a Record and a non-nil error (as nil is indistinguishable from a real null Record).
	// Once Next returns io.EOF, every subsequent call will also return io.EOF.
	// Next is safe for concurrent use by multiple goroutines, although no guarantee of Record ordering is given under
	// concurrent access.
	Next() (Record, error)
}

// CollectStream buffers a Stream into an Array of records.
func CollectStream(s Stream) (arr Array, err error) {
	for {
		switch r, err := s.Next(); err {
		case nil:
			arr = append(arr, r)
		case io.EOF:
			return arr, nil
		default:
			return arr, err
		}
	}
}

type SliceStream struct {
	Records []Record
	mtx     sync.Mutex
}

func (a *SliceStream) Next() (out Record, err error) {
	a.mtx.Lock()
	defer a.mtx.Unlock()
	if len(a.Records) == 0 {
		return nil, io.EOF
	}
	out, a.Records = a.Records[0], a.Records[1:]
	return out, nil
}

type SingletonStream struct {
	Rec  Record
	once atomic.Uintptr
}

func (t *SingletonStream) Next() (Record, error) {
	if t.once.Swap(1) != 0 {
		return nil, io.EOF
	}
	return t.Rec, nil
}
