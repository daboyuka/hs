package plugin

import (
	"github.com/daboyuka/hs/hsruntime"
)

// Plugin is called to install a plugin on a given Context, updating it to add/wrap funcs, cookiejars, etc.
type Plugin func(ctx *hsruntime.Context) error

var registered []Plugin

// Register adds a new Plugin, to be applied by Apply. Plugins should call this from init.
func Register(p Plugin) { registered = append(registered, p) }

// Apply applies all Plugins (added by Register), in order of registration, to ctx.
func Apply(ctx *hsruntime.Context) error {
	for _, p := range registered {
		if err := p(ctx); err != nil {
			return err
		}
	}
	return nil
}
