package bindings

import (
	"encoding/json"
	"slices"
	"strings"

	"github.com/daboyuka/hs/program/record"
	"github.com/daboyuka/hs/program/scope"
)

// Bindings is a set of runtime binds of Idents. Bindings are always read-only.
type Bindings struct {
	parent *Bindings

	binds map[scope.Ident]record.Record
}

// New creates a Bindings derived from parent (or root if parent == nil) with the given value binds.
// The caller must not mutate the binds map after calling New.
func New(parent *Bindings, binds map[scope.Ident]record.Record) *Bindings {
	return &Bindings{parent: parent, binds: binds}
}

// Get returns the value associated with id, if bound.
func (binds *Bindings) Get(id scope.Ident) (val record.Record, bound bool) {
	if binds == nil {
		return nil, false
	}
	val, bound = binds.binds[id]
	if !bound {
		val, bound = binds.parent.Get(id)
	}
	return
}

func (binds *Bindings) AllIdents() (out []scope.Ident) {
	if binds == nil {
		return nil
	}
	for id := range binds.binds {
		out = append(out, id)
	}
	for _, id := range binds.parent.AllIdents() {
		if _, ok := binds.binds[id]; !ok {
			out = append(out, id)
		}
	}
	return out
}

func (binds *Bindings) AllValues() map[string]record.Record {
	out := make(map[string]record.Record)
	binds.allValuesInto(out)
	return out
}
func (binds *Bindings) allValuesInto(into map[string]record.Record) {
	if binds == nil {
		return
	}
	binds.parent.allValuesInto(into)
	for id, val := range binds.binds {
		into[id.String()] = val
	}
}

func (binds *Bindings) String() string {
	buf := strings.Builder{}
	ids := binds.AllIdents()
	slices.SortFunc(ids, func(a, b scope.Ident) int { return strings.Compare(a.String(), b.String()) })
	for _, id := range ids {
		val, _ := binds.Get(id)
		buf.WriteString(id.String())
		buf.WriteString(" = ")
		j, _ := json.Marshal(val)
		buf.Write(j)
		buf.WriteByte('\n')
	}
	return buf.String()
}

func (binds *Bindings) Scoped(scp *scope.Scope) Scoped {
	return Scoped{Scope: scp, Binds: binds}
}

// Scoped combines a Scope and Bindings to give the visible symbols and their bound values at a particular point
// in a program.
type Scoped struct {
	Scope *scope.Scope
	Binds *Bindings
}

func (sb Scoped) Lookup(name string) (record.Record, bool) {
	return sb.Binds.Get(sb.Scope.Lookup(name))
}
