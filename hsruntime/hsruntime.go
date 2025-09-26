// Package hsruntime provides HTTPScript-specific program runtime context, such as config loading, cookie loading, etc.
package hsruntime

import (
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"

	"github.com/daboyuka/hs/hsruntime/config"
	"github.com/daboyuka/hs/hsruntime/cookie"
	"github.com/daboyuka/hs/program/scope"
)

type Context struct {
	Globals scope.ScopedBindings
	Funcs   *scope.FuncTable

	ConfigInit   ConfigInitFn
	DefaultHost  string
	HostAliasing HostAliasFn
	Client       *http.Client
}

// ConfigInitFn is a function that augments the default configuration
type ConfigInitFn func(cfg string) (string, error)

// HostAliasFn applies host aliasing rules, returning a new hostname (or "" if no aliasing is applied).
type HostAliasFn func(hostname string) (newHostname string)

func (f HostAliasFn) Apply(u *url.URL) error {
	if u.User == nil || u.User.String() != "" {
		return nil // not an host alias pattern (starting with bare @)
	}

	newHost := f(u.Host)
	if newHost == "" {
		return fmt.Errorf("unknown host alias @%s", u.Host)
	}

	u.User, u.Host = nil, newHost
	return nil
}

func noHostAliasing(hostname string) string { return "" }

type SimpleHostAliases map[string]string

func (sha SimpleHostAliases) GetAlias(hostname string) string {
	if h2, ok := sha[hostname]; ok {
		return h2
	}
	return ""
}

func ComposeHostAliasing(base, next HostAliasFn) HostAliasFn {
	return func(hostname string) string {
		if hostname2 := base(hostname); hostname2 != "" {
			return hostname2
		}
		return next(hostname)
	}
}

func NewContext() *Context {
	return &Context{
		ConfigInit:   config.DefaultConfiguration,
		HostAliasing: noHostAliasing,
		Client:       &http.Client{},
	}
}

type Options struct {
	CookieSpecs []string
}

// NewDefaultContext returns a default setup of Context, binding standard funcs, loading config, etc.
func NewDefaultContext(opts Options) (ctx *Context, err error) {
	ctx = NewContext()
	ctx.Globals.Scope, ctx.Globals.Binds, err = config.Load(nil, nil)
	if err != nil {
		return nil, err
	}

	if ctx.DefaultHost, err = getHost(ctx.Globals); err != nil {
		return nil, err
	}
	if ctx.HostAliasing, err = defaultHostAliasing(noHostAliasing, ctx.Globals); err != nil {
		return nil, err
	}

	ctx.Client.Jar = &DeferredLoadCookieJar{ // don't actually load cookies until needed
		LoadFn: func() (http.CookieJar, error) { return cookie.Load(opts.CookieSpecs, ctx.Globals) },
	}
	if ctx.Client.Jar, err = defaultCookieHostAliasing(ctx.Client.Jar, ctx.Globals); err != nil {
		return nil, err
	}

	return ctx, nil
}

// DeferredLoadCookieJar is an http.CookieJar that only constructs its underlying jar (using LoadFn) on use.
type DeferredLoadCookieJar struct {
	LoadFn func() (http.CookieJar, error)

	once sync.Once
	jar  http.CookieJar
}

func (d *DeferredLoadCookieJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	d.once.Do(d.load)
	d.jar.SetCookies(u, cookies)
}

func (d *DeferredLoadCookieJar) Cookies(u *url.URL) []*http.Cookie {
	d.once.Do(d.load)
	return d.jar.Cookies(u)
}

func (d *DeferredLoadCookieJar) load() {
	if j, err := d.LoadFn(); err != nil {
		log.Printf("error: failed to load cookiejar: %s", err)
		d.jar, _ = cookiejar.New(nil)
	} else {
		d.jar = j
	}
}

func defaultCookieHostAliasing(base http.CookieJar, globals scope.ScopedBindings) (http.CookieJar, error) {
	aliasesIntf, _ := globals.Lookup("COOKIE_HOST_ALIASES")

	switch aliases := aliasesIntf.(type) {
	default:
		return nil, fmt.Errorf("expected map for COOKIE_HOST_ALIASES, got %T", aliases)
	case nil:
		return base, nil
	case map[string]interface{}:
		aliasJar := make(cookie.HostAliasJarAdapter)
		for k, v := range aliases {
			host, ok := v.(string)
			if !ok {
				return nil, fmt.Errorf("expected string values for COOKIE_HOST_ALIASES mappings, got %T", v)
			}
			scheme := ""
			if s, h, ok := strings.Cut(host, "://"); ok {
				scheme, host = s, h
			}
			aliasJar[k] = &url.URL{Scheme: scheme, Host: host}
		}
		return cookie.Adapt(base, aliasJar.Adapt), nil
	}
}

func defaultHostAliasing(base HostAliasFn, globals scope.ScopedBindings) (HostAliasFn, error) {
	aliasesIntf, _ := globals.Lookup("HOST_ALIASES")

	switch aliases := aliasesIntf.(type) {
	default:
		return nil, fmt.Errorf("expected map for HOST_ALIASES, got %T", aliases)
	case nil:
		return base, nil
	case map[string]interface{}:
		simpleAliases := make(SimpleHostAliases, len(aliases))
		for k, v := range aliases {
			vStr, ok := v.(string)
			if !ok {
				return nil, fmt.Errorf("expected string values for HOST_ALIASES mappings, got %T", v)
			}
			simpleAliases[k] = vStr
		}
		return ComposeHostAliasing(base, simpleAliases.GetAlias), nil
	}
}

func getHost(globals scope.ScopedBindings) (string, error) {
	hostIntf, _ := globals.Lookup("HOST")
	switch host := hostIntf.(type) {
	default:
		return "", fmt.Errorf("expected string for HOST, got %T", hostIntf)
	case nil:
		return "", nil
	case string:
		return host, nil
	}
}
