package command

import (
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/daboyuka/hs/hsruntime"
	"github.com/daboyuka/hs/hsruntime/datafmt"
	"github.com/daboyuka/hs/hsruntime/hostalias"
	"github.com/daboyuka/hs/program/expr"
	"github.com/daboyuka/hs/program/expr/parser"
	"github.com/daboyuka/hs/program/record"
	"github.com/daboyuka/hs/program/scope"
)

type RetryFunc func(req RequestAndBody, resp ResponseAndBody, attempt int) (backoff time.Duration, retry bool)

type RequestAndBody struct {
	*http.Request
	BodyContent string
}

type ResponseAndBody struct {
	*http.Response
	BodyContent string
	HTTPError   error
}

type httpBuilder struct {
	method  string
	url     expr.Expr
	body    expr.Expr
	headers []expr.Expr

	defaultHost  string
	hostAliasing hostalias.HostAlias

	autoContentTypeOnce sync.Once
	autoContentType     string
}

type httpRunner struct {
	client *http.Client
	retry  RetryFunc

	dryrun atomic.Bool
}

type HttpCommand struct {
	httpBuilder
	httpRunner
}

func NewHttpCommand(method, url, body string, headers []string, scope *scope.Scope, hctx *hsruntime.Context, retry RetryFunc) (cmd *HttpCommand, nextScope *scope.Scope, err error) {
	cmd = &HttpCommand{
		httpRunner: httpRunner{
			client: hctx.Client,
			retry:  retry,
		},
	}

	cmd.httpBuilder, err = newHttpBuilder(method, url, body, headers, scope, hctx)
	return cmd, scope, err
}

func (h *HttpCommand) Run(ctx context.Context, in record.Record, binds *scope.Bindings) (out record.Stream, outBinds *scope.Bindings, err error) {
	req, err := h.buildRequest(in, binds)
	if err != nil {
		return nil, nil, err
	}

	out, err = h.httpRunner.run(ctx, req)
	return out, binds, err
}

func newHttpBuilder(method, url, body string, headers []string, scp *scope.Scope, hctx *hsruntime.Context) (builder httpBuilder, finalErr error) {
	builder.method = method
	builder.defaultHost = hctx.DefaultHost
	builder.hostAliasing = hctx.HostAliasing

	parse := func(src string, out *expr.Expr) bool {
		*out, finalErr = parser.ParseTemplate(src, scp, hctx.Funcs)
		return finalErr == nil
	}

	if !parse(url, &builder.url) {
		return
	} else if body != "" && !parse(body, &builder.body) {
		return
	}
	builder.headers = make([]expr.Expr, len(headers))
	for i, hdr := range headers {
		if !parse(hdr, &builder.headers[i]) {
			return
		}
	}
	return
}

func (h *httpBuilder) buildRequest(in record.Record, binds *scope.Bindings) (req RequestAndBody, err error) {
	req.Request = &http.Request{
		Method: h.method,
		Header: make(http.Header),
	}

	uStr, err := expr.EvalToString(h.url, in, binds)
	if err != nil {
		return RequestAndBody{}, err
	}
	req.URL, err = url.Parse(uStr)
	if err != nil {
		return RequestAndBody{}, err
	}
	if req.URL.Scheme == "" {
		req.URL.Scheme = "https"
		if req.URL.Host == "" {
			req.URL.Host = h.defaultHost
		}
		if req.URL.Host == "" {
			return RequestAndBody{}, fmt.Errorf("request build is missing host, and no global HOST variable is set")
		}
	}

	if err := h.hostAliasing.Apply(req.URL); err != nil {
		return RequestAndBody{}, err
	}

	if h.body != nil {
		bodyStr, err := expr.EvalToString(h.body, in, binds)
		if err != nil {
			return RequestAndBody{}, err
		}
		bodyStrToRequestBody(bodyStr, &req)
	}

	for _, hdr := range h.headers {
		s, err := expr.EvalToString(hdr, in, binds)
		if err != nil {
			return RequestAndBody{}, err
		}

		colonIdx := strings.IndexRune(s, ':')
		if colonIdx == -1 {
			return RequestAndBody{}, fmt.Errorf("header field missing colon: %s", s)
		}
		field, val := s[:colonIdx], strings.TrimSpace(s[colonIdx+1:])
		req.Header.Add(field, val)
	}

	h.autodetectContentTypeIfNeeded(&req)

	return req, nil
}

// autodetectContentTypeIfNeeded autodetects and (if successful) applies Content-Type to req, provided it has a body
// and is missing Content-Type. It only detects on the first bodyContent seen, applying the same Content-Type thereafter.
func (h *httpBuilder) autodetectContentTypeIfNeeded(req *RequestAndBody) {
	if req.BodyContent == "" || len(req.Header.Values("Content-Type")) > 0 {
		return
	}

	// Autodetect only once
	h.autoContentTypeOnce.Do(func() {
		h.autoContentType = datafmt.Autodetect(req.BodyContent).ContentType() // = "" if Unknown
	})

	if h.autoContentType != "" {
		req.Header.Add("Content-Type", h.autoContentType)
	}
}

var dryrunErr = errors.New("request not sent")

func (h *httpRunner) run(ctx context.Context, req RequestAndBody) (out record.Stream, err error) {
	req.Request = req.Request.WithContext(ctx)

	req.Header.Set("Accept-Encoding", "gzip")
	outReq := RequestAndBody{Request: req.Clone(ctx), BodyContent: req.BodyContent}

	if h.dryrun.Load() {
		outRec := requestResponseToRecord(outReq, ResponseAndBody{HTTPError: dryrunErr}, nil)
		return &record.SingletonStream{Rec: outRec}, nil
	}

	var resp ResponseAndBody
	var retries []ResponseAndBody
	for {
		resp.Response, err = h.client.Do(req.Request)
		if err != nil {
			resp.HTTPError = err
		} else {
			var rdr io.Reader
			if resp.Response.Header.Get("Content-Encoding") == "gzip" {
				if gzipRdr, gzipRdrErr := gzip.NewReader(resp.Body); gzipRdrErr != nil {
					_ = resp.Body.Close()
					return nil, fmt.Errorf("gzip error: %w", gzipRdrErr)
				} else {
					rdr = gzipRdr
				}
			} else {
				rdr = resp.Body
			}

			if respBytes, err := io.ReadAll(rdr); err != nil {
				return nil, fmt.Errorf("response read error: %w", err)
			} else {
				_ = resp.Body.Close()
				resp.BodyContent = string(respBytes)
			}
		}

		if h.retry == nil {
			break
		} else if backoff, retry := h.retry(req, resp, len(retries)); !retry {
			break
		} else {
			time.Sleep(backoff)
			if req.Body != nil { // rebuild Body if needed
				if req.Body, err = req.GetBody(); err != nil {
					return nil, fmt.Errorf("failed to retry request: %w", err)
				}
			}

			retries = append(retries, resp)
			continue
		}
	}

	outRec := requestResponseToRecord(outReq, resp, retries)
	return &record.SingletonStream{Rec: outRec}, nil
}

// SetDryRun causes subsequent calls to run to immediately respond with status code 000 and human-readable error message.
// In-flight requests are unaffected.
func (h *httpRunner) SetDryRun() { h.dryrun.Store(true) }

func headersToRecord(h http.Header) record.Record {
	out := make(record.Object, len(h))
	for k, vs := range h {
		vsrec := make(record.Array, len(vs))
		for i, v := range vs {
			vsrec[i] = v
		}
		out[k] = vsrec
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func requestToRecord(req RequestAndBody) record.Object {
	ret := record.Object{
		"method": req.Method,
		"url":    req.URL.String(),
	}
	if hdrs := headersToRecord(req.Header); hdrs != nil {
		ret["headers"] = hdrs
	}
	if req.BodyContent != "" {
		ret["body"] = req.BodyContent
	}
	return ret
}

func responseToRecord(resp ResponseAndBody) record.Object {
	if resp.HTTPError != nil {
		return record.Object{"error": resp.HTTPError.Error()}
	}
	ret := record.Object{
		"status": float64(resp.StatusCode),
	}
	if hdrs := headersToRecord(resp.Header); hdrs != nil {
		ret["headers"] = hdrs
	}
	if resp.BodyContent != "" {
		ret["body"] = resp.BodyContent
	}
	return ret
}

func requestResponseToRecord(req RequestAndBody, resp ResponseAndBody, retries []ResponseAndBody) record.Object {
	ret := requestToRecord(req)
	ret["response"] = responseToRecord(resp)
	if len(retries) > 0 {
		retryObjs := make(record.Array, len(retries))
		for i, retry := range retries {
			retryObjs[i] = responseToRecord(retry)
		}
		ret["response"].(record.Object)["retries"] = retryObjs
	}
	return ret
}

func bodyStrToRequestBody(bodyStr string, req *RequestAndBody) {
	req.BodyContent = bodyStr
	req.GetBody = func() (io.ReadCloser, error) { return io.NopCloser(strings.NewReader(bodyStr)), nil }
	req.Body, _ = req.GetBody()
	req.ContentLength = int64(len(bodyStr))
}
