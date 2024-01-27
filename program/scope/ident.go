package scope

// Ident is a unique identifier in a particular Scope, to be bound to a (real-only) value at runtime in a Bindings of
// that Scope.
type Ident struct{ handle *string }

func (id Ident) Valid() bool    { return id != Ident{} }
func (id Ident) String() string { return *id.handle }

func validIdentRune(r rune) bool { return validIdentStartRune(r) || (r >= '0' && r <= '9') }
func validIdentStartRune(r rune) bool {
	return (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || r == '_'
}

// ValidIdentPrefix returns length (in bytes) of the longest prefix of s that is a valid identifier.
func ValidIdentPrefix(s string) int {
	for i, r := range s {
		if i == 0 {
			if !validIdentStartRune(r) {
				return i
			}
		} else if !validIdentRune(r) {
			return i
		}
	}
	return len(s)
}

func ValidIdent(s string) bool { return ValidIdentPrefix(s) == len(s) }
