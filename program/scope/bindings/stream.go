package bindings

import (
	"iter"

	"github.com/daboyuka/hs/program/record"
)

type BoundRecord struct {
	Record record.Record
	Binds  *Bindings
}

type BoundStream = iter.Seq2[BoundRecord, error]

func NewBoundStream(stream record.Stream, binds *Bindings) BoundStream {
	return func(yield func(BoundRecord, error) bool) {
		for rec, err := range stream {
			if !yield(BoundRecord{Record: rec, Binds: binds}, err) {
				break
			}
		}
	}
}

func ErrorStream(err error) BoundStream {
	return func(yield func(BoundRecord, error) bool) { yield(BoundRecord{}, err) }
}
