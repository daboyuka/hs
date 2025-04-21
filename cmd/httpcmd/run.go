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

const maxInputBufferRecords = 1 << 16

func cmdRun(cmd *cobra.Command, args []string) (finalErr error) {
	hctx, err := hsruntime.NewDefaultContext(hsruntime.Options{CookieSpecs: runFlagVals.cookies})
	if err != nil {
		return err
	} else if err := plugin.Apply(hctx); err != nil {
		return err
	}
	scp, binds := hctx.Globals.Scope, hctx.Globals.Binds

	var retry hscommand.RetryFunc
	if runFlagVals.retries > 0 {
		retry = func(req hscommand.RequestAndBody, resp hscommand.ResponseAndBody, attempt int) (backoff time.Duration, retry bool) {
			return time.Second, (resp.HTTPError != nil || resp.StatusCode/100 == 5) && attempt < runFlagVals.retries
		}
	}

	hcmd, scp, err := hscommand.NewHttpRunCommand(scp, hctx, retry)
	if err != nil {
		return err
	}

	input, err := openInput(os.Stdin, commonFlagVals.infmt)
	if err != nil {
		return err
	}

	isNonFileOut := isNonFileOutput(os.Stdout)

	sink := openOutput(os.Stdout, os.Stdout, runFlagVals.outfmt, isNonFileOut)
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

	ctx, cancel := context.WithCancel(context.Background())

	enableProgress := runFlagVals.progress == "true" || (runFlagVals.progress == "auto" && !isNonFileOut)
	input, outCounter, awaitProgressLogger := attachProgressLogger(ctx, input, enableProgress, maxInputBufferRecords, time.Second/4, os.Stderr)
	defer awaitProgressLogger()

	attachInterruptForHttpRunner(ctx, hcmd.SetDryRun, cancel)

	defer cancel()
	return command.RunParallel(ctx, hcmd, binds, input, sink, runFlagVals.parallel, outCounter)
}
