package stream

import (
	"sync"
)

func Parallel[T any](s Stream[T]) Stream[T] {
	return func(yield func(T) error) error {
		var wg sync.WaitGroup
		errCh := make(chan error, 1)

		err := s(func(v T) error {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := yield(v); err != nil {
					select {
					case errCh <- err:
					default:
					}
				}
			}()
			select {
			case err := <-errCh:
				return err
			default:
				return nil
			}
		})
		wg.Wait()
		close(errCh)
		err, _ = <-errCh
		return err
	}
}

func LimitedParallel[T any](s Stream[T], limit int) Stream[T] {
	return func(yield func(T) error) (finalErr error) {
		r := parallelRunner{workCh: make(chan func() error, limit), errCh: make(chan error, limit)}
		r.Start(limit)
		defer func() { finalErr = r.Stop(finalErr) }()

		return s(func(v T) error {
			return r.Do(func() error { return yield(v) })
		})
	}
}

type parallelRunner struct {
	wg     sync.WaitGroup
	errCh  chan error
	workCh chan func() error

	finalErr error
}

func (r *parallelRunner) Do(f func() error) error {
	select {
	case r.workCh <- f:
		return nil
	case err := <-r.errCh:
		return r.Stop(err)
	}
}

func (r *parallelRunner) Start(n int) {
	r.wg.Add(n)
	for range n {
		go r.worker()
	}
}

func (r *parallelRunner) Stop(withErr error) error {
	if r.finalErr == nil && withErr != nil {
		r.finalErr = withErr
	}

	if r.workCh != nil {
		close(r.workCh)
		r.wg.Wait()
		close(r.errCh)
		r.finalErr = <-r.errCh
		r.workCh = nil
	}

	return r.finalErr
}

func (r *parallelRunner) worker() {
	defer r.wg.Done()
	defer func() {
		if err, ok := recover().(error); ok {
			r.errCh <- err
		}
	}()
	for work := range r.workCh {
		if err := work(); err != nil {
			r.errCh <- err
			break
		}
	}
}
