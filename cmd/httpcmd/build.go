package httpcmd

import (
	"fmt"
	"os"
	"slices"

	"github.com/daboyuka/hs/program/scope/bindings"
	"github.com/daboyuka/hs/program/stream"
	"github.com/spf13/cobra"

	cmdctx "github.com/daboyuka/hs/cmd/context"
	"github.com/daboyuka/hs/hsruntime"
	hscommand "github.com/daboyuka/hs/hsruntime/command"
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

	hctx, err := cmdctx.Init(hsruntime.Options{}, true)
	if err != nil {
		return err
	}

	for _, spec := range buildFlagVals.loadSpecs {
		hctx.Globals, err = loadJSONTable(spec, hctx.Globals, hctx.Funcs)
		if err != nil {
			return err
		}
	}

	scp, binds := hctx.Globals.Scope, hctx.Globals.Binds

	input, err := openInput(os.Stdin, commonFlagVals.infmt)
	if err != nil {
		return err
	}

	hcmd, scp, err := hscommand.NewHttpBuildCommand(method, urlSrc, bodySrc, buildFlagVals.headers, scp, hctx)
	if err != nil {
		return err
	}

	sink := record.StringWriterSink(os.Stdout)

	// Build pipeline
	s := bindings.BindStream(input, binds)
	s = stream.Apply(s, hcmd.Operate)
	return stream.Run(s, bindings.BindSink(sink))
}
