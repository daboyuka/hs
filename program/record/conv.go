package record

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

// CoerceString formats r as a string as follows:
//   - string is returned as-is
//   - all other types are marshaled and returned as raw JSON
func CoerceString(r Record) string {
	if s, ok := r.(string); ok {
		return s
	} else if j, err := json.Marshal(r); err != nil {
		panic(err)
	} else {
		return string(j)
	}
}

var ErrNotANumber = errors.New("not a number")
var ErrNotAnInt = errors.New("not an integer")

func NumberToInt(r Record) (int, error) {
	if f, ok := r.(float64); !ok {
		return 0, ErrNotANumber
	} else if i := int(f); float64(i) != f {
		return 0, ErrNotAnInt
	} else {
		return i, nil
	}
}

func StringEscape(s string) string {
	// TODO: this is not quite right, due to Go having more escapes (e.g. high Unicode characters, some control characters)
	s = strconv.Quote(s)
	return s[1 : len(s)-1]
}

// AsStringsArray converts val to an array of strings:
//   - If nil, returns empty array
//   - If array with all elements as strings, returns as string array
//   - If string and allowString == true, return as a single-element array
//   - Otherwise, return an error
func AsStringsArray(val Record, allowString bool) (strs []string, err error) {
	switch valTyped := val.(type) {
	case nil:
		return nil, nil
	case string:
		if allowString {
			return []string{valTyped}, nil
		}
	case Array:
		for _, elem := range valTyped {
			s, ok := elem.(string)
			if !ok {
				return nil, fmt.Errorf("expected only strings in array, got %T element", elem)
			}
			strs = append(strs, s)
		}
		return strs, nil
	}
	return nil, fmt.Errorf("expected string or array-of-strings, got %T", val)
}
