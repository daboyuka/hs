package cookie

import (
	"net/http"
	"os"
	"strings"

	cookiemonster "github.com/MercuryEngineering/CookieMonster"
)

func loadFile(filename string) ([]*http.Cookie, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return cookiemonster.Parse(f)
}

// LoadSpec resolves "cookie specs" to cookies, based on format:
//   - <name>=<value>: return as literal (no domain/path)
//   - <filename>    : loads filename from disk (Netscape format)
func LoadSpec(spec string) (cookies []*http.Cookie, err error) {
	if idx := strings.IndexRune(spec, '='); idx != -1 {
		return []*http.Cookie{{
			Name:  spec[:idx],
			Value: spec[idx+1:],
		}}, nil
	}

	return loadFile(spec)
}

func LoadSpecs(specs ...string) (cookies []*http.Cookie, err error) {
	for _, spec := range specs {
		c, err := LoadSpec(spec)
		if err != nil {
			return nil, err
		}
		cookies = append(cookies, c...)
	}
	return cookies, nil
}
