package bindings

import (
	"github.com/daboyuka/hs/program/record"
)

type BoundStream interface {
	// Next is as record.Stream.Next, except it adds a Bindings alongside each record (may be shared by many records).
	Next() (record.Record, *Bindings, error)
}

type withBinds struct {
	stream record.Stream
	binds  *Bindings
}

func NewBoundStream(stream record.Stream, binds *Bindings) BoundStream {
	return &withBinds{stream: stream, binds: binds}
}

func (w *withBinds) Next() (record.Record, *Bindings, error) {
	rec, err := w.stream.Next()
	return rec, w.binds, err
}

type errStream struct{ err error }

func ErrorStream(err error) BoundStream { return &errStream{err: err} }

func (e *errStream) Next() (record.Record, *Bindings, error) { return nil, nil, e.err }
