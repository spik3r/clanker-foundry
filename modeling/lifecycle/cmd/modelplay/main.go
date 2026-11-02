// modelplay is an offline example-fixture runner. It has no filesystem output,
// network, worker, scheduling or authentication authority.
package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/compiler"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/scenarios"
)

func run(args []string, input io.Reader, output, diagnostic io.Writer) int {
	fail := func(code int, message string) int {
		_, _ = fmt.Fprintln(diagnostic, "modelplay: "+message)
		return code
	}
	if len(args) != 0 {
		return fail(2, "usage: modelplay < fixture.json")
	}
	if runtime.Version() != compiler.ToolchainVersion {
		return fail(2, "unsupported Go toolchain; no automatic download")
	}
	raw, e := io.ReadAll(io.LimitReader(input, scenarios.MaxSourceBytes+1))
	if e != nil {
		return fail(2, "input read failed")
	}
	if len(raw) > scenarios.MaxSourceBytes {
		return fail(2, "invalid scenario fixture")
	}
	report, e := scenarios.Execute(raw)
	if e != nil {
		switch {
		case errors.Is(e, scenarios.ErrModelBinding):
			return fail(2, "unsupported scenario model binding")
		case errors.Is(e, scenarios.ErrReportLimit):
			return fail(2, "scenario report limit")
		default:
			return fail(2, "invalid scenario fixture")
		}
	}
	encoded, e := scenarios.EncodeReport(report)
	if e != nil {
		return fail(2, "scenario report limit")
	}
	n, e := output.Write(encoded)
	if e != nil || n != len(encoded) {
		return fail(3, "output write failed")
	}
	if !report.Matched {
		return 1
	}
	return 0
}
func main() { os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr)) }
