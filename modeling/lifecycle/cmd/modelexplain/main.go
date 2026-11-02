// modelexplain explains caller-supplied assertions, not execution or authority.
package main

import (
	"fmt"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/compiler"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/scenarioexplain"
	"io"
	"os"
	"runtime"
)

func run(args []string, in io.Reader, out, diag io.Writer) int {
	fail := func() int { _, _ = fmt.Fprintln(diag, "modelexplain: invalid input"); return 2 }
	if len(args) != 0 || runtime.Version() != compiler.ToolchainVersion {
		return fail()
	}
	b, e := io.ReadAll(io.LimitReader(in, scenarioexplain.MaxRequestBytes+1))
	if e != nil || len(b) > scenarioexplain.MaxRequestBytes {
		return fail()
	}
	result, e := scenarioexplain.Execute(b)
	if e != nil {
		return fail()
	}
	n, e := out.Write(result)
	if e != nil || n != len(result) {
		_, _ = fmt.Fprintln(diag, "modelexplain: output failed")
		return 3
	}
	return 0
}
func main() { os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr)) }
