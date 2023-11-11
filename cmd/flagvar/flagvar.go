package flagvar

import (
	"fmt"
	"strings"
)

type enumFlag struct {
	out      *string
	caseSens bool
	valid    []string
}

func NewEnumFlag(out *string, caseSens bool, valid ...string) enumFlag {
	return enumFlag{out: out, caseSens: caseSens, valid: valid}
}

func (e enumFlag) Type() string   { return "string" }
func (e enumFlag) String() string { return *e.out }
func (e enumFlag) Set(s string) error {
	sNorm := s
	if !e.caseSens {
		sNorm = strings.ToLower(sNorm)
	}

	for _, v := range e.valid {
		vNorm := v
		if !e.caseSens {
			vNorm = strings.ToLower(vNorm)
		}
		if sNorm == vNorm {
			*e.out = v
			return nil
		}
	}
	return fmt.Errorf("invalid arg '%s'", s)
}
