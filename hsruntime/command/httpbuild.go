package command

import (
	"context"

	"github.com/daboyuka/hs/hsruntime"
	"github.com/daboyuka/hs/program/record"
	"github.com/daboyuka/hs/program/scope"
)

type HttpBuildCommand struct {
	httpBuilder
}

func NewHttpBuildCommand(method, url, body string, headers []string, scope *scope.Scope, hctx *hsruntime.Context) (cmd *HttpBuildCommand, nextScope *scope.Scope, err error) {
	cmd = &HttpBuildCommand{}
	cmd.httpBuilder, err = newHttpBuilder(method, url, body, headers, scope, hctx)
	return cmd, scope, err
}

func (h *HttpBuildCommand) Run(ctx context.Context, in record.Record, binds *scope.Bindings) (out record.Stream, outBinds *scope.Bindings, err error) {
	req, err := h.buildRequest(in, binds)
	if err != nil {
		return nil, nil, err
	}
	return &record.SingletonStream{Rec: requestToRecord(req)}, binds, err
}
