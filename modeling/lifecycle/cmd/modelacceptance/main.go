// modelacceptance checks frozen author declarations, not approval or authority.
package main

import (
	"encoding/json"
	"fmt"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/acceptanceauthor"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/compiler"
	"io"
	"os"
	"runtime"
)

func run(args []string, in io.Reader, out, diag io.Writer) int {
	fail := func(code int, s string) int { _, _ = fmt.Fprintln(diag, "modelacceptance: "+s); return code }
	if len(args) != 0 {
		return fail(2, "usage: modelacceptance < request.json")
	}
	if runtime.Version() != compiler.ToolchainVersion {
		return fail(2, "unsupported Go toolchain")
	}
	raw, e := io.ReadAll(io.LimitReader(in, acceptanceauthor.MaxRequestBytes+1))
	if e != nil {
		return fail(2, "input read failed")
	}
	if len(raw) > acceptanceauthor.MaxRequestBytes {
		return fail(2, "invalid request")
	}
	b, e := acceptanceauthor.Execute(raw)
	if e != nil {
		return fail(2, "invalid request")
	}
	n, e := out.Write(b)
	if e != nil || n != len(b) {
		return fail(3, "output write failed")
	}

	var result acceptanceauthor.Report
	if json.Unmarshal(b, &result) != nil {
		return fail(2, "invalid report")
	}
	if result.ModelBinding != "current" || result.SuiteBinding != "current" || result.SuiteModelBinding != "current" || result.RunnerBinding != "current" || result.Replay == "mismatched" {
		return 1
	}
	return 0
}
func main() { os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr)) }
