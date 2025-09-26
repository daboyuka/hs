package hostalias

import (
	"fmt"
	"net/url"
)

// HostAlias applies host aliasing rules, returning a new hostname (or "" if no aliasing is applied).
type HostAlias func(hostname string) (newHostname string)

// Apply transforms u in-place using host aliasing rules. Error is returned if an unknown/invalid alias is used.
func (f HostAlias) Apply(u *url.URL) error {
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

// None is the no-op HostAlias.
var None HostAlias = func(hostname string) string { return "" }

func Compose(base, next HostAlias) HostAlias {
	return func(hostname string) string {
		if hostname2 := base(hostname); hostname2 != "" {
			return hostname2
		}
		return next(hostname)
	}
}

func Simple(mapping map[string]string) HostAlias {
	return func(hostname string) string {
		return mapping[hostname] // if missing, returns "" -> no alias
	}
}
