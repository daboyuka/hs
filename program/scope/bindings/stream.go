package bindings

import (
	"github.com/daboyuka/hs/program/record"
	"github.com/daboyuka/hs/program/stream"
)

type BoundRecord struct {
	Binds *Bindings
	Rec   record.Record
}

type BoundStream = stream.Stream[BoundRecord]
type BoundSink = stream.Sink[BoundRecord]

func BindStream(stream record.Stream, binds *Bindings) BoundStream {
	return func(yield func(BoundRecord) error) error {
		yield2 := func(r record.Record) error { return yield(BoundRecord{Binds: binds, Rec: r}) }
		return stream(yield2)
	}
}

func BindSink(sink record.Sink) BoundSink {
	return func(rec BoundRecord) error { return sink(rec.Rec) }
}
