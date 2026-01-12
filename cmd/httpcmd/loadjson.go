package httpcmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/daboyuka/hs/program/expr/parser"
	"github.com/daboyuka/hs/program/record"
	"github.com/daboyuka/hs/program/scope"
	"github.com/daboyuka/hs/program/scope/bindings"
)

func loadJSONTable(spec string, sb bindings.Scoped, fns *scope.FuncTable) (bindings.Scoped, error) {
	filename, rest, _ := strings.Cut(spec, ",")
	varname, keyexprStr, ok := strings.Cut(rest, ",")
	if !ok {
		return sb, fmt.Errorf("bad load spec '%s', should be of form 'filename,varname,keyexpr'", spec)
	}

	keyexpr, err := parser.ParseExpr(keyexprStr, sb.Scope, fns)
	if err != nil {
		return sb, fmt.Errorf("bad key expression '%s': %s", keyexprStr, err)
	}

	f, err := os.Open(filename)
	if err != nil {
		return sb, fmt.Errorf("could not open file '%s': %s", filename, err)
	}
	defer f.Close()

	table := make(record.Object)
	fs := record.NewJSONStream(f)
	err = fs(func(v record.Record) error {
		k, err := keyexpr.Eval(v, sb.Binds)
		if err != nil {
			return fmt.Errorf("error evaluating key expression: %s", err)
		}
		table[record.CoerceString(k)] = v
		return nil
	})
	if err != nil {
		return sb, fmt.Errorf("could not read file '%s': %s", filename, err)
	}

	s2, ids := scope.NewScope(sb.Scope, varname)
	b2 := bindings.New(sb.Binds, map[scope.Ident]record.Record{ids[0]: table})
	return bindings.Scoped{Scope: s2, Binds: b2}, nil
}
