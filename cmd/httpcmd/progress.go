package httpcmd

import (
	"context"
	"io"
	"sync/atomic"
	"time"

	"github.com/daboyuka/hs/program/record"
	"github.com/daboyuka/hs/stream"
	"github.com/schollz/progressbar/v3"
)

// attachProgressLogger begins rendering a terminal-based progress bar to w if enable == true. Refresh occurs every
// interval. This function returns immediately, and continues until ctx is cancelled. waitForShutdown blocks until the
// progress logger exits (shortly after ctx is cancelled).
//
// Progress is determined by outCounter (caller should increment it as records are processed); progress bar max is
// indefinite until stream 'in' hits EOF, when it becomes total records streamed.
//
// Rendering only occurs if  to render is based on mode ("true", "false", "auto") and isOutTTY (if auto, render only if !isOutTTY)
func attachProgressLogger(ctx context.Context, in record.Stream, enable bool, maxBuffer int, interval time.Duration, w io.Writer) (newIn record.Stream, outCounter *atomic.Uint64, waitForShutdown func()) {
	if !enable {
		return in, nil, func() {}
	}

	outCounter = &atomic.Uint64{}
	counter := &stream.CountingStream[record.Record]{Stream: in}
	outBuf := stream.ChannelStream[record.Record]{Ch: make(chan record.RecordAndError, maxBuffer)}
	go func() {
		doneCh := ctx.Done()
		for {
			r, err := counter.Next()
			if err == io.EOF {
				close(outBuf.Ch)
				return
			}

			select {
			case outBuf.Ch <- record.RecordAndError{Val: r, Err: err}:
			case <-doneCh:
				return
			}
		}
	}()

	doneCh := make(chan struct{})
	go runProgressLogger(ctx, doneCh, counter, outCounter, interval, w)

	return outBuf, outCounter, func() { <-doneCh }
}

func runProgressLogger(ctx context.Context, doneCh chan<- struct{}, in *stream.CountingStream[record.Record], out *atomic.Uint64, interval time.Duration, w io.Writer) {
	defer close(doneCh)
	abortCh := ctx.Done()

	t := time.NewTicker(interval)
	defer t.Stop()

	foundMax := false
	pb := progressbar.NewOptions(-1,
		progressbar.OptionSetWriter(w),
		progressbar.OptionEnableColorCodes(false),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSetItsString("req"),
		progressbar.OptionSetElapsedTime(true),
		//progressbar.OptionShowDescriptionAtLineEnd(),
		progressbar.OptionSetDescription("(buffering input)"),
	)

	for {
		select {
		case <-t.C:
			inCnt, inDone := in.Count()
			outCnt := out.Load()

			if !foundMax && inDone {
				_ = pb.Set(0) // hack to force progress bar to recompute when ChangeMax is called
				pb.ChangeMax(inCnt)
				pb.Describe("")
				foundMax = true
			}

			_ = pb.Set(int(outCnt))
		case <-abortCh:
			_ = pb.Finish()
			w.Write([]byte{'\n'})
			return
		}
	}
}
