package httpcmd

import (
	"context"
	"os"
	"time"

	"github.com/spf13/cobra"

	cmdctx "github.com/daboyuka/hs/cmd/context"
	"github.com/daboyuka/hs/hsruntime"
	hscommand "github.com/daboyuka/hs/hsruntime/command"
	"github.com/daboyuka/hs/program/command"
)

const maxInputBufferRecords = 1 << 16

func cmdRun(cmd *cobra.Command, args []string) (finalErr error) {
	hctx, err := cmdctx.Init(hsruntime.Options{CookieSpecs: runFlagVals.cookies}, true)
	if err != nil {
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

	isStdoutNormalFile := isFileOutput(os.Stdout)
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
	defer sink.Finish()

	ctx, cancel := context.WithCancel(context.Background())

	enableProgress := runFlagVals.progress == "true" || (runFlagVals.progress == "auto" && isStdoutNormalFile)
	input, outCounter, awaitProgressLogger := attachProgressLogger(ctx, input, enableProgress, maxInputBufferRecords, time.Second/4, os.Stderr)
	defer awaitProgressLogger()

	attachInterruptForHttpRunner(ctx, hcmd.SetDryRun, cancel)

	defer cancel()
	return command.RunParallel(ctx, hcmd, binds, input, sink, runFlagVals.parallel, outCounter)
}
