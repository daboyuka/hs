package parser

import (
	"fmt"
	"testing"

	"github.com/daboyuka/hs/program/expr/parser/lex"
	"github.com/daboyuka/hs/program/record"
	"github.com/daboyuka/hs/program/scope"
)

func TestParseTemplate(t *testing.T) {
	s, _ := scope.NewScope(nil, "world", "wor", "ld", "W0R", "Ld")
	stub := func(args ...record.Record) (record.Record, error) { return nil, nil }
	fn := scope.NewFuncTable(nil, map[string]scope.Func{"myfunc": stub, "otherfunc": stub})

	log := func(src string) {
		defer func() {
			if p := recover(); p != nil {
				fmt.Println("error:", p)
			}
		}()

		l := lex.NewLex(src, lex.TemplateMode)
		p := newParser(l, s, fn)

		tmpl := p.parseTemplate()
		fmt.Printf("%v\n", tmpl)
	}

	log(`hello ${1}!`)
	log(`hello ${"world"}!`)

	log(`hello world!`)
	log(`hello ${world}!`)
	log(`hello ${W0R}${Ld}!`)

	log(`hello ${.}!`)
	log(`hello ${.world}!`)
	log(`hello ${.W0rld}!`)
	log(`hello ${.w.or.ld}!`)
	log(`hello ${.["w"][0].l["\x72\u0064"]}!`)
	log(`hello ${.["world😊"]}!`)

	log(`hello ${.foo[.bar]}!`)
	log(`hello ${.["foo \(.bar)"]}!`)
	log(`hello ${"\("\(.foo)")"}!`)

	log(`hello ${myfunc "w" .["or"] ld}!`)
	log(`hello ${  myfunc  "w"   .[  "o\(  "r"   )" ]   ld  }!`)
	log(`hello ${ myfunc "w" (otherfunc "or" ("l") (.["d"])) }!`)

	log(`hello ${wor.ld}!`)
	log(`hello ${(myfunc "wor" ld).foo}!`)

	fmt.Println("expect errors now:")

	log(`hello $`)
	log(`hello $😊`)
	log(`hello $!`)
	log(`hello ${}!`)
	log(`hello ${..}!`)
	log(`hello ${sekai}!`)
}
