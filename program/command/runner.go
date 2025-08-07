package command

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/daboyuka/hs/program/record"
	"github.com/daboyuka/hs/program/scope/bindings"
)

func RunParallel(ctx context.Context, cmd Command, binds *bindings.Bindings, input record.Stream, output record.Sink, n int, counter *atomic.Uint64) (finalErr error) {
	n = max(n, 1)

	wg := sync.WaitGroup{}
	defer wg.Wait() // don't return until everything's shut down

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	recCh := make(chan record.Record, n)
	errCh := make(chan error, n+1) // +1 for input goroutine

	// Feed recCh from input (or delivery error to errCh)
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(recCh)
		for rec, err := range input {
			if err != nil {
				errCh <- err
				return
			}
			select {
			case recCh <- rec:
			case <-ctx.Done():
				return
			}
		}
	}()

	// Run n goroutines, running cmd on inputs from recCh
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for in := range recCh {
				out, _, err := cmd.Run(ctx, in, binds)
				if err != nil {
					errCh <- err
					return
				}
				if counter != nil {
					counter.Add(1)
				}

				for rec, err := range out {
					if err == nil {
						err = output.Sink(rec)
					}
					if err != nil {
						errCh <- err
						return
					}
				}
			}
		}()
	}

	// Await all worker goroutines, then close errCh to unblock loop below
	go func() {
		wg.Wait()
		close(errCh)
	}()

	for err := range errCh {
		return err // if error occurs, abort (calls cancel() on defer, shutting everything down)
	}
	return nil
}
