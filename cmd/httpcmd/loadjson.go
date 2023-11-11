package httpcmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/daboyuka/hs/program/expr/parser"
	"github.com/daboyuka/hs/program/record"
	"github.com/daboyuka/hs/program/scope"
)

func loadJSONTable(spec string, sb scope.ScopedBindings, fns *scope.FuncTable) (scope.ScopedBindings, error) {
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

	j := record.NewJSONStream(f)
	table := make(record.Object)

	for {
		if v, err := j.Next(); err == io.EOF {
			break
		} else if err != nil {
			return sb, fmt.Errorf("could not read file '%s': %s", filename, err)
		} else if k, err := keyexpr.Eval(v, sb.Binds); err != nil {
			return sb, fmt.Errorf("error evaluating key expression: %s", err)
		} else {
			table[record.CoerceString(k)] = v
		}
	}

	s2, ids := scope.NewScope(sb.Scope, varname)
	b2 := scope.NewBindings(sb.Binds, map[scope.Ident]record.Record{ids[0]: table})
	return scope.ScopedBindings{Scope: s2, Binds: b2}, nil
}
