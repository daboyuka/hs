package record

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"io"
)

func NewRawStream(r io.Reader) Stream {
	return func(yield func(Record, error) bool) {
		if b, err := io.ReadAll(r); err != nil {
			yield(nil, err)
		} else {
			yield(string(b), nil)
		}
	}
}

func NewLineStream(r io.Reader) Stream {
	br := bufio.NewReaderSize(r, 1<<10)
	return func(yield func(Record, error) bool) {
		var buf bytes.Buffer
		for {
			data, more, err := br.ReadLine()
			if err != nil {
				if err != io.EOF {
					yield(nil, err)
				}
				return
			} else if len(data) == 0 && !more {
				continue // ignore empty lines
			}

			if buf.Len() == 0 {
				if more {
					buf = *bytes.NewBuffer(data)
				} else if len(data) > 0 { // only return non-empty lines
					if !yield(string(data), nil) {
						return
					}
				}
			} else {
				_, _ = buf.Write(data)
				if !more {
					if !yield(string(buf.Bytes()), nil) {
						return
					}
					buf.Reset()
				}
			}
		}
	}
}

func NewJSONStream(r io.Reader) Stream {
	d := json.NewDecoder(r)
	return func(yield func(Record, error) bool) {
		var out Record
		for {
			if err := d.Decode(&out); err == io.EOF || !yield(out, err) {
				return
			}
		}
	}
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
func NewCsvReader(r io.Reader, comma rune, raw bool) Stream {
	out := &CsvStream{r: *csv.NewReader(r), raw: raw}
	out.r.Comma = comma
	out.r.ReuseRecord = true // lets csv.Reader recycle Read's return slice
	return out.iter
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

	return c.r.Read()
}

func (c *CsvStream) iter(yield func(Record, error) bool) {
	for {
		vals, err := c.nextVals()
		if err == io.EOF {
			return
		} else if err != nil {
			yield(nil, err)
			return
		}

		var out Record
		if c.raw {
			outArr := make(Array, len(vals))
			for i, v := range vals {
				outArr[i] = v
			}
			out = outArr
		} else {
			outObj := make(Object, len(vals))
			for i, v := range vals {
				outObj[c.fields[i]] = v
			}
			out = outObj
		}

		if !yield(out, nil) {
			return
		}
	}
}
