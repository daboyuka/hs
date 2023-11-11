package command

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/daboyuka/hs/hsruntime"
	"github.com/daboyuka/hs/program/record"
	"github.com/daboyuka/hs/program/scope"
)

type HttpRunCommand struct {
	httpRunner
}

func NewHttpRunCommand(scope *scope.Scope, hctx *hsruntime.Context, retry RetryFunc) (cmd *HttpRunCommand, nextScope *scope.Scope, finalErr error) {
	return &HttpRunCommand{
		httpRunner: httpRunner{
			client: hctx.Client,
			retry:  retry,
		},
	}, scope, nil
}

func (h *HttpRunCommand) Run(ctx context.Context, in record.Record, binds *scope.Bindings) (out record.Stream, outBinds *scope.Bindings, err error) {
	req, err := h.extractRequest(in)
	if err != nil {
		return nil, nil, err
	}

	out, err = h.run(ctx, req)
	return out, binds, err
}

func (h *HttpRunCommand) extractRequest(in record.Record) (req RequestAndBody, err error) {
	obj, _ := in.(record.Object)
	if method, ok := obj["method"].(string); !ok {
		return RequestAndBody{}, fmt.Errorf("missing HTTP method string in record: %s", record.CoerceString(in))
	} else if uStr, ok := obj["url"].(string); !ok {
		return RequestAndBody{}, fmt.Errorf("missing HTTP URL string in record: %s", record.CoerceString(in))
	} else if u, err := url.Parse(uStr); err != nil {
		return RequestAndBody{}, fmt.Errorf("malformed HTTP URL '%s' in record: %s", uStr, err)
	} else {
		req = RequestAndBody{
			Request: &http.Request{
				Method: method,
				URL:    u,
				Header: make(http.Header),
			},
		}
	}

	hdrs, _ := obj["headers"].(record.Object)
	for key, vals := range hdrs {
		if valsArr, ok := vals.(record.Array); ok {
			for _, val := range valsArr {
				if valStr, ok := val.(string); !ok {
					return RequestAndBody{}, fmt.Errorf("non-string value in array for header '%s' in record: %s", key, record.CoerceString(val))
				} else {
					req.Header.Add(key, valStr)
				}
			}
		} else if valStr, ok := vals.(string); ok {
			req.Header.Add(key, valStr)
		} else {
			return RequestAndBody{}, fmt.Errorf("non-string/non-array value for header '%s' in record: %s", key, record.CoerceString(vals))
		}
	}

	if bodyStr, ok := obj["body"].(string); ok {
		bodyStrToRequestBody(bodyStr, &req)
	}

	return req, nil
}
