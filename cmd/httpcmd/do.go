package httpcmd

import (
	"context"
	"os"
	"time"

	"github.com/daboyuka/hs/program/record"
	"github.com/daboyuka/hs/program/scope"
	"github.com/daboyuka/hs/program/scope/bindings"
	"github.com/daboyuka/hs/program/stream"
	"github.com/spf13/cobra"

	cmdctx "github.com/daboyuka/hs/cmd/context"
	"github.com/daboyuka/hs/hsruntime"
	hscommand "github.com/daboyuka/hs/hsruntime/command"
)

func cmdDo(cmd *cobra.Command, args []string) (finalErr error) {
	method := cmd.Name()
	urlSrc := args[0]

	var bodySrc string
	if len(args) >= 2 {
		bodySrc = args[1]
	}

	hctx, err := cmdctx.Init(hsruntime.Options{CookieSpecs: runFlagVals.cookies}, true)
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

	var retry hscommand.RetryFunc
	if runFlagVals.retries > 0 {
		retry = func(req hscommand.RequestAndBody, resp hscommand.ResponseAndBody, attempt int) (backoff time.Duration, retry bool) {
			return time.Second, (resp.HTTPError != nil || resp.StatusCode/100 == 5) && attempt < runFlagVals.retries
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	input, err := openInput(os.Stdin, commonFlagVals.infmt)
	if err != nil {
		return err
	}

	var inputIdent scope.Ident
	var ops []stream.Operator[bindings.BoundRecord]

	if runFlagVals.outfmt == "full" {
		scp2, ids := scope.NewScope(scp, "_input")
		inputIdent = ids[0]
		m := func(br bindings.BoundRecord) (bindings.BoundRecord, error) {
			br.Binds = bindings.New(br.Binds, map[scope.Ident]record.Record{inputIdent: br.Rec})
			return br, nil
		}
		scp, ops = scp2, append(ops, stream.Mapper[bindings.BoundRecord](m).Operate)
	}

	hcmd, scp, err := hscommand.NewHttpCommand(ctx, method, urlSrc, bodySrc, buildFlagVals.headers, scp, hctx, retry)
	if err != nil {
		return err
	}
	attachInterruptForHttpRunner(ctx, hcmd.SetDryRun, cancel)
	ops = append(ops, hcmd.Operate)

	if inputIdent.Valid() {
		m := func(br bindings.BoundRecord) (bindings.BoundRecord, error) {
			inputV, _ := br.Binds.Get(inputIdent)
			br.Rec.(record.Object)["input"] = inputV
			return br, nil
		}
		ops = append(ops, stream.Mapper[bindings.BoundRecord](m).Operate)
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
	defer sink.Finish()

	isStdoutNormalFile := isFileOutput(os.Stdout)
	enableProgress := runFlagVals.progress == "true" || (runFlagVals.progress == "auto" && isStdoutNormalFile)

	input, outCounter, awaitProgressLogger := attachProgressLogger(ctx, input, enableProgress, maxInputBufferRecords, time.Second/4, os.Stderr)
	defer awaitProgressLogger()

	_ = outCounter // TODO figure this out

	// Build pipeline
	s := bindings.BindStream(input, binds)
	if runFlagVals.parallel > 1 {
		s = stream.LimitedParallel(s, runFlagVals.parallel)
	}
	for _, op := range ops {
		s = stream.Apply(s, op)
	}
	return stream.Run(s, bindings.BindSink(sink.Sink))
}
