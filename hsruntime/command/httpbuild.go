package command

import (
	"context"

	"github.com/daboyuka/hs/hsruntime"
	"github.com/daboyuka/hs/program/record"
	"github.com/daboyuka/hs/program/scope"
	"github.com/daboyuka/hs/program/scope/bindings"
	"github.com/daboyuka/hs/stream"
)

type HttpBuildCommand struct {
	httpBuilder
}

func NewHttpBuildCommand(method, url, body string, headers []string, scope *scope.Scope, hctx *hsruntime.Context) (cmd *HttpBuildCommand, nextScope *scope.Scope, err error) {
	cmd = &HttpBuildCommand{}
	cmd.httpBuilder, err = newHttpBuilder(method, url, body, headers, scope, hctx)
	return cmd, scope, err
}

func (h *HttpBuildCommand) Run(ctx context.Context, in record.Record, binds *bindings.Bindings) (out record.Stream, outBinds *bindings.Bindings, err error) {
	req, err := h.buildRequest(in, binds)
	if err != nil {
		return nil, nil, err
	}
	return stream.Singleton[record.Record](requestToRecord(req)), binds, err
}
