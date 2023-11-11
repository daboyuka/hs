package datafmt

import (
	"testing"
)

func TestAutodetect(t *testing.T) {
	tests := []struct {
		Name   string
		Input  string
		Expect Format
	}{
		{Name: "JSON object", Input: `  {"key":"val"}`, Expect: JSON},
		{Name: "JSON object, truncated", Input: `  {"key":"va`, Expect: JSON},
		{Name: "JSON array", Input: `  ["val1", "val2"]`, Expect: JSON},
		{Name: "JSON array, truncated", Input: `  ["val1", "va`, Expect: JSON},
		{Name: "JSON string", Input: `  "value"`, Expect: JSON},
		{Name: "JSON number", Input: `  123.456`, Expect: JSON},

		{Name: "formdata, single", Input: `key=value`, Expect: FormData},
		{Name: "formdata, multi", Input: `key1=value1&key2=value2`, Expect: FormData},
		{Name: "formdata, multi, no vals", Input: `key1&key2`, Expect: FormData},

		{Name: "simple text", Input: `foobar`, Expect: Unknown},
		{Name: "JSON string, truncated (unknown)", Input: `  "val`, Expect: Unknown},
	}

	for _, tst := range tests {
		if got := Autodetect(tst.Input); got != tst.Expect {
			t.Errorf("test '%s' failed: got %s -> %s, expect %s", tst.Name, tst.Input, got, tst.Expect)
		}
	}
}
