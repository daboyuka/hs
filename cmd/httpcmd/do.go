package httpcmd

import (
	"context"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/daboyuka/hs/hsruntime"
	hscommand "github.com/daboyuka/hs/hsruntime/command"
	"github.com/daboyuka/hs/hsruntime/plugin"
	"github.com/daboyuka/hs/program/command"
)

func cmdDo(cmd *cobra.Command, args []string) (finalErr error) {
	method := cmd.Name()
	urlSrc := args[0]

	var bodySrc string
	if len(args) >= 2 {
		bodySrc = args[1]
	}

	ctx := context.Background()
	hctx, err := hsruntime.NewDefaultContext(hsruntime.Options{CookieSpecs: runFlagVals.cookies})
	if err != nil {
		return err
	} else if err := plugin.Apply(hctx); err != nil {
		return err
	}

	for _, spec := range buildFlagVals.loadSpecs {
		hctx.Globals, err = loadJSONTable(spec, hctx.Globals, hctx.Funcs)
		if err != nil {
			return err
		}
	}

	scp, binds := hctx.Globals.Scope, hctx.Globals.Binds

	var retry hscommand.RetryFunc
	if runFlagVals.retries > 0 {
		retry = func(req hscommand.RequestAndBody, resp hscommand.ResponseAndBody, attempt int) (backoff time.Duration, retry bool) {
			return time.Second, (resp.HTTPError != nil || resp.StatusCode/100 == 5) && attempt < runFlagVals.retries
		}
	}

	var hcmd command.Command
	hcmd, scp, err = hscommand.NewHttpCommand(method, urlSrc, bodySrc, buildFlagVals.headers, scp, hctx, retry)
	if err != nil {
		return err
	}

	if runFlagVals.outfmt == "full" {
		hcmd = addInputFieldCommand{Cmd: hcmd, Field: "input"}
	}

	input, err := openInput(os.Stdin, commonFlagVals.infmt)
	if err != nil {
		return err
	}

	sink := openOutput(os.Stdout, os.Stdout, runFlagVals.outfmt)
	if fn := runFlagVals.failfile; fn != "" && fn != "-" {
		f, err := os.Create(fn)
		if err != nil {
			return err
		}
		defer func() {
			if err := f.Close(); err != nil && finalErr == nil {
				finalErr = err
			}
		}()
		sink.Err = f
	}

	return command.RunParallel(ctx, hcmd, binds, input, sink, runFlagVals.parallel)
}
