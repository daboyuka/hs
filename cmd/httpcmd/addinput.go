package httpcmd

import (
	"context"

	"github.com/daboyuka/hs/program/command"
	"github.com/daboyuka/hs/program/record"
	"github.com/daboyuka/hs/program/scope/bindings"
)

type addInputFieldCommand struct {
	Cmd   command.Command
	Field string
}

func (cmd addInputFieldCommand) Run(ctx context.Context, in record.Record, binds *bindings.Bindings) (out record.Stream, outBinds *bindings.Bindings, err error) {
	out, outBinds, err = cmd.Cmd.Run(ctx, in, binds)
	if err == nil {
		out = addFieldStream(out, cmd.Field, in)
	}
	return
}

func addFieldStream(out record.Stream, field string, val record.Record) record.Stream {
	return func(yield func(record.Record, error) bool) {
		for rec, err := range out {
			if err == nil {
				rec.(record.Object)[field] = val
			}
			if !yield(rec, err) {
				break
			}
		}
	}
}
