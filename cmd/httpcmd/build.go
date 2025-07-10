package httpcmd

import (
	"context"
	"fmt"
	"os"
	"slices"

	"github.com/spf13/cobra"

	"github.com/daboyuka/hs/hsruntime"
	hscommand "github.com/daboyuka/hs/hsruntime/command"
	"github.com/daboyuka/hs/hsruntime/config"
	"github.com/daboyuka/hs/hsruntime/plugin"
	"github.com/daboyuka/hs/program/command"
	"github.com/daboyuka/hs/program/record"
)

func cmdBuild(cmd *cobra.Command, args []string) (finalErr error) {
	method, urlSrc := args[0], args[1]

	if !slices.Contains(allMethods, method) {
		return fmt.Errorf("bad HTTP method '%s'", method)
	}

	var bodySrc string
	if len(args) >= 3 {
		bodySrc = args[2]
	}

	ctx := context.Background()
	hctx, err := hsruntime.NewDefaultContext(hsruntime.Options{})
	if err != nil {
		return err
	} else if err := plugin.Apply(hctx); err != nil {
		return err
	} else if err := config.WarnMissingBaseConfiguration(); err != nil {
		os.Stderr.WriteString("Warning: " + err.Error() + "\n")
	}

	for _, spec := range buildFlagVals.loadSpecs {
		hctx.Globals, err = loadJSONTable(spec, hctx.Globals, hctx.Funcs)
		if err != nil {
			return err
		}
	}

	scp, binds := hctx.Globals.Scope, hctx.Globals.Binds

	hcmd, scp, err := hscommand.NewHttpBuildCommand(method, urlSrc, bodySrc, buildFlagVals.headers, scp, hctx)
	if err != nil {
		return err
	}

	input, err := openInput(os.Stdin, commonFlagVals.infmt)
	if err != nil {
		return err
	}
	sink := &record.StringWriterSink{Writer: os.Stdout}

	return command.RunParallel(ctx, hcmd, binds, input, sink, 1, nil)
}
