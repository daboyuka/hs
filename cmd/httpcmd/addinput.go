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
		out = addFieldStream{stream: out, field: cmd.Field, val: in}
	}
	return
}

type addFieldStream struct {
	stream record.Stream
	field  string
	val    record.Record
}

func (afs addFieldStream) Next() (record.Record, error) {
	out, err := afs.stream.Next()
	if err != nil {
		return nil, err
	}
	out.(record.Object)[afs.field] = afs.val
	return out, nil
}
