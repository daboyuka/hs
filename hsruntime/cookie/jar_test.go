package cookie

import (
	"net/url"
	"reflect"
	"testing"

	"github.com/daboyuka/hs/hsruntime/hostalias"
)

func TestHostAliasJarAdapter_doAlias(t *testing.T) {
	afterAlias := hostalias.HostAlias(func(host string) string {
		if host == "toC" {
			return "toCAlias"
		}
		return ""
	})
	haja := hostAliasJarAdapter{
		mapping: map[string]*url.URL{
			"fromA": {Host: "toA"},
			"fromB": {Scheme: "https", Host: "toB"},
			"fromC": {User: &url.Userinfo{}, Host: "toC"},
		},
		afterAlias: &afterAlias,
	}

	tests := map[*url.URL]*url.URL{
		{Scheme: "http", Host: "fromA", Path: "/path"}:  {Scheme: "http", Host: "toA", Path: "/path"},
		{Scheme: "https", Host: "fromA", Path: "/path"}: {Scheme: "https", Host: "toA", Path: "/path"},
		{Scheme: "http", Host: "fromB", Path: "/path"}:  {Scheme: "https", Host: "toB", Path: "/path"},
		{Scheme: "https", Host: "fromB", Path: "/path"}: {Scheme: "https", Host: "toB", Path: "/path"},
		{Scheme: "https", Host: "fromC", Path: "/path"}: {Scheme: "https", Host: "toCAlias", Path: "/path"},
		{Scheme: "http", Host: "other", Path: "/path"}:  nil,
	}
	for in, expect := range tests {
		out := haja.doAlias(in)
		if !reflect.DeepEqual(out, expect) {
			t.Fatalf("mapping mismatch: '%v' -> '%v', expected '%v'", in, out, expect)
		}
	}
}
