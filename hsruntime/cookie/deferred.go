package cookie

import (
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"sync"
)

// DeferredJarLoader returns a http.CookieJar that lazily constructs its underlying jar (using loadFn) only on use.
func DeferredJarLoader(loadFn func() (http.CookieJar, error)) http.CookieJar {
	return &deferredLoader{loadFn: loadFn}
}

type deferredLoader struct {
	loadFn func() (http.CookieJar, error)
	once   sync.Once
	jar    http.CookieJar
}

func (d *deferredLoader) SetCookies(u *url.URL, cookies []*http.Cookie) {
	d.once.Do(d.load)
	d.jar.SetCookies(u, cookies)
}

func (d *deferredLoader) Cookies(u *url.URL) []*http.Cookie {
	d.once.Do(d.load)
	return d.jar.Cookies(u)
}

func (d *deferredLoader) load() {
	if j, err := d.loadFn(); err != nil {
		log.Printf("error: failed to load cookiejar: %s", err)
		d.jar, _ = cookiejar.New(nil)
	} else {
		d.jar = j
	}
}
