package datafmt

import (
	"encoding/json"
	"io"
	"net/url"
	"strings"
)

//go:generate stringer -type=Format
type Format int

const (
	Unknown = Format(iota)
	JSON
	FormData
)

var contentType = [...]string{
	Unknown:  "",
	JSON:     "application/json",
	FormData: "application/x-www-form-urlencoded",
}

func (f Format) ContentType() string { return contentType[f] }

const maxAutodetectLen = 512

func Autodetect(data string) (format Format) {
	if len(data) > maxAutodetectLen {
		data = data[:maxAutodetectLen]
	}

	// Detect as JSON if first token is valid JSON (lone '{' or '[' after whitespace is enough, or a complete JSON
	// primitive value. Note that a broken string will _not_ detect as JSON.)
	if _, err := json.NewDecoder(strings.NewReader(data)).Token(); err == nil {
		return JSON
	}

	// Detect as form data if it parses as form data (truncation always leaves valid data), and either '&' is present
	// (there are multiple entries) or '=' is present (there is both key and value for one entry).
	// This avoids detecting "foobar" (which parses valid as a single, key-only entry) as FormData.
	// Syntax: https://url.spec.whatwg.org/#concept-urlencoded
	if vals, err := url.ParseQuery(data); err == nil {
		if len(vals) >= 2 {
			return FormData
		}
		for k, v := range vals {
			if k != "" && len(v) > 0 && v[0] != "" {
				return FormData
			}
		}
	}

	return Unknown
}

func AutodetectReader(r io.Reader) (format Format, r2 io.Reader, err error) {
	buf := strings.Builder{}

	// Read up to first 512 bytes for autodetection
	fullRead := false
	switch _, err := io.CopyN(&buf, r, maxAutodetectLen); err {
	case nil:
	case io.EOF:
		fullRead = true
	default:
		return 0, nil, err
	}

	s := buf.String()
	format, r2 = Autodetect(s), strings.NewReader(s)
	if !fullRead {
		r2 = io.MultiReader(r2, r) // if there's more bytes left, chain the remaining reader on
	}
	return format, r2, nil
}
