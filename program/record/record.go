package record

import (
	"github.com/daboyuka/hs/stream"
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
	Stream = stream.Stream[Record]
)
