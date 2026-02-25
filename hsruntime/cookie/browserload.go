package cookie

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"slices"
	"strings"

	"github.com/daboyuka/kooky"
	_ "github.com/daboyuka/kooky/browser/all" // enable all browsers

	"github.com/daboyuka/hs/program/record"
	"github.com/daboyuka/hs/program/scope"
)

const (
	browserLoadersConfigVar        = "BROWSER_LOADERS"
	browserLoaderPrefixesConfigVar = "BROWSER_LOADER_PREFIXES"
)

func AllSupportedBrowsers() (browsers []string) {
	browsersSet := make(map[string]bool)
	for _, store := range kooky.FindAllCookieStores(context.Background()) {
		browsersSet[store.Browser()] = true
	}
	for browser := range browsersSet {
		browsers = append(browsers, browser)
	}
	sort.Strings(browsers)
	return browsers
}

// loadBrowserCookies loads cookies from browser cookie stores. Only browsers listed in config var browserLoadersConfigVar
// (string or array of strings) are loaded. If config browserLoaderPrefixesConfigVar is set, only cookies with one of
// those name prefixes are loaded (if unset, no prefix filtering is done).
func loadBrowserCookies(globals scope.ScopedBindings) (cookies []*http.Cookie, err error) {
	browsersSet, allBrowsers, prefixes, err := getBrowserLoaderConfig(globals)
	if err != nil {
		return nil, err
	} else if !allBrowsers && len(browsersSet) == 0 {
		return nil, nil // if no browsers given, don't load browser cookies at all
	}

	for _, store := range kooky.FindAllCookieStores(context.Background()) {
		if !allBrowsers && !browsersSet[store.Browser()] {
			continue
		}
		for kcookie, err := range store.TraverseCookies() {
			if err != nil || kcookie == nil {
				continue
			} else if c := kcookie.Cookie; !cookieMatchesPrefixes(&c, prefixes) {
				continue
			} else {
				cookies = append(cookies, &c)
			}
		}
	}

	return cookies, nil
}

func cookieMatchesPrefixes(c *http.Cookie, prefixes []string) bool {
	return len(prefixes) == 0 || slices.ContainsFunc(prefixes, func(prefix string) bool { return strings.HasPrefix(c.Name, prefix) })
}

func getBrowserLoaderConfig(globals scope.ScopedBindings) (browsersSet map[string]bool, allBrowsers bool, prefixes []string, err error) {
	browsersCfgVal, _ := globals.Lookup(browserLoadersConfigVar)
	browsers, err := record.AsStringsArray(browsersCfgVal, true)
	if err != nil {
		return nil, false, nil, fmt.Errorf("bad %s config value: %w", browserLoadersConfigVar, err)
	}

	prefixesCfgVal, _ := globals.Lookup(browserLoaderPrefixesConfigVar)
	prefixes, err = record.AsStringsArray(prefixesCfgVal, true)
	if err != nil {
		return nil, false, nil, fmt.Errorf("bad %s config value: %w", browserLoaderPrefixesConfigVar, err)
	}

	browsersSet = make(map[string]bool, len(browsers))
	for _, browser := range browsers {
		browser = strings.ToLower(browser)
		if browser == "all" {
			browsersSet, allBrowsers = nil, true
			break
		}
		browsersSet[browser] = true
	}

	return browsersSet, allBrowsers, prefixes, nil
}
