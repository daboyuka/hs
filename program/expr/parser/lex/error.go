package lex

import (
	"fmt"
)

type SyntaxError struct {
	Pos  int // byte position
	RPos int // rune (character) position
	Msg  string
}

func (s SyntaxError) Error() string {
	return fmt.Sprintf("syntax error at position %d: %s", s.RPos, s.Msg)
}

// RecoverSyntaxError intercepts any SyntaxError panic and stores in *to, with no effect on other panics. Example:
//
//	func Example() (err error) {
//		defer RecoverSyntaxError(&err)
//		// ... code that may panic with SyntaxError here ...
//	}
func RecoverSyntaxError(to *error) {
	switch p := recover().(type) {
	case nil:
	case SyntaxError:
		*to = p
	default:
		panic(p)
	}
}
