package expr

import (
	"errors"
	"fmt"
	"strings"

	"github.com/daboyuka/hs/program/record"
	"github.com/daboyuka/hs/program/scope"
	"github.com/daboyuka/hs/program/scope/bindings"
)

var ErrNoBinding = errors.New("reference to unbound variable")

// Expr models an expression. Its String method returns a human-readable form.
type Expr interface {
	fmt.Stringer
	Eval(rec record.Record, binds *bindings.Bindings) (record.Record, error)
}

func EvalToString(e Expr, rec record.Record, binds *bindings.Bindings) (string, error) {
	if out, err := e.Eval(rec, binds); err != nil {
		return "", err
	} else {
		return record.CoerceString(out), nil
	}
}

//
// Expr impls.
//

type Const struct{ Val record.Record }

func (c Const) Eval(record.Record, *bindings.Bindings) (record.Record, error) { return c.Val, nil }
func (c Const) String() string                                                { return record.CoerceString(c.Val) }

type Var struct{ Id scope.Ident }

func (v Var) Eval(rec record.Record, binds *bindings.Bindings) (record.Record, error) {
	if val, ok := binds.Get(v.Id); ok {
		return val, nil
	}
	return nil, ErrNoBinding
}

func (v Var) String() string { return v.Id.String() }

type Func struct {
	Func     scope.Func
	FuncName string
	Args     []Expr
}

func (f Func) Eval(rec record.Record, binds *bindings.Bindings) (record.Record, error) {
	vals := make([]record.Record, len(f.Args))
	for i, arg := range f.Args {
		v, err := arg.Eval(rec, binds)
		if err != nil {
			return nil, fmt.Errorf("arg %d to func %s: %w", i+1, f.FuncName, err)
		}
		vals[i] = v
	}
	return f.Func(vals...)
}

func (f Func) String() string {
	buf := strings.Builder{}
	buf.WriteString("(")
	buf.WriteString(f.FuncName)
	for _, arg := range f.Args {
		buf.WriteString(" ")
		buf.WriteString(arg.String())
	}
	buf.WriteString(")")
	return buf.String()
}
