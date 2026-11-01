// modeltraceimpact summarizes declared trace changes, not verified impact.
package main

import (
	"fmt"
	"io"
	"os"
	"runtime"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/compiler"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/traceimpact"
)

func run(args []string, in io.Reader, out, diagnostic io.Writer) int {
	fail := func(code int, s string) int { _, _ = fmt.Fprintln(diagnostic, "modeltraceimpact: "+s); return code }
	if len(args) != 0 {
		return fail(2, "usage: modeltraceimpact < request.json")
	}
	if runtime.Version() != compiler.ToolchainVersion {
		return fail(2, "unsupported Go toolchain")
	}
	raw, err := io.ReadAll(io.LimitReader(in, traceimpact.MaxRequestBytes+1))
	if err != nil {
		return fail(2, "input read failed")
	}
	if len(raw) > traceimpact.MaxRequestBytes {
		return fail(2, "invalid request or report limit")
	}
	result, err := traceimpact.Execute(raw)
	if err != nil {
		return fail(2, "invalid request or report limit")
	}
	n, err := out.Write(result)
	if err != nil || n != len(result) {
		return fail(3, "output write failed")
	}
	return 0
}
func main() { os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr)) }
