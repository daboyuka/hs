package httpcmd

// TODO get rid of this

import (
	"context"
	"io"
	"sync/atomic"
	"time"

	"github.com/daboyuka/hs/program/record"
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
	return in, nil, func() {}
}
