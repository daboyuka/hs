package scope

import (
	"encoding/json"
	"strings"

	"golang.org/x/exp/slices"

	"github.com/daboyuka/hs/program/record"
)

// Ident is a unique identifier in a particular Scope, to be bound to a (real-only) value at runtime in a Bindings of
// that Scope.
type Ident struct{ handle *string }

func (id Ident) Valid() bool    { return id != Ident{} }
func (id Ident) String() string { return *id.handle }

// Scope is lexical scope, a collection of in-scope Idents at some location in a program source. It may have a parent,
// inheriting Idents. A Scope is "instantiated" at runtime with Bind by binding values to its Idents.
//
// Scope is always read-only.
type Scope struct {
	parent *Scope

	// ids are the Idents local to this Scope
	ids []Ident
}

// NewScope creates a Scope derived from parent (or root if parent == nil) with the given names.
// New Idents corresponding to the names are returned, as well.
// Invariant: for i, name in names: scope.Lookup(name) == ids[i] && parent.Lookup(name) != ids[i]
func NewScope(parent *Scope, names ...string) (scope *Scope, ids []Ident) {
	ids = make([]Ident, len(names))
	for i, name := range names {
		name := name
		ids[i] = Ident{handle: &name}
	}
	return &Scope{parent: parent, ids: ids}, ids
}

// Parent returns this Scope's parent, or nil if this is a root Scope.
func (s *Scope) Parent() *Scope {
	if s == nil {
		return nil
	}
	return s.parent
}

// Lookup returns the Ident that name references relative to the current Scope, or nil if no such Ident.
// An Ident in an ancestor Scope is shadowed by an Ident of the same name in a child Scope.
func (s *Scope) Lookup(name string) Ident {
	if s == nil {
		return Ident{}
	}
	for _, id := range s.ids {
		if id.String() == name {
			return id
		}
	}
	return s.parent.Lookup(name)
}

// Bindings is a set of runtime bindings of Idents. Bindings are always read-only.
type Bindings struct {
	parent *Bindings

	binds map[Ident]record.Record
}

// NewBindings creates a Bindings derived from parent (or root if parent == nil) with the given value binds.
// The caller must not mutate the binds map after calling NewBindings.
func NewBindings(parent *Bindings, binds map[Ident]record.Record) *Bindings {
	return &Bindings{parent: parent, binds: binds}
}

// Get returns the value associated with id, if bound.
func (binds *Bindings) Get(id Ident) (val record.Record, bound bool) {
	if binds == nil {
		return nil, false
	}
	val, bound = binds.binds[id]
	if !bound {
		val, bound = binds.parent.Get(id)
	}
	return
}

func (binds *Bindings) AllIdents() (out []Ident) {
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
	slices.SortFunc(ids, func(a, b Ident) bool { return a.String() < b.String() })
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

func (binds *Bindings) Scoped(scp *Scope) ScopedBindings {
	return ScopedBindings{Scope: scp, Binds: binds}
}

// ScopedBindings combines a Scope and Bindings to give the visible symbols and their bound values at a particular point
// in a program.
type ScopedBindings struct {
	Scope *Scope
	Binds *Bindings
}

func (sb ScopedBindings) Lookup(name string) (record.Record, bool) {
	return sb.Binds.Get(sb.Scope.Lookup(name))
}
