package record

import (
	"io"
	"sync"
)

func StringWriterSink(sw io.StringWriter) Sink {
	return (&stringWriterSink{Writer: sw}).Sink
}

type stringWriterSink struct {
	mtx    sync.Mutex
	Writer io.StringWriter
}

func (s *stringWriterSink) Sink(r Record) error {
	str := CoerceString(r)
	s.mtx.Lock()
	defer s.mtx.Unlock()
	_, err := s.Writer.WriteString(str + "\n")
	return err
}
