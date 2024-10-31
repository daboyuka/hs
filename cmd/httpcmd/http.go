package httpcmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/term"

	"github.com/daboyuka/hs/cmd/flagvar"
	"github.com/daboyuka/hs/hsruntime/datafmt"
	"github.com/daboyuka/hs/program/record"
)

var Group = &cobra.Group{ID: "http", Title: "HTTP request commands"}
var Commands []*cobra.Command

var (
	commonFlags    pflag.FlagSet
	commonFlagVals struct {
		infmt string
	}

	buildFlags    pflag.FlagSet
	buildFlagVals struct {
		headers   []string
		loadSpecs []string
	}

	runFlags    pflag.FlagSet
	runFlagVals struct {
		cookies  []string
		failfile string
		retries  int
		outfmt   string
		parallel int
	}
)

var allMethods = []string{
	http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch,
	http.MethodDelete, http.MethodOptions, http.MethodTrace,
}

func init() {
	infmts := []string{"auto", "null", "raw", "lines", "json", "csv", "rawcsv", "tsv", "rawtsv"}
	commonFlagVals.infmt = "auto" // default
	_ = commonFlags.VarPF(flagvar.NewEnumFlag(&commonFlagVals.infmt, false, infmts...), "in", "i", "set input mode (one of "+strings.Join(infmts, " ")+")")
}

func init() {
	buildFlags.StringArrayVarP(&buildFlagVals.headers, "header", "H", nil, "add an HTTP request header; flag may be repeated")
	buildFlags.StringArrayVarP(&buildFlagVals.loadSpecs, "loadjson", "L", nil, "load a JSON file as a lookup table; argument has syntax \"filename,varname,keyexpr\"\n"+
		"filename = file to load, varname = variable to load into (as an object record),\n"+
		"keyexpr = expression to extract the key for each loaded value, to store it as an entry in varname")

}

func init() {
	runFlags.StringArrayVarP(&runFlagVals.cookies, "cookie", "b", nil, "add an HTTP Cookie header; flag may be repeated\n"+
		"If the argument is as 'name=value', Cookie 'name' with value 'value' is added.\n"+
		"Otherwise, the argument is a cookiejar filename to read, in Netscape format.",
	)

	runFlags.StringVarP(&runFlagVals.failfile, "fails", "F", "-", "write fail responses (connection error or non-2xx response) to this file, or stdout if '-'")
	runFlags.IntVarP(&runFlagVals.retries, "retry", "r", 0, "num. retries on HTTP error or 5xx response")

	outfmts := []string{"auto", "full", "reqresp", "resp", "bodycode", "body"}
	runFlagVals.outfmt = "auto" // default
	_ = runFlags.VarPF(flagvar.NewEnumFlag(&runFlagVals.outfmt, false, outfmts...), "out", "o", "set output mode (one of "+strings.Join(outfmts, " ")+")")

	runFlags.IntVarP(&runFlagVals.parallel, "parallel", "P", 1, "request parallelism (no request ordering guaranteed when greater than 1)")
}

func init() {
	cmd := &cobra.Command{
		Use:     "build [flags] method url [body]",
		Short:   "build request(s) but do not run",
		Long:    "make request(s) but do not run; to be used later by 'run' command",
		GroupID: Group.ID,
		Args:    cobra.RangeArgs(2, 3),
		RunE:    cmdBuild,
	}
	cmd.Flags().AddFlagSet(&commonFlags)
	cmd.Flags().AddFlagSet(&buildFlags)

	Commands = append(Commands, cmd)
}

func init() {
	cmd := &cobra.Command{
		Use:     "run [flags]",
		Short:   "run pre-built or failed request(s)",
		Long:    "run pre-built requests, or failed requests from a prior run",
		GroupID: Group.ID,
		Args:    cobra.NoArgs,
		RunE:    cmdRun,
	}
	cmd.Flags().AddFlagSet(&commonFlags)
	cmd.Flags().AddFlagSet(&runFlags)

	Commands = append(Commands, cmd)
}

func init() {
	for _, method := range allMethods {
		cmd := &cobra.Command{
			Use:     method + " [flags] url [body]",
			Aliases: []string{strings.ToLower(method)},
			Short:   "make " + method + " request(s)",
			Long:    "make " + method + " request(s)",
			GroupID: Group.ID,
			Args:    cobra.RangeArgs(1, 2),
			RunE:    cmdDo,
		}
		cmd.Flags().AddFlagSet(&commonFlags)
		cmd.Flags().AddFlagSet(&buildFlags)
		cmd.Flags().AddFlagSet(&runFlags)

		Commands = append(Commands, cmd)
	}
}

func isFailResponse(rec record.Record) bool {
	recObj, _ := rec.(record.Object)
	respObj, _ := recObj["response"].(record.Object)
	errVal := respObj["error"]
	status, _ := respObj["status"].(float64)
	return errVal != nil || int(status)/100 != 2
}

func openInput(r io.Reader, infmt string) (parsed record.Stream, err error) {
	// Default: null if tty stdin, lines otherwise
	if infmt == "" || infmt == "auto" {
		if f, ok := r.(*os.File); ok && term.IsTerminal(int(f.Fd())) {
			infmt = "null"
		} else {
			infmt, r, err = autoInputFormat(r)
			if err != nil {
				return nil, err
			}
		}
	}

	switch infmt {
	case "null":
		return &record.SingletonStream{Rec: nil}, nil
	case "raw":
		return record.NewRawStream(r), nil
	case "lines":
		return record.NewLineStream(r), nil
	case "json":
		return record.NewJSONStream(r), nil
	case "rawcsv":
		return record.NewCsvReader(r, ',', true, false), nil
	case "csv":
		return record.NewCsvReader(r, ',', false, false), nil
	case "rawtsv":
		return record.NewCsvReader(r, '\t', true, false), nil
	case "tsv":
		return record.NewCsvReader(r, '\t', false, false), nil
	}
	return nil, fmt.Errorf("unsupported infmt '%s'", infmt)
}

func autoInputFormat(r io.Reader) (infmt string, r2 io.Reader, err error) {
	df, r2, err := datafmt.AutodetectReader(r)
	if err != nil {
		return "", nil, err
	}

	switch df {
	case datafmt.JSON:
		return "json", r2, nil
	default:
		return "lines", r2, nil
	}
}

type outputFormatter func(record.Record) (record.Record, error)

func newOutputFormatter(outfmt string, tty bool) outputFormatter {
	if outfmt == "auto" {
		if tty {
			outfmt = "body"
		} else {
			outfmt = "reqresp"
		}
	}

	switch outfmt {
	case "reqresp", "full":
		return func(rr record.Record) (record.Record, error) { return rr, nil }
	case "resp":
		return func(rr record.Record) (record.Record, error) {
			return rr.(record.Object)["response"], nil
		}
	case "body":
		return func(rr record.Record) (record.Record, error) {
			respObj := rr.(record.Object)["response"].(record.Object)
			if errVal, ok := respObj["error"]; ok {
				return errVal, nil
			}
			return respObj["body"], nil
		}
	case "bodycode":
		return func(rr record.Record) (record.Record, error) {
			respObj := rr.(record.Object)["response"].(record.Object)
			if errVal, ok := respObj["error"]; ok {
				return "000\n" + record.CoerceString(errVal), nil
			}
			return record.CoerceString(respObj["status"]) + "\n" + record.CoerceString(respObj["body"]), nil
		}
	}
	panic(fmt.Errorf("unsupported outfmt '%s'", outfmt))
}

func openOutput(out, err io.StringWriter, outfmt string) *responseSplitFileSink {
	isTTY := false
	if f, ok := out.(*os.File); ok && term.IsTerminal(int(f.Fd())) {
		isTTY = true
	}

	return &responseSplitFileSink{
		OutFmt: newOutputFormatter(outfmt, isTTY),
		Out:    out,
		Err:    err,
	}
}

type responseSplitFileSink struct {
	mtx    sync.Mutex
	OutFmt outputFormatter
	Out    io.StringWriter
	Err    io.StringWriter
}

func (w *responseSplitFileSink) Sink(rec record.Record) (err error) {
	var writeTo io.StringWriter
	if isFailResponse(rec) {
		writeTo = w.Err
	} else {
		writeTo = w.Out
		rec, err = w.OutFmt(rec) // apply output formatting only for success records
		if err != nil {
			return err
		}
	}

	recStr := record.CoerceString(rec)

	w.mtx.Lock()
	defer w.mtx.Unlock()
	_, err = writeTo.WriteString(recStr)
	_, err = writeTo.WriteString("\n")
	return err
}
