package cookie

import (
	"net/url"
	"reflect"
	"testing"
)

func TestHostAliasJarAdapter_doAlias(t *testing.T) {
	haja := HostAliasJarAdapter{
		"fromA": {Host: "toA"},
		"fromB": {Scheme: "https", Host: "toB"},
	}

	tests := map[*url.URL]*url.URL{
		{Scheme: "http", Host: "fromA", Path: "/path"}:  {Scheme: "http", Host: "toA", Path: "/path"},
		{Scheme: "https", Host: "fromA", Path: "/path"}: {Scheme: "https", Host: "toA", Path: "/path"},
		{Scheme: "http", Host: "fromB", Path: "/path"}:  {Scheme: "https", Host: "toB", Path: "/path"},
		{Scheme: "https", Host: "fromB", Path: "/path"}: {Scheme: "https", Host: "toB", Path: "/path"},
		{Scheme: "http", Host: "other", Path: "/path"}:  nil,
	}
	for in, expect := range tests {
		out := haja.doAlias(in)
		if !reflect.DeepEqual(out, expect) {
			t.Fatalf("mapping mismatch: '%v' -> '%v', expected '%v'", in, out, expect)
		}
	}
}
