package context

import (
	"os"

	"github.com/daboyuka/hs/hsruntime"
	"github.com/daboyuka/hs/hsruntime/config"
	"github.com/daboyuka/hs/hsruntime/plugin"
)

// Init initializes the hsruntime.Context for an hs command in the standard way:
//	* Build default context
//	* Apply any plugins
//	* (if warnMissingCfg) Warn the user if config in home directory is missing
func Init(opts hsruntime.Options, warnMissingCfg bool) (*hsruntime.Context, error) {
	hctx, err := hsruntime.NewDefaultContext(opts)
	if err != nil {
		return nil, err
	} else if err := plugin.Apply(hctx); err != nil {
		return nil, err
	}

	if warnMissingCfg {
		if err := config.WarnMissingBaseConfiguration(); err != nil {
			_, _ = os.Stderr.WriteString("warning: " + err.Error() + "\n")
		}
	}
	return hctx, nil
}
