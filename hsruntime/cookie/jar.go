package cookie

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"

	"golang.org/x/net/publicsuffix"
)

//
//type ReadOnlyJar interface{ Cookies(*url.URL) []*http.Cookie }
//
//func ReadOnly(jar http.CookieJar) http.CookieJar {
//	return struct {
//		ReadOnlyJar
//		noopSetCookies
//	}{ReadOnlyJar: jar}
//}
//
//type noopSetCookies struct{}
//
//func (noopSetCookies) SetCookies(*url.URL, []*http.Cookie) {}

type JarAdapterFunc func(u *url.URL, next http.CookieJar) []*http.Cookie

// Adapt wraps a CookieJar with an adaptor, which implements Cookies on top of base (SetCookies is provided by base)
func Adapt(base http.CookieJar, adapter JarAdapterFunc) http.CookieJar {
	return jarAdapter{CookieJar: base, adapter: adapter}
}

// Concat combines two CookieJar together, with Cookies as their concatenation, and SetCookies delegated to base.
func Concat(base http.CookieJar, append http.CookieJar) http.CookieJar {
	return jarConcat{CookieJar: base, append: append}
}

func SimpleJar(cookies []*http.Cookie) (http.CookieJar, error) {
	matched, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return nil, err
	}
	jar := simpleJar{matched: matched}

	for i, cookie := range cookies {
		if cookie.Domain == "" && cookie.Path == "" {
			jar.always = append(jar.always, cookie)
			continue
		}

		u := &url.URL{Scheme: "http", Host: cookie.Domain, Path: cookie.Path}
		if cookie.Secure {
			u.Scheme = "https"
		}
		jar.matched.SetCookies(u, cookies[i:i+1])
	}

	jar.always = jar.always[:len(jar.always):len(jar.always)] // fit capacity, so that appending in simpleJar.Cookies reallocates rather than clobbering

	return jar, nil
}

// simpleJar is as a *cookiejar.Jar, except also supporting "global"/"always" cookies, to be returned regardless of URL.
type simpleJar struct {
	matched *cookiejar.Jar // discriminate by domain, path, expiry, etc.
	always  []*http.Cookie // returned unconditionally
}

func (c simpleJar) SetCookies(u *url.URL, cookies []*http.Cookie) { c.matched.SetCookies(u, cookies) }
func (c simpleJar) Cookies(u *url.URL) []*http.Cookie {
	// Note: func SimpleJar ensures c.always has no extra capacity, so that appending (more than zero cookies)
	// reallocates and does not trash the slice.
	return append(c.always, c.matched.Cookies(u)...)
}

type jarAdapter struct {
	http.CookieJar // SetCookies passthrough
	adapter        JarAdapterFunc
}

func (j jarAdapter) Cookies(u *url.URL) []*http.Cookie {
	return j.adapter(u, j.CookieJar)
}

type jarConcat struct {
	http.CookieJar // SetCookies passthrough
	append         http.CookieJar
}

func (j jarConcat) Cookies(u *url.URL) []*http.Cookie {
	return append(j.CookieJar.Cookies(u), j.append.Cookies(u)...)
}

type HostAliasJarAdapter map[string]*url.URL

func (haja HostAliasJarAdapter) doAlias(u *url.URL) *url.URL {
	h2, ok := haja[u.Hostname()]
	if !ok {
		return nil
	}
	u2 := *u
	u2.Host = h2.Host
	if h2.Scheme != "" {
		u2.Scheme = h2.Scheme
	}
	return &u2
}

func (haja HostAliasJarAdapter) Adapt(u *url.URL, next http.CookieJar) []*http.Cookie {
	cookies := next.Cookies(u)
	if u2 := haja.doAlias(u); u2 != nil {
		cookies = append(cookies, next.Cookies(u2)...)
	}

	return cookies
}
