package cmd

import (
	"os"

	"github.com/spf13/cobra"

	cmdctx "github.com/daboyuka/hs/cmd/context"
	"github.com/daboyuka/hs/hsruntime"
	"github.com/daboyuka/hs/hsruntime/config"
)

func cmdInit(_ *cobra.Command, _ []string) error {
	if hctx, err := cmdctx.Init(hsruntime.Options{}, false); err != nil {
		return err
	} else if cfg, err := hctx.ConfigInit(""); err != nil {
		return err
	} else if cfgOutput, err := config.CreateConfigurationFile(cfg); err != nil {
		return err
	} else {
		_, _ = os.Stderr.Write([]byte(cfgOutput))
		return nil
	}
}
