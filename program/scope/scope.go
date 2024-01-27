package scope

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
