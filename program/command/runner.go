package command

import (
	"context"
	"io"
	"sync"

	"github.com/daboyuka/hs/program/record"
	"github.com/daboyuka/hs/program/scope"
)

func RunParallel(ctx context.Context, cmd Command, binds *scope.Bindings, input record.Stream, output record.Sink, n int) (finalErr error) {
	if n <= 0 {
		n = 1
	}

	wg := sync.WaitGroup{}
	defer wg.Wait() // don't return until everything's shut down

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	errCh := make(chan error, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				in, err := input.Next()
				if err == io.EOF {
					return
				} else if err != nil {
					errCh <- err
					return
				}

				out, _, err := cmd.Run(ctx, in, binds)
				if err != nil {
					errCh <- err
					return
				}

				for {
					rec, err := out.Next()
					if err == io.EOF {
						break
					} else if err != nil {
						errCh <- err
						return
					} else if err := output.Sink(rec); err != nil {
						errCh <- err
						return
					}
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(errCh)
	}()

	for err := range errCh {
		return err // if error occurs, abort (calls cancel() on defer, shutting everything down)
	}
	return nil
}
