package command

import (
	"context"

	"github.com/daboyuka/hs/program/record"
	"github.com/daboyuka/hs/program/scope"
)

// Command takes action on a sequence of Records, possibly returning more Records in response to each.
//
// Examples include an HTTP command, which runs a request per input Record and returns a response Record; or a parsing
// command, which reformats each input record into one (or more) Records in a new format.
type Command interface {
	// Run begins execution of the command on an input Record with given bindings, returning a record.Stream of results
	// and derivative bindings for that stream. Execution may continue asynchronously after return, until a call to
	// out.Next returns non-nil error (normally io.EOF).
	//
	// outBinds must be either binds itself or a derivative binding.
	//
	// Both Run and its returned record.Stream operate under ctx, and may return early with ctx.Err() on cancellation.
	//
	// Run is safe for concurrent use by multiple goroutines, and concurrent use with draining the record.Stream returned
	// by any previous call to Run.
	Run(ctx context.Context, in record.Record, binds *scope.Bindings) (out record.Stream, outBinds *scope.Bindings, err error)
}
