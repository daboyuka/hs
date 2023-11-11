package record

import (
	"io"
	"sync"
)

// A Sink consumes Records.
type Sink interface {
	// Sink accepts a Record, with implementation-dependent effect. It is safe for concurrent use by multiple goroutines.
	Sink(rec Record) error
}

type DiscardSink struct{}

func (DiscardSink) Sink(Record) error { return nil }

type StringWriterSink struct {
	mtx    sync.Mutex
	Writer io.StringWriter
}

func (s *StringWriterSink) Sink(r Record) error {
	str := CoerceString(r)
	s.mtx.Lock()
	defer s.mtx.Unlock()
	_, err := s.Writer.WriteString(str + "\n")
	return err
}
