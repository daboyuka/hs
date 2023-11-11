package lex

type TokenKind int

const (
	TokBad = TokenKind(-2 + iota)
	TokEOF

	// Multiple mode tokens

	TokLiteral // literal character sequence (e.g. characters in string, template text outside expr)

	TokTmplExprOpen
	TokExprOpen

	TokTmplExprClose
	TokExprClose

	// ExprMode tokens

	TokWhitespace
	TokFieldSep
	TokIdxOpen
	TokIdxClose
	TokIdent
	TokNumber
	TokStrOpen

	// TemplateMode tokens

	// StringMode tokens

	TokStrClose
)

type Token struct {
	Kind TokenKind
	Val  any
}
