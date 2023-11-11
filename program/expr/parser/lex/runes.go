package lex

const (
	runeFieldSep   = '.'
	runeTmplEsc    = '$'
	runeStrEsc     = '\\'
	runeFieldOpen  = '['
	runeFieldClose = ']'
	runeQuote      = '"'

	runeTmplExprOpen  = '{' // inside template (with escape)
	runeTmplExprClose = '}'
	runeExprOpen      = '(' // inside string constant (with escape)
	runeExprClose     = ')'

	quoteOrStrEsc = string(runeQuote) + string(runeStrEsc)
	whitespace    = " \t"
)
