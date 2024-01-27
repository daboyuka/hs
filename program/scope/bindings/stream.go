package bindings

import (
	"github.com/daboyuka/hs/program/record"
	"github.com/daboyuka/hs/stream"
)

type BoundStream = stream.Stream[BoundRecord]

func BindStream(stream record.Stream, binds *Bindings) BoundStream {
	return &boundStream{stream: stream, binds: binds}
}

type boundStream struct {
	stream record.Stream
	binds  *Bindings
}

func (w *boundStream) Next() (BoundRecord, error) {
	rec, err := w.stream.Next()
	return BoundRecord{Record: rec, Binds: w.binds}, err
}
