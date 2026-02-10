package stream

import (
	"fmt"
	"reflect"
	"testing"
)

func TestApply(t *testing.T) {
	op := func(in int, out Sink[string]) error {
		for range in {
			if err := out(fmt.Sprintf("V%d", in)); err != nil {
				return err
			}
		}
		return nil
	}

	s2 := Apply(Slice([]int{1, 2, 3}), op)
	v, err := Collect(s2)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	} else if !reflect.DeepEqual(v, []string{"V1", "V2", "V2", "V3", "V3", "V3"}) {
		t.Errorf("unexpected result: %v", v)
	}
}

func TestMap(t *testing.T) {
	m := func(in int) (string, error) {
		return fmt.Sprintf("V%d", in), nil
	}

	s2 := Map(Slice([]int{1, 2, 3}), m)
	v, err := Collect(s2)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	} else if !reflect.DeepEqual(v, []string{"V1", "V2", "V3"}) {
		t.Errorf("unexpected result: %v", v)
	}
}
