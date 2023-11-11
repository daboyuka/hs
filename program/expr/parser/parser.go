package parser

import (
	"fmt"

	"github.com/daboyuka/hs/program/expr"
	"github.com/daboyuka/hs/program/expr/parser/lex"
	"github.com/daboyuka/hs/program/scope"
)

//
// Grammar (https://www.w3.org/TR/REC-xml/#sec-notation)
//
//  WS  ::= [#x20#x09]
//  DIG ::= [0-9]
//  HEX ::= [0-9] | [a-f] | [A-F]
//
//	/* in "expr mode" */
//
//  IDENT-FIRST ::= [A-Z] | [a-z] | '_'
//  IDENT       ::= IDENT-FIRST (IDENT-FIRST | DIG)*
//
//  NUMBER ::= '-'? DIG+
//
//  field-idx        ::= '[' WS? EXPR WS? ']
//  field-comp-first ::= '.' ( IDENT | fieldidx )
//  field-comp       ::= field-comp-first | fieldidx
//  field-path       ::= field-comp-first field-comp*
//
//  func-call ::= IDENT (WS expr)+
//
//  grouping ::= '(' top-expr ')
//
//	expr ::= IDENT
//         | NUMBER
//         | field-path
//         | str
//	       | grouping
//
//	top-expr ::= WS? (expr | func-call) WS?
//
//	/* in "string mode" */
//
//	STR-LIT ::= [^"\]+
//  str-esc ::= '\' [0nrt"'\]
//	          | '\' 'x' HEX HEX
//	          | '\' 'u' HEX HEX HEX HEX
//	          | '\' '(' top-expr ')
//
//	str ::= '"' (STR-LIT | str-esc)* '"'
//
//	/* in "template mode" */
//
//	TMPL-LIT ::= [^$]+
//	tmpl-esc ::= '$' '$'
//             | '$' '{' top-expr '}'
//
//	tmpl = (TMPL-LIT | tmpl-esc)*
//

func ParseExpr(src string, scp *scope.Scope, fns *scope.FuncTable) (expr expr.Expr, err error) {
	defer lex.RecoverSyntaxError(&err)
	p := newParser(lex.NewLex(src, lex.ExprMode), scp, fns)
	return p.parseExpr(false, lex.TokBad, lex.ExprMode), nil
}

func ParseString(src string, scp *scope.Scope, fns *scope.FuncTable) (expr expr.Expr, err error) {
	defer lex.RecoverSyntaxError(&err)
	p := newParser(lex.NewLex(src, lex.StringMode), scp, fns)
	return p.parseString(lex.StringMode), nil
}

func ParseTemplate(src string, scp *scope.Scope, fns *scope.FuncTable) (expr expr.Expr, err error) {
	defer lex.RecoverSyntaxError(&err)
	p := newParser(lex.NewLex(src, lex.TemplateMode), scp, fns)
	return p.parseTemplate(), nil
}

type parser struct {
	lex *lex.Lex
	scp *scope.Scope
	fns *scope.FuncTable
}

func newParser(lex *lex.Lex, scp *scope.Scope, fns *scope.FuncTable) *parser {
	return &parser{lex: lex, scp: scp, fns: fns}
}

func (p *parser) assertMode(m lex.Mode) {
	if p.lex.Mode() != m {
		panic(fmt.Errorf("wrong mode %v, expected %v", p.lex.Mode(), m))
	}
}

func (p *parser) parseError(msg string) error { return p.lex.SyntaxError(msg) }

func (p *parser) parseTemplate() (e expr.Expr) {
	p.assertMode(lex.TemplateMode)

	var tmpl expr.Template
	var nextLit string
	for {
		switch t := p.lex.Peek(); t.Kind {
		case lex.TokLiteral:
			p.lex.Adv()
			nextLit += t.Val.(string)
		case lex.TokTmplExprOpen:
			tmpl.Lits = append(tmpl.Lits, nextLit)
			nextLit = ""

			p.lex.AdvMode(lex.ExprMode)
			tmpl.Exprs = append(tmpl.Exprs, p.parseExpr(true, lex.TokTmplExprClose, lex.TemplateMode))
		case lex.TokEOF:
			tmpl.Lits = append(tmpl.Lits, nextLit)
			return tmpl.Simplify()
		}
	}
}

// parseExpr parses an expression. If close != TokBad, it requires and consumes that token as a terminal, switching to
// closeMode as it does.
func (p *parser) parseExpr(topExpr bool, close lex.TokenKind, closeMode lex.Mode) (e expr.Expr) {
	p.assertMode(lex.ExprMode)

	if topExpr && p.lex.Peek().Kind == lex.TokWhitespace {
		p.lex.Adv()
	}

	allowFieldPath := false
	switch t := p.lex.Peek(); t.Kind {
	case lex.TokFieldSep:
		e = p.parseFieldPath()
	case lex.TokIdent:
		p.lex.Adv()
		name := t.Val.(string)

		// Only attempt to parse func args in top-expr mode
		var args []expr.Expr
		if topExpr {
			args = p.parseFuncArgs(close)
		}

		if len(args) != 0 {
			fn := p.fns.Get(name)
			if fn == nil {
				panic(p.parseError("reference to undeclared func '" + name + "'"))
			}
			e = expr.Func{Func: fn, FuncName: name, Args: args}
		} else {
			id := p.scp.Lookup(name)
			if !id.Valid() {
				panic(p.parseError("reference to undeclared variable '" + name + "'"))
			}
			e = expr.Var{Id: id}
			allowFieldPath = true
		}
	case lex.TokNumber:
		p.lex.Adv()
		e = expr.Const{Val: t.Val}
	case lex.TokStrOpen:
		p.lex.AdvMode(lex.StringMode)
		e = p.parseString(lex.ExprMode)
	case lex.TokExprOpen:
		p.lex.Adv()
		e = p.parseExpr(true, lex.TokExprClose, lex.ExprMode)
		allowFieldPath = true
	default:
		panic(p.parseError("unexpected token '" + p.lex.RawToken() + "'"))
	}

	if allowFieldPath {
		if fp := p.parseFieldPath(); fp != nil {
			e = expr.BaseFieldPath{Base: e, Path: fp}
		}
	}

	if topExpr && p.lex.Peek().Kind == lex.TokWhitespace {
		p.lex.Adv()
	}

	if close != lex.TokBad {
		if t := p.lex.Peek().Kind; t != close {
			panic(p.parseError(fmt.Sprintf("expected end of expression with %d, got %d", close, t)))
		}
		p.lex.AdvMode(closeMode)
	}
	return e
}

func (p *parser) parseFuncArgs(close lex.TokenKind) (args []expr.Expr) {
	// Keep trying to parse (non-top) expressions and spaces
	for {
		// Expect whitespace preceding next arg
		if p.lex.Peek().Kind != lex.TokWhitespace {
			return
		}
		p.lex.Adv()

		// Stop on closing token (do not consume)
		if p.lex.Peek().Kind == close {
			return
		}

		arg := p.parseExpr(false, lex.TokBad, lex.ExprMode)
		args = append(args, arg)
	}
}

func (p *parser) parseFieldPath() (fp expr.FieldPath) {
	for {
		hasSep := p.lex.Peek().Kind == lex.TokFieldSep
		if hasSep {
			p.lex.Adv()
		}

		if t := p.lex.Peek(); t.Kind == lex.TokIdent && hasSep {
			p.lex.Adv()
			fp = append(fp, expr.Const{Val: t.Val})
		} else if t.Kind == lex.TokIdxOpen {
			p.lex.Adv()
			next := p.parseExpr(true, lex.TokIdxClose, lex.ExprMode)
			if err := expr.CheckValidFieldComponent(next); err != nil {
				panic(p.parseError(err.Error()))
			}
			fp = append(fp, next)
		} else {
			break
		}
	}

	return fp
}

func (p *parser) parseString(closeMode lex.Mode) (ret expr.Expr) {
	p.assertMode(lex.StringMode)

	var tmpl expr.Template
	var nextLit string
	for {
		switch t := p.lex.Peek(); t.Kind {
		case lex.TokLiteral:
			p.lex.Adv()
			nextLit += t.Val.(string)
		case lex.TokExprOpen:
			tmpl.Lits = append(tmpl.Lits, nextLit)
			nextLit = ""

			p.lex.AdvMode(lex.ExprMode)
			tmpl.Exprs = append(tmpl.Exprs, p.parseExpr(true, lex.TokExprClose, lex.StringMode))
		case lex.TokStrClose:
			p.lex.AdvMode(closeMode)
			tmpl.Lits = append(tmpl.Lits, nextLit)
			return tmpl.Simplify()
		default:
			panic(p.parseError("unexpected token '" + p.lex.RawToken() + "'"))
		}
	}
}
