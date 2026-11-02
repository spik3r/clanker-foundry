// modeldiff prints a bounded model review comparison, not execution authority.
package main

import (
	"fmt"
	"io"
	"os"
	"runtime"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/compiler"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/impact"
)

func run(args []string, input io.Reader, output, diagnostic io.Writer) int {
	fail := func(code int, message string) int {
		_, _ = fmt.Fprintln(diagnostic, "modeldiff: "+message)
		return code
	}
	if len(args) != 0 {
		return fail(2, "usage: modeldiff < comparison.json")
	}
	if runtime.Version() != compiler.ToolchainVersion {
		return fail(2, "unsupported Go toolchain; no automatic download")
	}
	raw, err := io.ReadAll(io.LimitReader(input, impact.MaxSourceBytes+1))
	if err != nil {
		return fail(2, "input read failed")
	}
	if len(raw) > impact.MaxSourceBytes {
		return fail(2, "invalid model review request")
	}
	encoded, err := impact.Execute(raw)
	if err != nil {
		if err == impact.ErrReportLimit {
			return fail(2, "model review report limit")
		}
		return fail(2, "invalid model review request")
	}
	n, err := output.Write(encoded)
	if err != nil || n != len(encoded) {
		return fail(3, "output write failed")
	}
	return 0
}
func main() { os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr)) }
