package stream

import (
	"sync/atomic"
	"testing"
	"testing/synctest"
)

func signalWhenDone[T any](s Stream[T], done *bool) Stream[T] {
	return func(yield func(T) error) error {
		err := s(yield)
		*done = true
		return err
	}
}

func rangeStream(n int) Stream[int] {
	return func(yield func(int) error) error {
		for i := range n {
			if err := yield(i); err != nil {
				return err
			}
		}
		return nil
	}
}

func countSink[T any](s Sink[T], counter *atomic.Int64) Sink[T] {
	return func(v T) error { counter.Add(1); return s(v) }
}

func blockSink[T any](s Sink[T], blockUntil <-chan struct{}) Sink[T] {
	return func(v T) error { <-blockUntil; return s(v) }
}

func TestParallel(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		doneBeforePll, doneAfterPll := false, false

		s := rangeStream(10)
		s = signalWhenDone(s, &doneBeforePll)
		s = Parallel(s)
		s = signalWhenDone(s, &doneAfterPll)

		var sinkCnt atomic.Int64
		blocked := make(chan struct{})
		sink := DiscardSink[int]
		sink = blockSink(sink, blocked)
		sink = countSink(sink, &sinkCnt)

		go Run(s, sink)

		// First, ensure Parallel allows the inner stream to complete despite sink not returning
		synctest.Wait()
		if n := sinkCnt.Load(); n != 10 {
			t.Errorf("unexpected sink hit count: %d", n)
		} else if !doneBeforePll {
			t.Errorf("expected upstream done")
		} else if doneAfterPll {
			t.Errorf("expected downstream not done")
		}

		// Next, unblock the sinks and ensure everything drains and exits
		close(blocked)

		synctest.Wait()
		if !doneAfterPll {
			t.Errorf("expected stream to finish")
		}
	})
}

func TestLimitedParallel(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		doneBeforePll, doneAfterPll := false, false

		s := rangeStream(10)
		s = signalWhenDone(s, &doneBeforePll)
		s = LimitedParallel(s, 5)
		s = signalWhenDone(s, &doneAfterPll)

		var sinkCnt atomic.Int64
		blocked := make(chan struct{}, 5)
		sink := DiscardSink[int]
		sink = blockSink(sink, blocked)
		sink = countSink(sink, &sinkCnt)

		go Run(s, sink)

		// First, ensure LimitedParallel allows the inner stream to complete despite sink not returning
		synctest.Wait()
		if n := sinkCnt.Load(); n != 10 {
			t.Errorf("unexpected sink hit count: %d", n)
		} else if !doneBeforePll {
			t.Errorf("expected upstream done")
		} else if doneAfterPll {
			t.Errorf("expected downstream not done")
		}

		// Next, unblock the sinks and ensure everything drains and exits
		close(blocked)

		synctest.Wait()
		if !doneAfterPll {
			t.Errorf("expected stream to finish")
		}
	})
}
