package stream

import (
	"errors"
	"reflect"
	"testing"
)

func TestSlice(t *testing.T) {
	x := []int{1, 2, 3}
	s := Slice(x)

	var out []int
	sink := (*CollectSink[int])(&out)
	err := s(sink.Sink)
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(out, x) {
		t.Errorf("unexpected result: %v", out)
	}
}

func TestError(t *testing.T) {
	theErr := errors.New("boom")
	s := Error[int](theErr)

	err := s(func(int) error { panic("should not be called") })
	if err != theErr {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCollect(t *testing.T) {
	x := []int{1, 2, 3}
	y, err := Collect(Slice(x))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	} else if !reflect.DeepEqual(y, x) {
		t.Errorf("unexpected result: %v", y)
	}
}

func TestCollectError(t *testing.T) {
	theErr := errors.New("boom")
	s := Error[int](theErr)

	_, err := Collect(s)
	if err != theErr {
		t.Errorf("unexpected error: %v", err)
	}
}
