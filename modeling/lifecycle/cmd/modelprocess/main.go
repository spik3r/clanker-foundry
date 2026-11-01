// modelprocess emits an inert review projection, not an executable process.
package main

import (
	"fmt"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/compiler"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/processview"
	"io"
	"os"
	"runtime"
)

func run(args []string, in io.Reader, out, diagnostic io.Writer) int {
	fail := func(code int, message string) int {
		_, _ = fmt.Fprintln(diagnostic, "modelprocess: "+message)
		return code
	}
	if len(args) != 1 || (args[0] != "json" && args[0] != "svg") {
		return fail(2, "usage: modelprocess json|svg < model.json")
	}
	if runtime.Version() != compiler.ToolchainVersion {
		return fail(2, "unsupported Go toolchain")
	}
	raw, e := io.ReadAll(io.LimitReader(in, processview.MaxSourceBytes+1))
	if e != nil {
		return fail(2, "input read failed")
	}
	if len(raw) > processview.MaxSourceBytes {
		return fail(2, "invalid model or output limit")
	}
	result, e := processview.Build(raw)
	if e != nil {
		return fail(2, "invalid model or output limit")
	}
	b := result.JSON
	if args[0] == "svg" {
		b = result.SVG
	}
	n, e := out.Write(b)
	if e != nil || n != len(b) {
		return fail(3, "output write failed")
	}
	return 0
}
func main() { os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr)) }
