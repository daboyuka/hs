package scope

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
