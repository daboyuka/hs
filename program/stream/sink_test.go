package stream

import (
	"reflect"
	"testing"
)

func TestCollectSink(t *testing.T) {
	var x []int
	s := (*CollectSink[int])(&x)

	err1, err2, err3 := s.Sink(1), s.Sink(2), s.Sink(3)
	if err1 != nil || err2 != nil || err3 != nil {
		t.Errorf("unexpected errors: %v, %v, %v", err1, err2, err3)
	} else if !reflect.DeepEqual(x, []int{1, 2, 3}) {
		t.Errorf("unexpected result: %v", x)
	}
}
