package expr

import (
	"fmt"
	"strings"

	"github.com/daboyuka/hs/program/record"
	"github.com/daboyuka/hs/program/scope"
	"github.com/daboyuka/hs/program/scope/bindings"
)

type FieldPath []Expr

func CheckValidFieldComponent(expr Expr) error {
	c, _ := expr.(Const)
	if num, ok := c.Val.(float64); ok {
		_, err := record.NumberToInt(num)
		return err
	}
	return nil
}
func (fp FieldPath) Eval(rec record.Record, binds *bindings.Bindings) (record.Record, error) {
	return fp.evalWithCtx(rec, rec, binds)
}

// evalWithCtx separates the "base" record (what we are indexing into) from the "context" record (the record that index
// expressions reference). For example, if the following FieldPath is applied to a record X:
//
//	.foo.bar[.baz]
//
// then in the final indexing [.baz], baseRec is the value X.foo.bar, while ctxRec is just X.
func (fp FieldPath) evalWithCtx(baseRec, ctxRec record.Record, binds *bindings.Bindings) (record.Record, error) {
	if len(fp) == 0 {
		return baseRec, nil
	}

	idx, err := fp[0].Eval(ctxRec, binds)
	if err != nil {
		return nil, err
	}

	var nextRec record.Record
	switch idx := idx.(type) {
	case float64:
		if intIdx, err := record.NumberToInt(idx); err != nil {
			return nil, fmt.Errorf("non-integer array index %f", idx)
		} else if arr, ok := baseRec.(record.Array); !ok {
			return nil, fmt.Errorf("array lookup on non-array %T", baseRec)
		} else if intIdx < 0 || intIdx >= len(arr) {
			return nil, fmt.Errorf("array index %d out of bounds on array of length %d", intIdx, len(arr))
		} else {
			nextRec = arr[intIdx]
		}
	case string:
		if obj, ok := baseRec.(record.Object); !ok {
			return nil, fmt.Errorf("string field lookup on non-object %T", baseRec)
		} else {
			nextRec = obj[idx]
		}
	}

	return fp[1:].evalWithCtx(nextRec, ctxRec, binds)
}

func (fp FieldPath) String() string {
	comps := make([]string, 1, 1+len(fp)) // len==1 adds blank string to start to give Join a prefix RuneFieldSep
	for _, fc := range fp {
		c, _ := fc.(Const) // c == Const{nil} if fc not a Const
		if s, ok := c.Val.(string); ok && scope.ValidIdent(s) {
			comps = append(comps, s) // special case: simple identifier-like string indices don't need brackets
		} else {
			comps = append(comps, "["+fc.String()+"]")
		}
	}
	return strings.Join(comps, ".")
}

type BaseFieldPath struct {
	Base Expr
	Path FieldPath
}

func (bfp BaseFieldPath) String() string { return bfp.Base.String() + bfp.Path.String() }

func (bfp BaseFieldPath) Eval(rec record.Record, binds *bindings.Bindings) (record.Record, error) {
	base, err := bfp.Base.Eval(rec, binds)
	if err != nil {
		return nil, err
	}
	return bfp.Path.evalWithCtx(base, rec, binds)
}
