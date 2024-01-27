package cookie

import (
	"fmt"
	"net/http"
	"os"

	cookiemonster "github.com/MercuryEngineering/CookieMonster"
	"github.com/daboyuka/hs/program/scope/bindings"

	"github.com/daboyuka/hs/hsruntime/searchpath"
	"github.com/daboyuka/hs/program/record"
)

const (
	cookieFilename   = ".hscookie"
	cookiesConfigVar = "COOKIES"
)

func Load(extraSpecs []string, globals bindings.Scoped) (http.CookieJar, error) {
	if extraCookies, err := LoadSpecs(extraSpecs...); err != nil {
		return nil, err
	} else if dotFileCookies, err := loadFromDotFiles(); err != nil {
		return nil, err
	} else if cfgCookies, err := loadFromCfg(globals); err != nil {
		return nil, err
	} else if browserCookies, err := loadBrowserCookies(globals); err != nil {
		return nil, err
	} else {
		cookies := append([]*http.Cookie(nil), extraCookies...)
		cookies = append(cookies, dotFileCookies...)
		cookies = append(cookies, cfgCookies...)
		cookies = append(cookies, browserCookies...)
		return SimpleJar(cookies)
	}
}

func loadFromDotFiles() (cookies []*http.Cookie, err error) {
	err = searchpath.Visit(cookieFilename, func(f *os.File) error {
		c, cErr := cookiemonster.Parse(f)
		cookies = append(cookies, c...)
		return cErr
	})
	return cookies, err
}

func loadFromCfg(globals bindings.Scoped) (cookies []*http.Cookie, err error) {
	specsCfgVal, _ := globals.Lookup(cookiesConfigVar)
	specs, err := record.AsStringsArray(specsCfgVal, true)
	if err != nil {
		return nil, fmt.Errorf("bad %s config value: %w", cookiesConfigVar, err)
	}
	return LoadSpecs(specs...)
}
