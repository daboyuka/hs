package exprcmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/daboyuka/hs/hsruntime"
	"github.com/daboyuka/hs/hsruntime/plugin"
	"github.com/daboyuka/hs/program/expr/parser"
	"github.com/daboyuka/hs/program/record"
)

var Cmd *cobra.Command

func init() {
	Cmd = &cobra.Command{
		Use:   "expr expression",
		Short: "evaluate an expression",
		Long:  "evaluate the given expression on a stream of records; input and output are JSON",
		Args:  cobra.ExactArgs(1),
		RunE:  cmdExpr,
	}
}

func cmdExpr(cmd *cobra.Command, args []string) (err error) {
	hctx, err := hsruntime.NewDefaultContext(hsruntime.Options{})
	if err != nil {
		return err
	} else if err := plugin.Apply(hctx); err != nil {
		return err
	}

	expr, err := parser.ParseExpr(args[0], hctx.Globals.Scope, hctx.Funcs)
	if err != nil {
		return fmt.Errorf("bad expression: %w", err)
	}

	input := json.NewDecoder(os.Stdin)
	output := json.NewEncoder(os.Stdout)

	for {
		var rec record.Record
		if err := input.Decode(&rec); err == io.EOF {
			return nil
		} else if err != nil {
			return fmt.Errorf("error reading input: %w", err)
		}

		out, err := expr.Eval(rec, hctx.Globals.Binds)
		if err != nil {
			return fmt.Errorf("error evaluating expression: %w", err)
		}

		if err := output.Encode(out); err != nil {
			return fmt.Errorf("error writing output: %w", err)
		}
	}
}
