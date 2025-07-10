package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/daboyuka/hs/hsruntime"
	"github.com/daboyuka/hs/hsruntime/config"
	"github.com/daboyuka/hs/hsruntime/plugin"
)

func cmdInit(_ *cobra.Command, _ []string) error {
	hctx, err := hsruntime.NewDefaultContext(hsruntime.Options{})
	if err != nil {
		return err
	} else if err := plugin.Apply(hctx); err != nil {
		return err
	}
	cfg, err := hctx.ConfigInit("")
	if err != nil {
		return err
	}
	if cfgOutput, err := config.CreateConfigurationFile(cfg); err != nil {
		return err
	} else {
		os.Stderr.Write([]byte(cfgOutput))
	}
	return nil
}
