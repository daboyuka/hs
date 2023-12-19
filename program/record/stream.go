package record

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"io"
	"sync"
)

type RawStream struct {
	r io.Reader
}

func NewRawStream(r io.Reader) *RawStream {
	return &RawStream{r: r}
}

func (r *RawStream) Next() (Record, error) {
	if b, err := io.ReadAll(r.r); err != nil {
		return nil, err
	} else {
		return string(b), nil
	}
}

type LineStream struct {
	r bufio.Reader
}

func NewLineStream(r io.Reader) *LineStream {
	return &LineStream{r: *bufio.NewReaderSize(r, 1<<10)}
}

func (l *LineStream) Next() (out Record, err error) {
	var buf bytes.Buffer
	for {
		data, more, err := l.r.ReadLine()
		if err != nil {
			return nil, err // handles io.EOF too
		} else if len(data) == 0 && !more {
			continue // ignore empty lines
		}

		if buf.Len() == 0 {
			if more {
				buf = *bytes.NewBuffer(data)
			} else if len(data) > 0 { // only return non-empty lines
				return string(data), nil
			}
		} else {
			_, _ = buf.Write(data)
			if !more {
				out = string(buf.Bytes())
				buf.Reset()
				return out, nil
			}
		}
	}
}

type JSONStream struct {
	d json.Decoder
}

func NewJSONStream(r io.Reader) *JSONStream {
	return &JSONStream{d: *json.NewDecoder(r)}
}

func (j *JSONStream) Next() (out Record, err error) {
	err = j.d.Decode(&out)
	return out, err
}

type CsvStream struct {
	r csv.Reader

	raw    bool
	fields []string
}

// NewCsvReader creates a Stream by parsing the input io.Reader as comma-separated value format.
// comma is the separator (should be ',' for true CSV).
// If raw, no CSV header is expected, and each line becomes a simple Array from field values; otherwise, the first line
// is interpreted as a header defining field names, and each line becomes an Object using those field names.
func NewCsvReader(r io.Reader, comma rune, raw bool) *CsvStream {
	out := &CsvStream{r: *csv.NewReader(r), raw: raw}
	out.r.Comma = comma
	out.r.ReuseRecord = true // lets csv.Reader recycle Read's return slice
	return out
}

func (c *CsvStream) loadHeader() error {
	hdr, err := c.r.Read()
	c.fields = append([]string{}, hdr...) // copy due to c.r.ReuseRecord
	return err
}

func (c *CsvStream) nextVals() ([]string, error) {
	if !c.raw && c.fields == nil {
		if err := c.loadHeader(); err != nil {
			return nil, err
		}
	}

	return c.r.Read() // c.concurrent -> !c.r.ReuseRecord -> fresh memory, making the return safe for concurrent in that case
}

func (c *CsvStream) Next() (Record, error) {
	vals, err := c.nextVals()
	if err != nil {
		return nil, err // handles io.EOF too
	}

	if c.raw {
		out := make(Array, len(vals))
		for i, v := range vals {
			out[i] = v
		}
		return out, nil
	} else {
		out := make(Object, len(vals))
		for i, v := range vals {
			out[c.fields[i]] = v
		}
		return out, nil
	}
}

// SyncStream wraps another Stream to make it safe for concurrent access by multiple goroutines.
type SyncStream struct {
	mtx sync.Mutex
	s   Stream
}

func NewSyncStream(s Stream) Stream { return &SyncStream{s: s} }

func (ss *SyncStream) Next() (Record, error) {
	ss.mtx.Lock()
	defer ss.mtx.Unlock()
	return ss.Next()
}
