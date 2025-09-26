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
	"github.com/daboyuka/hs/hsruntime/hostalias"
	"github.com/daboyuka/hs/program/scope"
)

type Context struct {
	Globals scope.ScopedBindings
	Funcs   *scope.FuncTable

	ConfigInit   ConfigInitFn
	DefaultHost  string
	HostAliasing hostalias.HostAlias
	Client       *http.Client
}

// ConfigInitFn is a function that augments the default configuration
type ConfigInitFn func(cfg string) (string, error)

func NewContext() *Context {
	return &Context{
		ConfigInit:   config.DefaultConfiguration,
		HostAliasing: hostalias.None,
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
	if ctx.HostAliasing, err = defaultHostAliasing(hostalias.None, ctx.Globals); err != nil {
		return nil, err
	}

	ctx.Client.Jar = &DeferredLoadCookieJar{ // don't actually load cookies until needed
		LoadFn: func() (http.CookieJar, error) { return cookie.Load(opts.CookieSpecs, ctx.Globals) },
	}
	if ctx.Client.Jar, err = defaultCookieHostAliasing(ctx.Client.Jar, ctx); err != nil {
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

func defaultCookieHostAliasing(base http.CookieJar, ctx *Context) (http.CookieJar, error) {
	aliasesIntf, _ := ctx.Globals.Lookup("COOKIE_HOST_ALIASES")

	switch aliases := aliasesIntf.(type) {
	default:
		return nil, fmt.Errorf("expected map for COOKIE_HOST_ALIASES, got %T", aliases)
	case nil:
		return base, nil
	case map[string]interface{}:
		aliasMap := make(map[string]*url.URL, len(aliases))
		for k, v := range aliases {
			if targetStr, ok := v.(string); !ok {
				return nil, fmt.Errorf("expected string values for COOKIE_HOST_ALIASES mappings, got %T", v)
			} else if targetUrl, err := parseTargetCookieHostAlias(targetStr); err != nil {
				return nil, fmt.Errorf("bad target hostname %s for cookie host alias: %w", targetStr, err)
			} else {
				aliasMap[k] = targetUrl
			}
		}
		return cookie.Adapt(base, cookie.HostAliasJarAdapter(aliasMap, &ctx.HostAliasing)), nil
	}
}

func parseTargetCookieHostAlias(src string) (*url.URL, error) {
	// Allow src to be parsed as a bare hostname if // is missing
	if !strings.Contains(src, "//") {
		src = "//" + src
	}

	hostUrl, err := url.Parse(src)
	switch { // host is required, scheme is optional, and user may only be missing or blank; no other URL structure is permitted
	case err != nil:
		return nil, err
	case hostUrl.User != nil && hostUrl.User.String() != "":
		return nil, fmt.Errorf("must not contain username/password")
	case hostUrl.RawPath != "":
		return nil, fmt.Errorf("must not contain path")
	case hostUrl.OmitHost:
		return nil, fmt.Errorf("must contain hostname")
	case hostUrl.RawQuery != "" || hostUrl.ForceQuery:
		return nil, fmt.Errorf("must not contain query parameters")
	case hostUrl.RawFragment != "":
		return nil, fmt.Errorf("must not contain fragment")
	}
	return hostUrl, nil
}

func defaultHostAliasing(base hostalias.HostAlias, globals scope.ScopedBindings) (hostalias.HostAlias, error) {
	aliasesIntf, _ := globals.Lookup("HOST_ALIASES")

	switch aliases := aliasesIntf.(type) {
	default:
		return nil, fmt.Errorf("expected map for HOST_ALIASES, got %T", aliases)
	case nil:
		return base, nil
	case map[string]interface{}:
		aliasesStr := make(map[string]string, len(aliases))
		for k, v := range aliases {
			vStr, ok := v.(string)
			if !ok {
				return nil, fmt.Errorf("expected string values for HOST_ALIASES mappings, got %T", v)
			}
			aliasesStr[k] = vStr
		}
		return hostalias.Compose(base, hostalias.Simple(aliasesStr)), nil
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
