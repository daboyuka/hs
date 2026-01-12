package command

import (
	"github.com/daboyuka/hs/hsruntime"
	"github.com/daboyuka/hs/program/scope"
	"github.com/daboyuka/hs/program/scope/bindings"
)

type HttpBuildCommand struct {
	httpBuilder
}

func NewHttpBuildCommand(method, url, body string, headers []string, scope *scope.Scope, hctx *hsruntime.Context) (cmd *HttpBuildCommand, nextScope *scope.Scope, err error) {
	cmd = &HttpBuildCommand{}
	cmd.httpBuilder, err = newHttpBuilder(method, url, body, headers, scope, hctx)
	return cmd, scope, err
}

func (h *HttpBuildCommand) Operate(in bindings.BoundRecord, sink bindings.BoundSink) error {
	outRec, err := h.buildRequest(in.Rec, in.Binds)
	if err != nil {
		return err
	}
	return sink(bindings.BoundRecord{Binds: in.Binds, Rec: outRec})
}
