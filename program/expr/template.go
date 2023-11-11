package expr

import (
	"strings"

	"github.com/daboyuka/hs/program/record"
	"github.com/daboyuka/hs/program/scope"
)

// Template is a string template, alternating string literals and embedded expressions (with literal at begin and end).
// evals[i] occurs between lits[i] and lits[i+1].
type Template struct {
	Lits  []string
	Exprs []Expr
}

func (t Template) String() string {
	strs := make([]string, 0, len(t.Lits)+len(t.Exprs))
	for i, lit := range t.Lits {
		if lit != "" {
			strs = append(strs, record.StringEscape(lit))
		}
		if i < len(t.Exprs) {
			strs = append(strs, "${"+t.Exprs[i].String()+"}")
		}
	}
	return strings.Join(strs, "")
}

func (t Template) Eval(rec record.Record, binds *scope.Bindings) (record.Record, error) {
	strs := make([]string, 0, len(t.Lits)+len(t.Exprs))
	strs = append(strs, t.Lits[0])
	for i, e := range t.Exprs {
		if evaled, err := e.Eval(rec, binds); err != nil {
			return "", err
		} else {
			strs = append(strs, record.CoerceString(evaled), t.Lits[i+1])
		}
	}
	return strings.Join(strs, ""), nil
}

func (t Template) Simplify() Expr {
	switch len(t.Lits) {
	case 0:
		return Const{Val: ""}
	case 1:
		return Const{Val: t.Lits[0]}
	case 2:
		if t.Lits[0] == "" && t.Lits[1] == "" {
			return t.Exprs[0]
		}
	}
	return t
}
