package lex

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/daboyuka/hs/program/scope"
)

type Mode int

const (
	ExprMode = Mode(iota)
	StringMode
	TemplateMode
)

type Lex struct {
	iter runeStream
	next Token
	mode Mode
}

func NewLex(s string, mode Mode) *Lex {
	l := &Lex{iter: newRuneStream(s), mode: mode}
	l.refreshNext()
	return l
}

func (l *Lex) SyntaxError(msg string) SyntaxError { return l.iter.MarkSyntaxError(msg) }
func (l *Lex) RawToken() string                   { return l.iter.MarkStr() }

// Peek returns the next Token. Subsequent calls will return the sake Token until a call to Adv, SetMode, or AdvMode.
func (l *Lex) Peek() Token { return l.next }

func (l *Lex) Mode() Mode { return l.mode }

// Adv moves to lex the next Token. Subsequent calls to Peek return this Token.
func (l *Lex) Adv() { l.refreshNext() }

// AdvMode is as Adv, but first switches lex to a new lexer mode. It is equivalent to, but more efficient than,
// calling Adv then SetMode.
func (l *Lex) AdvMode(newMode Mode) {
	l.mode = newMode
	l.refreshNext()
}

// SetMode re-lexes the characters underlying the current Token in a new lexer mode, likely changing the lexed Token.
// Subsequent calls to Peek return this Token.
func (l *Lex) SetMode(newMode Mode) {
	if l.mode != newMode {
		l.iter.Reset()
		l.AdvMode(newMode)
	}
}

func (l *Lex) refreshNext() {
	if l.iter.Peek() == EOF {
		l.next = Token{Kind: TokEOF}
		return
	}

	l.iter.Mark()
	switch l.mode {
	case ExprMode:
		l.next = l.nextExpr()
	case StringMode:
		l.next = l.nextStr()
	case TemplateMode:
		l.next = l.nextTmpl()
	default:
		panic(fmt.Errorf("bad mode %d", l.mode))
	}
}

func (l *Lex) nextExpr() Token {
	switch r := l.iter.Peek(); r {
	case runeFieldSep:
		l.iter.Adv()
		return Token{Kind: TokFieldSep}
	case runeFieldOpen:
		l.iter.Adv()
		return Token{Kind: TokIdxOpen}
	case runeFieldClose:
		l.iter.Adv()
		return Token{Kind: TokIdxClose}
	case runeQuote:
		l.iter.Adv()
		return Token{Kind: TokStrOpen}
	case runeExprOpen:
		l.iter.Adv()
		return Token{Kind: TokExprOpen}
	case runeExprClose:
		l.iter.Adv()
		return Token{Kind: TokExprClose}
	case runeTmplExprClose:
		l.iter.Adv()
		return Token{Kind: TokTmplExprClose}
	default:
		if l.nextSpace() {
			return Token{Kind: TokWhitespace}
		} else if id := l.nextIdent(); id != "" {
			return Token{Kind: TokIdent, Val: id}
		} else if v, ok := l.nextNumber(); ok {
			return Token{Kind: TokNumber, Val: v}
		}
	}
	return Token{Kind: TokBad, Val: l.iter.Peek()}
}

func (l *Lex) nextSpace() (ws bool) {
	for strings.ContainsRune(whitespace, l.iter.Peek()) {
		ws = true
		l.iter.Adv()
	}
	return ws
}

func (l *Lex) nextIdent() string {
	s := l.iter.Rem()
	n := scope.ValidIdentPrefix(s)
	l.iter.AdvBy(n)
	return s[:n]
}

func (l *Lex) nextNumber() (float64, bool) {
	neg := false
	if l.iter.Peek() == '-' {
		neg = true
		l.iter.Adv()
	}

	empty := true
	v := 0
	for {
		r := l.iter.Peek()
		if r < '0' || r > '9' {
			break
		}
		l.iter.Adv()
		empty = false
		v = 10*v + int(r-'0')
	}

	if empty {
		if neg {
			panic(l.SyntaxError("bad number literal"))
		}
		return 0, false
	}
	return float64(v), true
}

func (l *Lex) nextTmpl() Token {
	s := l.iter.Rem()
	idx := strings.IndexByte(s, runeTmplEsc)

	// Literal case
	if idx != 0 {
		if idx == -1 {
			idx = len(s)
		}
		l.iter.AdvBy(idx)
		return Token{Kind: TokLiteral, Val: s[:idx]}
	}

	// Escape case
	l.iter.Adv() // skip the escape char
	switch r := l.iter.Adv(); r {
	case runeTmplEsc:
		return Token{Kind: TokLiteral, Val: string(runeTmplEsc)}
	case runeTmplExprOpen:
		return Token{Kind: TokTmplExprOpen}
	default:
		panic(l.SyntaxError(fmt.Sprintf("illegal escape '%c'", r)))
	}
}

func (l *Lex) nextStr() Token {
	s := l.iter.Rem()
	idx := strings.IndexAny(s, quoteOrStrEsc)

	// Literal case
	if idx != 0 {
		if idx == -1 {
			idx = len(s)
		}
		l.iter.AdvBy(idx)
		return Token{Kind: TokLiteral, Val: s[:idx]}
	}

	r := l.iter.Adv()
	if r == runeQuote {
		return Token{Kind: TokStrClose}
	}

	// r was an escape; get next char to find out which kind
	r = l.iter.Adv()
	switch r {
	case 'x':
		return Token{Kind: TokLiteral, Val: string(rune(l.parseHex(2)))}
	case 'u':
		return Token{Kind: TokLiteral, Val: string(rune(l.parseHex(4)))}
	case runeExprOpen:
		return Token{Kind: TokExprOpen}
	}

	if e := escapes[r]; e != "" {
		return Token{Kind: TokLiteral, Val: e}
	}

	panic(l.SyntaxError(fmt.Sprintf("illegal escape '\\%c'", r)))
}

var escapes = [256]string{'0': "\000", '\\': "\\", '\'': "'", '"': "\"", 'n': "\n", 'r': "\r", 't': "\t"}

func (l *Lex) parseHex(n int) (val uint64) {
	s := l.iter.Rem()
	l.iter.AdvBy(n)
	val, err := strconv.ParseUint(s[:n], 16, 64)
	if err != nil {
		panic(l.SyntaxError("bad hex: " + err.Error()))
	}
	return val
}
