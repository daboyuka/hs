package record

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"io"
)

func NewRawStream(r io.Reader) Stream {
	return func(yield func(Record) error) error {
		if b, err := io.ReadAll(r); err != nil {
			return err
		} else {
			return yield(string(b))
		}
	}
}

func NewLineStream(r io.Reader) Stream {
	return func(yield func(Record) error) error {
		br := bufio.NewReaderSize(r, 1<<10)
		var buf bytes.Buffer
		for {
			if data, more, err := br.ReadLine(); err == io.EOF {
				return nil
			} else if err != nil {
				return err
			} else if more { // if there's more, keep what we have
				buf.Write(data)
			} else if len(data) == 0 { // ignore empty lines
				continue
			} else if buf.Len() == 0 { // if there's nothing in the buffer, write the data directly
				if err := yield(string(data)); err != nil {
					return err
				}
			} else { // otherwise, append and dump the buffer
				buf.Write(data)
				if err := yield(string(buf.Bytes())); err != nil {
					return err
				}
				buf.Reset()
			}
		}
	}
}

func NewJSONStream(r io.Reader) Stream {
	return func(yield func(Record) error) error {
		d := json.NewDecoder(r)
		for {
			var out Record
			if err := d.Decode(&out); err == io.EOF {
				return nil
			} else if err != nil {
				return err
			} else if err := yield(out); err != nil {
				return err
			}
		}
	}
}

// NewCsvReader creates a Stream by parsing the input io.Reader as comma-separated value format.
// comma is the separator (should be ',' for true CSV).
// If raw, no CSV header is expected, and each line becomes a simple Array from field values; otherwise, the first line
// is interpreted as a header defining field names, and each line becomes an Object using those field names.
func NewCsvReader(r io.Reader, comma rune, raw bool) Stream {
	return func(yield func(Record) error) error {
		cr := csv.NewReader(r)
		cr.Comma, cr.ReuseRecord = comma, true

		var fields []string
		if !raw {
			if hdr, err := cr.Read(); err == io.EOF {
				return nil
			} else if err != nil {
				return err
			} else {
				fields = append([]string{}, hdr...) // copy due to ReuseRecord
			}
		}

		for {
			if vals, err := cr.Read(); err == io.EOF {
				return nil
			} else if err != nil {
				return err
			} else if raw {
				out := make(Array, len(vals))
				for i, v := range vals {
					out[i] = v
				}
				if err := yield(out); err != nil {
					return err
				}
			} else {
				out := make(Object, len(vals))
				for i, v := range vals {
					out[fields[i]] = v
				}
				if err := yield(out); err != nil {
					return err
				}
			}
		}
	}
}
