package lex

import (
	"unicode/utf8"
)

const EOF = rune(-1)

type runeStream struct {
	s string

	next    rune // next rune of s at pos
	nextLen int  // length of next rune (bytes)

	pos  int // position (bytes)
	rpos int // position (runes)

	mark  int
	rmark int
}

func newRuneStream(s string) runeStream {
	in := runeStream{s: s, mark: -1}
	in.refreshNext()
	return in
}

func (s *runeStream) refreshNext() {
	s.next, s.nextLen = utf8.DecodeRuneInString(s.s[s.pos:])
	if s.next == utf8.RuneError {
		if s.nextLen != 0 {
			panic(s.SyntaxError("invalid utf8 character"))
		}
		s.next = EOF
	}
}

func (s *runeStream) Peek() (r rune)   { return s.next }
func (s *runeStream) PeekLen() (l int) { return s.nextLen }

func (s *runeStream) Rem() string { return s.s[s.pos:] }

func (s *runeStream) Adv() rune {
	if s.next == EOF {
		panic(s.SyntaxError("unexpected eof"))
	}
	r := s.next
	s.rpos++
	s.pos += s.nextLen
	s.refreshNext()
	return r
}

func (s *runeStream) AdvBy(n int) {
	s.expect(n)
	s.rpos += utf8.RuneCountInString(s.s[s.pos : s.pos+n])
	s.pos += n
	s.refreshNext()
}

func (s *runeStream) Mark() { s.mark, s.rmark = s.pos, s.rpos }

func (s *runeStream) MarkStr() string {
	s.expectMark()
	return s.s[s.mark:s.pos]
}

func (s *runeStream) Reset() {
	s.expectMark()
	s.pos, s.rpos, s.mark = s.mark, s.rmark, -1
}

func (s *runeStream) expect(n int) {
	if s.pos+n > len(s.s) {
		panic(s.SyntaxError("unexpected eof"))
	}
}

func (s *runeStream) expectMark() {
	if s.mark == -1 {
		panic("no active Mark")
	}
}

// SyntaxError returns a SyntaxError at the current position.
func (s *runeStream) SyntaxError(msg string) SyntaxError {
	return SyntaxError{Pos: s.pos, RPos: s.rpos, Msg: msg}
}

// MarkSyntaxError returns a SyntaxError with a message at Mark's position.
func (s *runeStream) MarkSyntaxError(msg string) SyntaxError {
	s.expectMark()
	return SyntaxError{Pos: s.mark, RPos: s.rmark, Msg: msg}
}
